package decoder

var mod = InstructionBits{Usage: BitsMod, BitCount: 2}
var reg = InstructionBits{Usage: BitsReg, BitCount: 3}
var rm = InstructionBits{Usage: BitsRm, BitCount: 3}
var w = InstructionBits{Usage: BitsW, BitCount: 1}
var s = InstructionBits{Usage: BitsS, BitCount: 1}
var d = InstructionBits{Usage: BitsD, BitCount: 1}
var sr = InstructionBits{Usage: BitsSR, BitCount: 2}
var data = InstructionBits{Usage: BitsData}
var dataIfW = InstructionBits{Usage: BitsWMakesDataWide, Value: 1}
var ipInc = InstructionBits{Usage: BitsIpInc}

// The group of ADD, SUB,CMP, AND, TEST, OR, XOR. affect these flags only:
// CF, ZF, SF, OF, PF, AF.
// https://yassinebridi.github.io/asm-docs/
var arithmeticAndLogicFlags = FlagCF | FlagZF | FlagSF | FlagOF | FlagPF | FlagAF

// Some operations dont have some fields, we implied them so decoder works as expected
func impliedReg(value uint8) InstructionBits {
	return InstructionBits{
		Usage:    BitsReg,
		BitCount: 0,
		Shift:    0,
		Value:    value,
	}
}

func impliedMod(value uint8) InstructionBits {
	return InstructionBits{
		Usage:    BitsMod,
		BitCount: 0,
		Shift:    0,
		Value:    value,
	}
}

func impliedRm(value uint8) InstructionBits {
	return InstructionBits{
		Usage:    BitsRm,
		BitCount: 0,
		Shift:    0,
		Value:    value,
	}
}

func impliedD(value uint8) InstructionBits {
	return InstructionBits{
		Usage:    BitsD,
		BitCount: 0,
		Shift:    0,
		Value:    value,
	}
}

func impliedW(value uint8) InstructionBits {
	return InstructionBits{
		Usage:    BitsW,
		BitCount: 0,
		Shift:    0,
		Value:    value,
	}
}

// List all instructions in 8086 processor according to manual page 261
var instructionsTable = [...]InstructionEncoding{
	// Register/memory to/from register
	{
		op:       OpMov,
		mnemonic: "mov",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 6, Value: 0b100010},
			d,
			w,
			mod,
			reg,
			rm,
		},
	},
	// Immediate to register/memory
	{
		op:       OpMov,
		mnemonic: "mov",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 7, Value: 0b1100011},
			w,
			mod,
			{Usage: BitsLiteral, BitCount: 3, Value: 0b000},
			rm,
			data,
			dataIfW,
			// NOTE(jose): Casey's implementation uses ImpD(0), what is this?
		},
	},
	// Immediate to register
	{
		op:       OpMov,
		mnemonic: "mov",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 4, Value: 0b1011},
			w,
			reg,
			data,
			dataIfW,
			// NOTE(jose): Casey's implementation uses ImpD(1), what is this?
		},
	},
	// Memory to accumulator
	{
		op:       OpMov,
		mnemonic: "mov",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 7, Value: 0b1010000},
			w,
			// 000 -> AX when w is 1. Or AL when w is 0.
			impliedReg(0b000),
			// Memory mode, no displacement follows...
			impliedMod(0b00),
			// ...except when R/M = 110. Then 16 bit displacement follows
			impliedRm(0b110),
			// Instruction destination is always in the reg field.
			impliedD(0b1),
		},
	},
	// Accumulator to memory
	{
		op:       OpMov,
		mnemonic: "mov",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 7, Value: 0b1010001},
			w,
			// 000 -> AX when w is 1. Or AL when w is 0.
			impliedReg(0b000),
			// Memory mode, no displacement follows...
			impliedMod(0b00),
			// ...except when R/M = 110. Then 16 bit displacement follows
			impliedRm(0b110),
			// Instruction source is always in the reg field.
			impliedD(0b0),
		},
	},
	// Register/Memory to segment register
	{
		op:       OpMov,
		mnemonic: "mov",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 8, Value: 0b1000_1110},
			mod,
			{Usage: BitsLiteral, BitCount: 1, Value: 0b0},
			sr,
			rm,
			// Segment register field acts as destination
			impliedD(0b1),
			// We assumme always wide, as segment register only support 16 bits data.
			impliedW(0b1),
		},
	},
	// Segment register to Register/Memory
	{
		op:       OpMov,
		mnemonic: "mov",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 8, Value: 0b1000_1100},
			mod,
			{Usage: BitsLiteral, BitCount: 1, Value: 0b0},
			sr,
			rm,
			// Segment register field acts as source
			impliedD(0b0),
			// We assumme always wide, as segment register only support 16 bits data.
			impliedW(0b1),
		},
	},
	// Arithmetic - Add Reg/memory with register to either
	{
		op:       OpAdd,
		mnemonic: "add",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 6, Value: 0b000_000},
			d,
			w,
			mod,
			reg,
			rm,
		},
		affectedFlags: arithmeticAndLogicFlags,
	},
	// Arithmetic - Immediate to register/memory
	{
		op:       OpAdd,
		mnemonic: "add",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 6, Value: 0b100_000},
			s,
			w,
			mod,
			{Usage: BitsLiteral, BitCount: 3, Value: 0b000},
			rm,
			data,
			dataIfW,
		},
		affectedFlags: arithmeticAndLogicFlags,
	},
	// Arithmetic - Immediate to accumulator
	{
		op:       OpAdd,
		mnemonic: "add",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 7, Value: 0b000_001_0},
			w,
			// 000 -> AX when w is 1. Or AL when w is 0.
			impliedReg(0b000),
			data,
			dataIfW,
		},
		affectedFlags: arithmeticAndLogicFlags,
	},
	// Arithmetic - Sub Reg/memory with register to either
	{
		op:       OpSub,
		mnemonic: "sub",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 6, Value: 0b001_010},
			d,
			w,
			mod,
			reg,
			rm,
		},
		affectedFlags: arithmeticAndLogicFlags,
	},
	// Arithmetic - Sub Immediate from register/memory
	{
		op:       OpSub,
		mnemonic: "sub",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 6, Value: 0b100_000},
			s,
			w,
			mod,
			{Usage: BitsLiteral, BitCount: 3, Value: 0b101},
			rm,
			data,
			dataIfW,
		},
		affectedFlags: arithmeticAndLogicFlags,
	},
	// Arithmetic - Sub Immediate to accumulator
	{
		op:       OpSub,
		mnemonic: "sub",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 7, Value: 0b00_101_10},
			w,
			// 000 -> AX when w is 1. Or AL when w is 0.
			impliedReg(0b000),
			data,
			dataIfW,
		},
		affectedFlags: arithmeticAndLogicFlags,
	},
	// Arithmetic - Cmp Reg/memory with register to either
	{
		op:       OpCmp,
		mnemonic: "cmp",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 6, Value: 0b00_111_0},
			d,
			w,
			mod,
			reg,
			rm,
		},
		affectedFlags: arithmeticAndLogicFlags,
	},
	// Arithmetic - Cmp Immediate from register/memory
	{
		op:       OpCmp,
		mnemonic: "cmp",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 6, Value: 0b100_000},
			s,
			w,
			mod,
			{Usage: BitsLiteral, BitCount: 3, Value: 0b111},
			rm,
			data,
			dataIfW,
		},
		affectedFlags: arithmeticAndLogicFlags,
	},
	// Arithmetic - Sub Immediate to accumulator
	{
		op:       OpCmp,
		mnemonic: "cmp",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 7, Value: 0b00_111_10},
			w,
			// 000 -> AX when w is 1. Or AL when w is 0.
			impliedReg(0b000),
			data,
			dataIfW,
		},
		affectedFlags: arithmeticAndLogicFlags,
	},
	// JNE/JNZ = Jump on not equal/not zero
	{
		op:       OpJNZ,
		mnemonic: "jnz",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 8, Value: 0b0111_0101},
			// // 000 -> AX when w is 1. Or AL when w is 0.
			// impliedReg(0b000),
			// impliedW(1),
			// // Instruction destination is always in the reg field.
			// impliedD(0b1),

			// Altough this is x86 reference, the jmp instructions has a single destination operand.
			// See: https://www.felixcloutier.com/x86/jmp
			// For this case, the target operand specifies a relative offset (a signed displacement relative to the current value of the instruction pointer in the IP register).
			// A near jump to a relative offset of 8-bits (rel8) is referred to as a short jump. The CS register is not changed on near and short jumps.
			//
			// The BitsIpInc is to indicate that the destination operand is an Instruction Pointer Increment
			// However the actual data is extracted from data.
			ipInc,
			// 8-bit IP increment.
			impliedW(0),
			data,
		},
	},
	// JNE/JNZ = Jump on not equal/not zero
	{
		op:       OpJZ,
		mnemonic: "jz",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 8, Value: 0b0111_0100},
			ipInc,
			// 8-bit IP increment.
			impliedW(0),
			data,
		},
	},
	{
		op:       OpJP,
		mnemonic: "jp",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 8, Value: 0b0111_1010},
			ipInc,
			// 8-bit IP increment.
			impliedW(0),
			data,
		},
	},
	{
		op:       OpJB,
		mnemonic: "jb",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 8, Value: 0b0111_0010},
			ipInc,
			// 8-bit IP increment.
			impliedW(0),
			data,
		},
	},
	{
		op:       OpLoop,
		mnemonic: "loop",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 8, Value: 0b1110_0010},
			ipInc,
			// 8-bit IP increment.
			impliedW(0),
			data,
		},
	},
	{
		op:       OpLoopZ,
		mnemonic: "loopz",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 8, Value: 0b1110_0001},
			ipInc,
			// 8-bit IP increment.
			impliedW(0),
			data,
		},
	},
	{
		op:       OpLoopNZ,
		mnemonic: "loopnz",
		bits: [16]InstructionBits{
			{Usage: BitsLiteral, BitCount: 8, Value: 0b1110_0000},
			ipInc,
			// 8-bit IP increment.
			impliedW(0),
			data,
		},
	},
}

// NOTE(casey): This is the "Intel-specified" maximum length of an instruction, including prefixes\
const MaxInstructionByteCount = 15
