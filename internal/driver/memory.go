package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

const (
	read    = BrightBlue
	written = BrightRed
	normal  = Blue
)

var memory [65536]byte
var address    uint16
var lastAction string = normal
var Disassemly map[uint16] string

func LoadRom(filename string) DisplayMessage {
	memSize := len(memory)
	if bytes, err := ioutil.ReadFile(filename); err != nil {
		return DisplayMessage{ fmt.Sprintf("Failed to read ROM: %s", err), true }
	} else {
		percent := -1
		for i := 0; i < memSize; i++ {
			if i < len(bytes) {
				memory[i] = bytes[i]
			} else {
				memory[i] = 0
			}
			if i * 100 / memSize > percent {
				percent = i * 100 / memSize
				display.Progress(fmt.Sprintf("Loading ROM: %s", filename), percent)
			}
		}
		Disassemly = Disassemble(0, uint16(len(bytes)))
		return DisplayMessage{ fmt.Sprintf("%d byte(s) read.", len(bytes)), false }
	}
}

func SetAddress(hi byte, lo byte) DisplayMessage {
	if addr, err := binary.ReadUvarint(bytes.NewBuffer([]byte{hi, lo})); err != nil {
		return DisplayMessage {
			fmt.Sprintf("Invalid address: %s", err.Error()),
			true,
		}
	} else {
		address = uint16(addr)
		return DisplayMessage {
			fmt.Sprintf("Address set to %s",  HexAddress(address)),
			false,
		}
	}
}

func ReadData() (byte, DisplayMessage) {
	lastAction = read
	return memory[address], DisplayMessage{
		fmt.Sprintf("Memory[%s] returned %s", HexAddress(address), HexData(memory[address])),
		false,
	}
}

func WriteData(data byte) DisplayMessage {
	memory[address] = data
	lastAction = written
	return DisplayMessage{
		fmt.Sprintf("Memory[%s] set to %s", HexAddress(address), HexData(memory[address])),
		false,
	}
}

func MemoryBlock(start uint16) (lines []string) {
	if uint32(start) + 256 > 65535 {
		start = 65535 - 256
	}

	lines = append(lines, Yellow + "     0  1  2  3  4  5  6  7   8  9  A  B  C  D  E  F" + Reset)

	var colour, last, line string
	var second = 0
	for i := 0; i < 16; i++ {
		line = fmt.Sprintf("%s%s%s%s ", Yellow, HEX[address >> 12], HEX[address >> 12 & 15], HEX[address >> 8 & 15])
		for j := 0; j < 16; j++ {
			if address == start {
				colour = lastAction
			} else if last != normal {
				colour = normal
			} else {
				colour = ""
			}
			last = colour
			line += fmt.Sprintf("%s%s ", colour, HexData(memory[start]))
			if j == 7 {
				line += " "
			}
			start++
		}
		last = ""
		lines = append(lines, fmt.Sprintf("%s%s", line, Reset))
		if i == 7 {
			second++
			lines = append(lines, "")
		}
	}
	lastAction = normal
	return lines
}

func Disassemble(nStart, nStop uint16) map[uint16] string {
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
		sInst := fmt.Sprintf("$%s: ", HexAddress(lineAddr))

		// Read instruction, and get its readable name
		opcode := memory[addr]
		addr++

		sInst += lookup[opcode].Name + " "

		// Get oprands from desired locations, and form the
		// instruction based upon its addressing mode. These
		// routines mimmick the actual fetch routine of the
		// 6502 in order to get accurate data as part of the
		// instruction
		if lookup[opcode].AddrMode == IMP {
			sInst += " {IMP}"
		} else if lookup[opcode].AddrMode == IMM {
			value = memory[addr]
			addr++
			sInst += "#$" + HexData(value) + " {IMM}"
		} else if lookup[opcode].AddrMode == ZPG {
			lo = memory[addr]
			addr++
			hi = 0x00
			sInst += "$" + HexData(lo) + " {ZPG}"
		} else if lookup[opcode].AddrMode == ZPX {
			lo = memory[addr]
			addr++
			hi = 0x00
			sInst += "$" + HexData(lo) + ", X {ZPX}"
		} else if lookup[opcode].AddrMode == ZPY {
			lo = memory[addr]
			addr++
			hi = 0x00
			sInst += "$" + HexData(lo) + ", Y {ZPY}"
		} else if lookup[opcode].AddrMode == IZX {
			lo = memory[addr]
			addr++
			hi = 0x00
			sInst += "($" + HexData(lo) + ", X) {IZX}"
		} else if lookup[opcode].AddrMode == IZY {
			lo = memory[addr]
			addr++
			hi = 0x00
			sInst += "($" + HexData(lo) + "), Y {IZY}"
		} else if lookup[opcode].AddrMode == ABS {
			lo = memory[addr]
			addr++
			hi = memory[addr]
			addr++
			sInst += "$" + HexAddress(uint16(hi << 8) | uint16(lo)) + " {ABS}"
		} else if lookup[opcode].AddrMode == ABX {
			lo = memory[addr]
			addr++
			hi = memory[addr]
			addr++
			sInst += "$" + HexAddress(uint16(hi << 8) | uint16(lo)) + ", X {ABX}"
		} else if lookup[opcode].AddrMode == ABY {
			lo = memory[addr]
			addr++
			hi = memory[addr]
			addr++
			sInst += "$" + HexAddress(uint16(hi << 8) | uint16(lo)) + ", Y {ABY}"
		} else if lookup[opcode].AddrMode == IND {
			lo = memory[addr]
			addr++
			hi = memory[addr]
			addr++
			sInst += "($" + HexAddress(uint16(hi << 8) | uint16(lo)) + ") {IND}"
		} else if lookup[opcode].AddrMode == REL {
			value = memory[addr]
			addr++
			sInst += "$" + HexData(value) + " [$" + HexAddress(uint16(addr) + uint16(value)) + "] {REL}"
		}

		// Add the formed string to a std::map, using the instruction's
		// address as the key. This makes it convenient to look for later
		// as the instructions are variable in length, so a straight up
		// incremental index is not sufficient.
		mapLines[lineAddr] = sInst

	}
	return mapLines
}