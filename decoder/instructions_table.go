package decoder

var mod = InstructionBits{Usage: BitsMod, BitCount: 2}
var reg = InstructionBits{Usage: BitsReg, BitCount: 3}
var rm = InstructionBits{Usage: BitsRm, BitCount: 3}
var w = InstructionBits{Usage: BitsW, BitCount: 1}
var s = InstructionBits{Usage: BitsS, BitCount: 1}
var d = InstructionBits{Usage: BitsD, BitCount: 1}
var data = InstructionBits{Usage: BitsData}
var dataIfW = InstructionBits{Usage: BitsWMakesDataWide, Value: 1}

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
	},
	// Arithmetic - Immediate to accumulator
	{
		// TODO: This case is not working occrectly
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
	},
}

// NOTE(casey): This is the "Intel-specified" maximum length of an instruction, including prefixes\
const MaxInstructionByteCount = 15
