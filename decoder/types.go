package decoder

type OperationType uint16

const (
	OpNone OperationType = iota
	OpMov
	OpAdd
	OpSub
	OpCmp

	// Jump if Not Zero (Not Equal).
	OpJNZ

	// Jump if Zero (Equal).
	OpJZ
)

type InstructionBitsUsage uint8

const (
	// NOTE(casey): The 0 value, indicating the end of the instruction encoding array
	BitsEnd InstructionBitsUsage = iota
	BitsLiteral
	BitsD
	BitsS
	BitsW
	BitsMod
	BitsReg
	BitsRm
	BitsDisp
	BitsData
	BitsWMakesDataWide
	// Segment register, 00=ES, 01=CS, 10=SS, 11=DS
	BitsSR
	// Signed increment value to instruction pointer (IP). It's a signed number because we can jump back and forth.
	BitsIpInc
	BitsCount
)

func (v InstructionBitsUsage) String() string {
	switch v {
	case BitsEnd:
		return "BitsEnd"
	case BitsLiteral:
		return "BitsLiteral"
	case BitsS:
		return "BitsS"
	case BitsD:
		return "BitsD"
	case BitsW:
		return "BitsW"
	case BitsMod:
		return "BitsMod"
	case BitsReg:
		return "BitsReg"
	case BitsRm:
		return "BitsRm"
	case BitsDisp:
		return "BitsDisp"
	case BitsData:
		return "BitsData"
	case BitsWMakesDataWide:
		return "BitsWMakesDataWide"
	case BitsCount:
		return "BitsCount"
	}

	return "Unknown"
}

type InstructionBits struct {
	// The usage of these bits, ef reg, rm, mod, etc.
	Usage InstructionBitsUsage

	// Number of bits for this part, eg reg field have 3 bits
	BitCount uint8

	// Ammount we need to left shift the original byte (8 bits)
	// so we extract the field, eg 00reg000. Needs a 3 bits shift
	Shift uint8

	// The actual bytes.
	Value uint8
}

type InstructionEncoding struct {
	op OperationType

	// The instruction mnemonic
	mnemonic string

	// Each item represent a part of the entire instruction encoding. Eg:
	// reg field which are 3 bytes
	//
	// This array has 16 items because there is no operation that have more than 16 parts.
	// summing reg, opcode, d, w, etc.
	//
	// NOTE(casey): This is the "Intel-specified" maximum length of an instruction, including prefixes
	// NOTE(jose): We set the array to 16 elements instead of 15 because when the array is initialized all elements
	// get initialized to zero value (0 = BitsEnd), so we know we read the entire instruction.
	//
	// This is a clever design pattern - by making the sentinel value 0, you automatically get termination
	// without explicitly adding end markers to every instruction definition.
	bits [16]InstructionBits

	// This holds all the flags that are affected by this instruction. This is used by the simulator to compute corresponding flags.
	affectedFlags Flag
}

type OperandsUsage struct {
	// The destination operand
	destination Operand

	// The source operand
	source Operand
}

type Instruction struct {
	Op       OperationType
	Mnemonic string
	RawBits  []byte
	Parts    [16]InstructionBits
	// Size in bytes
	Size              int
	Operands          OperandsUsage
	InstructionExtras InstructionExtra
	AffectedFlags     Flag
}

type InstructionExtra int

const (
	InstructionFlagWide InstructionExtra = 1 << iota
)

type RegisterName int

const (
	RegisterNone RegisterName = iota
	RegisterA
	RegisterC
	RegisterD
	RegisterB
	RegisterSP
	RegisterBP
	RegisterSI
	RegisterDI
	RegisterCount
)

func (r RegisterName) String() string {
	switch r {
	case 1:
		return "ax"
	case 2:
		return "cx"
	case 3:
		return "dx"
	case 4:
		return "bx"
	case 5:
		return "sp"
	case 6:
		return "bp"
	case 7:
		return "si"
	case 8:
		return "di"
	}

	return ""
}

type SegmentRegisterName int

const (
	// Extra Segment (00)
	SegmentRegisterES SegmentRegisterName = iota
	// Code Segment (01)
	SegmentRegisterCS
	// Stack Segment (10)
	SegmentRegisterSS
	// Data Segment (11)
	SegmentRegisterDS
)

func (r SegmentRegisterName) String() string {
	switch r {
	case SegmentRegisterES:
		return "es"
	case SegmentRegisterCS:
		return "cs"
	case SegmentRegisterSS:
		return "ss"
	case SegmentRegisterDS:
		return "ds"
	}

	return ""
}

type RegisterInfo struct {
	// Which register (A, B, C, D, SP, BP, SI, DI)
	RegisterName RegisterName

	// Byte offset within the 16-bit register (0 = low byte, 1 = high byte)
	Offset int

	// Number of bytes accessed (1 = 8-bit, 2 = 16-bit)
	Count int
}

type Operand interface {
	operandMarker()
}

type MemoryOperand struct {
	Terms        [2]RegisterInfo
	Displacement int
}

func (MemoryOperand) operandMarker() {}

type RegisterOperand struct {
	// TODO: Remove unnecesary nested Register
	Register RegisterInfo
}

func (RegisterOperand) operandMarker() {}

type ImmediateOperand struct {
	Value int16
}

func (ImmediateOperand) operandMarker() {}

type SegmentRegisterOperand struct {
	SegmentRegister SegmentRegisterName
}

func (SegmentRegisterOperand) operandMarker() {}

// We store both 8-bit increments and 16-bit increments in the same operand type.
type InstructionPointerIncrementOperand struct {
	Increment int
}

func (InstructionPointerIncrementOperand) operandMarker() {}

// TODO: Just create a new type of operand?

type Flag int

const (
	// No flags.
	FlagNone Flag = 0

	// Carry Flag (CF) - this flag is set to 1 when there is an unsigned overflow. For example when you add bytes 255 + 1 (result is not in range 0...255). When there is no overflow this flag is set to 0.
	FlagCF Flag = 1

	// Parity Flag (PF) - this flag is set to 1 when there is even number of one bits in result, and to 0 when there is odd number of one bits. Even if result is a word only 8 low bits are analyzed!
	FlagPF Flag = 3

	// Auxiliary Flag (AF) - set to 1 when there is an unsigned overflow for low nibble (4 bits).
	FlagAF Flag = 5

	// Zero Flag (ZF) - set to 1 when result is zero. For none zero result this flag is set to 0.
	FlagZF Flag = 7

	// Sign Flag (SF) - set to 1 when result is negative. When result is positive it is set to 0. Actually this flag take the value of the most significant bit.
	FlagSF Flag = 8

	// Overflow Flag (OF) - set to 1 when there is a signed overflow. For example, when you add bytes 100 + 50 (result is not in range -128...127).
	FlagOF Flag = 12
)

func (f Flag) String() string {
	switch f {
	case FlagNone:
		return "FlagNone"
	case FlagCF:
		return "FlagCF"
	case FlagPF:
		return "FlagPF"
	case FlagAF:
		return "FlagAF"
	case FlagZF:
		return "FlagZF"
	case FlagSF:
		return "FlagSF"
	case FlagOF:
		return "FlagOF"
	}

	return ""
}
