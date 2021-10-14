package memory

import (
	"encoding/hex"
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/instructionSet"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
	"io/ioutil"
	"strings"
)

const (
	read      = common.BrightYellow
	written   = common.BrightRed
	normal    = common.Blue
	current   = common.BrightGreen
	lineCount = 11
)

var colorSet = [][]interface{}{
		{common.BrightMagenta, common.BrightYellow, common.BrightMagenta, "", "", "", "", common.Grey, common.Reset},
		{common.BrightMagenta, "", "", common.BrightYellow, common.BrightMagenta, "", "", common.Grey, common.Reset},
		{common.BrightMagenta, "", "", "", "", common.BrightYellow, common.BrightMagenta, common.Grey, common.Reset},
		{common.Magenta, "", "", "", "", "", "", common.Grey, common.Reset},
	}

type Memory struct {
	filename       string
	size           uint16
	memory         [65536]byte
	lastAction     string
	disassembly    map[uint16]string
	opCodes        *instructionSet.OpCodes
	log            *logging.Log
	baseAddress    uint16
	displayAddress uint16
	terminal       *display.Terminal
	xOffset        []int
	yOffset        []int
	cursor         common.Coord
	redraw         func(bool)
	inputMode      bool
	input          string
	lastInput      byte
	lastAddress    uint16
	hasLastInput   bool
}
func New(log *logging.Log, opCodes *instructionSet.OpCodes, terminal *display.Terminal, redraw func(bool)) *Memory {
	return &Memory{
		lastAction:  normal,
		opCodes:     opCodes,
		log:         log,
		terminal:    terminal,
		xOffset:     []int{5, 6},
		yOffset:     []int{2, 3},
		cursor:      common.Coord{X:0, Y:0},
		input:       "xx",
		redraw:      redraw,
	}
}

func (m *Memory) LoadRom(l *logging.Log, filename string, baseAddress uint16) bool {
	m.baseAddress = baseAddress
	m.filename = filename
	if bs, err := ioutil.ReadFile(filename); err != nil {
		m.log.Errorf("Failed to read ROM: %s", err)
		return false
	} else {
		for i := uint16(0); i < uint16(len(bs)); i++ {
			m.memory[i+baseAddress] = bs[i]
		}
		m.memory[0xfffc] = 0x00
		m.memory[0xfffd] = 0x80
		m.disassembly = m.disassemble(uint16(len(bs)))
		m.log.Infof("%d byte(s) read.", len(bs))
		return true
	}
}
func (m *Memory) disassemble(size uint16) map[uint16]string {
	m.size = size
	addr := m.baseAddress
	var lo, hi uint8 = 0, 0
	mapLines := map[uint16]string{}
	var lineAddr uint16 = 0

	for addr >= addr + size {
		lineAddr = addr

		// Prefix line with instruction address
		sInst := fmt.Sprintf("%%s$%s: ", display.HexAddress(lineAddr))

		// Read instruction, and get its readable name
		opCode := m.opCodes.Lookup(m.memory[addr])
		sInst = fmt.Sprintf("%s%%s%s%%s ", sInst, opCode.Name)
		addr++

		// Get operands from desired locations, and form the
		// instruction based upon its addressing mode. These
		// routines mimic the actual fetch routine of the
		// 6502 in order to get accurate data as part of the
		// instruction
		if opCode.AddrMode == instructionSet.IMP {
			sInst = fmt.Sprintf("%s%%s%%s%%s%%s          %%sIMP", sInst)
		} else if opCode.AddrMode == instructionSet.IMM {
			lo = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s#$%%s%s%%s%%s%%s      %%sIMM", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ZPG {
			lo = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%s%s%%s%%s%%s       %%sZPG", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ZPX {
			lo = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%s%s,X%%s%%s%%s     %%sZPX", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ZPY {
			lo = m.memory[addr]
			addr++
			//sInst += "$" + display.HexData(lo) + ", Y {ZPY}"
			sInst = fmt.Sprintf("%s$%%s%s,Y%%s%%s%%s     %%sZPY", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.IZX {
			lo = m.memory[addr]
			addr++
			//sInst += "($" + display.HexData(lo) + ", X) {IZX}"
			sInst = fmt.Sprintf("%s($%%s%s,X)%%s%%s%%s   %%sIZX", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.IZY {
			lo = m.memory[addr]
			addr++
			//sInst += "($" + display.HexData(lo) + "), Y {IZY}"
			sInst = fmt.Sprintf("%s($%%s%s,Y)%%s%%s%%s   %%sIZY", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ABS {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%[6]s%s%%[7]s%%[4]s%s%%[5]s     %%[8]sABS", sInst, display.HexData(hi), display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ABX {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%[6]s%s%%[7]s%%[4]s%s,X%%[5]s   %%[8]sABX", sInst, display.HexData(hi), display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ABY {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%[6]s%s%%[7]s%%[4]s%s,Y%%[5]s   %%[8]sABY", sInst, display.HexData(hi), display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.IND {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s($%%[6]s%s%%[7]s%%[4]s%s)%%[5]s   %%[8]sIND", sInst, display.HexData(hi), display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.REL {
			lo = m.memory[addr]
			addr++
			//sInst += "$" + display.HexData(value) + " [$" + display.HexAddress(uint16(addr) + uint16(value)) + "] {REL}"
			sInst = fmt.Sprintf("%s$%%s%s%%s%%s%%s       %%sREL", sInst, display.HexData(lo))
		}

		// Add the formed string to a std::map, using the instruction's
		// address as the key. This makes it convenient to look for later
		// as the instructions are variable in length, so a straight up
		// incremental index is not sufficient.
		mapLines[lineAddr] = fmt.Sprintf("%s%%s", sInst)
	}
	return mapLines
}

func (m *Memory) ReadMemory(address uint16) (byte, bool) {
	m.lastAction = read
	m.log.Debugf("Memory[%s] returned %s", display.HexAddress(address), display.HexData(m.memory[address]))
	return m.memory[address], true
}
func (m *Memory) WriteMemory(address uint16, data byte) bool {
	//if address < 0x6000 {
	//	m.log.Errorf("Cannot write %s to ROM address %s", display.HexData(data), display.HexAddress(address))
	//	return false
	//}
	m.memory[address] = data
	m.lastAction = written
	m.log.Infof("Memory[%s] set to %s", display.HexAddress(address), display.HexData(m.memory[address]))
	return true
}

func (m *Memory) MemoryBlock(address uint16) (lines []string) {
	// Round down to nearest block
	start := address - address % 256
	if m.displayAddress != start {
		m.hasLastInput = false
		m.displayAddress = start
	}
	lines = append(lines, common.Yellow+ "     0  1  2  3  4  5  6  7   8  9  A  B  C  D  E  F" +common.Reset)

	colour, lastColour, line := normal, "", ""
	var second = 0
	for i := 0; i < 16; i++ {
		line = fmt.Sprintf("%s%s%s%s ", common.Yellow, display.HEX[start >> 12], display.HEX[start >> 8 & 15], display.HEX[start >> 4 & 15])
		for j := 0; j < 16; j++ {
			colour = normal
			if address == start {
				colour = m.lastAction
			}
			if colour == lastColour { colour = "" } else { lastColour = colour }
			value := display.HexData(m.memory[start])
			if m.inputMode && m.cursor.X == j && m.cursor.Y == i {
				lastColour = common.BrightRed
				colour = common.BrightRed
				value = (m.input + "__")[:2]
			}
			line += fmt.Sprintf("%s%s ", colour, value)
			if j == 7 {
				line += " "
			}
			start++
		}
		lastColour = ""
		lines = append(lines, fmt.Sprintf("%s%s", line, common.Reset))
		if i == 7 {
			second++
			lines = append(lines, "")
		}
	}
	m.lastAction = current
	return lines
}
func (m *Memory) InstructionBlock(instrAddr, address uint16) (lines []string) {

	totalLines := lineCount
	addrBefore := instrAddr
	addrAfter  := instrAddr

	colorSetIndex := address - instrAddr
	if colorSetIndex < 0 || colorSetIndex > 2 {
		colorSetIndex = 0
	}
	if line, ok := m.disassembly[instrAddr]; ok {
		lines = append(lines, fmt.Sprintf(line, colorSet[colorSetIndex]...))

		for len(lines) < totalLines {
			addrBefore-- // wraps around to bottom (0xffff) of memory
			if addrBefore < instrAddr {
				if line, ok = m.disassembly[addrBefore]; ok {
					lines = append([]string{fmt.Sprintf(line, colorSet[3]...)}, lines...)
				}
			}

			addrAfter++ // wraps around to top (0) of memory
			if addrAfter > instrAddr {
				if line, ok = m.disassembly[addrAfter]; ok {
					lines = append(lines, fmt.Sprintf(line, colorSet[3]...))
				}
			}
		}
	}
	return
}

func (m *Memory) Up(n int) {
	if m.cursor.Y - n >= 0 {
		m.cursor.Y -= n
		m.PositionCursor()
		m.redraw(false)
	} else {
		m.terminal.Bell()
	}
}
func (m *Memory) Down(n int) {
	if m.cursor.Y + n <= 15 {
		m.cursor.Y += n
		m.PositionCursor()
		m.redraw(false)
	} else {
		m.terminal.Bell()
	}
}
func (m *Memory) Left(n int) {
	if m.cursor.X - n >= 0 {
		m.cursor.X -= n
		m.PositionCursor()
		m.redraw(false)
	} else {
		m.terminal.Bell()
	}
}
func (m *Memory) Right(n int) {
	if m.cursor.X + n <= 15 {
		m.cursor.X += n
		m.PositionCursor()
		m.redraw(false)
	} else {
		m.terminal.Bell()
	}
}
func (m *Memory) PositionCursor() {
	m.terminal.At(m.cursor.X * 3 + m.xOffset[(m.cursor.X)/8] +len(m.input), m.cursor.Y + m.yOffset[(m.cursor.Y)/8])
}
func (m *Memory) CursorPosition() string {
	return display.HexAddress(m.displayAddress + uint16(m.cursor.X) + uint16(m.cursor.Y * 16))
}
func (m *Memory) KeyIntercept(input common.Input) bool {
	if input.KeyCode != 0 && !m.inputMode {
		switch input.KeyCode {
		case display.CursorUp:
			m.Up(1)
		case display.CursorDown:
			m.Down(1)
		case display.CursorLeft:
			m.Left(1)
		case display.CursorRight:
			m.Right(1)
		default:
			// keycode not processed
			return false
		}
	} else {
		switch input.Ascii {
		case '0','1','2','3','4','5','6','7','8','9','a','b','c','d','e','f','A','B','C','D','E','F':
			m.input += strings.ToUpper(string(input.Ascii))
			if len(m.input) == 2 {
				bs, _ := hex.DecodeString(m.input)
				m.lastAddress = m.displayAddress + uint16(m.cursor.X) + uint16(m.cursor.Y * 16)
				m.lastInput = m.memory[m.lastAddress]
				m.memory[m.lastAddress] = bs[0]
				m.inputMode = false
				m.hasLastInput = true
				m.disassembly = m.disassemble(m.size)
			}
			m.redraw(true)

		case 13, 127:
			if !m.inputMode {
				m.input = ""
				m.inputMode = true
				m.redraw(false)
			}
		case 26:
			if m.hasLastInput {
				m.memory[m.lastAddress] = m.lastInput
				m.disassembly = m.disassemble(m.size)
				m.hasLastInput = false
				m.redraw(false)
			}
		case 27:
			m.inputMode = false
			m.input = ""
			m.redraw(false)
		default:
			// key not processed
			return false
		}
	}
	// key processed
	return true
}