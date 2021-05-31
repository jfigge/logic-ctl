package instructions

import "fmt"

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
	FlgA        // Set source of flags

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

type Codes struct {
	line             uint
	showBlock        bool
	lines[]          string
	lineNames        []string
	lineDescriptions []string
	setDirty         func()

}
func NewCodes(setDirty func()) *Codes {
	l := Codes{
		lineNames: []string{
			/* EPROM 1a */ "TRST", "AHD1", "AHD0", "AHC1", "AHC0", "DBD2", "DBD1", "DBD0",
			/* EPROM 1b */ "ALD2", "ALD1", "ALD0", "DOUT", "IRLD", "PCIN", "PCLL", "PCLH",
			/* EPROM 2a */ "ADLH", "ADLL", "SPLD", "ALC2", "ALC1", "ALC0", "NMIR", "UNU1",
			/* EPROM 2b */ "HALT", "PAUS", "ALIB", "ALAS", "ALBS", "ASSA", "ASOX", "SHLR",
			/* EPROM 3a */ "ShSA", "ShSB", "SblA", "SblY", "SblX", "SBD0", "SBD1", "SBD2",
			/* EPROM 3b */ "FlgA", "FlgN", "FlgV", "FlgC", "FlgZ", "FlgI", "Clk1", "Clk2",
		},
		lineDescriptions: []string{
			// EPROM 1a
			"TRST Timer reset",
			"Address High driver 2-bit",
			"Address High driver 1-bit",
			"Address Bus High Constant (1-7)",
			"Address Bus High Constant (0)",
			"Data Bus driver 4-bit",
			"Data Bus driver 2-bit",
			"Data Bus driver 1-bit",
			// EPROM 1b
			"Address Bus driver 4-bit",
			"Address Bus driver 2-bit",
			"Address Bus driver 1-bit",
			"Enable data out",
			"Instruction counter load",
			"Increment program counter",
			"Load program counter from ADL",
			"Load program counter from ADH",
			// EPROM 2a
			"Load address bus high from ADH",
			"Load address bus low from ADL",
			"Stack pointer load",
			"Address Low constant (2)",
			"Address Low constant (1)",
			"Address Low constant (0)",
			"Reset NMI latch",
			"Unused",
			// EPROM 2b
			"Stop clock until reset",
			"Set clock manual step mode",
			"ALU Load Invert data bus",
			"ALU Load A Selector (0 = DB, 1 = ADL)",
			"ALU Load B Selector (0 = 0, 1 = Special Bus)",
			"ALU SUM / AND selector",
			"ALU OR / XOR selector",
			"Shift logic Left / Right selector",
			// EPROM 3a
			"Shift logic selector A (bit 1)",
			"Shift logic selector B (bit 2)",
			"Special Bus load Accumulator",
			"Special Bus load Y",
			"Special Bus load X",
			"Special Bus driver 1-bit",
			"Special Bus driver 2-bit",
			"Special Bus driver 4-bit",
			// EPROM 3b
			"Set source of flags",
			"Load flag Z/I from bus (bit 6)",
			"Load flag C from bus (bit 7)",
			"Load flag V from bus (bit 1)",
			"Load flag N from bus (bit 0 / XOR FlagA)",
			"Set bit for flag Interrupt disabling",
			"Clock Phi-1",
			"Clock Phi-2",
		},
		setDirty: setDirty,
	}

	for i := 0; i < 4; i++ {
		line := ""
		for j := 1; j <= 48; j++ {
			line += fmt.Sprintf(" %c", l.lineNames[j-1][i])
			if j % 8 == 0 {
				line += " "
			}
		}
		l.lines = append(l.lines, line)
	}

	return &l
}

func (l *Codes) IsShowBlock() bool{
	return l.showBlock
}
func (l *Codes) ShowBlock(enable bool) {
	if l.showBlock != enable {
		l.showBlock = enable
		l.setDirty()
	}
}
func (l *Codes) SetLine(line uint) {
	if line >= 0 && line < 48 {
		l.line = line
	}
}
func (l *Codes) Block() []string {
	if l.showBlock {
		return l.lines
	}
	return []string{ fmt.Sprintf(" %-40s", l.lineDescriptions[l.line ])}
}