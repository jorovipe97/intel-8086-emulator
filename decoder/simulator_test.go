package decoder

import (
	"testing"
)

type registerExpectedValue struct {
	Name  RegisterName
	Value uint16
}

type segmentRegisterExpectedValue struct {
	Name  SegmentRegisterName
	Value uint16
}

type flagsExpectedValue struct {
	Flag  Flag
	Value int
}

type simulatorTestCase struct {
	Name                            string
	InputBinaryStream               []byte
	RegistersExpectedValues         []registerExpectedValue
	SegmentRegisterExpectedValues   []segmentRegisterExpectedValue
	InstructionPointerExpectedValue int
	FlagsExpectedValues             []flagsExpectedValue
}

var simulatorCases = []simulatorTestCase{
	{
		Name: "non memory movs",
		// Non-Memory movs
		InputBinaryStream: []byte{
			0b10111000, 0b00100010, 0b00100010, 0b10111011, 0b01000100, 0b01000100, 0b10111001, 0b01100110,
			0b01100110, 0b10111010, 0b10001000, 0b10001000, 0b10001110, 0b11010000, 0b10001110, 0b11011011,
			0b10001110, 0b11000001, 0b10110000, 0b00010001, 0b10110111, 0b00110011, 0b10110001, 0b01010101,
			0b10110110, 0b01110111, 0b10001000, 0b11011100, 0b10001000, 0b11110001, 0b10001110, 0b11010000,
			0b10001110, 0b11011011, 0b10001110, 0b11000001, 0b10001100, 0b11010100, 0b10001100, 0b11011101,
			0b10001100, 0b11000110, 0b10001001, 0b11010111,
		},
		InstructionPointerExpectedValue: 44,
		RegistersExpectedValues: []registerExpectedValue{
			{Name: RegisterA, Value: 0x4411},
			{Name: RegisterB, Value: 0x3344},
			{Name: RegisterC, Value: 0x6677},
			{Name: RegisterD, Value: 0x7788},
			{Name: RegisterSP, Value: 0x4411},
			{Name: RegisterBP, Value: 0x3344},
			{Name: RegisterSI, Value: 0x6677},
			{Name: RegisterDI, Value: 0x7788},
		},
		SegmentRegisterExpectedValues: []segmentRegisterExpectedValue{
			{Name: SegmentRegisterES, Value: 0x6677},
			{Name: SegmentRegisterSS, Value: 0x4411},
			{Name: SegmentRegisterDS, Value: 0x3344},
		},
	},
	{
		Name: "some memory movs",
		InputBinaryStream: []byte{
			0b10000001, 0b11000011, 0b00110000, 0b01110101, 0b10000001, 0b11000011, 0b00010000,
			0b00100111, 0b10000001, 0b11101011, 0b10001000, 0b00010011, 0b10000001, 0b11101011,
			0b10001000, 0b00010011, 0b10111011, 0b00000001, 0b00000000, 0b10111001, 0b01100100,
			0b00000000, 0b00000001, 0b11001011, 0b10111010, 0b00001010, 0b00000000, 0b00101001,
			0b11010001, 0b10000001, 0b11000011, 0b01000000, 0b10011100, 0b10000011, 0b11000001,
			0b10100110, 0b10111100, 0b01100011, 0b00000000, 0b10111101, 0b01100010, 0b00000000,
			0b00111001, 0b11100101,
		},
		InstructionPointerExpectedValue: 44,
		RegistersExpectedValues: []registerExpectedValue{
			{Name: RegisterB, Value: 0x9ca5},
			{Name: RegisterD, Value: 0x000a},
			{Name: RegisterSP, Value: 0x0063},
			{Name: RegisterBP, Value: 0x0062},
		},
		FlagsExpectedValues: []flagsExpectedValue{
			{Flag: FlagCF, Value: 1},
			{Flag: FlagZF, Value: 0},
			{Flag: FlagSF, Value: 1},
			// {flag: FlagAF, expectedValue: 1}, // We dont support this flag.
			{Flag: FlagOF, Value: 0},
			{Flag: FlagPF, Value: 1},
		},
	},
	{
		// bits 16
		// mov cx, 3
		// mov bx, 1000
		// loop_start:
		// add bx, 10
		// sub cx, 1
		// jnz loop_start
		Name: "loop using jnz",
		InputBinaryStream: []byte{
			0b10111001, 0b00000011, 0b00000000, 0b10111011, 0b11101000, 0b00000011, 0b10000011,
			0b11000011, 0b00001010, 0b10000011, 0b11101001, 0b00000001, 0b01110101, 0b11111000,
		},
		RegistersExpectedValues: []registerExpectedValue{
			{Name: RegisterB, Value: 0x0406},
		},
		InstructionPointerExpectedValue: 14,
		FlagsExpectedValues: []flagsExpectedValue{
			{Flag: FlagZF, Value: 1},
			{Flag: FlagPF, Value: 1},
		},
	},
}

func TestSimulator(t *testing.T) {
	// Label to represent the test cases
	for _, testCase := range simulatorCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_memory := NewMemory(testCase.InputBinaryStream)
			_decoder := NewDecoder(_memory)
			_simulator := NewSimulator(_memory)

			for {
				if !_memory.HasMoreInstructions() {
					break
				}

				instruction, err := _decoder.NextInstruction()
				if err != nil {
					t.Errorf("error decoding instruction with: %v", err)
					// Use label to continue outer loop instead of the inner for {} loop.
					return
				}

				err = _simulator.ExecInstruction(instruction)
				if err != nil {
					t.Errorf("error executing extensions with: %v", err)
					return
				}
			}

			// Checks registers have the expected value
			for _, expected := range testCase.RegistersExpectedValues {
				if _simulator.registers[expected.Name] != expected.Value {
					t.Errorf("Register %v expected %04x, got %04x", expected.Name, expected.Value, _simulator.registers[expected.Name])
					return
				}
			}

			if _memory.AbsolutePosition != testCase.InstructionPointerExpectedValue {
				t.Errorf("Expected instruction pointer (%v) do not match (%v)", testCase.InstructionPointerExpectedValue, _memory.AbsolutePosition)
				return
			}

			// Checks segment registers have the expected value
			for _, expected := range testCase.SegmentRegisterExpectedValues {
				if _simulator.segmentRegisters[expected.Name] != expected.Value {
					t.Errorf("Register %v expected %04x, got %04x", expected.Name, expected.Value, _simulator.segmentRegisters[expected.Name])
					return
				}
			}
		})
	}
}

type testSimulatorFlagsExpectedValues struct {
	flag          Flag
	expectedValue int
}

type testCaseSimulatorFlags struct {
	operation                        OperationType
	destinationValue                 uint16
	sourceValue                      uint16
	isByte                           bool
	testSimulatorFlagsExpectedValues []testSimulatorFlagsExpectedValues
}

var testCasesSimulatorFlags = []testCaseSimulatorFlags{
	{
		operation:        OpAdd,
		destinationValue: 0b00000000_01100100, // 100
		sourceValue:      0b11111111_00111000, // -200
		testSimulatorFlagsExpectedValues: []testSimulatorFlagsExpectedValues{
			{flag: FlagCF, expectedValue: 0},
			{flag: FlagZF, expectedValue: 0},
			{flag: FlagSF, expectedValue: 1},
			{flag: FlagOF, expectedValue: 0},
			{flag: FlagPF, expectedValue: 1},
		},
	},
	{
		operation:        OpAdd,
		destinationValue: 61443,
		sourceValue:      3841,
		testSimulatorFlagsExpectedValues: []testSimulatorFlagsExpectedValues{
			{flag: FlagCF, expectedValue: 0},
			{flag: FlagZF, expectedValue: 0},
			{flag: FlagSF, expectedValue: 1},
			{flag: FlagOF, expectedValue: 0},
			{flag: FlagPF, expectedValue: 0},
		},
	},
	{
		operation:        OpSub,
		destinationValue: 61443, // 100
		sourceValue:      61443, // 100
		testSimulatorFlagsExpectedValues: []testSimulatorFlagsExpectedValues{
			{flag: FlagCF, expectedValue: 0},
			{flag: FlagZF, expectedValue: 1},
			{flag: FlagSF, expectedValue: 0},
			{flag: FlagOF, expectedValue: 0},
			{flag: FlagPF, expectedValue: 1},
		},
	},
	{
		operation:        OpSub,
		destinationValue: 10,
		sourceValue:      13,
		testSimulatorFlagsExpectedValues: []testSimulatorFlagsExpectedValues{
			{flag: FlagCF, expectedValue: 1},
			{flag: FlagZF, expectedValue: 0},
			{flag: FlagSF, expectedValue: 1},
			{flag: FlagOF, expectedValue: 0},
			{flag: FlagPF, expectedValue: 0},
		},
		isByte: true,
	},
}

func TestSimulatorFlags(t *testing.T) {
	for _, testCase := range testCasesSimulatorFlags {
		memory := NewMemory([]byte{})
		simulator := NewSimulator(memory)
		simulator.registers[RegisterB] = testCase.destinationValue
		simulator.registers[RegisterC] = testCase.sourceValue

		instruction := Instruction{
			Op: testCase.operation,
			Operands: OperandsUsage{
				destination: RegisterOperand{
					Register: RegisterInfo{
						RegisterName: RegisterB,
						Offset:       0,
						Count:        2,
					},
				},
				source: RegisterOperand{
					Register: RegisterInfo{
						RegisterName: RegisterC,
						Offset:       0,
						Count:        2,
					},
				},
			},
			AffectedFlags:     arithmeticAndLogicFlags,
			InstructionExtras: InstructionFlagWide,
		}

		if testCase.isByte {
			// Adds a garbage value to BH to test edge cases, we use BL
			simulator.registers[RegisterB] = (11 << 8) | testCase.destinationValue

			// Adds garbage to CL to test edge cases, we use CH
			simulator.registers[RegisterC] = (testCase.sourceValue << 8) | 11

			instruction.Operands = OperandsUsage{
				destination: RegisterOperand{
					Register: RegisterInfo{
						RegisterName: RegisterB, // BL
						Offset:       0,
						Count:        1,
					},
				},
				source: RegisterOperand{
					Register: RegisterInfo{
						RegisterName: RegisterC, // CH
						Offset:       1,
						Count:        1,
					},
				},
			}
			instruction.InstructionExtras = 0
		}

		err := simulator.ExecInstruction(instruction)
		if err != nil {
			t.Error(err)
			continue
		}

		// TODO: If is a byte then validates untouched parts of the registry to verify everything i ok
		if testCase.isByte {
			// A garbage value was added to bh. we check that value continues untouched.
			// Adds a garbage value to BH to test edge cases, we use BL
			var garbageValue uint16 = (11 << 8)
			if simulator.registers[RegisterB]&garbageValue != garbageValue {
				t.Error("value in bh was modified even tough it should continue invariant.")
			}

			if simulator.registers[RegisterC]&11 != 11 {
				t.Error("value in cl was modified even tough it should continue invariant.")
			}
		}

		for _, flagCase := range testCase.testSimulatorFlagsExpectedValues {
			if simulator.getFlagValue(flagCase.flag) != flagCase.expectedValue {
				t.Errorf("%v expected %v, but got: %v", flagCase.flag, flagCase.expectedValue, simulator.getFlagValue(flagCase.flag))
				continue
			}
		}
	}
}
