package instructionSet

import (
	"encoding/json"
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
	"os"
	"strings"
)

// This is the disassembly function. Its workings are not required for emulation.
// It is merely a convenience function to turn the binary instruction code into
// human readable form. Its included as part of the emulator because it can take
// advantage of many of the CPUs internal operations to do this.

const (
	///////////////////////////////////////////////////////////////////////////////
	// ADDRESSING MODES
	//
	// The 6502 can address between 0x0000 - 0xFFFF. The high byte is often referred
	// to as the "page", and the low byte is the offset into that page. This implies
	// there are 256 pages, each containing 256 bytes.
	//
	// Several addressing modes have the potential to require an additional clock
	// cycle if they cross a page boundary. This is combined with several instructions
	// that enable this additional clock cycle. So each addressing function returns
	// a flag saying it has potential, as does each instruction. If both instruction
	// and address function return 1, then an additional clock cycle is required.

	// IMM Address Mode: IMM,
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
	//
	// Wrap-Around
	// Use caution with indexed zero page operations as they are subject to wrap-around. For example, if the X
	// register holds $FF and you execute LDA $80,X you will not access $017F as you might expect; instead you
	// access $7F i.e. $80-1. This characteristic can be used to advantage but make sure your code is well commented.
	//
	// It is possible, however, to access $017F when X = $FF by using the ABX, addressing mode of LDA $80,X. That is,
	//instead of:
	//    LDA $80,X    ; ZeroPage,X - the resulting object code is: B5 80
	// which accesses $007F when X=$FF, use:
	//    LDA $0080,X  ; ABX, - the resulting object code is: BD 80 00
    // which accesses $017F when X = $FF (a at cost of one additional byte and one additional cycle). All of the
    // ZeroPage,X and ZeroPage,Y instructions except STX ZeroPage,Y and STY ZeroPage,X have a corresponding ABX, and
    // ABY, instruction. Unfortunately, a lot of 6502 assemblers don't have an easy way to force Absolute addressing,
    // i.e. most will assemble a LDA $0080,X as B5 80. One way to overcome this is to insert the bytes using the .BYTE
    // pseudo-op (on some 6502 assemblers this pseudo-op is called DB or DFB, consult the assembler documentation) as
    // follows:
	//    .BYTE $BD,$80,$00  ; LDA $0080,X (absolute,X addressing mode)
	// The comment is optional, but highly recommended for clarity.
	// In cases where you are writing code that will be relocated you must consider wrap-around when assigning dummy
	// values for addresses that will be adjusted. Both zero and the semi-standard $FFFF should be avoided for dummy
	// labels. The use of zero or zero page values will result in assembled code with zero page opcodes when you wanted
	// absolute codes. With $FFFF, the problem is in addresses+1 as you wrap around to page 0.
	IZX
	IZY

	// ZPG Address Mode: ZPG, 
	// To save program bytes, zero page addressing allows you to absolutely address
	// a location in first 0xFF bytes of address range. Clearly this only requires
	// one byte instead of the usual two.
	ZPG

	// ZPX Address Mode: ZPG,  with X/Y Offset
	// Fundamentally the same as ZPG,  addressing, but the contents of the X Register
	// is added to the supplied single byte address. This is useful for iterating through
	// ranges within the fi— page.
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

	// Operates on the Accumulator and not any address

	ACC

	opCodes     = "internal/services/instructionSet/opCodes.bin"
	timingColor = common.Yellow
	clockColour = common.Cyan
	lineColor   = common.Blue
	PresetChg   = common.BrightYellow
	defaultChg  = common.BrightRed
	activeLine  = common.BrightCyan
	activeClock = common.BrightGreen
	timeMarker  = common.BrightWhite
)

var (
	addressModeNames = []string{"", "IMM", "IMP", "IZX", "IZY", "ZPG", "ZPX", "ZPY", "REL", "ABS", "ABX", "ABY", "IND", "ACC"}
)

type OperationCodes struct {
	opCodes      []OpCode
	lookup       map[uint8]*OpCode
	log          *logging.Log
}
func New(log *logging.Log) *OperationCodes {
	operationCodes := &OperationCodes{
		log:    log,
		lookup: defineOpCodes(),
	}
	for _, oc := range operationCodes.lookup {
		oc.Presets = oc.Lines
		//for flags := uint8(0); flags < 16; flags++ {
		//	for timing := uint8(0); timing < 8; timing++ {
		//		oc.Presets[flags][timing][PHI1] = defaults
		//		oc.Presets[flags][timing][PHI2] = defaults
		//	}
		//}
	}
	return operationCodes
}

func (op *OperationCodes) ReadInstructions() (result bool) {
	f, err := os.Open(opCodes)
	if err != nil {
		op.log.Errorf("Failed to open opCodes file: %v", err)
		return false
	}
	defer func() {
		if err := f.Close(); err != nil {
			op.log.Errorf("Trouble closing opCodes file: %v", err)
			result = false
		}
	}()

	fi, err := f.Stat()
	if err != nil {
		op.log.Errorf("Failed to retrieve file info: %v", err)
		return false
	}

	bs := make([]byte, fi.Size())
	n, err := f.Read(bs)
	if err != nil {
		op.log.Errorf("Failed to retrieve file info: %v", err)
		return false
	} else if n != int(fi.Size()) {
		op.log.Errorf("Expected %d bytes, read %d bytes", fi.Size(), n)
		return false
	}

	err = json.Unmarshal(bs, &op.opCodes)
	if err != nil {
		op.log.Errorf("Failed to unmarshal opCodes: %v", err)
		return false
	}

	op.lookup = map[uint8]*OpCode{}
	for _, instruction := range op.opCodes {
		op.lookup[instruction.OpCode] = &instruction
	}

	op.log.Info("OpCodes loaded")
	return true
}
func (op *OperationCodes) WriteInstructions() (result bool) {
	f, err := os.Create(opCodes)
	if err != nil {
		op.log.Errorf("Failed to create instruction file: %v", err)
		return false
	}
	defer func() {
		if err := f.Close(); err != nil {
			op.log.Errorf("Trouble closing opCodes file: %v", err)
			result = false
		}
	}()

	bs, err := json.Marshal(op.opCodes)
	if err != nil {
		op.log.Errorf("Failed to marshal opCodes: %v", err)
		return false
	}
	_, err = f.Write(bs)
	if err != nil {
		op.log.Errorf("Failed to write opCodes: %v", err)
		return false
	}
	op.log.Info("OpCodes saved")
	return true
}
func (op *OperationCodes) Lookup(opcode uint8) *OpCode {
	return op.lookup[opcode]
}

func defineOpCodes() map[uint8]*OpCode {
	ocs := map[uint8]*OpCode {
		// Program Counter
		// When the 6502 is ready for the next instruction it increments the program counter before fetching the
		// instruction. Once it has the op code, it increments the program counter by the length of the operand, if
		// any. This must be accounted for when calculating branches or when pushing bytes to create a false return
		// address (i.e. jump table addresses are made up of addresses-1 when it is intended to use an RTS rather than
		// a JMP).
		//
		// The program counter is loaded least signifigant byte first. Therefore the most signifigant byte must be
		// pushed first when creating a false return address.
		//
		// When calculating branches a forward branch of 6 skips the following 6 bytes so, effectively the program
		// counter points to the address that is 8 bytes beyond the address of the branch opcode; and a backward
		// branch of $FA (256-6) goes to an address 4 bytes before the branch instruction.

		// Execution Times
		// Op code execution times are measured in machine cycles; one machine cycle equals one clock cycle. Many
		// instructions require one extra cycle for execution if a page boundary is crossed


		// ADC (ADd with Carry)
		// Affects Flags: N V Z C
		// ADC results are dependant on the setting of the decimal flag. In decimal mode, addition is carried out on the assumption that the values involved are packed BCD (Binary Coded Decimal).
		// There is no way to add without carry.
		0x69 : mop(IMM, "ADC", "#$44",    0x69, 2, 2, false),
		0x65 : mop(ZPG, "ADC", "$44",     0x65, 2, 3, false),
		0x75 : mop(ZPX, "ADC", "$44,X",   0x75, 2, 4, false),
		0x6D : mop(ABS, "ADC", "$4400",   0x6D, 3, 4, false),
		0x7D : mop(ABX, "ADC", "$4400,X", 0x7D, 3, 4, true),
		0x79 : mop(ABY, "ADC", "$4400,Y", 0x79, 3, 4, true),
		0x61 : mop(IZX, "ADC", "($44,X)", 0x61, 2, 6, false),
		0x71 : mop(IZY, "ADC", "($44),Y", 0x71, 2, 5, true),


		// AND (bitwise AND with accumulator)
		// Affects Flags: N Z
		// + add 1 cycle if page boundary crossed
		0x29 : mop(IMM, "AND", "#$44",    0x29, 2, 2, false),
		0x25 : mop(ZPG, "AND", "$44",     0x25, 2, 3, false),
		0x35 : mop(ZPX, "AND", "$44,X",   0x35, 2, 4, false),
		0x2D : mop(ABS, "AND", "$4400",   0x2D, 3, 4, false),
		0x3D : mop(ABX, "AND", "$4400,X", 0x3D, 3, 4, true),
		0x39 : mop(ABY, "AND", "$4400,Y", 0x39, 3, 4, true),
		0x21 : mop(IZX, "AND", "($44,X)", 0x21, 2, 6, false),
		0x31 : mop(IZY, "AND", "($44),Y", 0x31, 2, 5, true),


		// ASL (Arithmetic Shift Left)
		// Affects Flags: N Z C
		// ASL shifts all bits left one position. 0 is shifted into bit 0 and the original bit 7 is shifted into the Carry.
		0x0A : mop(ACC, "ASL", "A",       0x0A,  1,   2, false),
		0x06 : mop(ZPG, "ASL", "$44",     0x06,  2,   5, false),
		0x16 : mop(ZPX, "ASL", "$44,X",   0x16,  2,   6, false),
		0x0E : mop(ABS, "ASL", "$4400",   0x0E,  3,   6, false),
		0x1E : mop(ABX, "ASL", "$4400,X", 0x1E,  3,   7, false),


		// BIT (test BITs)
		// Affects Flags: N V Z
		// BIT sets the Z flag as though the value in the address tested were ANDed with the accumulator. The N and V flags are set to match bits 7 and 6 respectively in the value stored at the tested address.
		// BIT is often used to skip one or two following bytes as in:
		//
		// CLOSE1 LDX #$10   If entered here, we
		// .BYTE $2C  effectively perform
		// CLOSE2 LDX #$20   a BIT test on $20A2,
		// .BYTE $2C  another one on $30A2,
		// CLOSE3 LDX #$30   and end up with the X
		// CLOSEX LDA #12    register still at $10
		// STA ICCOM,X upon arrival here.
		//
		// Beware: a BIT instruction used in this way as a NOP does have effects: the flags may be modified, and the read of the absolute address, if it happens to access an I/O device, may cause an unwanted action.
		0x24 : mop(ZPG, "BIT", "$44",   0x24, 2, 3, false),
		0x2C : mop(ABS, "BIT", "$4400", 0x2C, 3, 4, false),


		// Branch Instructions
		// Affect Flags: none
		//
		// All branches are relative mode and have a length of two bytes. Syntax is "Bxx Displacement" or (better) "Bxx Label". See the notes on the Program Counter for more on displacements.
		// Branches are dependant on the status of the flag bits when the op code is encountered. A branch not taken requires two machine cycles. Add one if the branch is taken and add one more if the branch crosses a page boundary.
		//
		// // There is no BRA (BRanch Always) instruction but it can be easily emulated by branching on the basis of a known condition. One of the best flags to use for this purpose is the oVerflow which is unchanged by all but addition and subtraction operations.
		// A page boundary crossing occurs when the branch destination is on a different page than the instruction AFTER the branch instruction. For example:
		//
		// SEC
		// BCS LABEL
		// NOP
		// A page boundary crossing occurs (i.e. the BCS takes 4 cycles) when (the address of) LABEL and the NOP are on different pages. This means that
		// CLV
		// BVC LABEL
		// LABEL NOP
		// the BVC instruction will take 3 cycles no matter what address it is located at.
		0x10 : brc("BPL", 0x10, 8, false), // Branch on PLus
		0x30 : brc("BMI", 0x30, 8, true),  // Branch on MInus
		0x50 : brc("BVC", 0x50, 7, false), // Branch on oVerflow Clear
		0x70 : brc("BVS", 0x70, 7, true),  // Branch on oVerflow Set
		0x90 : brc("BCC", 0x90, 1, false), // Branch on Carry Clear
		0xB0 : brc("BCS", 0xB0, 1, true),  // Branch on Carry Set
		0xD0 : brc("BNE", 0xD0, 2, false), // Branch on Not Equal
		0xF0 : brc("BEQ", 0xF0, 2, true),  // Branch on EQual


		// BRK (BReaK)
		// Affects Flags: B
		// BRK causes a non-maskable interrupt and increments the program counter by one. Therefore an RTI will go to the address of the BRK +2 so that BRK may be used to replace a two-byte instruction for debugging and the subsequent RTI will be correct.
		0x00 : brk(IMP, "BRK", "", 0x00, 1, 7, false),
		0x02 : brk(IMP, "RST", "", 0x02, 1, 7, false), // Pseudo instruction
		0x12 : brk(IMP, "NMI", "", 0x12, 1, 7, false), // Pseudo instruction
		0x22 : brk(IMP, "IRQ", "", 0x22, 1, 7, false), // Pseudo instruction


		// CMP (CoMPare accumulator)
		// Affects Flags: N Z C
		// + add 1 cycle if page boundary crossed
		// Compare sets flags as if a subtraction had been carried out. If the value in the accumulator is equal or
		// greater than the compared value, the Carry will be set. The equal (Z) and negative (N) flags will be set
		// based on equality or lack thereof and the sign (i.e. A>=$80) of the accumulator.
		0xC9 : mop(IMM, "CMP", "#$44",    0xC9, 2, 2, false),
		0xC5 : mop(ZPG, "CMP", "$44",     0xC5, 2, 3, false),
		0xD5 : mop(ZPX, "CMP", "$44,X",   0xD5, 2, 4, false),
		0xCD : mop(ABS, "CMP", "$4400",   0xCD, 3, 4, false),
		0xDD : mop(ABX, "CMP", "$4400,X", 0xDD, 3, 4, true),
		0xD9 : mop(ABY, "CMP", "$4400,Y", 0xD9, 3, 4, true),
		0xC1 : mop(IZX, "CMP", "($44,X)", 0xC1, 2, 6, false),
		0xD1 : mop(IZY, "CMP", "($44),Y", 0xD1, 2, 5, true),


		// CPX (ComPare X register)
		// Affects Flags: N Z C
		// Operation and flag results are identical to equivalent mode accumulator CMP ops.
		0xE0 : mop(IMM, "CPX", "#$44",  0xE0, 2, 2, false),
		0xE4 : mop(ZPG, "CPX", "$44",   0xE4, 2, 3, false),
		0xEC : mop(ABS, "CPX", "$4400", 0xEC, 3, 4, false),


		// CPY (ComPare Y register)
		// Affects Flags: N Z C
		// Operation and flag results are identical to equivalent mode accumulator CMP ops.
		0xC0 : mop(IMM, "CPY", "#$44",  0xC0, 2, 2, false),
		0xC4 : mop(ZPG, "CPY", "$44",   0xC4, 2, 3, false),
		0xCC : mop(ABS, "CPY", "$4400", 0xCC, 3, 4, false),


		// DEC (DECrement memory)
		// Affects Flags: N Z
		0xC6 : mop(ZPG, "DEC", "$44",     0xC6, 2, 5, false),
		0xD6 : mop(ZPX, "DEC", "$44,X",   0xD6, 2, 6, false),
		0xCE : mop(ABS, "DEC", "$4400",   0xCE, 3, 6, false),
		0xDE : mop(ABX, "DEC", "$4400,X", 0xDE, 3, 7, false),

		// EOR (bitwise Exclusive OR)
		// Affects Flags: N Z
		// add 1 cycle if page boundary crossed
		0x49 : mop(IMM, "EOR", "#$44",    0x49, 2, 2, false),
		0x45 : mop(ZPG, "EOR", "$44",     0x45, 2, 3, false),
		0x55 : mop(ZPX, "EOR", "$44,X",   0x55, 2, 4, false),
		0x4D : mop(ABS, "EOR", "$4400",   0x4D, 3, 4, false),
		0x5D : mop(ABX, "EOR", "$4400,X", 0x5D, 3, 4, true),
		0x59 : mop(ABY, "EOR", "$4400,Y", 0x59, 3, 4, true),
		0x41 : mop(IZX, "EOR", "($44,X)", 0x41, 2, 6, false),
		0x51 : mop(IZY, "EOR", "($44),Y", 0x51, 2, 5, true),


		// Flag (Processor Status) Instructions
		// Affect Flags: as noted
		// These instructions are implied mode, have a length of one byte and require two machine cycles.
		// Notes:
		// The Interrupt flag is used to prevent (SEI) or enable (CLI) maskable interrupts (aka IRQ's). It does not
		// signal the presence or absence of an interrupt condition. The 6502 will set this flag automatically in
		// response to an interrupt and restore it to its prior status on completion of the interrupt service routine.
		// If you want your interrupt service routine to permit other maskable interrupts, you must clear the I flag
		// in your code.
		//
		// The Decimal flag controls how the 6502 adds and subtracts. If set, arithmetic is carried out in packed
		// binary coded decimal. This flag is unchanged by interrupts and is unknown on power-up. The implication is
		// that a CLD should be included in boot or interrupt coding.
		//
		// The Overflow flag is generally misunderstood and therefore under-utilised. After an ADC or SBC instruction,
		// the overflow flag will be set if the twos complement result is less than -128 or greater than +127, and it
		// will cleared otherwise. In twos complement, $80 through $FF represents -128 through -1, and $00 through $7F
		// represents 0 through +127. Thus, after:
		// CLC
		// LDA #$7F ;   +127
		// ADC #$01 ; +   +1
		// the overflow flag is 1 (+127 + +1 = +128), and after:
		// CLC
		// LDA #$81 ;   -127
		// ADC #$FF ; +   -1
		// the overflow flag is 0 (-127 + -1 = -128). The overflow flag is not affected by increments, decrements,
		// shifts and logical operations i.e. only ADC, BIT, CLV, PLP, RTI and SBC affect it. There is no op code to
		// set the overflow but a BIT test on an RTS instruction will do the trick.
		0x18 : ups("CLC", 0x18, 1, false), // CLear Carry
		0xD8 : ups("CLD", 0xD8, 4, false), // CLear Decimal
		0x58 : ups("CLI", 0x58, 3, false), // CLear Interrupt
		0xB8 : ups("CLV", 0xB8, 7, false), // CLear oVerflow
		0x38 : ups("SEC", 0x38, 1, true), // SEt Carry
		0xF8 : ups("SED", 0xF8, 4, true), // SEt Decimal
		0x78 : ups("SEI", 0x78, 3, true), // SEt Interrupt


		// INC (Increment memory)
		// Affects Flags: N Z
		0xE6 : mop(ZPG, "INC", "$44",     0xE6, 2, 5, false),
		0xF6 : mop(ZPX, "INC", "$44,X",   0xF6, 2, 6, false),
		0xEE : mop(ABS, "INC", "$4400",   0xEE, 3, 6, false),
		0xFE : mop(ABX, "INC", "$4400,X", 0xFE, 3, 7, false),


		// JMP (JuMP)
		// Affects Flags: none
		//
		// JMP transfers program execution to the following address (absolute) or to the location contained in the
		// following address (indirect). Note that there is no carry associated with the indirect jump so:
		// AN INDIRECT JUMP MUST NEVER USE A VECTOR BEGINNING ON THE LAST BYTE OF A PAGE
		// For example if address $3000 contains $40, $30FF contains $80, and $3100 contains $50, the result of JMP
		// ($30FF) will be a transfer of control to $4080 rather than $5080 as you intended i.e. the 6502 took the low
		// byte of the address from $30FF and the high byte from $3000.
		0x4C : mop(ABS, "JMP", "$5597", 0x4C, 3, 3, false),
		0x6C : mop(IND, "JMP", "($5597)", 0x6C, 3, 5, false),


		// JSR (Jump to SubRoutine)
		// Affects Flags: none
		// JSR pushes the address-1 of the next operation on to the stack before transferring program control to the
		// following address. Subroutines are normally terminated by a RTS op code.
		0x20 : jsr(mop(ABS, "JSR", "$5597", 0x20, 3, 6, false)),


		// LDA (Load Accumulator)
		// Affects Flags: N Z
		// + add 1 cycle if page boundary crossed
		0xA9 : lda(mop(IMM, "LDA", "#$44",    0xA9, 2, 2, false)),
		0xA5 : lda(mop(ZPG, "LDA", "$44",     0xA5, 2, 3, false)),
		0xB5 : lda(mop(ZPX, "LDA", "$44,X",   0xB5, 2, 4, false)),
		0xAD : lda(mop(ABS, "LDA", "$4400",   0xAD, 3, 4, false)),
		0xBD : lda(mop(ABX, "LDA", "$4400,X", 0xBD, 3, 4, true)),
		0xB9 : lda(mop(ABY, "LDA", "$4400,Y", 0xB9, 3, 4, true)),
		0xA1 : lda(mop(IZX, "LDA", "($44,X)", 0xA1, 2, 6, false)),
		0xB1 : lda(mop(IZY, "LDA", "($44),Y", 0xB1, 2, 5, true)),


		// LDX (LoaD X register)
		// Affects Flags: N Z
		// + add 1 cycle if page boundary crossed
		0xA2 : mop(IMM, "LDX", "#$44",    0xA2, 2, 2, false),
		0xA6 : mop(ZPG, "LDX", "$44",     0xA6, 2, 3, false),
		0xB6 : mop(ZPY, "LDX", "$44,Y",   0xB6, 2, 4, false),
		0xAE : mop(ABS, "LDX", "$4400",   0xAE, 3, 4, false),
		0xBE : mop(ABY, "LDX", "$4400,Y", 0xBE, 3, 4, true),


		// LDY (LoaD Y register)
		// Affects Flags: N Z
		// + add 1 cycle if page boundary crossed
		0xA0 : mop(IMM, "LDY", "#$44",    0xA0, 2, 2, false),
		0xA4 : mop(ZPG, "LDY", "$44",     0xA4, 2, 3, false),
		0xB4 : mop(ZPX, "LDY", "$44,X",   0xB4, 2, 4, false),
		0xAC : mop(ABS, "LDY", "$4400",   0xAC, 3, 4, false),
		0xBC : mop(ABX, "LDY", "$4400,X", 0xBC, 3, 4, true),


		// LSR (Logical Shift Right)
		// Affects Flags: N Z C
		// LSR shifts all bits right one position. 0 is shifted into bit 7 and the original bit 0 is shifted into the
		// Carry.
		0x4A : mop(ACC, "LSR", "A",       0x4A, 1, 2, false),
		0x46 : mop(ZPG, "LSR", "$44",     0x46, 2, 5, false),
		0x56 : mop(ZPX, "LSR", "$44,X",   0x56, 2, 6, false),
		0x4E : mop(ABS, "LSR", "$4400",   0x4E, 3, 6, false),
		0x5E : mop(ABX, "LSR", "$4400,X", 0x5E, 3, 7, false),


		// NOP (No OPeration)
		// Affects Flags: none
		// NOP is used to reserve space for future modifications or effectively REM out existing code.
		0xEA : mop(IMP, "NOP", "", 0xEA, 1, 2, false),


		// ORA (bitwise OR with Accumulator)
		// Affects Flags: N Z
		// + add 1 cycle if page boundary crossed
		0x09 : mop(IMM, "ORA", "#$44",    0x09, 2, 2, false),
		0x05 : mop(ZPG, "ORA", "$44",     0x05, 2, 3, false),
		0x15 : mop(ZPX, "ORA", "$44,X",   0x15, 2, 4, false),
		0x0D : mop(ABS, "ORA", "$4400",   0x0D, 3, 4, false),
		0x1D : mop(ABX, "ORA", "$4400,X", 0x1D, 3, 4, true),
		0x19 : mop(ABY, "ORA", "$4400,Y", 0x19, 3, 4, true),
		0x01 : mop(IZX, "ORA", "($44,X)", 0x01, 2, 6, false),
		0x11 : mop(IZY, "ORA", "($44),Y", 0x11, 2, 5, true),


		// Register Instructions
		// Affect Flags: N Z
		// These instructions are implied mode, have a length of one byte and require two machine cycles.
		0xCA : reg("DEX", 0xCA), // Decrement X
		0x88 : reg("DEY", 0x88), // Decrement Y
		0xE8 : reg("INX", 0xE8), // Increment X
		0xC8 : reg("INY", 0xC8), // Increment Y
		0xAA : reg("TAX", 0xAA), // Transfer A to X
		0x8A : reg("TXA", 0x8A), // Transfer X to A
		0xA8 : reg("TAY", 0xA8), // Transfer A to Y
		0x98 : reg("TYA", 0x98), // Transfer Y to A


		// ROL (ROtate Left)
		// Affects Flags: N Z C
		// ROL shifts all bits left one position. The Carry is shifted into bit 0 and the original bit 7 is shifted into the Carry.
		0x2A : mop(ACC, "ROL", "A",       0x2A, 1, 2, false),
		0x26 : mop(ZPG, "ROL", "$44",     0x26, 2, 5, false),
		0x36 : mop(ZPX, "ROL", "$44,X",   0x36, 2, 6, false),
		0x2E : mop(ABS, "ROL", "$4400",   0x2E, 3, 6, false),
		0x3E : mop(ABX, "ROL", "$4400,X", 0x3E, 3, 7, false),


		// ROR (ROtate Right)
		// Affects Flags: N Z C
		// ROR shifts all bits right one position. The Carry is shifted into bit 7 and the original bit 0 is shifted into the Carry.
		0x6A : mop(ACC, "ROR", "A",       0x6A, 1, 2, false),
		0x66 : mop(ZPG, "ROR", "$44",     0x66, 2, 5, false),
		0x76 : mop(ZPX, "ROR", "$44,X",   0x76, 2, 6, false),
		0x6E : mop(ABS, "ROR", "$4400",   0x6E, 3, 6, false),
		0x7E : mop(ABX, "ROR", "$4400,X", 0x7E, 3, 7, false),


		// RTI (ReTurn from Interrupt)
		// Affects Flags: all
		// RTI retrieves the Processor Status Word (flags) and the Program Counter from the stack in that order
		// (interrupts push the PC first and then the PSW).
		// Note that unlike RTS, the return address on the stack is the actual address rather than the address-1.
		0x40 : mop(IMP, "RTI", "", 0x40, 1, 6, false),


		//RTS (ReTurn from Subroutine)
		// Affects Flags: none
		// RTS pulls the top two bytes off the stack (low byte first) and transfers program control to that address+1.
		// It is used, as expected, to exit a subroutine invoked via JSR which pushed the address-1.
		// RTS is frequently used to implement a jump table where addresses-1 are pushed onto the stack and accessed
		// via RTS eg. to access the second of four routines:
		// LDX #1
		// JSR EXEC
		// JMP SOMEWHERE
		//
		// LOBYTE
		// .BYTE <ROUTINE0-1,<ROUTINE1-1
		// .BYTE <ROUTINE2-1,<ROUTINE3-1
		//
		// HIBYTE
		// .BYTE >ROUTINE0-1,>ROUTINE1-1
		// .BYTE >ROUTINE2-1,>ROUTINE3-1
		//
		// EXEC
		// LDA HIBYTE,X
		// PHA
		// LDA LOBYTE,X
		// PHA
		// RTS
		0x60 : mop(IMP, "RTS", "", 0x60, 1, 6, false),


		// SBC (SuBtract with Carry)
		// Affects Flags: N V Z C
		//+ add 1 cycle if page boundary crossed
		//
		// SBC results are dependant on the setting of the decimal flag. In decimal mode, subtraction is carried out on
		// the assumption that the values involved are packed BCD (Binary Coded Decimal).
		//
		// There is no way to subtract without the carry which works as an inverse borrow. i.e, to subtract you set the
		// carry before the operation. If the carry is cleared by the operation, it indicates a borrow occurred.
		0xE9 : mop(IMM, "SBC", "#$44",    0xE9, 2, 2, false),
		0xE5 : mop(ZPG, "SBC", "$44",     0xE5, 2, 3, false),
		0xF5 : mop(ZPX, "SBC", "$44,X",   0xF5, 2, 4, false),
		0xED : mop(ABS, "SBC", "$4400",   0xED, 3, 4, false),
		0xFD : mop(ABX, "SBC", "$4400,X", 0xFD, 3, 4, true),
		0xF9 : mop(ABY, "SBC", "$4400,Y", 0xF9, 3, 4, true),
		0xE1 : mop(IZX, "SBC", "($44,X)", 0xE1, 2, 6, false),
		0xF1 : mop(IZY, "SBC", "($44),Y", 0xF1, 2, 5, true),


		// STA (STore Accumulator)
		// Affects Flags: none
		0x85 : str(mop(ZPG, "STA", "$44",     0x85, 2, 3, false), CL_DBD0 | CL_DBD1 | CL_DBD2),
		0x95 : str(mop(ZPX, "STA", "$44,X",   0x95, 2, 4, false), CL_DBD0 | CL_DBD1 | CL_DBD2),
		0x8D : str(mop(ABS, "STA", "$4400",   0x8D, 3, 4, false), CL_DBD0 | CL_DBD1 | CL_DBD2),
		0x9D : str(mop(ABX, "STA", "$4400,X", 0x9D, 3, 5, false), CL_DBD0 | CL_DBD1 | CL_DBD2),
		0x99 : str(mop(ABY, "STA", "$4400,Y", 0x99, 3, 5, false), CL_DBD0 | CL_DBD1 | CL_DBD2),
		0x81 : str(mop(IZX, "STA", "($44,X)", 0x81, 2, 6, false), CL_DBD0 | CL_DBD1 | CL_DBD2),
		0x91 : str(mop(IZY, "STA", "($44),Y", 0x91, 2, 6, false), CL_DBD0 | CL_DBD1 | CL_DBD2),


		// Stack Instructions
		// These instructions are implied mode, have a length of one byte and require machine cycles as indicated.
		// The "PuLl" operations are known as "POP" on most other microprocessors. With the 6502, the stack is always
		// on page one ($100-$1FF) and works top down.
		0x9A : stk("TXS", 0x9A, 2), // Transfer X to Stack ptr
		0xBA : stk("TSX", 0xBA, 2), // Transfer Stack ptr to X
		0x48 : stk("PHA", 0x48, 3), // PusH Accumulator
		0x68 : stk("PLA", 0x68, 4), // PuLl Accumulator
		0x08 : stk("PHP", 0x08, 3), // PusH Processor status
		0x28 : stk("PLP", 0x28, 4), // PuLl Processor status


		// STX (STore X register)
		// Affects Flags: none
		0x86 : str(mop(ZPG, "STX", "$44",   0x86, 2, 3, false), CL_DBD0 | CL_DBD2 | CL_SBD0 | CL_SBD2),
		0x96 : str(mop(ZPY, "STX", "$44,Y", 0x96, 2, 4, false), CL_DBD0 | CL_DBD2 | CL_SBD0 | CL_SBD2),
		0x8E : str(mop(ABS, "STX", "$4400", 0x8E, 3, 4, false), CL_DBD0 | CL_DBD2 | CL_SBD0 | CL_SBD2),


		// STY (STore Y register)
		// Affects Flags: none
		0x84 : str(mop(ZPG, "STY", "$44",   0x84, 2, 3, false), CL_DBD0 | CL_DBD2 | CL_SBD1 | CL_SBD2),
		0x94 : str(mop(ZPX, "STY", "$44,X", 0x94, 2, 4, false), CL_DBD0 | CL_DBD2 | CL_SBD1 | CL_SBD2),
		0x8C : str(mop(ABS, "STY", "$4400", 0x8C, 3, 4, false), CL_DBD0 | CL_DBD2 | CL_SBD1 | CL_SBD2),
	}

	for i := 0; i < 256; i++ {
		oc := uint8(i)
		if ocs[oc] == nil {
			ocs[oc] = mop(IMP, "x" + display.HexData(oc), "", oc, 1, 1, false)
		}
	}

	return ocs
}
func mop(addrMode uint8, name string, syntax string, opcode uint8, length uint8, timing uint8, pageCross bool) *OpCode {
	oc := new(OpCode)
	oc.AddrMode  = addrMode
	oc.Name      = name
	oc.Syntax    = fmt.Sprintf("%s %s", name, syntax)
	oc.OpCode    = opcode
	oc.Operands  = length - 1
	oc.Steps     = timing
	oc.PageCross = pageCross
	oc.Virtual   = false
	oc.BranchBit = 0
	oc.BranchSet = false
	setDefaultLines(oc)
	return oc
}
func brk(addrMode uint8, name string, syntax string, opcode uint8, length uint8, timing uint8, pageCross bool) *OpCode {
	oc := new(OpCode)
	oc.AddrMode  = addrMode
	oc.Name      = name
	oc.Syntax    = fmt.Sprintf("%s %s", name, syntax)
	oc.OpCode    = opcode
	oc.Operands  = length - 1
	oc.Steps     = timing
	oc.PageCross = pageCross
	oc.Virtual   = opcode != 0
	oc.BranchBit = 0
	oc.BranchSet = false
	setDefaultLines(oc)

	for flags := uint8(0); flags < 16; flags++ {
		oc.Lines[flags][0][PHI1] ^= 0
		oc.Lines[flags][0][PHI2] ^= CL_PCIN | CL_FSIB | CL_FMAN
		oc.Lines[flags][1][PHI1] ^= CL_AHC1 | CL_DBD1 | CL_ALD2 | CL_ALLD | CL_AHLD | CL_AULB | CL_AULA | CL_AUSB
		oc.Lines[flags][1][PHI2] ^= CL_DBD1 | CL_FSIB
		oc.Lines[flags][2][PHI1] ^= CL_DBD0 | CL_DBD1 | CL_ALD0 | CL_ALD1 | CL_ALLD | CL_AULB | CL_AUSB
		oc.Lines[flags][2][PHI2] ^= CL_DBD0 | CL_DBD1
		oc.Lines[flags][3][PHI1] ^= CL_DBD2 | CL_ALD0 | CL_ALD1 | CL_ALLD | CL_AULB | CL_AUSB
		oc.Lines[flags][3][PHI2] ^= CL_DBD2
		oc.Lines[flags][4][PHI1] ^= CL_ALD0 | CL_ALD2 | CL_ALLD | CL_AHLD | CL_SPLD | CL_AULB | CL_SBD2
		oc.Lines[flags][4][PHI2] ^= CL_ALD0 | CL_ALD1 | CL_ALD2 | CL_PCLL
		oc.Lines[flags][5][PHI1] ^= CL_ALD0 | CL_ALD2 | CL_ALLD
		oc.Lines[flags][5][PHI2] ^= CL_AHD0 | CL_PCLH
		oc.Lines[flags][6][PHI1] ^= CL_AHD0 | CL_AHD1 | CL_ALD1 | CL_ALD2 | CL_ALLD | CL_AHLD
		oc.Lines[flags][6][PHI2] ^= CL_PCIN

		switch opcode {
		case 0x00: // Break
			oc.Lines[flags][1][PHI2] ^= CL_DBRW
			oc.Lines[flags][2][PHI2] ^= CL_DBRW
			oc.Lines[flags][3][PHI2] ^= CL_DBRW
			oc.Lines[flags][4][PHI1] ^= CL_ALC0

		case 0x02: // Reset
			oc.Lines[flags][0][PHI1] ^= CL_CRST
			oc.Lines[flags][0][PHI2] ^= CL_FMAN
			oc.Lines[flags][1][PHI2] ^= CL_FSCB | CL_FSVB
			oc.Lines[flags][4][PHI1] ^= CL_ALC1 | CL_ALC0
			oc.Lines[flags][5][PHI1] ^= CL_ALC1

		case 0x12: // NMI
			oc.Lines[flags][0][PHI1] ^= CL_CRST
			oc.Lines[flags][0][PHI2] ^= CL_FMAN
			oc.Lines[flags][1][PHI2] ^= CL_DBRW
			oc.Lines[flags][2][PHI2] ^= CL_DBRW
			oc.Lines[flags][3][PHI2] ^= CL_DBRW
			oc.Lines[flags][4][PHI1] ^= CL_ALC2 | CL_ALC0
			oc.Lines[flags][5][PHI1] ^= CL_ALC2

		case 0x22: // IRQ
			oc.Lines[flags][1][PHI2] ^= CL_DBRW
			oc.Lines[flags][2][PHI2] ^= CL_DBRW
			oc.Lines[flags][3][PHI2] ^= CL_DBRW
			oc.Lines[flags][4][PHI1] ^= CL_ALC0
		}
	}
	return oc
}
func brc(name string, opcode uint8, bit uint8, value bool) *OpCode {
	// Branch Instructions
	oc := new(OpCode)
	oc.AddrMode  = REL
	oc.Name      = name
	oc.Syntax    = fmt.Sprintf("%s Label (Displayment: -128 to +127)", name)
	oc.OpCode    = opcode
	oc.Operands  = 1
	oc.Steps     = 2
	oc.PageCross = true
	oc.Virtual   = false
	oc.BranchBit = bit
	oc.BranchSet = value
	setDefaultLines(oc)

	for flags := 0; flags < 16; flags++ {
		oc.Lines[flags][0][PHI1] ^= 0
		oc.Lines[flags][0][PHI1] ^= CL_PCIN

	}
	return oc
}
func ups(name string, opcode uint8, bit uint8, value bool) *OpCode {
	// Flag (Processor Status) Instructions
	oc := new(OpCode)
	oc.AddrMode  = IMP
	oc.Name      = name
	oc.Syntax    = name
	oc.OpCode    = opcode
	oc.Operands  = 0
	oc.Steps     = 2
	oc.PageCross = false
	oc.Virtual   = false
	oc.BranchBit = bit
	oc.BranchSet = value
	return oc
}
func reg(name string, opcode uint8) *OpCode {
	// Register Instructions
	oc := new(OpCode)
	oc.AddrMode  = IMP
	oc.Name      = name
	oc.Syntax    = name
	oc.OpCode    = opcode
	oc.Operands  = 0
	oc.Steps     = 2
	oc.PageCross = false
	oc.Virtual   = false
	oc.BranchBit = 0
	oc.BranchSet = false
	return oc
}
func stk(name string, opcode uint8, timing uint8) *OpCode {
	// Stack Instructions
	oc := new(OpCode)
	oc.AddrMode  = IMP
	oc.Name      = name
	oc.Syntax    = name
	oc.OpCode    = opcode
	oc.Operands  = 0
	oc.Steps     = timing
	oc.PageCross = false
	oc.Virtual   = false
	oc.BranchBit = 0
	oc.BranchSet = false
	return oc
}

func lda(oc *OpCode) *OpCode {
	for flags := uint8(0); flags < 16; flags++ {
		switch oc.AddrMode {
		case IMM:
			oc.Lines[flags][0][PHI1] ^= CL_AHD0 | CL_AHD1 | CL_ALD1 | CL_ALD2 | CL_AHLD | CL_ALLD
			oc.Lines[flags][0][PHI2] ^= CL_PCIN
			oc.Lines[flags][1][PHI1] ^= CL_AULA | CL_AULB | CL_SBD1 | CL_SBLA
			oc.Lines[flags][1][PHI2] ^= 0

		case ZPG:
		case ZPX:
		case ABS:
		case ABX:
		case ABY:
		case IZX:
		case IZY:

			//0xA9 : lda(mop(IMM, "LDA", "#$44",    0xA9, 2, 2, false)),
			//0xA5 : lda(mop(ZPG, "LDA", "$44",     0xA5, 2, 3, false)),
			//0xB5 : lda(mop(ZPX, "LDA", "$44,X",   0xB5, 2, 4, false)),
			//0xAD : lda(mop(ABS, "LDA", "$4400",   0xAD, 3, 4, false)),
			//0xBD : lda(mop(ABX, "LDA", "$4400,X", 0xBD, 3, 4, true)),
			//0xB9 : lda(mop(ABY, "LDA", "$4400,Y", 0xB9, 3, 4, true)),
			//0xA1 : lda(mop(IZX, "LDA", "($44,X)", 0xA1, 2, 6, false)),
			//0xB1 : lda(mop(IZY, "LDA", "($44),Y", 0xB1, 2, 5, true)),
		}
		loadNextInstruction(oc, flags)
	}
	return oc
}
func str(oc *OpCode, dbSource uint64) *OpCode {
	for flags := uint8(0); flags < 16; flags++ {
		switch oc.AddrMode {
		case IMM:
		case ZPG:
		case ZPX:
		case ABS:
			oc.Lines[flags][0][PHI1] ^= CL_AHD0 | CL_AHD1 | CL_ALD1 | CL_ALD2 | CL_AHLD | CL_ALLD
			oc.Lines[flags][0][PHI2] ^= CL_PCIN
			oc.Lines[flags][1][PHI1] ^= CL_AHD0 | CL_AHD1 | CL_ALD1 | CL_ALD2 | CL_ALLD | CL_AHLD | CL_AULA | CL_AULB | CL_AUSA
			oc.Lines[flags][1][PHI2] ^= CL_PCIN
			oc.Lines[flags][2][PHI1] ^= CL_AHD0 | CL_ALD0 | CL_ALD1 | CL_DBRW | CL_ALLD | CL_AHLD
			oc.Lines[flags][2][PHI2] ^= dbSource | CL_DBRW
			oc.Lines[flags][3][PHI1] ^= 0
			oc.Lines[flags][3][PHI2] ^= 0
		case ABX:
		case ABY:
		case IZX:
		case IZY:

			//0xA9 : lda(mop(IMM, "LDA", "#$44",    0xA9, 2, 2, false)),
			//0xA5 : lda(mop(ZPG, "LDA", "$44",     0xA5, 2, 3, false)),
			//0xB5 : lda(mop(ZPX, "LDA", "$44,X",   0xB5, 2, 4, false)),
			//0xAD : lda(mop(ABS, "LDA", "$4400",   0xAD, 3, 4, false)),
			//0xBD : lda(mop(ABX, "LDA", "$4400,X", 0xBD, 3, 4, true)),
			//0xB9 : lda(mop(ABY, "LDA", "$4400,Y", 0xB9, 3, 4, true)),
			//0xA1 : lda(mop(IZX, "LDA", "($44,X)", 0xA1, 2, 6, false)),
			//0xB1 : lda(mop(IZY, "LDA", "($44),Y", 0xB1, 2, 5, true)),
		}
		loadNextInstruction(oc, flags)
	}
	return oc
}
func jsr(oc *OpCode) *OpCode {
	for flags := uint8(0); flags < 16; flags++ {
		oc.Lines[flags][0][PHI1] ^= CL_AHD0 | CL_AHD1 | CL_ALD1 | CL_ALD2 | CL_ALLD | CL_AHLD | CL_AULA
		oc.Lines[flags][0][PHI2] ^= CL_PCIN
		oc.Lines[flags][1][PHI1] ^= CL_AHC1 | CL_DBD1 | CL_ALD2 | CL_ALLD | CL_AHLD | CL_SPLD | CL_AULB | CL_AUSB | CL_SBD1
		oc.Lines[flags][1][PHI2] ^= CL_DBD1 | CL_DBRW
		oc.Lines[flags][2][PHI1] ^= CL_DBD0 | CL_DBD1 | CL_ALD0 | CL_ALD1 | CL_ALLD | CL_AULB | CL_AUSB
		oc.Lines[flags][2][PHI2] ^= CL_DBD0 | CL_DBD1 | CL_DBRW
		oc.Lines[flags][3][PHI1] ^= CL_AHD0 | CL_AHD1 | CL_ALD1 | CL_ALD2 | CL_ALLD | CL_AHLD
		oc.Lines[flags][3][PHI2] ^= CL_AHD0 | CL_PCLH
		oc.Lines[flags][4][PHI1] ^= CL_AHD0 | CL_ALD2 | CL_ALLD | CL_AHLD | CL_SPLD | CL_SBD2
		oc.Lines[flags][4][PHI2] ^= CL_ALD1 | CL_PCLL
		oc.Lines[flags][5][PHI1] ^= 0
		oc.Lines[flags][5][PHI2] ^= 0
	}
	return oc
}

func setDefaultLines(oc *OpCode) {
	for flags := 0; flags < 16; flags++ {
		for timing := uint8(0); timing < 8; timing++ {
			oc.Lines[flags][timing][PHI1] = Defaults[PHI1]
			oc.Lines[flags][timing][PHI2] = Defaults[PHI2]
			if  timing >= oc.Steps - 1 {
				// Add a clock reset to every PHI2 step on or after the last instruction
				oc.Lines[flags][timing][PHI2] ^= CL_CTMR
			}

		}
	}
}
func loadNextInstruction(oc *OpCode, flags uint8 ) {
	oc.Lines[flags][oc.Steps - 1][PHI1] ^= CL_AHD0 | CL_AHD1 | CL_ALD1 | CL_ALD2 | CL_AHLD | CL_ALLD
	oc.Lines[flags][oc.Steps - 1][PHI2] ^= CL_PCIN
}

type OpCode struct {
	Name      string           `json:"name"`
	OpCode    uint8            `json:"opCode"`
	Syntax    string           `json:"syntax"`
	AddrMode  uint8            `json:"addrMode"`
	Operands  uint8            `json:"operands"`
	Steps     uint8            `json:"steps"`
	PageCross bool             `json:"pageCross"`
	BranchBit uint8            `json:"branchBit"`
	BranchSet bool             `json:"branchSet"`
	Virtual   bool             `json:"Virtual"`
	Lines     [16][8][2]uint64 `json:"lines,omitempty"`
	Presets   [16][8][2]uint64 `json:"presets,omitempty"`
	// Flags, Timing, Clock 1/0
}
func (op *OpCode) Block(flags uint8, step uint8, clock uint8, editStep uint8, editPhase uint8) ([]string, []string, [4]string, [4]string) {
	var lines         []string
	var lines2        []string
	var outputs       [4]string
	var aluOperations [4]string

	if op != nil {
		for i := uint8(0); i < op.Steps; i++ {
			colour := lineColor
			for j := uint8(0); j < 2; j++ {
				chevron := " "
				timing := "  "
				if j == 0 {
					timing = fmt.Sprintf("%sT%d", timingColor, i + 2)
				} else if i == op.Steps - 1 {
					timing = fmt.Sprintf("%sT1", timingColor)
				}

				if i == step {
					colour = activeLine
					if j == clock {
						chevron = ">"
						colour = activeClock
					}
				}
				str := op.uint64ToBinary(op.Lines[flags][i][j], op.Presets[flags][i][j], Defaults[j], colour, j)
				line := fmt.Sprintf("%s%s Φ%d%s %s %s%s%s", timing, clockColour, j+1, timeMarker, chevron, colour, str, common.Reset)
				lines = append(lines, line)
			}
		}
		lines2 = op.ActiveLines(flags, editStep, editPhase, 8, " ", "")

		outputs[0] = OutputsDB [op.Lines[flags][editStep][editPhase] & (CL_DBD0|CL_DBD1|CL_DBD2)]
		outputs[1] = OutputsADL[op.Lines[flags][editStep][editPhase] & (CL_ALD0|CL_ALD1|CL_ALD2)]
		outputs[2] = OutputsADH[op.Lines[flags][editStep][editPhase] & (CL_AHD0|CL_AHD1)]
		outputs[3] = OutputsSB [op.Lines[flags][editStep][editPhase] & (CL_SBD0|CL_SBD1|CL_SBD2)]

		aluOperations[0] = AluA  [op.Lines[flags][editStep][editPhase] & (CL_AUSA)]
		aluOperations[1] = AluB  [op.Lines[flags][editStep][editPhase] & (CL_AUSB)]
		aluOperations[2] = AluOp [op.Lines[flags][editStep][editPhase] & (CL_AUIB|CL_AUS1|CL_AUS2|CL_AUO1|CL_AUO2)]
		aluOperations[3] = AluDir[op.Lines[flags][editStep][editPhase] & (CL_AUS1|CL_AUS2|CL_AULR)]
	} else {
		lines = append(lines, "-------- -------- -------- -------- -------- --------")
	}

	return lines, lines2, outputs, aluOperations
}
func (op *OpCode) ActiveLines(flags uint8, step uint8, clock uint8, groupSize int, join string, prefix string) []string {
	var collector []string
	var lines []string
	var index = 0
	bit := uint64(140737488355328)
	l := op.Lines[flags][step][clock] ^ Defaults[clock]
	for bit > 0 {
		if l & bit > 0 {
			collector = append(collector, fmt.Sprintf("%s%s", prefix, mnemonics[index][clock]))
		}
		index++
		if index % groupSize == 0 {
			if len(collector) > 0 {
				lines = append(lines, strings.Join(collector, join))
			}
			collector = make([]string, 0)
		}
		bit >>= 1
	}
	if len(collector) > 0 {
		lines = append(lines, strings.Join(collector, join))
	}
	return lines}
func (op *OpCode) uint64ToBinary(qword uint64, presetQword uint64, defaultQword uint64, lineColor string, clock uint8) string {

	str1 := fmt.Sprintf("%s%%s%s", PresetChg, lineColor)
	str2 := fmt.Sprintf("%s%%s%s", defaultChg, lineColor)
	bs := ""
	for i := 0; i < 48; i++ {
		c := qword & 140737488355328         // current
		p := presetQword & 140737488355328   // preset
		d := defaultQword & 140737488355328  // default

		state := "0"
		if c > 0 {
			state = "1"
		}

		if c == p && p != d {
			state = fmt.Sprintf(str1, state)
		} else if c != d && p == d {
			state = fmt.Sprintf(str2, state)
		}
		bs += state
		if (i + 1) % 8 == 0 { bs += " " }
		if (i + 1) % 16 == 0 { bs += " " }
		qword <<= 1
		presetQword <<= 1
		defaultQword <<= 1
	}
	return bs
}
func (op *OpCode) ValidateLine(step uint8, clock uint8, bit uint64 ) (string, bool) {
	// Validation on which bits can be set when.
	switch uint64(1 << bit) {
	case CL_CTMR:
		return "Timer reset cannot be changed", false
	case CL_ALLD:
		if clock != PHI1 {
			return "Address bus low can only be loaded on phi-1", false
		}
	case CL_AHLD:
		if clock != PHI1 {
			return "Address bus high can only be loaded on phi-1", false
		}
	case CL_SPLD:
		if clock != PHI1 {
			return "Stack pointer can only be loaded on phi-1", false
		}
	case CL_PCIN:
		if clock != PHI2 {
			return "Program counter can only be incremented on phi-2", false
		}
	case CL_PCLL:
		if clock != PHI2 {
			return "Program counter low can only be loaded on phi-2", false
		}
	case CL_PCLH:
		if clock != PHI2 {
			return "Program counter high can only be loaded on phi-2", false
		}
	case CL_FSCA, CL_FSCB, CL_FSIA, CL_FSIB, CL_FSVA, CL_FSVB:
		if clock != PHI2 {
			return "Flag updates can only be performed on phase 2", false
		}
	case CL_SBLA:
		if clock != PHI1 {
			return "Accumulator can only be loaded on phi-1", false
		}
	case CL_SBLX:
		if clock != PHI1 {
			return "X register can only be loaded on phi-1", false
		}
	case CL_SBLY:
		if clock != PHI1 {
			return "Y register can only be loaded on phi-1", false
		}
	}
	return "Ok", true
}
