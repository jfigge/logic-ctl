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
	E1 = 0
	E2 = 1
	E3 = 2
) // EPROMS
const (
	E3_CLK2 = 1 << iota // Clock Phi-2
	E3_CLK1             // Clock Phi-1
	E3_FLGI             // Custom flag value (0=Off, 1=On)
	E3_FLGZ             // Flag Z/I selector bit 1 (0=Reg/I, 1=Reg/Bus)
	E3_FLGC             // Flag C selector bit 0 (0=first, 1=second)
	E3_FLGV             // Flag V selector bit 6 (0=first, 1=second)
	E3_FLGN             // Flag N selector bit 7 (0=CPU, 1=Bus)
	E3_FLGS             // Flag selector (0=Register/Custom, 1=CPU/Bus)

	E3_SBD2             // Special Bus driver 4-bit
	E3_SBD1             // Special Bus driver 2-bit
	E3_SBD0             // Special Bus driver 1-bit
	E3_SBLX             // Special Bus load X
	E3_SBLY             // Special Bus load Y
	E3_SBLA             // Special Bus load Accumulator
	E3_AULR             // Shift direction selector (0=Left, 1=Right)
	E3_AUS2             // ALU Shift #2 selector (0=first, 1=second)

	E2_AUS1             // ALU Shift #1 selector (0=Log/Rot 1=Arth/Sum)
	E2_AUO2             // ALU Op Selector #2 (0=first, 1=second)
	E2_AUO1             // ALU Op Selector #1 (0=Sum/Or, 1=And/Xor)
	E2_ALSA             // ALU Load A Selector (0=Special Bus, 1=zeros)
	E2_ALSB             // ALU Load B Selector (0=DB, 1=ADL)
	E2_AUIB             // ALU Load Invert data bus
	E2_PAUS             // Set clock manual step mode
	E2_HALT             // Stop clock until reset

	E2_UNU2             // Unused #2
	E2_UNU1             // Unused #1
	E2_ALC0             // Address Low constant (0)
	E2_ALC1             // Address Low constant (1)
	E2_ALC2             // Address Low constant (2)
	E2_SPLD             // Stack pointer load
	E2_AHLD		        // Load address bus high from ADH
	E2_ALLD		        // Load address bus low from ADL

	E1_PCLH             // Load program counter from ADH
	E1_PCLL             // Load program counter from ADL
	E1_PCIN             // Increment program counter
	E1_DBRW             // Data bus Read/Write (0=Read, 1=Write)
	E1_IRLD             // Instruction counter load
	E1_ALD2             // Address low driver 4-bit
	E1_ALD1             // Address low driver 2-bit
	E1_ALD0             // Address low driver 1-bit

	E1_DBD2             // Data bus driver 4-bit
	E1_DBD1             // Data bus driver 2-bit
	E1_DBD0             // Data bus driver 1-bit
	E1_AHC0             // Address bus high Constant (0)
	E1_AHC1             // Address bus high Constant (1-7)
	E1_AHD1             // Address high driver 2-bit
	E1_AHD0             // Address high driver 1-bit
	E1_TRST             // Timer reset
)

const (
	DB_Accumulator = /* 1 */ E1_DBD0
	DB_Flags       = /* 2 */ E1_DBD1
	DB_SB          = /* 3 */ E1_DBD0 | E1_DBD1
	DB_PC_High     = /* 4 */ E1_DBD2
	DB_PC_Low      = /* 5 */ E1_DBD0 | E1_DBD2
	DB_Input       = /* 6 */ E1_DBD1 | E1_DBD2
) // Data Bus driver
const (
	ADH_Input      = /* 0 */ 0
	ADH_Constants  = /* 1 */ E1_AHD0
	ADH_PC_High    = /* 2 */ E1_AHD1
	ADH_SB         = /* 3 */ E1_AHD0 | E1_AHD1
) // Address bus high driver
const (
	ADL_Input      = /* 0 */ 0
	ADL_PC_Low     = /* 1 */ E1_ALD0
	ADL_Constants  = /* 2 */ E1_ALD1
	ADL_SP         = /* 3 */ E1_ALD0 | E1_ALD1
	ADL_ADD        = /* 4 */ E1_ALD2
) // Address bus low driver
const (
	SB_ACC         = /* 0 */ 0
	SB_Y_REG       = /* 1 */ E3_SBD0
	SB_X_REG       = /* 2 */ E3_SBD1
	SB_ADD         = /* 3 */ E3_SBD0 | E3_SBD1
	SB_SP          = /* 4 */ E3_SBD2
	SB_DB          = /* 5 */ E3_SBD0 | E3_SBD2
	SB_ADH         = /* 6 */ E3_SBD1 | E3_SBD2
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
		"ALU Op Selector #1 (0=Sum/Or, 1=And/Xor)",
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
		"Flag selector (0=Register/Custom, 1=CPU/Bus)",
		"Flag N selector bit 7 (0=CPU, 1=Bus)",
		"Flag V selector bit 6 (0=first, 1=second)",
		"Flag C selector bit 0 (0=first, 1=second)",
		"Flag Z/I selector bit 1 (0=Reg/I, 1=Reg/Bus)",
		"Custom flag value (0=Off, 1=On)",
		"Clock Phi-1",
		"Clock Phi-2",
	}
	defaults = uint64(E1_TRST | E1_AHD0 | E1_AHC0 | E1_AHC1 | E1_PCLH | E1_PCLL | E1_IRLD | E1_DBRW | E1_PCIN | E1_ALD0 | E1_ALD1 | E1_ALD2 |
					  E2_HALT | E2_UNU1 | E2_ALC0 | E2_ALC1 | E2_ALC2 | E2_SPLD | E2_ALLD | E2_AHLD | E2_AUS1 |
					  E3_AUS2 | E3_SBD2 | E3_SBD1 | E3_SBD0 | E3_SBLX | E3_SBLY | E3_SBLA | E3_FLGS)

	OutputsDB  = map[uint64]string {
		0: "None (1)",
		E1_DBD0 : "Accumulator",
		E1_DBD1 : "Processor status",
		E1_DBD0 | E1_DBD1 : "Special bus",
		E1_DBD2 : "Program counter high",
		E1_DBD0 | E1_DBD2 : "Program counter low",
		E1_DBD1 | E1_DBD2 : "Input data latch",
		E1_DBD0 | E1_DBD1 | E1_DBD2 : "None (8)",
	}
	OutputsADH = map[uint64]string{
		0 : "Input data latch",
		E1_AHD0 : "Constants",
		E1_AHD1 : "Program counter",
		E1_AHD0 | E1_AHD1 : "Serial bus",
	}
	OutputsADL = map[uint64]string{
		0 : "Input data latch",
		E1_ALD0 : "Program counter",
		E1_ALD1 : "Constants",
		E1_ALD0 | E1_ALD1 : "Stack pointer",
		E1_ALD2 : "Address Hold register",
		E1_ALD0 | E1_ALD2 : "None (6)",
		E1_ALD1 | E1_ALD2 : "None (7)",
		E1_ALD0 | E1_ALD1 | E1_ALD2 : "None (8)",
	}
	OutputsSB  = map[uint64]string{
		0 : "Accumulator",
		E3_SBD0 : "Y register",
		E3_SBD1 : "X register",
		E3_SBD0 | E3_SBD1 : "Address hold reg",
		E3_SBD2 : "Stack pointer",
		E3_SBD0 | E3_SBD2 : "Data bus",
		E3_SBD1 | E3_SBD2 : "Address high bus",
		E3_SBD0 | E3_SBD1 | E3_SBD2 : "None (8)",
	}
	OutputsALU = map[uint64]string{
		0 : "Logical Shift",
		E2_AUS1 : "Rotation Shift",
		E3_AUS2 : "Arithmetic Shift",
		E2_AUS1 | E3_AUS2 : "Sum",
		E2_AUS1 | E3_AUS2 | E2_AUO1 : "OR",
		E2_AUS1 | E3_AUS2 | E2_AUO2 : "AND",
		E2_AUS1 | E3_AUS2 | E2_AUO1 | E2_AUO2 : "XOR",
	}
	OutputsLR = map[uint64]string{
		0 : "Left",
		E3_AULR : "Right",
	}
)

type coord struct {
	x,y int
}
type ControlLines struct {
	showBlock bool
	lines     []          string
	setDirty  func(bool)
	xOffset   []int
	yOffset   int
	cursor    coord
	terminal  *display.Terminal
	log       *logging.Log
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
		l.setDirty(false)
		l.PositionCursor()
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) Down(n int) {
	if l.cursor.y + n <= 16 {
		l.cursor.y += n
		l.setDirty(false)
		l.PositionCursor()
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) Left(n int) {
	if l.cursor.x - n >= 1 {
		l.cursor.x -= n
		l.setDirty(false)
		l.PositionCursor()
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) Right(n int) {
	if l.cursor.x + n <= 48 {
		l.cursor.x += n
		l.setDirty(false)
		l.PositionCursor()
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) PositionCursor() {
	l.terminal.At(l.cursor.x + l.xOffset[(l.cursor.x-1)/8], l.cursor.y + l.yOffset)
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