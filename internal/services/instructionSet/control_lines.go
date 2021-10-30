package instructionSet

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

const (
	PHI1 = 0
	PHI2 = 1
) // Clock stages
const (
	CL_CENB = 1 << iota // Carry enable
	CL_FSIA // Flag select NZI-A (0=Reg/Manual(i), 1=Reg/Bus)
	CL_FMAN // Flag manual setting (0=Off, 1=On)
	CL_FSCA // Flag select C-A (0=Reg/Manual, 1=CPU/Bus)
	CL_FSCB // Flag select C-B (0=first, 1=second)
	CL_FSVB // Flag select V-B (0=first, 1=second)
	CL_FSIB // Flag select NZI-B (0=first, 1=second)
	CL_FSVA // Flag select V-A (0=Reg/Manual, 1=CPU/Bus)

	CL_SBD2 // Special Bus driver 4-bit
	CL_SBD1 // Special Bus driver 2-bit
	CL_SBD0 // Special Bus driver 1-bit
	CL_SBLX // Special Bus load X
	CL_SBLY // Special Bus load Y
	CL_SBLA // Special Bus load Accumulator
	CL_AULR // Shift direction selector (0=Left, 1=Right)
	CL_AUS2 // ALU Shift #2 selector (0=first, 1=second)

	CL_AUS1 // ALU Shift #1 selector (0=Log/Rot 1=Arth/Sum)
	CL_AUO1 // ALU Op Selector #1 (0=Sum/And, 1=Or/Xor)
	CL_AUO2 // ALU Op Selector #2 (0=first, 1=second)
	CL_AUSA // ALU Load A Selector (0=Special Bus, 1=zeros)
	CL_AUSB // ALU Load B Selector (0=DB, 1=ABL)
	CL_AUIB // ALU Load Invert data bus
	CL_PAUS // Set clock manual step mode
	CL_AULA // ALU Input A Load

	CL_AULB // ALU Input B Load
	CL_CRST // Clear reset
	CL_ALC0 // Address Low constant (0)
	CL_ALC1 // Address Low constant (1)
	CL_ALC2 // Address Low constant (2)
	CL_SPLD // Stack pointer load
	CL_AHLD // Load address bus high from ABH
	CL_ALLD // Load address bus low from ABL

	CL_PCLH // Load program counter from ABH
	CL_PCLL // Load program counter from ABL
	CL_PCIN // Increment program counter
	CL_UNU1 // Unused #1
	CL_DBRW // Data bus Read/Write (0=Write, 1=Read)
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
	CL_CTMR // Timer reset

	CL_CIOV = CL_AULA
)

type Ref struct{
	Name string
	Index int
}
var (
	mnemonics = [][]string{
		/* EPROM 1a */ {"CTMR", ""}, {"AHD0", ""}, {"AHD1", ""}, {"AHC1", ""}, {"AHC0", ""},     {"DBD0", ""}, {"DBD1", ""}, {"DBD2", ""},
		/* EPROM 1b */ {"ALD0", ""}, {"ALD1", ""}, {"ALD2", ""}, {"DBRW", ""}, {"UNU1", "CIOV"}, {"PCIN", ""}, {"PCLL", ""}, {"PCLH", ""},
		/* EPROM 2a */ {"ALLD", ""}, {"AHLD", ""}, {"SPLD", ""}, {"ALC2", ""}, {"ALC1", ""},     {"ALC0", ""}, {"CRST", ""}, {"AULB", ""},
		/* EPROM 2b */ {"AULA", ""}, {"PAUS", ""}, {"AUIB", ""}, {"AUSB", ""}, {"AUSA", ""},     {"AUO2", ""}, {"AUO1", ""}, {"AUS1", ""},
		/* EPROM 3a */ {"AUS2", ""}, {"AULR", ""}, {"SBLA", ""}, {"SBLY", ""}, {"SBLX", ""},     {"SBD0", ""}, {"SBD1", ""}, {"SBD2", ""},
		/* EPROM 3b */ {"FSVA", ""}, {"FSIB", ""}, {"FSVB", ""}, {"FSCB", ""}, {"FSCA", ""},     {"FMAN", ""}, {"FSIA", ""}, {"CENB", ""},
	}
	lineDescriptions = [][]string{
		// EPROM 1a
		{"Timer reset",""},
		{"Address High driver 1-bit",""},
		{"Address High driver 2-bit",""},
		{"Address Bus High Constant (1-7)",""},
		{"Address Bus High Constant (0)",""},
		{"Data Bus driver 1-bit",""},
		{"Data Bus driver 2-bit",""},
		{"Data Bus driver 4-bit",""},
		// EPROM 1b
		{"Address Low driver 1-bit",""},
		{"Address Low driver 2-bit",""},
		{"Address Low driver 4-bit",""},
		{"Data bus Read/Write (0=Write, 1=Read)",""},
		{"Unused #1", ""},
		{"Increment program counter",""},
		{"Load program counter from ABL",""},
		{"Load program counter from ABH",""},
		// EPROM 2a
		{"Load address bus low from ABL",""},
		{"Load address bus high from ABH",""},
		{"Stack pointer load",""},
		{"Address Low constant (2)",""},
		{"Address Low constant (1)",""},
		{"Address Low constant (0)",""},
		{"Clear Reset",""},
		{"ALU Input B Load",""},
		// EPROM 2b
		{"ALU Input A Load","Carry-in override (0=off, 1=on)"},
		{"Set clock manual step mode",""},
		{"ALU Load Invert data bus",""},
		{"ALU Load B Selector (0=DB, 1=ABL)",""},
		{"ALU Load A Selector (0=Special Bus, 1=zeros)",""},
		{"ALU Op Selector #2 (0=first, 1=second)",""},
		{"ALU Op Selector #1 (0=Sum/And, 1=Or/Xor)",""},
		{"ALU Shift #1 selector (0=Log/Rot 1=Arth/Sum)",""},
		// EPROM 3a
		{"ALU Shift #2 selector (0=first, 1=second)",""},
		{"Shift direction selector (0=Left, 1=Right)",""},
		{"Special Bus load Accumulator",""},
		{"Special Bus load Y",""},
		{"Special Bus load X",""},
		{"Special Bus driver 1-bit",""},
		{"Special Bus driver 2-bit",""},
		{"Special Bus driver 4-bit",""},
		// EPROM 3b
		{"Flag select V-A (0=Reg/Manual, 1=CPU/Bus)",""},
		{"Flag select NZI-B (0=first, 1=second)",""},
		{"Flag select V-B (0=first, 1=second)",""},
		{"Flag select C-B (0=first, 1=second)",""},
		{"Flag select C-A (0=Reg/Manual, 1=CPU/Bus)",""},
		{"Manual setting line for flags",""},
		{"Flag select NZI A (0=Reg/Manual(I), 1=Reg/Bus)",""},
		{"Enable carry-in for ALU",""},
	}

	x = uint64(CL_AHD0 | CL_AHC0 | CL_AHC1 | CL_DBD1 | CL_DBD2 | CL_PCLH | CL_PCLL | CL_DBRW | CL_PCIN | CL_ALD0 | CL_ALD1 | CL_ALD2 | CL_CRST |
		CL_ALC0 | CL_ALC1 | CL_ALC2 | CL_SPLD | CL_ALLD | CL_AHLD | CL_AUS1 | CL_AULA | CL_AULB |
		CL_AUS2 | CL_SBD2 | CL_SBD1 | CL_SBD0 | CL_SBLX | CL_SBLY | CL_SBLA)
	Defaults = [2]uint64 {x, x ^ CL_CIOV}

	OutputsDB  = map[uint64]Ref {
		0:                           {"None (0)",0},
		CL_DBD0:                     {"Accumulator",1},
		CL_DBD1:                     {"Processor status",2},
		CL_DBD0 | CL_DBD1:           {"Special bus", 3},
		CL_DBD2:                     {"Program counter high", 4},
		CL_DBD0 | CL_DBD2:           {"Program counter low", 5},
		CL_DBD1 | CL_DBD2:           {"Input data latch*", 6},
		CL_DBD0 | CL_DBD1 | CL_DBD2: {"None (7)", 7},
	}
	OutputsABH = map[uint64]Ref{
		0 :                {"Input data latch", 0},
		CL_AHD0:           {"Constants*", 1},
		CL_AHD1:           {"Program counter", 2},
		CL_AHD0 | CL_AHD1: {"Serial bus", 3},
	}
	OutputsABL = map[uint64]Ref{
		0 :                          {"Input data latch", 0},
		CL_ALD0:                     {"Program counter", 1},
		CL_ALD1:                     {"Constants", 2},
		CL_ALD0 | CL_ALD1:           {"Stack pointer", 3},
		CL_ALD2:                     {"ALU", 4},
		CL_ALD0 | CL_ALD2:           {"PC Low Register", 5},
		CL_ALD1 | CL_ALD2:           {"None (6)", 6},
		CL_ALD0 | CL_ALD1 | CL_ALD2: {"None* (7)", 7},
	}
	OutputsSB  = map[uint64]Ref{
		0 :                          {"Accumulator", 0},
		CL_SBD0:                     {"Y register", 1},
		CL_SBD1:                     {"X register", 2},
		CL_SBD0 | CL_SBD1:           {"ALU", 3},
		CL_SBD2:                     {"Stack pointer", 4},
		CL_SBD0 | CL_SBD2:           {"Data bus", 5},
		CL_SBD1 | CL_SBD2:           {"Address high bus", 6},
		CL_SBD0 | CL_SBD1 | CL_SBD2: {"None* (7)", 7},
	}

	AluA = map[uint64]Ref{
		0:       {"Special Bus*", 0},
		CL_AUSA: {"Zeros", 1},
	}
	AluB = map[uint64]Ref{
		0 :      {"Data bus*", 0},
		CL_AUSB: {"Address bus low", 1},
	}
	AluOp = map[uint64]Ref{
		0 :                                              {"Logical Shift", 0},
		CL_AUS1:                                         {"Rotation Shift", 1},
		CL_AUS2:                                         {"Arithmetic Shift", 2},
		CL_AUS1 | CL_AUS2:                               {"Add*", 3},
		CL_AUS1 | CL_AUS2 | CL_AUO1:                     {"OR", 4},
		CL_AUS1 | CL_AUS2 | CL_AUO2:                     {"AND", 5},
		CL_AUS1 | CL_AUS2 | CL_AUO1 | CL_AUO2:           {"XOR", 6},
		CL_AUIB:                                         {"Logical Shift", 7},
		CL_AUIB | CL_AUS1:                               {"Rotation Shift", 8},
		CL_AUIB | CL_AUS2:                               {"Arithmetic Shift", 9},
		CL_AUIB | CL_AUS1 | CL_AUS2:                     {"Subtract", 10},
		CL_AUIB | CL_AUS1 | CL_AUS2 | CL_AUO1:           {"OR", 11},
		CL_AUIB | CL_AUS1 | CL_AUS2 | CL_AUO2:           {"AND", 12},
		CL_AUIB | CL_AUS1 | CL_AUS2 | CL_AUO1 | CL_AUO2: {"XOR", 13},
	}
	AluDir  = map[uint64]Ref{
		0 :                          {"Left",  0},
		CL_AULR:                     {"Right", 1},
		CL_AUS1:                     {"Left",  2},
		CL_AUS1 | CL_AULR:           {"Right", 3},
		CL_AUS2:                     {"Left",  4},
		CL_AUS2 | CL_AULR:           {"Right", 5},
		CL_AUS1 | CL_AUS2:           {"",      6},
		CL_AUS1 | CL_AUS2 | CL_AULR: {"",      7},
	}

	busNames = []string {" DB", "ABH", "ABL", " SB"}//, "ALU-B", "ALU-A", "   OP", "  Dir"}
	busMaps  = []map[uint64]Ref {OutputsDB, OutputsABH, OutputsABL, OutputsSB}//, AluB, AluB, AluOp, AluDir}
	busLines = []uint64 {
		CL_DBD0 | CL_DBD1 | CL_DBD2,
		CL_AHD0 | CL_AHD1,
		CL_ALD0 | CL_ALD1 | CL_ALD2,
		CL_SBD0 | CL_SBD1 | CL_SBD2,
		//CL_AUSA, CL_AUSB, CL_AUIB | CL_AUS1 | CL_AUS2 | CL_AUO1 | CL_AUO2, CL_AUS1 | CL_AUS2 | CL_AULR,
	}
)

type BusController struct {
	xOffset   int
	yOffset   []int
	cursor    common.Coord
	terminal  *display.Terminal
	redraw    func(bool)
	ctrlLines uint64
	step      uint8
	clock     uint8
	setLines  func(step uint8, clock uint8, bit uint64, value uint8)
}
type ControlLines struct {
	xOffset   []int
	yOffset   int
	cursor    common.Coord
	terminal  *display.Terminal
	log       *logging.Log
	steps     int
	redraw    func(bool)
	setLine   func(step uint8, clock uint8, bit uint64, value uint8)
	busCntrl  *BusController
}
func NewControlLines(log *logging.Log, terminal *display.Terminal, redraw func(bool),
	setLine func(step uint8, clock uint8, bit uint64, value uint8)) *ControlLines {
	l := ControlLines{
		log:      log,
		terminal: terminal,
		cursor:   common.Coord{X:1,Y:1},
		xOffset:  []int{8,9,11,12,14,15},
		yOffset:  20,
		setLine:  setLine,
		redraw:   redraw,
		busCntrl: &BusController{
			terminal: terminal,
			redraw:   redraw,
			xOffset:  90,
			yOffset:  []int{8, 10},
			cursor:   common.Coord{X: 0, Y: 0},
			setLines: setLine,
		},
	}
	for i := 0; i < 48; i++ {
		if lineDescriptions[i][1] == "" { lineDescriptions[i][1] = lineDescriptions[i][0] }
		if mnemonics[i][1] == "" { mnemonics[i][1] = mnemonics[i][0] }
	}
	return &l
}
func (l *ControlLines) BusController() *BusController {
	return l.busCntrl
}

func (l *ControlLines) Up(n int) {
	if l.cursor.Y - n >= 1 {
		l.cursor.Y -= n
		l.PositionCursor()
		l.redraw(false)
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) Down(n int) {
	if l.cursor.Y + n <= l.steps * 2 {
		l.cursor.Y += n
		l.PositionCursor()
		l.redraw(false)
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) Left(n int) {
	if l.cursor.X - n >= 1 {
		l.cursor.X -= n
		l.PositionCursor()
		l.redraw(false)
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) Right(n int) {
	if l.cursor.X + n <= 48 {
		l.cursor.X += n
		l.PositionCursor()
		l.redraw(false)
	} else {
		l.terminal.Bell()
	}
}
func (l *ControlLines) PositionCursor() {
	l.terminal.At(l.cursor.X + l.xOffset[(l.cursor.X-1)/8], l.cursor.Y + l.yOffset)
	l.busCntrl.step  = uint8((l.cursor.Y - 1) / 2)
	l.busCntrl.clock = uint8((l.cursor.Y - 1) % 2)
}
func (l *ControlLines) CursorPosition() string {
	return fmt.Sprintf("    %02d,%02d", l.cursor.X, l.cursor.Y)
}
func (l *ControlLines) EditStep() uint8 {
	return uint8(l.cursor.Y)
}
func (l *ControlLines) SetEditStep(y uint8) {
	l.cursor.Y  = int(y)
	l.busCntrl.step  = uint8((l.cursor.Y - 1) / 2)
	l.busCntrl.clock = uint8((l.cursor.Y - 1) % 2)
}
func (l *ControlLines) SetControlLines(ctrlLines [8][2]uint64) {
	l.busCntrl.ctrlLines = ctrlLines[uint8((l.cursor.Y - 1) / 2)][uint8((l.cursor.Y - 1) % 2)]
}

func (l *ControlLines) KeyIntercept(input common.Input) bool {
	if input.KeyCode != 0 {
		switch input.KeyCode {
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
		switch input.Ascii {
		case '1', '0', 0x7F, ' ':
			if input.Ascii == '0' {
				value = 0
			} else if input.Ascii == '1' {
				value = 1
			} else if input.Ascii == 0x7F {
				value = 2
			}
			step  := uint8((l.cursor.Y - 1) / 2)
			clock := uint8((l.cursor.Y - 1) % 2)
			bit   := uint64(47 - (l.cursor.X - 1) % 64)
			l.setLine(step, clock, bit, value)
		default:
			// key not processed
			return false
		}
	}
	// key processed
	return true
}

func (l *ControlLines) LineNamesBlock(clock uint8) []string {
	return []string{ fmt.Sprintf("%s%s", lineDescriptions[l.cursor.X - 1][clock], display.ClearEnd)}
}
func (l *ControlLines) SetSteps(steps uint8) {
	l.steps = int(steps)
	if l.cursor.Y > l.steps * 2 {
		l.cursor.Y = l.steps * 2
	}
}

func (b *BusController) Up(n int) {
	if b.cursor.Y - n >= 0 {
		b.cursor.Y -= n
		b.PositionCursor()
		b.redraw(false)
	} else {
		b.terminal.Bell()
	}
}
func (b *BusController) Down(n int) {
	if b.cursor.Y + n < len(busMaps) {
		b.cursor.Y += n
		b.PositionCursor()
		b.redraw(false)
	} else {
		b.terminal.Bell()
	}
}
func (b *BusController) Left(n int) {
	if next, ok := b.findNext(-n); ok {
		b.ctrlLines = (b.ctrlLines &^ busLines[b.cursor.Y]) ^ next
		b.setLines(b.step, b.clock, b.ctrlLines, 4)
		b.PositionCursor()
	} else {
		b.terminal.Bell()
	}
}
func (b *BusController) Right(n int) {
	if next, ok := b.findNext(n); ok {
		b.ctrlLines = (b.ctrlLines &^ busLines[b.cursor.Y]) ^ next
		b.setLines(b.step, b.clock, b.ctrlLines, 4)
		b.PositionCursor()
	} else {
		b.terminal.Bell()
	}
}
func (b *BusController) findNext(offset int) (uint64, bool)  {
	busMap := busMaps[b.cursor.Y]
	currentRef := busMap[b.ctrlLines & busLines[b.cursor.Y]]
	next := currentRef.Index + offset
	for k, ref := range busMap {
		if ref.Index == next {
			return k, true
		}
	}
	return 0, false
}
func (b *BusController) KeyIntercept(input common.Input) bool {
	if input.KeyCode != 0 {
		switch input.KeyCode {
		case display.CursorUp:
			b.Up(1)
		case display.CursorDown:
			b.Down(1)
		case display.CursorLeft:
			b.Left(1)
		case display.CursorRight:
			b.Right(1)
		default:
			// keycode not processed
			return false
		}
	} else {
		return false
	}
	// key processed
	return true
}
func (b *BusController) PositionCursor() {
	b.terminal.At(b.xOffset, b.cursor.Y + b.yOffset[b.cursor.Y/4])
}
func (b *BusController) CursorPosition() string {
	return fmt.Sprintf("      %s", busNames[b.cursor.Y])
}
