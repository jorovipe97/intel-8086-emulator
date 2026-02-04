package decoder

import "testing"

type registerExpectedValue struct {
	Name  RegisterName
	Value uint16
}

type segmentRegisterExpectedValue struct {
	Name  SegmentRegisterName
	Value uint16
}

type simulatorTestCase struct {
	InputBinaryStream             []byte
	RegistersExpectedValues       []registerExpectedValue
	SegmentRegisterExpectedValues []segmentRegisterExpectedValue
}

var simulatorCases = []simulatorTestCase{
	{
		// Non-Memory movs
		InputBinaryStream: []byte{
			0b10111000, 0b00100010, 0b00100010, 0b10111011, 0b01000100, 0b01000100, 0b10111001, 0b01100110,
			0b01100110, 0b10111010, 0b10001000, 0b10001000, 0b10001110, 0b11010000, 0b10001110, 0b11011011,
			0b10001110, 0b11000001, 0b10110000, 0b00010001, 0b10110111, 0b00110011, 0b10110001, 0b01010101,
			0b10110110, 0b01110111, 0b10001000, 0b11011100, 0b10001000, 0b11110001, 0b10001110, 0b11010000,
			0b10001110, 0b11011011, 0b10001110, 0b11000001, 0b10001100, 0b11010100, 0b10001100, 0b11011101,
			0b10001100, 0b11000110, 0b10001001, 0b11010111,
		},
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
}

func TestSimulator(t *testing.T) {
	// Label to represent the test cases
nextTextCase:
	for _, testCase := range simulatorCases {
		_decoder := NewDecoder(testCase.InputBinaryStream)
		_simulator := NewSimulator(nil)

		for {
			if !_decoder.HasNext() {
				break
			}

			instruction, err := _decoder.NextInstruction()
			if err != nil {
				t.Errorf("error decoding instruction with: %v", err)
				// Use label to continue outer loop instead of the inner for {} loop.
				continue nextTextCase
			}

			err = _simulator.ExecInstruction(instruction)
			if err != nil {
				t.Errorf("error executing extensions with: %v", err)
				continue nextTextCase
			}
		}

		// Checks registers have the expected value
		for _, expected := range testCase.RegistersExpectedValues {
			if _simulator.registers[expected.Name] != expected.Value {
				t.Errorf("Register %v expected %04x, got %04x", expected.Name, expected.Value, _simulator.registers[expected.Name])
				continue nextTextCase
			}
		}

		// Checks segment registers have the expected value
		for _, expected := range testCase.SegmentRegisterExpectedValues {
			if _simulator.segmentRegisters[expected.Name] != expected.Value {
				t.Errorf("Register %v expected %04x, got %04x", expected.Name, expected.Value, _simulator.segmentRegisters[expected.Name])
				continue nextTextCase
			}
		}
	}
}
