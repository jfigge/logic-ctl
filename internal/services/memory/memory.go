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
	memory [65536]byte
	lastAction  string
	disassembly map[uint16]string
	opCodes     *instructionSet.OpCodes
	log         *logging.Log
	baseAddress uint16
}
func New(log *logging.Log, opCodes *instructionSet.OpCodes) *Memory {
	return &Memory{
		lastAction: normal,
		opCodes:    opCodes,
		log: log,
	}
}

func (m *Memory) LoadRom(l *logging.Log, filename string, baseAddress uint16) bool {
	m.baseAddress = baseAddress
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
			sInst = fmt.Sprintf("%s%%s%%s%%s%%s        %%s{IMP}", sInst)
		} else if opCode.AddrMode == instructionSet.IMM {
			lo = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s#$%%s%s%%s%%s%%s    %%s{IMM}", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ZPG {
			lo = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%s%s%%s%%s%%s     %%s{ZPG}", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ZPX {
			lo = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%s%s,X%%s%%s%%s   %%s{ZPX}", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ZPY {
			lo = m.memory[addr]
			addr++
			//sInst += "$" + display.HexData(lo) + ", Y {ZPY}"
			sInst = fmt.Sprintf("%s$%%s%s,Y%%s%%s%%s   %%s{ZPY}", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.IZX {
			lo = m.memory[addr]
			addr++
			//sInst += "($" + display.HexData(lo) + ", X) {IZX}"
			sInst = fmt.Sprintf("%s($%%s%s,X)%%s%%s%%s %%s{IZX}", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.IZY {
			lo = m.memory[addr]
			addr++
			//sInst += "($" + display.HexData(lo) + "), Y {IZY}"
			sInst = fmt.Sprintf("%s($%%s%s,Y)%%s%%s%%s %%s{IZY}", sInst, display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ABS {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%[6]s%s%%[7]s%%[4]s%s%%[5]s   %%[8]s{ABS}", sInst, display.HexData(hi), display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ABX {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%[6]s%s%%[7]s%%[4]s%s,X%%[5]s %%[8]s{ABX}", sInst, display.HexData(hi), display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.ABY {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s$%%[6]s%s%%[7]s%%[4]s%s,Y%%[5]s %%[8]s{ABY}", sInst, display.HexData(hi), display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.IND {
			lo = m.memory[addr]
			addr++
			hi = m.memory[addr]
			addr++
			sInst = fmt.Sprintf("%s($%%[6]s%s%%[7]s%%[4]s%s)%%[5]s %%[8]s{IND}", sInst, display.HexData(hi), display.HexData(lo))
		} else if opCode.AddrMode == instructionSet.REL {
			lo = m.memory[addr]
			addr++
			//sInst += "$" + display.HexData(value) + " [$" + display.HexAddress(uint16(addr) + uint16(value)) + "] {REL}"
			sInst = fmt.Sprintf("%s$%%s%s%%s%%s%%s       %%s{REL}", sInst, display.HexData(lo))
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