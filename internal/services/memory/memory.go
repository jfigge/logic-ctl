package memory

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/instructions"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
	"io/ioutil"
)

const (
	read      = common.BrightBlue
	written   = common.BrightRed
	normal    = common.Blue
	current   = common.BrightGreen
	lineCount = 11
)

type Memory struct {
	memory [65536]byte
	lastAction  string
	disassembly map[uint16] string
	opCodes     *instructions.OperationCodes
	log         *logging.Log
}
func New(log *logging.Log, opCodes *instructions.OperationCodes) *Memory {
	return &Memory{
		lastAction: normal,
		opCodes:    opCodes,
		log: log,
	}
}

func (m *Memory) LoadRom(l *logging.Log, filename string) bool {
	memSize := len(m.memory)
	if bs, err := ioutil.ReadFile(filename); err != nil {
		m.log.Error(fmt.Sprintf("Failed to read ROM: %s", err))
		return false
	} else {
		percent := -1
		for i := 0; i < memSize; i++ {
			if i < len(bs) {
				m.memory[i] = bs[i]
			} else {
				m.memory[i] = 0
			}
			if i * 100 / memSize > percent {
				percent = i * 100 / memSize
				l.Progress(fmt.Sprintf("Loading ROM: %s", filename), percent)
			}
		}
		m.disassembly = m.disassemble(0, uint16(len(bs)))
		m.log.Info(fmt.Sprintf("%d byte(s) read.", len(bs)))
		return true
	}
}
func (m *Memory) disassemble(nStart, nStop uint16) map[uint16] string {
	addr := uint32(nStart)
	var value, lo, hi uint8 = 0, 0, 0

	mapLines := map[uint16]string{}
	var lineAddr uint16 = 0

	// Starting at the specified address we read an instruction
	// byte, which in turn yields information from the lookup table
	// as to how many additional bytes we need to read and what the
	// addressing mode is. I need this info to assemble human readable
	// syntax, which is different depending upon the addressing mode

	// As the instruction is decoded, a string is assembled
	// with the readable output
	for addr <= uint32(nStop) {
		lineAddr = uint16(addr)

		// Prefix line with instruction address
		sInst := fmt.Sprintf("$%s: ", display.HexAddress(lineAddr))

		// Read instruction, and get its readable name
		opcode := m.memory[addr]
		addr++

		sInst += m.opCodes.Lookup(opcode).Name + " "

		// Get operands from desired locations, and form the
		// instruction based upon its addressing mode. These
		// routines mimic the actual fetch routine of the
		// 6502 in order to get accurate data as part of the
		// instruction
		if m.opCodes.Lookup(opcode).AddrMode == instructions.IMP {
			sInst += " {IMP}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.IMM {
			value = m.memory[addr]
			addr++
			sInst += "#$" + display.HexData(value) + " {IMM}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.ZPG {
			lo = m.memory[addr]
			addr++
			hi = 0x00
			sInst += "$" + display.HexData(lo) + " {ZPG}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.ZPX {
			lo = m.memory[addr]
			addr++
			hi = 0x00
			sInst += "$" + display.HexData(lo) + ", X {ZPX}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.ZPY {
			lo = m.memory[addr]
			addr++
			hi = 0x00
			sInst += "$" + display.HexData(lo) + ", Y {ZPY}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.IZX {
			lo = m.memory[addr]
			addr++
			hi = 0x00
			sInst += "($" + display.HexData(lo) + ", X) {IZX}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.IZY {
			lo = m.memory[addr]
			addr++
			hi = 0x00
			sInst += "($" + display.HexData(lo) + "), Y {IZY}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.ABS {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst += "$" + display.HexAddress(uint16(hi) << 8 | uint16(lo)) + " {ABS}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.ABX {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst += "$" + display.HexAddress(uint16(hi) << 8 | uint16(lo)) + ", X {ABX}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.ABY {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst += "$" + display.HexAddress(uint16(hi) << 8 | uint16(lo)) + ", Y {ABY}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.IND {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst += "($" + display.HexAddress(uint16(hi) << 8 | uint16(lo)) + ") {IND}"
		} else if m.opCodes.Lookup(opcode).AddrMode == instructions.REL {
			value = m.memory[addr]
			addr++
			sInst += "$" + display.HexData(value) + " [$" + display.HexAddress(uint16(addr) + uint16(value)) + "] {REL}"
		}

		// Add the formed string to a std::map, using the instruction's
		// address as the key. This makes it convenient to look for later
		// as the instructions are variable in length, so a straight up
		// incremental index is not sufficient.
		mapLines[lineAddr] = fmt.Sprintf("%-34s", sInst)

	}
	return mapLines
}

func (m *Memory) ReadMemory(address uint16) (byte, bool) {
	m.lastAction = read
	m.log.Info(fmt.Sprintf("Memory[%s] returned %s", display.HexAddress(address), display.HexData(m.memory[address])))
	return m.memory[address], true
}
func (m *Memory) WriteMemory(address uint16, data byte) bool {
	m.memory[address] = data
	m.lastAction = written
	m.log.Info(fmt.Sprintf("Memory[%s] set to %s", display.HexAddress(address), display.HexData(m.memory[address])))
	return true
}

func (m *Memory) MemoryBlock(address uint16) (lines []string) {
	// Round down to nearest block
	start := address - address % 256
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
			line += fmt.Sprintf("%s%s ", colour, display.HexData(m.memory[start]))
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
func (m *Memory) InstructionBlock(current uint16) (lines []string) {
	highlight := uint16(lineCount / 2)
	i := int(current) - int(highlight)
	if i < 0 {
		highlight = uint16(int(highlight) + i)
		i = 9
	}

	memSize :=len(m.memory)
	for len(lines) < lineCount {
		if i > memSize || i < 0 {
			lines = append(lines, "")
		} else if m.disassembly[uint16(i)] > "" {
			if len(lines) == int(highlight) {
				lines = append(lines, common.BrightMagenta+m.disassembly[uint16(i)]+common.Reset)
			} else {
				lines = append(lines, common.Magenta+m.disassembly[uint16(i)]+common.Reset)
			}
		}
		i++
	}
	return
}