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
}

// NOTE(casey): This is the "Intel-specified" maximum length of an instruction, including prefixes\
const MaxInstructionByteCount = 15
