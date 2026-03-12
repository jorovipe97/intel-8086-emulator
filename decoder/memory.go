package decoder

import "errors"

// TODO: How to manage CS and IP register in this type, For now probably I can just ignore it.
type Memory struct {
	// The program data, this holds the program binary instructions.
	Data []byte

	// The absolute address position at wich CPU is reading instructions from.
	// Note the IP register is 16 bits and a 8086 program can access 1mb of data.
	// With 16 bits we only are able to address 2^16 = 65536 different address locations.
	// So we they solved the problem combining code segment and IP registers. Convention is (CS:IP)
	// To produce a 20 bits numbers. Formula is the following:
	// AbsolutePosition = (cs << 4) | ip.
	//
	// Internally we use an int as is faster in modern CPUs, and either way we dont
	// have a int20 type in go.
	//
	// See: https://stackoverflow.com/q/12263720/4086981
	// See: https://qr.ae/pCQvax
	// See: https://qr.ae/pCQvgD
	AbsolutePosition int
}

func NewMemory(data []byte) *Memory {
	return &Memory{
		Data:             data,
		AbsolutePosition: 0,
	}
}

func (m *Memory) HasMoreInstructions() bool {
	return (m.AbsolutePosition + 1) < len(m.Data)
}

func (m *Memory) GetIPRegister() uint16 {
	return uint16(m.AbsolutePosition)
}

func (m *Memory) GetByteAtPosition(position int) (byte, error) {
	if position >= len(m.Data) && position < 0 {
		return 0, errors.New("position is out of instructions range")
	}
	return m.Data[position], nil
}

func (m *Memory) ResetAbsolutePosition() {
	m.AbsolutePosition = 0
}

// Not that increment can be negative and it will be a decrement
func (m *Memory) IncrementPosition(increment int) {
	m.AbsolutePosition += increment
}
