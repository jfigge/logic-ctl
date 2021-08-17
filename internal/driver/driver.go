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
	"time"
)

type Driver struct {
	instrAddr    uint16
	address      uint16
	opCode       *instructionSet.OpCode
	display      *display.Terminal
	clock        *status.Clock
	irq          *status.Irq
	nmi          *status.Nmi
	reset        *status.Reset
	log          *logging.Log
	serial       *serial.Serial
	opCodes      *instructionSet.OperationCodes
	codes        *instructionSet.ControlLines
	errorPage    *ErrorPage
	helpPage     *HelpPage
	memory       *memory.Memory
	step         *status.Steps
	flags        *status.Flags
	ready        bool
	UIs          []common.UI
	dirty        bool
	initialize   bool
	keyIntercept []common.Intercept
	ignoreFlags  bool
	lines        uint64
	writer       func()
	quit         chan struct{}
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
	d.lines        = instructionSet.Defaults[0]
	d.ignoreFlags  = true
	d.UIs          = append(d.UIs, &d)
	d.errorPage    = NewErrorPage()
	d.helpPage     = NewHelpPage()
	d.log          = logging.New(d.redraw)
	d.step         = status.NewSteps(d.log)
	d.flags        = status.NewFlags(d.log)
	d.opCodes      = instructionSet.New(d.log)
	d.clock        = status.NewClock(d.log, d.tick)
	d.irq          = status.NewIrq(d.log, d.redraw)
	d.nmi          = status.NewNmi(d.log, d.redraw)
	d.reset        = status.NewReset(d.log, d.redraw)
	d.codes        = instructionSet.NewControlLines(d.log, d.display, d.SetDirty, d.setLine)
	d.serial       = serial.New(d.log, d.clock, d.irq, d.nmi, d.reset, d.redraw, d.flags.SetFlags, d.startDataCapture, d.stopDataCapture)
	d.memory       = memory.New(d.log, d.opCodes)
	d.instrAddr    = 0
	d.keyIntercept = append(d.keyIntercept, d.codes)
	return &d
}

func (d *Driver) Run() {
	d.display.Cls()
	if !d.memory.LoadRom(d.log, config.CLIConfig.RomFile, 0x8000) {
		d.log.Dump()
		os.Exit(1)
	}
	if d.serial.Connect(false) {
		tickFunc(d, false)
	}
	d.ready = true

	for len(d.UIs) > 0 {
		connected := d.serial.IsConnected()
		if !connected {
			if connected = d.serial.Reconnect(); connected {
				tickFunc(d, false)
				d.SetDirty(true)
			}
		}
		d.UIs[0].Draw(d.display, connected)
		a, k, _ := d.ReadChar()
		if a != 0 || k != 0 {
			if d.UIs[0].Process(a, k, connected) {
				d.UIs = d.UIs[1:]
				if len(d.UIs) > 0 {
					d.UIs[0].SetDirty(true)
				}
			}
		}
	}
	d.display.Cls()
	d.serial.Disconnect()
}

func (d *Driver) ReadChar() (ascii int, keyCode int, err error) {
	x, _ := term.Open("/dev/tty")
	if x == nil {
		return 0, 0, fmt.Errorf("unavailable")
	} else {
		defer func() {
			if err := x.Restore(); err != nil {
				d.log.Errorf("Failed to restore terminal mode: %v", err)
			}
			if err := x.Close(); err != nil {
				d.log.Errorf("Failed to close terminal input: %v", err)
			}
		}()
	}

	if err := term.RawMode(x); err != nil {
		str := fmt.Sprintf("Failed to access terminal RawMode: %v", err)
		d.log.Error(str)
		d.UIs = append([]common.UI{d.errorPage.ErrorViewer(str)}, d.UIs...)
	}
	bs := make([]byte, 3)

	if err := x.SetReadTimeout(50 * time.Millisecond); err != nil {
		d.log.Warn("Failed to set read timeout")
	}
	if numRead, err := x.Read(bs); err != nil {
		if err.Error() != "EOF" {
			if err := x.Restore(); err != nil {
				d.log.Errorf("Failed to restore terminal mode: %v", err)
			}
			if err := x.Close(); err != nil {
				d.log.Errorf("Failed to close terminal input: %v", err)
			}
		}
		return 0, 0, nil
	} else if numRead == 3 && bs[0] == 27 && bs[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bs[2] == 65 {
			// Up
			keyCode = 38
		} else if bs[2] == 66 {
			// Down
			keyCode = 40
		} else if bs[2] == 67 {
			// Right
			keyCode = 39
		} else if bs[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bs[0])
	} else {
		d.log.Warn("Two character read unexpected")
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
func (d *Driver) SetOpCode(opCode uint8) {
	if d.opCode == nil || d.opCode.OpCode != opCode {
		d.opCode = d.opCodes.Lookup(opCode)
		d.codes.SetSteps(d.opCode.Steps)
	}
	if !d.opCode.Virtual {
		d.instrAddr = d.address
	}
	d.log.Debugf("Loaded OpCode: %s", d.opCode.Name)
}

func (d *Driver) redraw() {
	if len(d.UIs) > 0 {
		d.UIs[0].SetDirty(false)
	}
}
func (d *Driver) reinitialize() {
	if len(d.UIs) > 0 {
		d.UIs[0].SetDirty(true)
	}
}
func (d *Driver) tick() {
	go tickFunc(d, true)
}
func (d *Driver) setLine(step uint8, clock uint8, bit uint64, value uint8) {

	flags := uint8(0)
	if !d.ignoreFlags {
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
		}

		if mask == instructionSet.CL_DBRW {
			if d.opCode.Lines[flags][step][clock] & mask == 0 {
				d.opCode.Lines[flags][step][1-clock] = d.opCode.Lines[flags][step][1-clock] &^ mask
			} else {
				d.opCode.Lines[flags][step][1-clock] = d.opCode.Lines[flags][step][1-clock] | mask
			}
		}

		d.lines = d.opCode.Lines[flags][step][clock]
		d.redraw()
	}

	 if step == d.step.CurrentStep() &&
		(clock == d.clock.CurrentState() || mask == instructionSet.CL_DBRW) &&
	   d.serial.IsConnected() {
	 	d.serial.SetLines(d.opCode.Lines[flags][step][d.clock.CurrentState()])
	}
}

func (d *Driver) startDataCapture() {
	if d.quit == nil {
		if d.clock.CurrentState() == instructionSet.PHI2 {
			d.log.Info("Capture started")
			d.quit = make(chan struct{})
			go func(d *Driver, ticker *time.Ticker) {
				for {
					select {
					case <-ticker.C:
						address := d.address
						if data, ok := d.serial.ReadData(); ok {
							d.log.Infof("Capture writer(%s, %s)", display.HexAddress(address), display.HexData(data))
							d.writer = func() {
								if ok := d.memory.WriteMemory(address, data); !ok {
									d.log.Warnf("Failed to write %s @ %s", display.HexAddress(address), display.HexData(data))
								}
							}
						}
					case <-d.quit:
						d.log.Info("Capture stopped")
						ticker.Stop()
						d.quit = nil
						return
					}
				}
			}(d, time.NewTicker(200*time.Millisecond))
		} else {
			d.log.Info("Wrong phase to start capturing")
		}
	} else {
		d.log.Info("Capture already running")
	}
}

func (d *Driver) stopDataCapture() {
	if d.quit != nil {
		close(d.quit)
		if d.writer != nil {
			d.writer()
		}
		d.quit = nil
	}
}


func (d *Driver) Draw(t *display.Terminal, connected bool) {

	// Skip a redraw if we/re not ready or already drawn
	if !d.ready || (!d.dirty && !d.initialize)  {
		return
	}

	d.display.HideCursor()

	if d.initialize {
		t.Cls()
		d.initialize = false
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
	steps := uint8(2)
	if d.opCode != nil {
		steps = d.opCode.Steps
	}
	t.PrintAt(55, 5, d.step.StepBlock(steps))

	// Instructions
	t.PrintAtf(58, 7, "%sInstructions", common.Yellow)
	lines = d.memory.InstructionBlock(d.instrAddr)
	for i := 0; i < 11; i++ {
		t.PrintAt(55, 8+i, lines[uint16(i)])
	}

	// Indicate if the board is connected
	colour := common.BGGreen + common.White
	if !d.serial.IsConnected() {
		colour = common.BGRed + common.White
	}

	// Control lines
	offset := d.display.Rows() - 5
	if !connected {
		d.display.PrintAtf(1, 1, "%s   %s" , colour, common.Reset)
	} else {
		d.display.PrintAtf(1, 1, "%s%-3s%s", colour, d.opCode.Name, common.Reset)
	}

	var aLines[]string
	var outputs[4]string
	var AluOperations[4]string
	flags := uint8(0)
	if !d.ignoreFlags {
		flags = d.flags.CurrentFlags()
	}
	if d.ignoreFlags {
		t.PrintAtf(1, 20, "%sControl Lines (Ignoring flags)", common.Yellow)
	} else {
		t.PrintAtf(1, 20, "%sControl Lines", common.Yellow)
	}
	t.PrintAtf(66, 20, "%sActiveLines", common.Yellow)

	lines, aLines, outputs, AluOperations = d.opCode.Block(
		flags, d.step.CurrentStep(), d.clock.CurrentState(), (d.codes.EditStep() - 1) / 2, (d.codes.EditStep() - 1) % 2)
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
	lines = d.codes.LineNamesBlock((d.codes.EditStep() - 1) % 2)
	for i, line := range lines {
		t.PrintAt(1, 21 + offset + i, "        " + line)
	}
	offset = 20 + offset + len(lines)

	// Bus Content
	t.PrintAtf(90,  7, "%sBus Content", common.Yellow)
	t.PrintAtf(85,  8, "%sADH: %s%s%s", common.Yellow, common.White, outputs[2], display.ClearEnd)
	t.PrintAtf(86,  9, "%sDB: %s%s%s", common.Yellow, common.White, outputs[0], display.ClearEnd)
	t.PrintAtf(85, 10, "%sADL: %s%s%s", common.Yellow, common.White, outputs[1], display.ClearEnd)
	t.PrintAtf(86, 11, "%sSB: %s%s%s", common.Yellow, common.White, outputs[3], display.ClearEnd)

	// ALU Operation
	t.PrintAtf(90, 13, "%sALU", common.Yellow)
	t.PrintAtf(87, 14, "%sB: %s%s%s", common.Yellow, common.White, AluOperations[1], display.ClearEnd)
	t.PrintAtf(87, 15, "%sA: %s%s%s", common.Yellow, common.White, AluOperations[0], display.ClearEnd)
	t.PrintAtf(86, 16, "%sOp: %s%s%s", common.Yellow, common.White, AluOperations[2], display.ClearEnd)
	t.PrintAtf(85, 17, "%sDir: %s%-10s%s", common.Yellow, common.White, AluOperations[3], display.ClearEnd)

	d.display.ShowCursor()

	// X and Y coordinates of cursor
	str := d.codes.CursorPosition()
	d.display.PrintAt(d.display.Cols() - len(str), 1, str)

	// Notifications
	lines = d.log.LogBlock()
	max := d.display.Rows() - offset
	if len(lines) > max {
		lines = lines[:max]
	}
	for i := 0; i < max; i++ {
		line := display.ClearLine
		if i < len(lines) {
			line = lines[i]
		}
		d.display.PrintAt(1, d.display.Rows()-i, line)
	}

	// Restore cursor position
	d.codes.PositionCursor()
	d.dirty = false
}
func (d *Driver) SetDirty(initialize bool) {
	d.dirty = true
	if initialize {
		d.initialize = true
	}
}
func (d *Driver) Process(a int, k int, connected bool) bool {
	for _, ki := range d.keyIntercept {
		if ki.KeyIntercept(a,k, connected) {
			return false
		}
	}
	if k != 0 {
		d.log.Warnf("Unknown code: [%v]", k)
	} else {
		switch a {
		case 'f':
			d.ignoreFlags = !d.ignoreFlags
			d.reinitialize()
		case 'q':
			return true
		case 'D':
			d.log.SetDebug(false)
		case 'd':
			d.log.SetDebug(true)
		case 'c':
			flags := uint8(0)
			if !d.ignoreFlags {
				flags = d.flags.CurrentFlags()
			}
			mnemonics := d.opCode.ActiveLines(flags, (d.codes.EditStep() - 1) / 2, (d.codes.EditStep() - 1) % 2, 64, " | ", "CL_")
			if len(mnemonics) > 0 && strings.HasPrefix(mnemonics[0], "CL_CTMR") {
				if strings.HasPrefix(mnemonics[0], "CL_CTMR | ") {
					mnemonics = []string{mnemonics[0][10:]}
				} else {
					mnemonics = []string{mnemonics[0][7:]}
				}
			}
			if len(mnemonics) > 0 {
				clipboard.WriteAll(mnemonics[0])
				d.log.Info("Mnemonics copied to clipboard")
			} else {
				clipboard.WriteAll("")
				d.log.Info("No lines set")
			}

		case 'h':
			d.UIs = append([]common.UI{d.helpPage.Help()}, d.UIs...)
		case 'l':
			d.UIs = append([]common.UI{d.log.HistoryViewer()}, d.UIs...)
		case 'p':
			d.UIs = append([]common.UI{d.serial.PortViewer()}, d.UIs...)
		default:
			d.log.Warnf("Unmapped ascii code: [%c]", a)
		}
	}
	return false
}

func tickFunc(d *Driver, phaseChange bool) {

	if state, ok := d.serial.ReadStatus();  ok {
		d.step.SetStep(state)
		d.flags.SetFlags(state)
	} else {
		return
	}

	if d.opCode == nil || d.step.CurrentStep() == 0 && d.clock.CurrentState() == 0 {
		if opCode, ok := d.serial.ReadOpCode(); ok {
			d.SetOpCode(opCode)
		} else {
			return
		}
	}

	flags := uint8(0)
	if !d.ignoreFlags {
		flags = d.flags.CurrentFlags()
	}
	d.lines = d.opCode.Lines[flags][d.step.CurrentStep()][d.clock.CurrentState()]
	d.serial.SetLines(d.lines)

	time.Sleep(50 * time.Millisecond)
	if address, ok := d.serial.ReadAddress(); ok {
		d.SetAddress(address)
	}

	if d.clock.CurrentState() == instructionSet.PHI1 || d.lines & instructionSet.CL_DBRW != 0 {
		if data, ok := d.memory.ReadMemory(d.address); ok {
			d.serial.SetData(data)
		}
	}

	d.codes.SetEditStep(d.step.CurrentStep() * 2 + d.clock.CurrentState() + 1)
	d.log.Tracef("tickFunc. PhaseChange: %v. Clock: %v. Flags: %v. Phase %v", phaseChange, d.step.CurrentStep(), d.flags.CurrentFlags(), d.clock.CurrentState())
	d.SetDirty(false)
}