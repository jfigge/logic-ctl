package driver

import (
	"encoding/json"
	"fmt"
	"os"
)

// This is the disassembly function. Its workings are not required for emulation.
// It is merely a convenience function to turn the binary instruction code into
// human readable form. Its included as part of the emulator because it can take
// advantage of many of the CPUs internal operations to do this.

const (
	///////////////////////////////////////////////////////////////////////////////
	// ADDRESSING MODES

	// The 6502 can address between 0x0000 - 0xFFFF. The high byte is often referred
	// to as the "page", and the low byte is the offset into that page. This implies
	// there are 256 pages, each containing 256 bytes.
	//
	// Several addressing modes have the potential to require an additional clock
	// cycle if they cross a page boundary. This is combined with several instructions
	// that enable this additional clock cycle. So each addressing function returns
	// a flag saying it has potential, as does each instruction. If both instruction
	// and address function return 1, then an additional clock cycle is required.

	//Address Mode: Immediate
	// The instruction expects the next byte to be used as a value, so we'll prep
	// the read address to point to the next byte
	IMM = iota + 1

	// Address Mode: Implied
	// There is no additional data required for this instruction. The instruction
	// does something very simple like like sets a status bit. However, we will
	// target the accumulator, for instructions like PHA
	IMP

	// Address Mode: Indirect X/Y
	// The supplied 8-bit address indexes a location in page 0x00. From
	// here the actual 16-bit address is read, and the contents of
	// Y Register is added to it to offset it. If the offset causes a
	// change in page then an additional clock cycle is required.
	IZX
	IZY

	// Address Mode: Zero Page
	// To save program bytes, zero page addressing allows you to absolutely address
	// a location in first 0xFF bytes of address range. Clearly this only requires
	// one byte instead of the usual two.
	ZPG

	// Address Mode: Zero Page with X/Y Offset
	// Fundamentally the same as Zero Page addressing, but the contents of the X Register
	// is added to the supplied single byte address. This is useful for iterating through
	// ranges within the first page.
	ZPX
	ZPY

	// Address Mode: Relative
	// This address mode is exclusive to branch instructions. The address
	// must reside within -128 to +127 of the branch instruction, i.e.
	// you cant directly branch to any address in the addressable range.
	REL

	// Address Mode: Absolute
	// A full 16-bit address is loaded and used
	ABS

	// Address Mode: Absolute with X/Y Offset
	// Fundamentally the same as absolute addressing, but the contents of the Y Register
	// is added to the supplied two byte address. If the resulting address changes
	// the page, an additional clock cycle is required
	ABX
	ABY

	// Address Mode: Indirect
	// The supplied 16-bit address is read to get the actual 16-bit address. This is
	// instruction is unusual in that it has a bug in the hardware! To emulate its
	// function accurately, we also need to emulate this bug. If the low byte of the
	// supplied address is 0xFF, then to read the high byte of the actual address
	// we need to cross a page boundary. This doesnt actually work on the chip as
	// designed, instead it wraps back around in the same page, yielding an
	// invalid actual address
	IND
)

var (
	instructions Instructions
	lookup map[uint8]Instruction
)



type Instructions []Instruction
type Instruction struct {
	Name     string       `json:"name"`
	OpCode   uint8        `json:"opCode"`
	AddrMode uint8        `json:"addrMode"`
	Steps    uint8        `json:"steps"`
	Lines    [7][3]uint16 `json:"lines"`
}

func ReadInstructions() *DisplayMessage {
	f, err := os.Open("instructions/instructions.bin")
	if err != nil {
		return &DisplayMessage{ fmt.Sprintf("Failed to open instruction file: %v", err), true }
	}
	defer func() {
		if err := f.Close(); err != nil {
			display.Warn(fmt.Sprintf("Trouble closing instruction file: %v", err))
		}
	}()

	fi, err := f.Stat()
	if err != nil {
		return &DisplayMessage{ fmt.Sprintf("Failed to retrieve file info: %v", err), true }
	}

	bs := make([]byte, fi.Size())
	n, err := f.Read(bs)
	if err != nil {
		return &DisplayMessage{ fmt.Sprintf("Failed to read instructions: %v", err), true }
	} else if n != int(fi.Size()) {
		return &DisplayMessage{ fmt.Sprintf("Expected %d bytes, read %d bytes", fi.Size(), n), true }
	}

	err = json.Unmarshal(bs, &instructions)
	if err != nil {
		return &DisplayMessage{ fmt.Sprintf("Failed to unmarshal instructions: %v", err), true }
	}

	lookup = map[uint8]Instruction{}
	for _, instruction := range instructions {
		lookup[instruction.OpCode] = instruction
	}

	return &DisplayMessage { "Instructions loaded", false }
}

func WriteInstructions() *DisplayMessage {
	f, err := os.Create("instructions/instructions.bin")
	if err != nil {
		return &DisplayMessage{ fmt.Sprintf("Failed to create instruction file: %v", err), true }
	}
	defer func() {
		if err := f.Close(); err != nil {
			display.Warn(fmt.Sprintf("Trouble closing instruction file: %v", err))
		}
	}()

	bs, err := json.Marshal(instructions)
	if err != nil {
		return &DisplayMessage{ fmt.Sprintf("Failed to marshal instructions: %v", err), true }
	}
	_, err = f.Write(bs)
	if err != nil {
		return &DisplayMessage{ fmt.Sprintf("Failed to write instructions: %v", err), true }
	}
	return &DisplayMessage { "Instructions saved", false }
}


func InstructionsBlock(start, length uint16) (lines []string) {
	i := start
	for uint16(len(lines)) < length {
		if int(i) > len(memory) {
			lines = append(lines, "")
		} else if Disassemly[i] > "" {
			if len(lines) == 5 {
				lines = append(lines, BrightMagenta+Disassemly[i]+Reset)
			} else {
				lines = append(lines, Magenta+Disassemly[i]+Reset)
			}
		}
		i++
	}
	return
}

