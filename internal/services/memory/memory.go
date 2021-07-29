package memory

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/instructionSet"
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
	opCodes     *instructionSet.OperationCodes
	log         *logging.Log
}
func New(log *logging.Log, opCodes *instructionSet.OperationCodes) *Memory {
	return &Memory{
		lastAction: normal,
		opCodes:    opCodes,
		log: log,
	}
}

func (m *Memory) LoadRom(l *logging.Log, filename string) bool {
	memSize := len(m.memory)
	if bs, err := ioutil.ReadFile(filename); err != nil {
		m.log.Errorf("Failed to read ROM: %s", err)
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
		m.log.Infof("%d byte(s) read.", len(bs))
		return true
	}
}
func (m *Memory) disassemble(nStart, nStop uint16) map[uint16]string {
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
		opCode := m.opCodes.Lookup(m.memory[addr])
		if opCode == nil {
			sInst = "xxx"
			addr++
		} else {
			sInst += opCode.Name + " "
			addr++

			// Get operands from desired locations, and form the
			// instruction based upon its addressing mode. These
			// routines mimic the actual fetch routine of the
			// 6502 in order to get accurate data as part of the
			// instruction
			if opCode.AddrMode == instructionSet.IMP {
				sInst += " {IMP}"
			} else if opCode.AddrMode == instructionSet.IMM {
				value = m.memory[addr]
				addr++
				sInst += "#$" + display.HexData(value) + " {IMM}"
			} else if opCode.AddrMode == instructionSet.ZPG {
				lo = m.memory[addr]
				addr++
				hi = 0x00
				sInst += "$" + display.HexData(lo) + " {ZPG}"
			} else if opCode.AddrMode == instructionSet.ZPX {
				lo = m.memory[addr]
				addr++
				hi = 0x00
				sInst += "$" + display.HexData(lo) + ", X {ZPX}"
			} else if opCode.AddrMode == instructionSet.ZPY {
				lo = m.memory[addr]
				addr++
				hi = 0x00
				sInst += "$" + display.HexData(lo) + ", Y {ZPY}"
			} else if opCode.AddrMode == instructionSet.IZX {
				lo = m.memory[addr]
				addr++
				hi = 0x00
				sInst += "($" + display.HexData(lo) + ", X) {IZX}"
			} else if opCode.AddrMode == instructionSet.IZY {
				lo = m.memory[addr]
				addr++
				hi = 0x00
				sInst += "($" + display.HexData(lo) + "), Y {IZY}"
			} else if opCode.AddrMode == instructionSet.ABS {
				lo = m.memory[addr]
				addr++
				hi = m.memory[addr]
				addr++
				sInst += "$" + display.HexAddress(uint16(hi) << 8 | uint16(lo)) + " {ABS}"
			} else if opCode.AddrMode == instructionSet.ABX {
				lo = m.memory[addr]
				addr++
				hi = m.memory[addr]
				addr++
				sInst += "$" + display.HexAddress(uint16(hi) << 8 | uint16(lo)) + ", X {ABX}"
			} else if opCode.AddrMode == instructionSet.ABY {
				lo = m.memory[addr]
				addr++
				hi = m.memory[addr]
				addr++
				sInst += "$" + display.HexAddress(uint16(hi) << 8 | uint16(lo)) + ", Y {ABY}"
			} else if opCode.AddrMode == instructionSet.IND {
				lo = m.memory[addr]
				addr++
				hi = m.memory[addr]
				addr++
				sInst += "($" + display.HexAddress(uint16(hi) << 8 | uint16(lo)) + ") {IND}"
			} else if opCode.AddrMode == instructionSet.REL {
				value = m.memory[addr]
				addr++
				sInst += "$" + display.HexData(value) + " [$" + display.HexAddress(uint16(addr) + uint16(value)) + "] {REL}"
			}
		}

		// Add the formed string to a std::map, using the instruction's
		// address as the key. This makes it convenient to look for later
		// as the instructions are variable in length, so a straight up
		// incremental index is not sufficient.
		mapLines[lineAddr] = fmt.Sprintf("%-30s", sInst)

	}
	return mapLines
}

func (m *Memory) ReadMemory(address uint16) (byte, bool) {
	m.lastAction = read
	m.log.Debugf("Memory[%s] returned %s", display.HexAddress(address), display.HexData(m.memory[address]))
	return m.memory[address], true
}
func (m *Memory) WriteMemory(address uint16, data byte) bool {
	m.memory[address] = data
	m.lastAction = written
	m.log.Infof("Memory[%s] set to %s", display.HexAddress(address), display.HexData(m.memory[address]))
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
func (m *Memory) InstructionBlock(address uint16) (lines []string) {

	totalLines := lineCount
	if totalLines > len(m.disassembly) {
		totalLines = len(m.disassembly)
	}

	addrBefore := address
	addrAfter  := address
	lines = append(lines, common.BrightMagenta+m.disassembly[address]+common.Reset)

	for len(lines) < totalLines {
		addrBefore--
		if addrBefore < address {
			if line, ok := m.disassembly[addrBefore]; ok {
				lines = append([]string{common.Magenta + line + common.Reset}, lines...)
			}
		}

		addrAfter++
		if addrAfter > address {
			if line, ok := m.disassembly[addrAfter]; ok {
				lines = append(lines, common.Magenta+line+common.Reset)
			}
		}
	}
	return
}