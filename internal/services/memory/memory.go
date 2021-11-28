package memory

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/instructionSet"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
	"io/ioutil"
	"path"
	"path/filepath"
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

type disassemblyEntry struct {
	line string
	address uint16
}
type memoryEntry struct {
	data             byte
	opCode           bool
	breakpoint       bool
	disassembleIndex uint16
	void             bool
}
type Memory struct {
	filename       string
	size           int
	memory         []*memoryEntry
	lastAction     string
	disassembly    []disassemblyEntry
	opCodes        *instructionSet.OpCodes
	log            *logging.Log
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
	bpfilename     string
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

func (m *Memory) LoadRom(l *logging.Log, filename string) bool {
	m.memory = make([]*memoryEntry, 65536, 65536)
	m.memory[0xfffa] = &memoryEntry{data: 0x00, opCode: false}
	m.memory[0xfffb] = &memoryEntry{data: 0x02, opCode: false}
	m.memory[0xfffc] = &memoryEntry{data: 0x00, opCode: false}
	m.memory[0xfffd] = &memoryEntry{data: 0x02, opCode: false}
	m.memory[0xfffe] = &memoryEntry{data: 0x00, opCode: false}
	m.memory[0xffff] = &memoryEntry{data: 0x02, opCode: false}
	m.filename = filename
	m.bpfilename = m.makeBPFile()
	if bs, err := ioutil.ReadFile(filename); err != nil {
		m.log.Errorf("Failed to read ROM: %s", err)
		return false
	} else if len(bs) > len(m.memory) {
		m.log.Errorf("Program too large for memory: %d", len(bs))
		return false
	} else {
		for i := 0; i < len(bs); i++ {
			m.memory[uint16(i)] = &memoryEntry{data: bs[i]}
		}
		m.size = len(bs)
		m.disassembly = m.disassemble()
		m.log.Infof("%d byte(s) read.", len(bs))
		m.loadBreakPoints()
		return true
	}
}

func (m* Memory) getEntry(address uint16) *memoryEntry {
	if entry := m.memory[address]; entry != nil {
		return entry
	} else {
		m.memory[address] = &memoryEntry{void: true}
		return m.memory[address]
	}
}
func (m *Memory) disassemble() []disassemblyEntry {
	addr := 0x200
	var lo, hi uint8 = 0, 0
	var lines []disassemblyEntry
	breakCount := 5
	for addr < int(m.size) {
		lineAddr := uint16(addr)

		// Prefix line with instruction address
		sInst := fmt.Sprintf("%%s$%s: ", display.HexAddress(lineAddr))

		// Read instruction, and get its readable name
		me := m.getEntry(uint16(addr))
		opCode := m.opCodes.Lookup(me.data)
		if opCode.Virtual {
			m.log.Infof("Encountered an invalid opcode [%s] at address [%s]. Disassembly terminated", opCode.Name, display.HexAddress(uint16(addr)))
			return lines
		} else if opCode.Name == "BRK" {
				breakCount--
				if breakCount == 0 {
				m.log.Infof("Encountered 5 consecutive breaks at address [%s]. Disassembly terminated", display.HexAddress(uint16(addr - 4)))
				return lines[:len(lines) - 4]
			}
		} else {
			breakCount = 3
		}
		me.opCode = true
		me.disassembleIndex = uint16(len(lines))
		sInst = fmt.Sprintf("%s%%s%s%%s ", sInst, opCode.Name)
		addr++

		// Get operands from desired locations, and form the
		// instruction based upon its addressing mode. These
		// routines mimic the actual fetch routine of the
		// 6502 in order to get accurate data as part of the
		// instruction
		switch opCode.AddrMode {
		case instructionSet.ACC:
			fallthrough
		case instructionSet.IMP:
			sInst = fmt.Sprintf("%s%%s%%s%%s%%s          %%sIMP", sInst)
		case instructionSet.IMM:
			lo = m.getEntry(uint16(addr)).data
			addr++
			sInst = fmt.Sprintf("%s#$%%s%s%%s%%s%%s      %%sIMM", sInst, display.HexData(lo))
		case instructionSet.ZPG:
			lo = m.getEntry(uint16(addr)).data
			addr++
			sInst = fmt.Sprintf("%s$%%s%s%%s%%s%%s       %%sZPG", sInst, display.HexData(lo))
		case instructionSet.ZPX:
			lo = m.getEntry(uint16(addr)).data
			addr++
			sInst = fmt.Sprintf("%s$%%s%s,X%%s%%s%%s     %%sZPX", sInst, display.HexData(lo))
		case instructionSet.ZPY:
			lo = m.getEntry(uint16(addr)).data
			addr++
			//sInst += "$" + display.HexData(lo) + ", Y {ZPY}"
			sInst = fmt.Sprintf("%s$%%s%s,Y%%s%%s%%s     %%sZPY", sInst, display.HexData(lo))
		case instructionSet.IZX:
			lo = m.getEntry(uint16(addr)).data
			addr++
			//sInst += "($" + display.HexData(lo) + ", X) {IZX}"
			sInst = fmt.Sprintf("%s($%%s%s,X)%%s%%s%%s   %%sIZX", sInst, display.HexData(lo))
		case instructionSet.IZY:
			lo = m.getEntry(uint16(addr)).data
			addr++
			//sInst += "($" + display.HexData(lo) + "), Y {IZY}"
			sInst = fmt.Sprintf("%s($%%s%s,Y)%%s%%s%%s   %%sIZY", sInst, display.HexData(lo))
		case instructionSet.ABS:
			lo = m.getEntry(uint16(addr)).data
			addr++
			hi = m.getEntry(uint16(addr)).data
			addr++
			sInst = fmt.Sprintf("%s$%%[6]s%s%%[7]s%%[4]s%s%%[5]s     %%[8]sABS", sInst, display.HexData(hi), display.HexData(lo))
		case instructionSet.ABX:
			lo = m.getEntry(uint16(addr)).data
			addr++
			hi = m.getEntry(uint16(addr)).data
			addr++
			sInst = fmt.Sprintf("%s$%%[6]s%s%%[7]s%%[4]s%s,X%%[5]s   %%[8]sABX", sInst, display.HexData(hi), display.HexData(lo))
		case instructionSet.ABY:
			lo = m.getEntry(uint16(addr)).data
			addr++
			hi = m.getEntry(uint16(addr)).data
			addr++
			sInst = fmt.Sprintf("%s$%%[6]s%s%%[7]s%%[4]s%s,Y%%[5]s   %%[8]sABY", sInst, display.HexData(hi), display.HexData(lo))
		case instructionSet.IND:
			lo = m.getEntry(uint16(addr)).data
			addr++
			hi = m.getEntry(uint16(addr)).data
			addr++
			sInst = fmt.Sprintf("%s($%%[6]s%s%%[7]s%%[4]s%s)%%[5]s   %%[8]sIND", sInst, display.HexData(hi), display.HexData(lo))
		case instructionSet.REL:
			lo = m.getEntry(uint16(addr)).data
			addr++
			//sInst += "$" + display.HexData(value) + " [$" + display.HexAddress(uint16(addr) + uint16(value)) + "] {REL}"
			sInst = fmt.Sprintf("%s$%%s%s%%s%%s%%s       %%sREL", sInst, display.HexData(lo))
		}

		// Add the formed string to a std::map, using the instruction's
		// address as the key. This makes it convenient to look for later
		// as the instructions are variable in length, so a straight up
		// incremental index is not sufficient.
		lines = append(lines, disassemblyEntry{
			line: fmt.Sprintf("%s%%s", sInst),
			address: lineAddr,
		})
	}
	return lines
}

func (m *Memory) ReadMemory(address uint16) (byte, bool) {
	m.lastAction = read
	me := m.getEntry(address)
	m.log.Debugf("Memory[%s] returned %s", display.HexAddress(address), display.HexData(me.data))
	return me.data, true
}
func (m *Memory) WriteMemory(address uint16, data byte) bool {
	me := m.getEntry(address)
	//if !me.opCode {
		me.data = data
		m.lastAction = written
		m.log.Infof("Memory[%s] set to %s", display.HexAddress(address), display.HexData(me.data))
		return true
	//} else {
	//	m.log.Errorf("Memory[%s] represents an opCode and cannot be changed", display.HexAddress(address))
	//	return false
	//}
}
func (m *Memory) ToggleBreakPoint(address uint16) {
	if me, found := m.getRootInstruction(address); !found {
		m.log.Warn("No valid opcode")
	} else if !me.opCode {
		m.log.Info("Selected value is data, not an opcode")
	} else {
		me.breakpoint = !me.breakpoint
		m.redraw(false)
	}
}
func (m *Memory) HasBreakPoint(address uint16) bool {
	if me, found := m.getRootInstruction(address); found {
		return me.breakpoint
	} else {
		return false
	}
}
func (m * Memory) getRootInstruction(address uint16) (*memoryEntry, bool) {
	maxLoops := 3
	var me *memoryEntry
	for me == nil || !me.opCode {
		me = m.getEntry(address)
		if me.void {
			return nil, false
		}
		address--
		maxLoops--
		if maxLoops == 0 {
			return nil, false
		}
	}
	return me, true
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
			me := m.getEntry(start)
			colour = normal
			if address == start {
				colour = m.lastAction
			}
			if colour == lastColour { colour = "" } else { lastColour = colour }
			value := display.HexData(me.data)
			if m.inputMode && m.cursor.X == j && m.cursor.Y == i {
				lastColour = common.BrightRed
				colour = common.BrightRed
				value = (m.input + "__")[:2]
			} else {
				if me.breakpoint {
					lastColour = lastColour + common.BGRed
					value = common.BGRed + value + common.Reset
				}
				if me.opCode && !me.void {
					lastColour = lastColour + common.Underline
					value = common.Underline + value + common.Reset
				}
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
func (m *Memory) InstructionBlock(instrAddr, address uint16) []string {

	me := m.getEntry(instrAddr)
	center := int(me.disassembleIndex)
	preIndex := center - lineCount / 2
	postIndex := center + lineCount / 2

	colorSetIndex := address - instrAddr
	if colorSetIndex < 0 || colorSetIndex > 2 {
		colorSetIndex = 0
	}

	var lines []string
	for i := preIndex; i <= postIndex; i++ {
		if i < 0 || i >= len(m.disassembly) {
			lines = append(lines, "                        ")
		} else {
			de := m.disassembly[i]
			line := de.line
			le := m.getEntry(de.address)
			if i != center {
				line = fmt.Sprintf(line, colorSet[3]...)
			} else {
				line = fmt.Sprintf(line, colorSet[colorSetIndex]...)
			}
			if le.breakpoint {
				line = common.BGRed + line
			}
			lines = append(lines, line)
		}
	}
	return lines
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
	address := m.displayAddress + uint16(m.cursor.X) + uint16(m.cursor.Y * 16)
	me := m.getEntry(address)
	return display.HexAddress(address) + "->" + m.opCodes.Lookup(me.data).Name
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
		return true
	} else  {
		switch input.Ascii {
		case '0','1','2','3','4','5','6','7','8','9','a','b','c','d','e','f','A','B','C','D','E','F':
			if m.inputMode {
				m.input += strings.ToUpper(string(input.Ascii))
				if len(m.input) == 2 {
					bs, _ := hex.DecodeString(m.input)
					m.lastAddress = m.displayAddress + uint16(m.cursor.X) + uint16(m.cursor.Y*16)
					me := m.getEntry(m.lastAddress)
					m.lastInput = me.data
					me.data = bs[0]
					m.inputMode = false
					m.hasLastInput = true
					m.disassembly = m.disassemble()
				}
				m.redraw(true)
			} else if input.Ascii == 'b' {
				m.ToggleBreakPoint(m.displayAddress + uint16(m.cursor.X) + uint16(m.cursor.Y*16))
				m.saveBreakPoints()
			} else {
				return false
			}

		case 13, 127:
			if !m.inputMode {
				m.input = ""
				m.inputMode = true
				m.redraw(false)
			}
		case 26:
			if m.hasLastInput {
				m.getEntry(m.lastAddress).data = m.lastInput
				m.disassembly = m.disassemble()
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
	// Key processed
	return true
}
func (m* Memory) makeBPFile() string {
	dir, filename := path.Split(m.filename)
	return path.Join(dir, fmt.Sprintf(".%s.bp", filename[:len(filename)-len(filepath.Ext(filename))]))
}
func (m *Memory) loadBreakPoints() {
	if bs, err := ioutil.ReadFile(m.bpfilename); err == nil {
		for i := uint16(0); i < uint16(len(bs)); i += 2 {
			bp := binary.LittleEndian.Uint16(bs[i : i+2])
			entry := m.getEntry(bp)
			if entry.opCode {
				entry.breakpoint = true
			}
		}
	}
}
func (m *Memory) saveBreakPoints() {
	bs := make([]byte, 0, len(m.disassembly) * 2)
	for _, de := range m.disassembly {
		me := m.getEntry(de.address)
		if me.void {
			break
		} else  if me.breakpoint {
			bs = bs[:len(bs) + 2]
			binary.LittleEndian.PutUint16(bs[len(bs) - 2:], de.address)
		}
	}
	if len(bs) > 0 {
		if err := ioutil.WriteFile(m.bpfilename, bs, 0644); err != nil {
			m.log.Warnf("Failed to write break point file: %v", err)
		} else {
			m.log.Info("Break point captured")
		}
	}
}