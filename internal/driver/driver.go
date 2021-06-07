package driver

import (
	"fmt"
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
	"time"
)

type Driver struct {
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
	memory       *memory.Memory
	status       *status.Status
	ready        bool
	UIs          []common.UI
	dirty        bool
	initialize   bool
	keyIntercept []common.Intercept
	ignoreFlags  bool
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
	d.ignoreFlags = false
	d.UIs        = append(d.UIs, &d)
	d.errorPage  = NewErrorPage()
	d.log        = logging.New(d.redraw)
	d.status     = status.NewStatus(d.log)
	d.irq        = status.NewIrq(d.log)
	d.nmi        = status.NewNmi(d.log)
	d.reset      = status.NewReset(d.log)
	d.opCodes    = instructionSet.New(d.log)
	d.codes      = instructionSet.NewControlLines(d.log, d.display, d.SetDirty, d.setLine)
	d.clock      = status.NewClock(d.log, d.tick)
	d.serial     = serial.New(d.log, d.clock, d.irq, d.nmi, d.reset, d.redraw, d.status.SetStatus, d.tick)
	d.memory     = memory.New(d.log, d.opCodes)

	d.keyIntercept = append(d.keyIntercept, d.codes)
	return &d
}

func (d *Driver) Run() {
	d.display.Cls()
	if !d.memory.LoadRom(d.log, config.CLIConfig.RomFile) {
		d.log.Dump()
		os.Exit(1)
	}
	d.serial.Connect(false)
	d.tick(true)
	d.ready = true

	for len(d.UIs) > 0 {
		if !d.serial.IsConnected() {
			d.serial.Reconnect()
		}
		d.UIs[0].Draw(d.display)
		if d.serial.IsConnected() {
			a, k, _ := d.ReadChar()
			if a != 0 || k != 0 {
				if d.UIs[0].Process(a, k) {
					d.UIs = d.UIs[1:]
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
			d.log.Warn("Input error.  Resetting")
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
		hex := display.HexData(opCode)
		str := fmt.Sprintf("OpCode set to %s (%s)", hex, d.opCode.Name)
		d.log.Debug(str)
		d.SetDirty(true)
	}
}

func (d *Driver) redraw() {
	d.UIs[0].SetDirty(false)
}
func (d *Driver) reinitialize() {
	d.UIs[0].SetDirty(true)
}
func (d *Driver) tick(synchronized bool) {
	if synchronized {
		tickFunc(d)
	} else {
		go tickFunc(d)
	}
}
func (d *Driver) setLine(step uint8, clock uint8, bit uint64, value uint8) {
	flags := 0
	if !d.ignoreFlags {
		d.status.CurrentFlags()
	}
	mask := uint64(1 << bit)
	switch value{
		case 0: d.opCode.Lines[flags][step][clock] = d.opCode.Lines[flags][step][clock] &^ mask
		case 1: d.opCode.Lines[flags][step][clock] = d.opCode.Lines[flags][step][clock] | mask
		case 2: d.opCode.Lines[flags][step][clock] = d.opCode.Lines[flags][step][clock] &^ mask | d.opCode.Defaults[flags][step][clock] & mask
		case 3: d.opCode.Lines[flags][step][clock] = d.opCode.Lines[flags][step][clock] ^ mask
	}
	if step == d.status.CurrentStep() &&
	   clock == d.clock.CurrentState() {
		d.serial.SetLines(d.opCode.Lines[flags][step][clock])
	}
	d.redraw()
}

func (d *Driver) Draw(t *display.Terminal) {

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

	// Indicate if the board is connected
	colour := common.BGGreen
	if !d.serial.IsConnected() {
		colour = common.BGRed
	}
	d.display.PrintAtf(1, 1, "%s   %s" , colour, common.Reset)

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
	t.PrintAt(55, 2, d.status.FlagsBlock())

	// Timing
	t.PrintAtf(61, 4, "Step%s", common.Yellow)
	t.PrintAt(55, 5, d.status.StepBlock())

	// Instructions
	t.PrintAtf(58, 7, "%sInstructions", common.Yellow)
	lines = d.memory.InstructionBlock(d.address)
	for i := 0; i < 11; i++ {
		t.PrintAt(55, 8+i, lines[uint16(i)])
	}

	// Control lines
	offset := d.display.Rows() - 5
	if d.serial.IsConnected() {
		var aLines  []string
		var outputs [6]string
		flags := uint8(0)
		if !d.ignoreFlags {
			flags = d.status.CurrentFlags()
		}
		if d.ignoreFlags {
			t.PrintAtf(1, 20, "%sControl Lines (Ignoring flags)", common.Yellow)
		} else {
			t.PrintAtf(1, 20, "%sControl Lines", common.Yellow)
		}

		lines, aLines, outputs = d.opCode.Block(flags, d.status.CurrentStep(), d.clock.CurrentState())
		for i := 0; i < len(lines); i++ {
			t.PrintAt(1, 21+i, lines[i])
		}
		t.PrintAtf(66, 20, "%sActiveLines", common.Yellow)
		for i := 0; i < 12; i++ {
			str := ""
			if i < len(aLines) {
				str = aLines[i]
			}
			t.PrintAtf(66, 21 + i, "%s%s%s%s", display.ClearEnd, common.Magenta , str, common.Reset)
		}

		// Control line names
		offset := len(lines)
		lines = d.codes.LineNamesBlock()
		for i, line := range lines {
			t.PrintAt(9, 21 + offset + i, line)
		}
		offset = 20 + offset + len(lines)

		// Bus Content
		t.PrintAtf(90,  7, "%sBus Content", common.Yellow)
		t.PrintAtf(85,  8, "%sDB:  %s%s%s", common.Yellow, common.White, outputs[0], display.ClearEnd)
		t.PrintAtf(85,  9, "%sADL: %s%s%s", common.Yellow, common.White, outputs[1], display.ClearEnd)
		t.PrintAtf(85, 10, "%sAHD: %s%s%s", common.Yellow, common.White, outputs[2], display.ClearEnd)
		t.PrintAtf(85, 11, "%sSB:  %s%s%s", common.Yellow, common.White, outputs[3], display.ClearEnd)

		// ALU Operation
		t.PrintAtf(94, 13, "%sALU", common.Yellow)
		t.PrintAtf(85, 14, "%sOperation: %s", common.White, outputs[4], display.ClearEnd)
		t.PrintAtf(85, 14, "%sOperation: %s", common.White, outputs[5], display.ClearEnd)

		d.display.ShowCursor()
	}

	// X and Y coordinates of cursor
	str := d.codes.CursorPosition()
	d.display.PrintAt(d.display.Cols() - len(str), d.display.Rows(), str)

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
		d.display.PrintAtf(1, d.display.Rows() - i, "%s%s", display.ClearLine, line)
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
func (d *Driver) Process(a int, k int) bool {
	for _, ki := range d.keyIntercept {
		if ki.KeyIntercept(a,k) {
			return false
		}
	}
	if k != 0 {
		d.log.Warnf("Unknown code: [%v]", k)
	} else {
		switch a {
		case 'a':
			if a, ok := d.serial.ReadAddress(); ok {
				d.SetAddress(a)
				d.SetDirty(false)
			}
		case 's':
			if s, ok := d.serial.ReadStatus(); ok {
				d.status.SetStatus(s)
				d.SetDirty(false)
			}
		case 'q':
			return true
		case 'd':
			d.log.SetDebug(false)
		case 'D':
			d.log.SetDebug(true)
		case 'H':
			d.codes.ShowNames(!d.codes.IsShowNames())
		case 'h':
			d.UIs = append([]common.UI{d.log.HistoryViewer()}, d.UIs...)
		case 'm':
			d.opCodes.ToggleMnemonic()
			d.redraw()
		case 'M':
			d.ignoreFlags = !d.ignoreFlags
			d.reinitialize()
		case 'p':
			d.UIs = append([]common.UI{d.serial.PortViewer()}, d.UIs...)
		default:
			d.log.Warnf("Unmapped ascii code: [%c]", a)
		}
	}
	return false
}

func tickFunc(d *Driver) {
	if state, ok := d.serial.ReadStatus(); ok { d.status.SetStatus(state) } else { return }
	if address, ok := d.serial.ReadAddress(); ok { d.SetAddress(address) } else { return }
	if opCode, ok := d.serial.ReadOpCode(); ok { d.SetOpCode(opCode) } else { return }

	flags := uint8(0)
	if !d.ignoreFlags {
		flags = d.status.CurrentFlags()
	}
	d.log.Debug(fmt.Sprintf("Ocdode: %d, Flags: %d, Step: %d, State: %d", d.opCode.OpCode, flags, d.status.CurrentStep(), d.clock.CurrentState()))
	d.serial.SetLines(d.opCode.Lines[flags][d.status.CurrentStep()][d.clock.CurrentState()])
}
