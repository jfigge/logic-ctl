package instructionSet

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

const (

) // Flags  (
const (
	PHI1 = 0
	PHI2 = 1
) // Clock stages
const (
	CL_CLK2 = 1 << iota // Clock Phi-2
	CL_CLK1 // Clock Phi-1
	CL_FLGI // Custom flag value (0=Off, 1=On)
	CL_FLGZ // Flag Z/I selector bit 1 (0=Reg/I, 1=Reg/Bus)
	CL_FLGC // Flag C selector bit 0 (0=first, 1=second)
	CL_FLGV // Flag V selector bit 6 (0=first, 1=second)
	CL_FLGN // Flag N selector bit 7 (0=CPU, 1=Bus)
	CL_FLGS // Flag selector (0=Register/Custom, 1=CPU/Bus)

	CL_SBD2 // Special Bus driver 4-bit
	CL_SBD1 // Special Bus driver 2-bit
	CL_SBD0 // Special Bus driver 1-bit
	CL_SBLX // Special Bus load X
	CL_SBLY // Special Bus load Y
	CL_SBLA // Special Bus load Accumulator
	CL_AULR // Shift direction selector (0=Left, 1=Right)
	CL_AUS2 // ALU Shift #2 selector (0=first, 1=second)

	CL_AUS1 // ALU Shift #1 selector (0=Log/Rot 1=Arth/Sum)
	CL_AUO2 // ALU Op Selector #2 (0=first, 1=second)
	CL_AUO1 // ALU Op Selector #1 (0=Sum/And, 1=Or/Xor)
	CL_AUSA // ALU Load A Selector (0=Special Bus, 1=zeros)
	CL_AUSB // ALU Load B Selector (0=DB, 1=ADL)
	CL_AUIB // ALU Load Invert data bus
	CL_PAUS // Set clock manual step mode
	CL_HALT // Stop clock until reset

	CL_UNU2 // Unused #2
	CL_UNU1 // Unused #1
	CL_ALC0 // Address Low constant (0)
	CL_ALC1 // Address Low constant (1)
	CL_ALC2 // Address Low constant (2)
	CL_SPLD // Stack pointer load
	CL_AHLD // Load address bus high from ADH
	CL_ALLD // Load address bus low from ADL

	CL_PCLH // Load program counter from ADH
	CL_PCLL // Load program counter from ADL
	CL_PCIN // Increment program counter
	CL_DBRW // Data bus Read/Write (0=Read, 1=Write)
	CL_IRLD // Instruction counter load
	CL_ALD2 // Address low driver 4-bit
	CL_ALD1 // Address low driver 2-bit
	CL_ALD0 // Address low driver 1-bit

	CL_DBD2 // Data bus driver 4-bit
	CL_DBD1 // Data bus driver 2-bit
	CL_DBD0 // Data bus driver 1-bit
	CL_AHC0 // Address bus high Constant (0)
	CL_AHC1 // Address bus high Constant (1-7)
	CL_AHD1 // Address high driver 2-bit
	CL_AHD0 // Address high driver 1-bit
	CL_TRST // Timer reset
)

const (
	DB_Accumulator = /* 1 */ CL_DBD0
	DB_Flags       = /* 2 */ CL_DBD1
	DB_SB          = /* 3 */ CL_DBD0 | CL_DBD1
	DB_PC_High     = /* 4 */ CL_DBD2
	DB_PC_Low      = /* 5 */ CL_DBD0 | CL_DBD2
	DB_Input       = /* 6 */ CL_DBD1 | CL_DBD2
) // Data Bus driver
const (
	ADH_Input      = /* 0 */ 0
	ADH_Constants  = /* 1 */ CL_AHD0
	ADH_PC_High    = /* 2 */ CL_AHD1
	ADH_SB         = /* 3 */ CL_AHD0 | CL_AHD1
) // Address bus high driver
const (
	ADL_Input      = /* 0 */ 0
	ADL_PC_Low     = /* 1 */ CL_ALD0
	ADL_Constants  = /* 2 */ CL_ALD1
	ADL_SP         = /* 3 */ CL_ALD0 | CL_ALD1
	ADL_ADD        = /* 4 */ CL_ALD2
) // Address bus low driver
const (
	SB_ACC         = /* 0 */ 0
	SB_Y_REG       = /* 1 */ CL_SBD0
	SB_X_REG       = /* 2 */ CL_SBD1
	SB_ADD         = /* 3 */ CL_SBD0 | CL_SBD1
	SB_SP          = /* 4 */ CL_SBD2
	SB_DB          = /* 5 */ CL_SBD0 | CL_SBD2
	SB_ADH         = /* 6 */ CL_SBD1 | CL_SBD2
) // Special Bus driver

var (
	lineNames = []string{
		/* EPROM 1a */ "TRST", "AHD0", "AHD1", "AHC1", "AHC0", "DBD0", "DBD1", "DBD2",
		/* EPROM 1b */ "ALD0", "ALD1", "ALD2", "IRLD", "DBRW", "PCIN", "PCLL", "PCLH",
		/* EPROM 2a */ "ALLD", "AHLD", "SPLD", "ALC2", "ALC1", "ALC0", "UNU1", "UNU2",
		/* EPROM 2b */ "HALT", "PAUS", "AUIB", "AUSB", "AUSA", "AUO1", "AUO2", "AUS1",
		/* EPROM 3a */ "AUS2", "AULR", "SBLA", "SBLY", "SBLX", "SBD0", "SBD1", "SBD2",
		/* EPROM 3b */ "FLGS", "FLGN", "FLGV", "FLGC", "FLGZ", "FLGI", "CLK1", "CLK2",
	}
	lineDescriptions = []string{
		// EPROM 1a
		"Timer reset",
		"Address High driver 1-bit",
		"Address High driver 2-bit",
		"Address Bus High Constant (1-7)",
		"Address Bus High Constant (0)",
		"Data Bus driver 1-bit",
		"Data Bus driver 2-bit",
		"Data Bus driver 4-bit",
		// EPROM 1b
		"Address Low driver 1-bit",
		"Address Low driver 2-bit",
		"Address Low driver 4-bit",
		"Instruction counter load",
		"Data bus Read/Write (0=Read, 1=Write)",
		"Increment program counter",
		"Load program counter from ADL",
		"Load program counter from ADH",
		// EPROM 2a
		"Load address bus low from ADL",
		"Load address bus high from ADH",
		"Stack pointer load",
		"Address Low constant (2)",
		"Address Low constant (1)",
		"Address Low constant (0)",
		"Unused #1",
		"Unused #2",
		// EPROM 2b
		"Stop clock until reset",
		"Set clock manual step mode",
		"ALU Load Invert data bus",
		"ALU Load B Selector (0=DB, 1=ADL)",
		"ALU Load A Selector (0=Special Bus, 1=zeros)",
		"ALU Op Selector #1 (0=Sum/And, 1=Or/Xor)",
		"ALU Op Selector #2 (0=first, 1=second)",
		"ALU Shift #1 selector (0=Log/Rot 1=Arth/Sum)",
		// EPROM 3a
		"ALU Shift #2 selector (0=first, 1=second)",
		"Shift direction selector (0=Left, 1=Right)",
		"Special Bus load Accumulator",
		"Special Bus load Y",
		"Special Bus load X",
		"Special Bus driver 1-bit",
		"Special Bus driver 2-bit",
		"Special Bus driver 4-bit",
		// EPROM 3b
		"Flag selector (0=Register/Direct, 1=CPU/Bus)",
		"Flag N selector bit 7 (0=CPU, 1=Bus)",
		"Flag V selector bit 6 (0=first, 1=second)",
		"Flag C selector bit 0 (0=first, 1=second)",
		"Flag I selector bit 3 (0=first, 1=second)",
		"Direct flag value (0=Off, 1=On)",
		"Clock Phi-1",
		"Clock Phi-2",
	}
	defaults = uint64(CL_TRST | CL_AHD0 | CL_AHC0 | CL_AHC1 | CL_PCLH | CL_PCLL | CL_IRLD | CL_DBRW | CL_PCIN | CL_ALD0 | CL_ALD1 | CL_ALD2 |
		CL_HALT | CL_UNU1 | CL_ALC0 | CL_ALC1 | CL_ALC2 | CL_SPLD | CL_ALLD | CL_AHLD | CL_AUS1 |
		CL_AUS2 | CL_SBD2 | CL_SBD1 | CL_SBD0 | CL_SBLX | CL_SBLY | CL_SBLA | CL_FLGS)

	OutputsDB  = map[uint64]string {
		0:                           "None (1)",
		CL_DBD0:                     "Accumulator",
		CL_DBD1:                     "Processor status",
		CL_DBD0 | CL_DBD1:           "Special bus",
		CL_DBD2:                     "Program counter high",
		CL_DBD0 | CL_DBD2:           "Program counter low",
		CL_DBD1 | CL_DBD2:           "Input data latch",
		CL_DBD0 | CL_DBD1 | CL_DBD2: "None (8)",
	}
	OutputsADH = map[uint64]string{
		0 :                "Input data latch",
		CL_AHD0:           "Constants",
		CL_AHD1:           "Program counter",
		CL_AHD0 | CL_AHD1: "Serial bus",
	}
	OutputsADL = map[uint64]string{
		0 :                          "Input data latch",
		CL_ALD0:                     "Program counter",
		CL_ALD1:                     "Constants",
		CL_ALD0 | CL_ALD1:           "Stack pointer",
		CL_ALD2:                     "Address Hold register",
		CL_ALD0 | CL_ALD2:           "None (6)",
		CL_ALD1 | CL_ALD2:           "None (7)",
		CL_ALD0 | CL_ALD1 | CL_ALD2: "None (8)",
	}
	OutputsSB  = map[uint64]string{
		0 :                          "Accumulator",
		CL_SBD0:                     "Y register",
		CL_SBD1:                     "X register",
		CL_SBD0 | CL_SBD1:           "Address hold reg",
		CL_SBD2:                     "Stack pointer",
		CL_SBD0 | CL_SBD2:           "Data bus",
		CL_SBD1 | CL_SBD2:           "Address high bus",
		CL_SBD0 | CL_SBD1 | CL_SBD2: "None (8)",
	}

	AluA = map[uint64]string{
		0 :      "Zeros",
		CL_AUSA: "Special Bus",
	}
	AluB = map[uint64]string{
		0 :      "Data bus",
		CL_AUSB: "Address bus low",
	}
	AluOp = map[uint64]string{
		0 :                                    "Logical Shift",
		CL_AUS1:                               "Rotation Shift",
		CL_AUS2:                               "Arithmetic Shift",
		CL_AUS1 | CL_AUS2:                     "Add",
		CL_AUS1 | CL_AUS2 | CL_AUO1:           "OR",
		CL_AUS1 | CL_AUS2 | CL_AUO2:           "AND",
		CL_AUS1 | CL_AUS2 | CL_AUO1 | CL_AUO2: "XOR",
		CL_AUIB:                               "Logical Shift",
		CL_AUIB | CL_AUS1:                     "Rotation Shift",
		CL_AUIB | CL_AUS2:                     "Arithmetic Shift",
		CL_AUIB | CL_AUS1 | CL_AUS2:           "Subtract",
		CL_AUIB | CL_AUS1 | CL_AUS2 | CL_AUO1: "OR",
		CL_AUIB | CL_AUS1 | CL_AUS2 | CL_AUO2: "AND",
		CL_AUIB | CL_AUS1 | CL_AUS2 | CL_AUO1 | CL_AUO2: "XOR",
	}
	AluDir  = map[uint64]string{
		0 :                          "Left",
		CL_AULR:                     "Right",
		CL_AUS1:                     "Left",
		CL_AUS1 | CL_AULR:           "Right",
		CL_AUS2:                     "Left",
		CL_AUS2 | CL_AULR:           "Right",
		CL_AUS1 | CL_AUS2:           "",
		CL_AUS1 | CL_AUS2 | CL_AULR: "",
	}
)

type coord struct {
	x,y int
}
type ControlLines struct {
	showBlock bool
	lines     []string
	setDirty  func(bool)
	xOffset   []int
	yOffset   int
	cursor    coord
	terminal  *display.Terminal
	log       *logging.Log
	steps     int
	setLine   func(step uint8, clock uint8, bit uint64, value uint8)
}
func NewControlLines(log *logging.Log, terminal *display.Terminal, setDirty func(bool),
	                 setLine func(step uint8, clock uint8, bit uint64, value uint8)) *ControlLines {
	l := ControlLines{
		log:      log,
		terminal: terminal,
		cursor:   coord{1,1},
		xOffset:  []int{8,9,11,12,14,15},
		yOffset:  20,
		setDirty: setDirty,
		setLine:  setLine,
	}

	for i := 0; i < 4; i++ {
		line := ""
		for j := 1; j <= 48; j++ {
			colour := common.BGBrightCyan
			if j % 2 == 0 {
				colour = common.BGBrightGreen
			}
			line += fmt.Sprintf("%s%c", colour, lineNames[j-1][i])
			if j % 16 == 0 {
				line += common.Reset + "  "
			} else if j % 8 == 0 {
				line += common.Reset + " "
			}
		}
		line += common.Reset
		l.lines = append(l.lines, line)
	}

	return &l
}

func (l *ControlLines) Up(n int) {
	if l.cursor.y - n >= 1 {
		l.cursor.y -= n
		l.PositionCursor()
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) Down(n int) {
	if l.cursor.y + n <= l.steps * 2 {
		l.cursor.y += n
		l.PositionCursor()
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) Left(n int) {
	if l.cursor.x - n >= 1 {
		l.cursor.x -= n
		l.PositionCursor()
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) Right(n int) {
	if l.cursor.x + n <= 46 {
		l.cursor.x += n
		l.PositionCursor()
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) PositionCursor() {
	l.terminal.At(l.cursor.x + l.xOffset[(l.cursor.x-1)/8], l.cursor.y + l.yOffset)
	l.setDirty(false)
}
func (l *ControlLines) CursorPosition() string {
	return fmt.Sprintf("  %d,%d", l.cursor.x, l.cursor.y)
}

func (l *ControlLines) KeyIntercept(a int, k int) bool {
	if k != 0 {
		switch k {
		case display.CursorUp:
			l.Up(1)
		case display.CursorDown:
			l.Down(1)
		case display.CursorLeft:
			l.Left(1)
		case display.CursorRight:
			l.Right(1)
		default:
			// keycode not processed
			return false
		}
	} else {
		value := uint8(3)
		switch a {
		case '1', '0', 0x7F, ' ':
			if a == '0' {
				value = 0
			} else if a == '1' {
				value = 1
			} else if a == 0x7F {
				value = 2
			}
			step  := uint8((l.cursor.y - 1) / 2)
			clock := uint8((l.cursor.y - 1) % 2)
			bit   := uint64(47 - (l.cursor.x - 1) % 64)
			l.setLine(step, clock, bit, value)
		default:
			// key not processed
			return false
		}
	}
	// key processed
	return true
}

func (l *ControlLines) IsShowNames() bool{
	return l.showBlock
}
func (l *ControlLines) ShowNames(enable bool) {
	if l.showBlock != enable {
		l.showBlock = enable
		l.setDirty(true)
	}
}
func (l *ControlLines) LineNamesBlock() []string {
	if l.showBlock {
		return l.lines
	}
	return []string{ fmt.Sprintf("%s%s", lineDescriptions[l.cursor.x - 1], display.ClearEnd)}
}
func (l *ControlLines) SetSteps(steps uint8) {
	l.steps = int(steps)
}