package instructions

import (
	"encoding/json"
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
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

	// IMM Address Mode: Immediate
	// The instruction expects the next byte to be used as a value, so we'll prep
	// the read address to point to the next byte
	IMM = iota + 1

	// IMP Address Mode: Implied
	// There is no additional data required for this instruction. The instruction
	// does something very simple like like sets a status bit. However, we will
	// target the accumulator, for instructions like PHA
	IMP

	// IZX Address Mode: Indirect X/Y
	// The supplied 8-bit address indexes a location in page 0x00. From
	// here the actual 16-bit address is read, and the contents of
	// Y Register is added to it to offset it. If the offset causes a
	// change in page then an additional clock cycle is required.
	IZX
	IZY

	// ZPG Address Mode: Zero Page
	// To save program bytes, zero page addressing allows you to absolutely address
	// a location in first 0xFF bytes of address range. Clearly this only requires
	// one byte instead of the usual two.
	ZPG

	// ZPX Address Mode: Zero Page with X/Y Offset
	// Fundamentally the same as Zero Page addressing, but the contents of the X Register
	// is added to the supplied single byte address. This is useful for iterating through
	// ranges within the first page.
	ZPX
	ZPY

	// REL Address Mode: Relative
	// This address mode is exclusive to branch instructions. The address
	// must reside within -128 to +127 of the branch instruction, i.e.
	// you cant directly branch to any address in the addressable range.
	REL

	// ABS Address Mode: Absolute
	// A full 16-bit address is loaded and used
	ABS

	// ABX Address Mode: Absolute with X/Y Offset
	// Fundamentally the same as absolute addressing, but the contents of the Y Register
	// is added to the supplied two byte address. If the resulting address changes
	// the page, an additional clock cycle is required
	ABX
	ABY

	// IND Address Mode: Indirect
	// The supplied 16-bit address is read to get the actual 16-bit address. This is
	// instruction is unusual in that it has a bug in the hardware! To emulate its
	// function accurately, we also need to emulate this bug. If the low byte of the
	// supplied address is 0xFF, then to read the high byte of the actual address
	// we need to cross a page boundary. This doesnt actually work on the chip as
	// designed, instead it wraps back around in the same page, yielding an
	// invalid actual address
	IND

	opCodes = "internal/services/instructions/opCodes.bin"
)

type OperationCodes struct {
	opCodes  []OpCode
	lookup  map[uint8]OpCode
	log     *logging.Log
}
func New(log *logging.Log) *OperationCodes {
	return &OperationCodes{
		log: log,
	}
}
type OpCode struct {
	Name     string        `json:"name"`
	OpCode   uint8         `json:"opCode"`
	AddrMode uint8         `json:"addrMode"`
	Steps    uint8         `json:"steps"`
	Lines    [3][16]uint16 `json:"lines"`
}

func (op *OperationCodes) Lookup(opcode uint8) OpCode {
	return op.lookup[opcode]
}
func (op *OperationCodes) ReadInstructions() (result bool) {
	f, err := os.Open(opCodes)
	if err != nil {
		op.log.Error(fmt.Sprintf("Failed to open opCodes file: %v", err))
		return false
	}
	defer func() {
		if err := f.Close(); err != nil {
			op.log.Error(fmt.Sprintf("Trouble closing opCodes file: %v", err))
			result = false
		}
	}()

	fi, err := f.Stat()
	if err != nil {
		op.log.Error(fmt.Sprintf("Failed to retrieve file info: %v", err))
		return false
	}

	bs := make([]byte, fi.Size())
	n, err := f.Read(bs)
	if err != nil {
		op.log.Error(fmt.Sprintf("Failed to retrieve file info: %v", err))
		return false
	} else if n != int(fi.Size()) {
		op.log.Error(fmt.Sprintf("Expected %d bytes, read %d bytes", fi.Size(), n))
		return false
	}

	err = json.Unmarshal(bs, &op.opCodes)
	if err != nil {
		op.log.Error(fmt.Sprintf("Failed to unmarshal opCodes: %v", err))
		return false
	}

	op.lookup = map[uint8]OpCode{}
	for _, instruction := range op.opCodes {
		op.lookup[instruction.OpCode] = instruction
	}

	op.log.Info("OpCodes loaded")
	return true
}
func (op *OperationCodes) WriteInstructions() (result bool) {
	f, err := os.Create(opCodes)
	if err != nil {
		op.log.Error(fmt.Sprintf("Failed to create instruction file: %v", err))
		return false
	}
	defer func() {
		if err := f.Close(); err != nil {
			op.log.Error(fmt.Sprintf("Trouble closing opCodes file: %v", err))
			result = false
		}
	}()

	bs, err := json.Marshal(op.opCodes)
	if err != nil {
		op.log.Error(fmt.Sprintf("Failed to marshal opCodes: %v", err))
		return false
	}
	_, err = f.Write(bs)
	if err != nil {
		op.log.Error(fmt.Sprintf("Failed to write opCodes: %v", err))
		return false
	}
	op.log.Info("OpCodes saved")
	return true
}