package decoder

type Opcode byte

const (
	// MOV destination, sourcce
	// Register/memory to/from register
	MovRegisterMemoryToFromRegister Opcode = 0b0010_0010

	// Immediate to register/memory
	MovImmediateToRegisterMemory Opcode = 0b0110_0011

	// Immediate to register.
	MovImmediateToRegister Opcode = 0b1011

	// Memory to accumulator
	MovMemoryToAccumulator Opcode = 0b0101_0000

	// Accumulator to memory
	MovAccumulatorToMemory Opcode = 0b0101_0001

	// Add - Reg/Memory with register to either
	AddRegMemoryWithRegisterToEither Opcode = 0b0000_0000

	// Add - Immediate to register/memory
	AddImmediateToRegisterMemory Opcode = 0b0010_0000

	// Add - Immediate to accumulator
	AddImmediateToAccumulator Opcode = 0b00000010
)
