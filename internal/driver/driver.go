package driver

import (
	"fmt"
	"github.com/pkg/term"
	"github.td.teradata.com/sandbox/logic-ctl/internal/config"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/instructions"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/memory"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/serial"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/status"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/timing"
	"os"
	"time"
)

// Data Bus driver
const (
	DB_DEFAULT     = 0
	DB_Accumulator = 1
	DB_Flags       = 2
	DB_SB          = 3
	DB_PC_High     = 4
	DB_PC_Low      = 5
	DB_Input       = 6
)

// Address bus high driver
const (
	ADH_DEFAULT    = 1
	ADH_Input      = 0
	ADH_Constants  = 1
	ADH_PC_High    = 2
	ADH_SB         = 3
)

// Address bus low driver
const (
	ADL_DEFAULT    = 2
	ADL_Input      = 0
	ADL_PC_Low     = 1
	ADL_Constants  = 2
	ADL_SP         = 3
	ADL_ADD        = 4
)

// Special Bus driver
const (
	SB_DEFAULT     = 7
	SB_ACC         = 0
	SB_Y_REG       = 1
	SB_X_REG       = 2
	SB_ADD         = 3
	SB_SP          = 4
	SB_DB          = 5
	SB_ADH         = 6
)

const (
	// EPROM 3b
	Clk2 = iota // Clock Phi-2
	Clk1        // Clock Phi-1
	FlgI        // Set bit for flag Interrupt disabling
	FlgZ        // Load flag N from bus (bit 0 / XOR FlagA)
	FlgC        // Load flag V from bus (bit 1)
	FlgV        // Load flag C from bus (bit 7)
	FlgN        // Load flag Z/I from bus (bit 6)
	FlgSA       // Set source of flags

	// EPROM 3a
	SBD2        // Special Bus driver 4-bit
	SBD1        // Special Bus driver 2-bit
	SBD0        // Special Bus driver 1-bit
	SblX        // Special Bus load X
	SblY        // Special Bus load Y
	SblA        // Special Bus load Accumulator
	ShSB        // Shift logic selector B (bit 2)
	ShSA        // Shift logic selector A (bit 1)

	// EPROM 2b
	SHLR        // Shift logic Left / Right selector
	ASOX        // ALU OR / XOR selector
	ASSA        // ALU SUM / AND selector
	ALBS    	// ALU Load B Selector (0 = 0, 1 = Special Bus)
	ALAS		// ALU Load A Selector (0 = DB, 1 = ADL)
	ALIB		// ALU Load Invert data bus
	PAUS        // Set clock manual step mode
	HALT        // Stop clock until reset

	// EPROM 2a
	UNU1		// Unused
	NMIR        // Reset NMI latch
	ALC0        // Address Low constant (0)
	ALC1        // Address Low constant (1)
	ALC2        // Address Low constant (2)
	SPLD        // Stack pointer load
	ADLL		// Load address bus low from ADL
	ADLH		// Load address bus high from ADH

	// EPROM 1b
	PCLH        // Load program counter from ADH
	PCLL        // Load program counter from ADL
	PCIN        // Increment program counter
	IRLD        // Instruction counter load
	DOUT        // Enable data out
	ALD0		// Address Bus driver 1-bit
	ALD1		// Address Bus driver 2-bit
	ALD2		// Address Bus driver 4-bit

	// EPROM 1a
	DBD0        // Data Bus driver 1-bit
	DBD1        // Data Bus driver 2-bit
	DBD2        // Data Bus driver 4-bit
	AHC0		// Address Bus High Constant (0)
	AHC1		// Address Bus High Constant (1-7)
	AHD0        // Address High driver 1-bit
	AHD1        // Address High driver 2-bit
	TRST        // TRST Timer reset
)

type Driver struct {
	address      uint16
	display      *display.Terminal
	clock        *timing.Clock
	log          *logging.Log
	serial       *serial.Serial
	opCodes      *instructions.OperationCodes
	memory       *memory.Memory
	status       *status.Status
	ready        bool
	task         uint16
	UIs          []common.UI
	dirty        bool
	initialize   bool
}
func New() *Driver {
	d := Driver{}

	var err error
	if d.display, err = display.New(); err != nil {
		fmt.Printf("Failed to gain control of terminal: %v", err)
		os.Exit(1)
	}
	d.UIs = append(d.UIs, &d)

	d.log     = logging.New(d.reinitialize)
	d.status  = status.New(d.log)
	d.clock   = timing.New(d.log)
	d.opCodes = instructions.New(d.log)
	d.serial  = serial.New(d.log, d.clock, d.redraw)
	d.memory  = memory.New(d.log, d.opCodes)

	d.task = 0
	return &d
}

func (d *Driver) Run() {
	d.ready = true
	d.display.Cls()
	if !d.opCodes.ReadInstructions() {
		d.log.Dump()
		os.Exit(1)
	}
	if !d.memory.LoadRom(d.log, config.CLIConfig.RomFile) {
		d.log.Dump()
		os.Exit(1)
	}
	if !d.serial.Connect() {
		d.log.Dump()
		os.Exit(1)
	}

	for len(d.UIs) > 0 {
		d.UIs[0].Draw(d.display)
		a, k, e := d.ReadChar()
		if e != nil {
			fmt.Printf("Unexpected error: %v", e)
			os.Exit(1)
		}
		if a != 0 || k != 0 {
			if d.UIs[0].Process(a, k) {
				d.UIs = d.UIs[1:]
				d.UIs[0].SetDirty(true)
			}
		}
	}
	d.display.Cls()
	d.serial.Disconnect()
}

func (d *Driver) ReadChar() (ascii int, keyCode int, err error) {
	x, _ := term.Open("/dev/tty")
	if err := term.RawMode(x); err != nil {
		d.log.Error(fmt.Sprintf("Failed to access terminal RawMode: %v", err))
	}
	bs := make([]byte, 3)

	if err := x.SetReadTimeout(50 * time.Millisecond); err != nil {
		d.log.Warn("Failed to set read timeout")
	}
	if numRead, err := x.Read(bs); err != nil {
		if err.Error() != "EOF" {
			d.log.Warn("Input error.  Resetting")
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
	if err := x.Restore(); err != nil {
		d.log.Error(fmt.Sprintf("Failed to restore terminal mode: %v", err))
	}
	if err := x.Close(); err != nil {
		d.log.Error(fmt.Sprintf("Failed to close terminal input: %v", err))
	}
	return
}
func (d *Driver) SetAddress(address uint16) {
	d.address = address
	d.log.Info(fmt.Sprintf("Address set to %s", display.HexAddress(d.address)))
	d.SetDirty(false)
}
func (d *Driver) redraw() {
	d.UIs[0].SetDirty(false)
}
func (d *Driver) reinitialize() {
	d.UIs[0].SetDirty(true)
}

func (d *Driver) Draw(t *display.Terminal) {

	// Skip a redraw if we/re not ready or already drawn
	if !d.ready || (!d.dirty && !d.initialize)  {
		return
	}

	if d.initialize {
		t.Cls()
		d.initialize = false
	}

	c := t.Col()
	r := t.Row()

	// Memory
	lines := d.memory.MemoryBlock(d.address)
	for row, line := range lines {
		if ok := t.PrintAt(line, 1, row+1); !ok {
			break
		}
	}
	xOffset := len(display.StripFormatting(lines[0])) + 3

	// Clock
	t.PrintAt(common.Yellow+"Clock", xOffset+28, 1)
	t.PrintAt(d.clock.Block(), xOffset+30, 2)

	// FLags
	t.PrintAt(common.Yellow+"Flags", xOffset+6, 1)
	t.PrintAt(d.status.FlagsBlock(), xOffset, 2)

	// Timing
	t.PrintAt(common.Yellow+"Step", xOffset+6, 4)
	t.PrintAt(d.status.StepBlock(), xOffset, 5)

	// Instr
	t.PrintAt(common.Yellow+"Instructions", xOffset+3, 7)
	lines = d.memory.InstructionBlock(d.task)
	for i := 0; i < 11; i++ {
		t.PrintAt(lines[uint16(i)], xOffset, 8+i)
	}

	// Control lines
	t.PrintAt(common.Yellow+"Control Lines", 1, 20)
	for i := uint8(0); i < 7; i++ {
		colour := common.Red
		if i+1 == d.status.CurrentStep() {
			colour = common.Cyan
		}
		t.PrintAt(fmt.Sprintf("%sT%d %s 1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1 %s", common.Yellow, i+1, colour, common.Reset), 1, 21+int(i))
	}

	// Notifications
	lines = d.log.LogBlock()
	if len(lines) > 5 {
		lines = lines[:5]
	}

	str := fmt.Sprintf("%%-%ds", d.display.Cols())
	for i, line := range lines {
		d.display.PrintAt(fmt.Sprintf(str, line), 1, d.display.Rows() - i)
	}

	t.At(c, r)
	d.dirty = false
}
func (d *Driver) SetDirty(initialize bool) {
	d.dirty = true
	if initialize {
		d.initialize = true
	}
}
func (d *Driver) Process(a int, k int) bool {
	if k != 0 {
		switch k {
		case display.CursorUp:
			d.display.Up(1)
		case display.CursorDown:
			d.display.Down(1)
		case display.CursorLeft:
			d.display.Left(1)
		case display.CursorRight:
			d.display.Right(1)
		default:
			d.log.Warn(fmt.Sprintf("Unknown code: [%v]", k))
		}
	} else {
		switch a {
		case 'a':
			if a, ok := d.serial.ReadAddress(); ok {
				d.SetAddress(a)
			}
		case 'q':
			return true
		case 'h':
			d.UIs = append([]common.UI{d.log.HistoryViewer()}, d.UIs...)
		case 'p':
			d.UIs = append([]common.UI{d.serial.PortViewer()}, d.UIs...)
		case 'n':
			d.log.Info(fmt.Sprintf("Hello: %d", d.address))
			d.SetAddress(d.address + 1)
		default:
			d.log.Warn(fmt.Sprintf("Unmapped ascii code: [%c]", a))
		}
	}

	return false
}
