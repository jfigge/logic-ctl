package driver

import (
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/pkg/term"
	"github.td.teradata.com/sandbox/logic-ctl/internal/config"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/instructionSet"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/memory"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/serial"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/status"
	"os"
	"strings"
	"sync"
	"time"
)

type Driver struct {
	instrAddr   uint16
	address     uint16
	buses       [7]uint64
	opCode      *instructionSet.OpCode
	display     *display.Terminal
	clock       *status.Clock
	irq         *status.Irq
	nmi         *status.Nmi
	reset       *status.Reset
	log         *logging.Log
	serial      *serial.Serial
	opCodes     *instructionSet.OpCodes
	lines       *instructionSet.ControlLines
	errorPage   *ErrorPage
	helpPage    *HelpPage
	memory      *memory.Memory
	step        *status.Steps
	flags       *status.Flags
	UIs         []common.UI
	dispChan    chan bool
	monitorChan chan bool
	clockChan   chan bool
	inputChan   chan common.Input
	connected    bool
	keyIntercept []common.Intercept
	wg           *sync.WaitGroup
	editor       int
	xTerm        *term.Term
}
func New() *Driver {
	d := Driver{}
	time.Sleep(1 * time.Second)

	var err error
	if d.display, err = display.New(); err != nil {
		fmt.Printf("%sFailed to gain control of terminal: %v%s\n", common.Red, err, common.Reset)
		os.Exit(1)
	} else if d.display.Rows() < 38 || d.display.Cols() < 100 {
		fmt.Printf("%sMinimum console size must be 100x38.  Currently at %dx%d%s\n", common.Red, d.display.Cols(), d.display.Rows(), common.Reset)
		os.Exit(1)
	}

	d.wg           = &sync.WaitGroup{}
	d.buses        = [7]uint64 {1, 6, 7, 7, 0, 3, 0}
	d.UIs          = append(d.UIs, &d)
	d.errorPage    = NewErrorPage()
	d.helpPage     = NewHelpPage()
	d.log          = logging.New(d.redraw)
	d.step         = status.NewSteps(d.log)
	d.opCodes      = instructionSet.New(d.log)
	d.clock        = status.NewClock(d.log, d.tick)
	d.irq          = status.NewIrq(d.log, d.redraw)
	d.nmi          = status.NewNmi(d.log, d.redraw)
	d.reset        = status.NewReset(d.log, d.redraw)
	d.flags        = status.NewFlags(d.log, d.display, d.redraw)
	d.memory       = memory.New(d.log, d.opCodes, d.display, d.redraw)
	d.lines        = instructionSet.NewControlLines(d.log, d.display, d.redraw, d.setLine)
	d.serial       = serial.New(d.log, d.clock, d.irq, d.nmi, d.reset, d.connectionStatus, d.wg)
	d.instrAddr    = 0
	d.keyIntercept = append(d.keyIntercept, d.lines, d.memory, d.lines.BusController())
	d.editor       = 0
	d.dispChan     = make(chan bool)
	d.monitorChan  = make(chan bool)
	d.clockChan    = make(chan bool)
	d.inputChan    = make(chan common.Input)
	return &d
}

func (d *Driver) Run() {
	go d.output(d.wg)
	go d.input(d.wg)

	if !d.memory.LoadRom(d.log, config.CLIConfig.RomFile, 0x8000) {
		d.log.Dump()
		os.Exit(1)
	}

	d.dispChan <- true

	for len(d.UIs) > 0 {
		if a, k, e := d.ReadChar(); e != nil {
			d.log.Warn(e.Error())
		} else if a != 0 || k != 0 && len(d.UIs) > 0 {
			d.inputChan <- common.Input{Ascii: a, KeyCode: k, Connected: d.connected}
		}
	}

	close(d.monitorChan)
	close(d.dispChan)

	d.wg.Wait()
	fmt.Println("Wait Done")
}

func (d *Driver) output(wg *sync.WaitGroup) {
	wg.Add(1)
	defer func() {
		fmt.Println("Output Done")
		wg.Done()
	}()
	initialize, ok := false, true
	for ok {
		select {
		case initialize, ok = <- d.dispChan:
			if len(d.UIs) > 0 {
				d.UIs[0].Draw(d.display, d.connected, initialize)
			}
		}
	}
	d.display.Cls()
}
func (d *Driver) input(wg *sync.WaitGroup) {
	wg.Add(1)
	defer func() {
		fmt.Println("Input Done")
		wg.Done()
	}()
	input, connected, phaseChange, ok := common.Input{}, false, false, true

	// By monitoring all state changes in one thread we can
	// ensure all updates are serialized
	for ok {
		select {
		case connected, ok = <-d.monitorChan:
			d.connected = connected
			d.redraw(true)

		case input, ok = <-d.inputChan:

			if len(d.UIs) > 0 && d.UIs[0].Process(input) {
				d.UIs = d.UIs[1:]
				d.redraw(true)
			}
		case phaseChange, ok = <- d.clockChan:
			d.tickFunc(phaseChange)
		}
	}
	d.serial.Terminate()
}

func (d *Driver) ReadChar() (ascii int, keyCode int, err error) {
	d.xTerm, _ = term.Open("/dev/tty")
	if d.xTerm == nil {
		return 0, 0, fmt.Errorf("terminal unavailable")
	} else {
		defer func() {
			if err := d.xTerm.Restore(); err != nil {
				d.log.Errorf("Failed to restore terminal mode: %v", err)
			}
			if err := d.xTerm.Close(); err != nil {
				d.log.Errorf("Failed to close terminal input: %v", err)
			}
		}()
	}

	if err := term.RawMode(d.xTerm); err != nil {
		str := fmt.Sprintf("Failed to access terminal RawMode: %v", err)
		d.log.Error(str)
		d.UIs = append([]common.UI{d.errorPage.ErrorViewer(str)}, d.UIs...)
		d.redraw(true)
	}
	bs := make([]byte, 5)

	if err := d.xTerm.SetReadTimeout(20 * time.Second); err != nil {
		d.log.Warn("Failed to set read timeout")
	}
	if numRead, err := d.xTerm.Read(bs); err != nil {
		if err.Error() != "EOF" {
			if err := d.xTerm.Restore(); err != nil {
				d.log.Errorf("Failed to restore terminal mode: %v", err)
			}
			if err := d.xTerm.Close(); err != nil {
				d.log.Errorf("Failed to close terminal input: %v", err)
			}
		}
		return 0, 0, nil
	} else if numRead == 3 && bs[0] == 27 && bs[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII lines for arrow keys, we use
		// Javascript key lines.
		switch bs[2] {
			case 65: keyCode = 38 // Up
			case 66: keyCode = 40 // Down
			case 67: keyCode = 39 // Right
			case 68: keyCode = 37 // Left
		}
	} else if numRead == 3 && bs[0] == 0x1B && bs[1] == 0x4F {
		switch bs[2] {
			case 50: keyCode = 101 // Option+1
			case 51: keyCode = 102 // Option+2
			case 52: keyCode = 102 // Option+3
			case 53: keyCode = 103 // Option+4
		}
	} else if numRead == 1 {
		ascii = int(bs[0])
	} else {
		text := fmt.Sprintf("%d characters read.", numRead)
		for i := 0; i < numRead; i++ {
			text = fmt.Sprintf("%s %d:%s", text, i, display.HexData(bs[i]))
		}
		d.log.Warnf(text)
		// Two characters read??
	}
	return
}
func (d *Driver) SetAddress(address uint16) {
	if d.address != address {
		d.address = address
		d.log.Debugf("Address set to %s", display.HexAddress(d.address))
	}
}
func (d *Driver) setLine(step uint8, clock uint8, bit uint64, value uint8) {

	flags := d.flags.DevFlags()
	if !d.flags.Ignore {
		flags = d.flags.CurrentFlags()
	}

	mask := uint64(0)
	if value != 99 {
		if str, ok := d.opCode.ValidateLine(step, clock, bit); !ok {
			d.log.Warn(str)
			return
		}

		mask = uint64(1 << bit)
		switch value {
		case 0:
			d.opCode.Lines[flags][step][clock] = d.opCode.Lines[flags][step][clock] &^ mask
		case 1:
			d.opCode.Lines[flags][step][clock] = d.opCode.Lines[flags][step][clock] | mask
		case 2:
			d.opCode.Lines[flags][step][clock] = d.opCode.Lines[flags][step][clock]&^mask | d.opCode.Presets[flags][step][clock]&mask
		case 3:
			d.opCode.Lines[flags][step][clock] = d.opCode.Lines[flags][step][clock] ^ mask
		case 4:
			d.opCode.Lines[flags][step][clock] = bit
		}

		if mask == instructionSet.CL_DBRW {
			if d.opCode.Lines[flags][step][clock] & mask == 0 {
				d.opCode.Lines[flags][step][1-clock] = d.opCode.Lines[flags][step][1-clock] &^ mask
			} else {
				d.opCode.Lines[flags][step][1-clock] = d.opCode.Lines[flags][step][1-clock] | mask
			}
		}

		d.redraw(false)
	}

	if step == d.step.CurrentStep() &&
		(clock == d.clock.CurrentState() || mask == instructionSet.CL_DBRW) &&
		d.connected {
		d.serial.SetLines(d.opCode.Lines[flags][step][d.clock.CurrentState()])
	}
}

func (d *Driver) redraw(clearScreen bool) {
	if len(d.UIs) > 0 {
		// Pull any previous requires
		previous := false
		if clearScreen {}
			select {
			case previous = <- d.dispChan:
			default:
		}
		// Write back the higher of the requested or previous value
		d.dispChan <- clearScreen || previous
	}
}
func (d *Driver) connectionStatus(connected bool) {
	// Remove any previous value before it can be read
	select {
	case <-d.monitorChan:
	default: // don't block if there is no value
	}

	// Push the current connected status
	d.monitorChan <- connected
}
func (d *Driver) tick(phaseChange bool) {
	select {
	case d.clockChan <- phaseChange:
	default:
		d.log.Debug("Tick ignored. phase change already queued")
	}
}

func (d *Driver) Draw(t *display.Terminal, connected, initialize bool) {

	if d.opCode == nil {
		opCode, ok := d.serial.ReadOpCode()
		if !ok {
			opCode = 0xff
		}
		d.SetOpCode(opCode)
	}

	d.display.HideCursor()
	if initialize {
		t.Cls()
	}

	// Memory
	lines := d.memory.MemoryBlock(d.address)
	for row, line := range lines {
		if ok := t.PrintAt(1, row+1, line); !ok {
			break
		}
	}

	// IRQ
	t.PrintAtf(84, 1, "%sIRQ", common.Yellow)
	t.PrintAt(85, 2, d.irq.IrqBlock())

	// NMI
	t.PrintAtf(94, 1, "%sNMI", common.Yellow)
	t.PrintAt(95, 2, d.nmi.NmiBlock())

	// Clock
	t.PrintAtf(83, 4, "%sClock", common.Yellow)
	t.PrintAt(85, 5, d.clock.Block())

	// Reset
	t.PrintAtf(93, 4, "%sReset", common.Yellow)
	t.PrintAt(95, 5, d.reset.ResetBlock())

	// FLags
	t.PrintAtf(61, 1, "%sFlags", common.Yellow)
	t.PrintAt(55, 2, d.flags.FlagsBlock())

	// Step
	t.PrintAtf(61, 4, "%sStep", common.Yellow)
	steps := d.opCode.Steps
	t.PrintAt(55, 5, d.step.StepBlock(steps))

	// Instructions
	t.PrintAtf(58, 7, "%sInstructions", common.Yellow)
	lines = d.memory.InstructionBlock(d.instrAddr, d.address)
	for i := 0; i < 11; i++ {
		if i < len(lines) {
			t.PrintAt(55, 8+i, lines[uint16(i)])
		} else {
			t.PrintAt(55, 8+i, "                             ")
		}
	}

	// Control lines
	offset := d.display.Rows() - 5
	if !connected {
		d.display.PrintAtf(1, 1, "%s%s   %s" , common.BGRed, common.White, common.Reset)
	} else {
		d.display.PrintAtf(1, 1, "%s%s%-3s%s", common.BGGreen, common.White, d.opCode.Name, common.Reset)
	}

	var aLines[]string
	var outputs[4]string
	var AluOperations[4]string
	flags := d.flags.DevFlags()
	if d.flags.Ignore {
		t.PrintAtf(1, 20, "%sControl Lines (%s%s)%s", common.Yellow, d.flags.DevBlock(), common.Yellow, common.Reset)
	} else {
		flags = d.flags.CurrentFlags()
		t.PrintAtf(1, 20, "%sControl Lines (Following flags)", common.Yellow)
	}
	t.PrintAtf(66, 20, "%sActiveLines", common.Yellow)

	lines, aLines, outputs, AluOperations = d.opCode.Block(
		flags, d.step.CurrentStep(), d.clock.CurrentState(), (d.lines.EditStep() - 1) / 2, (d.lines.EditStep() - 1) % 2)
	d.lines.SetControlLines(d.opCode.Lines[flags])
	for i := 0; i < 14; i++ {
		str := ""
		if i < len(lines) {
			t.PrintAt(1, 21+i, lines[i])
		}
		if i < len(aLines) {
			str = aLines[i]
		}
		t.PrintAtf(66, 21 + i, "%s%s%s%s", display.ClearEnd, common.Magenta , str, common.Reset)
	}

	// Control line names
	offset = len(lines)
	lines = d.lines.LineNamesBlock((d.lines.EditStep() - 1) % 2)
	for i, line := range lines {
		t.PrintAt(1, 21 + offset + i, "        " + line)
	}
	offset = 20 + offset + len(lines)

	// Bus Content
	t.PrintAtf(90,  7, "%sBus Content", common.Yellow)
	t.PrintAtf(86,  8, "%sDB: %s%s%s", common.Yellow, common.White, outputs[0], display.ClearEnd)
	t.PrintAtf(85,  9, "%sABH: %s%s%s", common.Yellow, common.White, outputs[2], display.ClearEnd)
	t.PrintAtf(85, 10, "%sABL: %s%s%s", common.Yellow, common.White, outputs[1], display.ClearEnd)
	t.PrintAtf(86, 11, "%sSB: %s%s%s", common.Yellow, common.White, outputs[3], display.ClearEnd)

	// ALU Operation
	t.PrintAtf(90, 13, "%sALU", common.Yellow)
	t.PrintAtf(87, 14, "%sB: %s%s%s", common.Yellow, common.White, AluOperations[1], display.ClearEnd)
	t.PrintAtf(87, 15, "%sA: %s%s%s", common.Yellow, common.White, AluOperations[0], display.ClearEnd)
	t.PrintAtf(86, 16, "%sOp: %s%s%s", common.Yellow, common.White, AluOperations[2], display.ClearEnd)
	t.PrintAtf(85, 17, "%sDir: %s%-10s%s", common.Yellow, common.White, AluOperations[3], display.ClearEnd)

	// X and Y coordinates of cursor
	str := d.keyIntercept[d.editor].CursorPosition()
	d.display.PrintAt(d.display.Cols()-9, 1, str)

	// Notifications
	max := d.display.Rows() - offset
	lines = d.log.LogBlock(max)
	for i := 0; i < max; i++ {
		line := display.ClearLine
		if i < len(lines) {
			line = lines[i]
		}
		d.display.PrintAt(1, d.display.Rows()-i, line)
	}

	// Restore cursor position
	d.keyIntercept[d.editor].PositionCursor()
	d.display.ShowCursor()
}
func (d *Driver) Process(input common.Input) bool {
	if d.editor >= 0 && d.editor < len(d.keyIntercept) && d.keyIntercept[d.editor].KeyIntercept(input) {
		return false
	}

	if input.KeyCode != 0 {
		switch input.KeyCode {
		default:
			d.log.Warnf("Unknown code: [%v]", input.KeyCode)
		}
	} else {
		switch input.Ascii {
		case 'a':
			if address, ok := d.serial.ReadAddress(); ok {
				d.log.Infof("Read address: %s", display.HexAddress(address))
			} else {
				d.log.Warn("Failed to read address")
			}
		case 'q':
			return true
		case 'D':
			d.log.SetDebug(false)
		case 'd':
			d.log.SetDebug(true)
		case 'c', 'C':
			flags := d.flags.DevFlags()
			if !d.flags.Ignore {
				flags = d.flags.CurrentFlags()
			}
			mnemonics := d.opCode.DescribeLine(flags, (d.lines.EditStep() - 1) / 2, (d.lines.EditStep() - 1) % 2, 64, " | ", "CL_", input.Ascii == 'c')
			if len(mnemonics) > 0 && strings.HasPrefix(mnemonics[0], "CL_CTMR") {
				if strings.HasPrefix(mnemonics[0], "CL_CTMR | ") {
					mnemonics = []string{mnemonics[0][10:]}
				} else {
					mnemonics = []string{mnemonics[0][7:]}
				}
			}
			if len(mnemonics) > 0 {
				clipboard.WriteAll(mnemonics[0])
				if input.Ascii == 'c' {
					d.log.Info("Mnemonics copied to clipboard without address mode lines")
				} else {
					d.log.Info("Mnemonics copied to clipboard")
				}
			} else {
				clipboard.WriteAll("0")
				if input.Ascii == 'c' {
					d.log.Info("No lines set outside of address mode lines")
				} else {
					d.log.Info("No lines set")
				}
			}

		case 'h':
			d.UIs = append([]common.UI{d.helpPage.Help()}, d.UIs...)
			d.redraw(true)
		case 'l':
			d.UIs = append([]common.UI{d.log.HistoryViewer()}, d.UIs...)
			d.redraw(true)
		case 'p':
			d.UIs = append([]common.UI{d.serial.PortViewer()}, d.UIs...)
			d.redraw(true)
		case 'L':
			d.editor = 0
			d.redraw(false)
		case 'M':
			d.editor = 1
			d.redraw(false)
		case 'B':
			d.editor = 2
			d.redraw(false)
		case 'F':
			if len(d.keyIntercept) == 3 {
				d.flags.Ignore = true
				d.keyIntercept = append(d.keyIntercept, d.flags)
			}
			d.editor = 3
			d.redraw(false)
		case 'f':
			d.flags.Ignore = !d.flags.Ignore
			if d.flags.Ignore {
				d.keyIntercept = append(d.keyIntercept, d.flags)
				d.editor = len(d.keyIntercept) - 1
			} else {
				if d.editor == len(d.keyIntercept) - 1 { d.editor = 0 }
				d.keyIntercept = d.keyIntercept[:len(d.keyIntercept) - 1]
			}
			d.redraw(true)
		case 's':
			d.flags.SyncFlags()
		case '\t':
			if d.editor + 1 >= len(d.keyIntercept) {
				d.editor = 0
			} else {
				d.editor += 1
			}
			d.redraw(false)
		default:
			d.log.Warnf("Unmapped ascii code: [%c]", input.Ascii)
		}
	}
	return false
}

func (d *Driver) tickFunc(phaseChange bool) {

	if state, ok := d.serial.ReadStatus(); ok {
		d.step.SetStep(state)
		d.flags.SetFlags(state)
	} else {
		d.log.Errorf("Failed to read status during tick")
		d.ResetChannels()
		return
	}

	if d.opCode == nil || (d.step.CurrentStep() == 0 && d.clock.CurrentState() == 0) {
		if opCode, ok := d.serial.ReadOpCode(); ok {
			d.SetOpCode(opCode)
		} else {
			d.log.Errorf("Failed to read OpCode during tick")
			d.ResetChannels()
			return
		}
	}

	if d.step.CurrentStep() > d.opCode.Steps {
		d.log.Errorf("Invalid state. Step %d of %d", d.step.CurrentStep(), d.opCode.Steps)
		return
	}

	flags := d.flags.DevFlags()
	if !d.flags.Ignore {
		flags = d.flags.CurrentFlags()
	}
	lines := d.opCode.Lines[flags][d.step.CurrentStep()][d.clock.CurrentState()]
	d.serial.SetLines(lines)

	time.Sleep(50 * time.Millisecond)
	if address, ok := d.serial.ReadAddress(); ok {
		d.SetAddress(address)
	} else {
		d.log.Errorf("Failed to retrieve address")
		d.ResetChannels()
		return
	}

	if d.clock.CurrentState() == instructionSet.PHI1 || lines & instructionSet.CL_DBRW != 0 {
		if data, ok := d.memory.ReadMemory(d.address); ok {
			d.serial.SetData(data)
		} else {
			d.log.Errorf("Failed to read memory address %s during tick", display.HexAddress(d.address))
			return
		}
	} else {
		if data, ok := d.serial.ReadData(); ok {
			if ok = d.memory.WriteMemory(d.address, data); !ok {
				d.log.Warnf("Failed to write %s @ %s", display.HexAddress(d.address), display.HexData(data))
				return
			}
		} else {
			d.log.Errorf("Failed to read data during tick")
			d.ResetChannels()
			return
		}
	}

	d.lines.SetEditStep(d.step.CurrentStep() * 2 + d.clock.CurrentState() + 1)
	d.log.Tracef("tickFunc. PhaseChange: %v. Clock: %v. Flags: %v. Phase %v", phaseChange, d.step.CurrentStep(), d.flags.CurrentFlags(), d.clock.CurrentState())
	d.editor = 0
	d.redraw(false)
}
func (d *Driver) SetOpCode(opCode uint8) {
	if d.opCode == nil || d.opCode.OpCode != opCode {
		d.opCode = d.opCodes.Lookup(opCode)
		d.lines.SetSteps(d.opCode.Steps)
	}
	if !d.opCode.Virtual {
		d.instrAddr = d.address
	}
	d.log.Debugf("Loaded OpCode: %s", d.opCode.Name)
}
func (d *Driver) ResetChannels() {
	for {
		select {
		case <-d.dispChan:
		case <-d.monitorChan:
		case <-d.clockChan:
		case <-d.inputChan:
		default:
			d.serial.ResetChannels()
			return
		}
	}
}