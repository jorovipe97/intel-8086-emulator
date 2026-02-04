package decoder

import "fmt"

type Simulator struct {
	// First item is unnuced, registers start from index 1 to 8
	registers        [9]uint16
	segmentRegisters [4]uint16
	asmPrinter       *AsmPrinter
}

func NewSimulator(printer *AsmPrinter) *Simulator {
	return &Simulator{
		asmPrinter: printer,
	}
}

func (s *Simulator) getOperandValue(operand Operand) uint16 {
	switch specificOperand := operand.(type) {
	case ImmediateOperand:
		return uint16(specificOperand.Value)
	case RegisterOperand:
		// If is a byte operand, eg: al, bl, cl, dl, ah, bh, ch, dh
		// then we need to write the appropiate part of the register
		if specificOperand.Register.Count == 1 {
			// This is used to remove the mart of the register we are not interested in.
			var mask uint16 = 0b00000000_11111111
			var rightShift uint16 = uint16(specificOperand.Register.Offset) * 8
			return (s.registers[specificOperand.Register.RegisterName] >> rightShift) & mask
		}

		// If reaches here we are in a word operand, eg: ax, bx, cx, dx
		return s.registers[specificOperand.Register.RegisterName]
	case SegmentRegisterOperand:
		return s.segmentRegisters[specificOperand.SegmentRegister&0b11]
	}

	return 0
}

func (s *Simulator) ExecInstruction(instruction Instruction) error {
	fmt.Println("#### Executing Instruction")
	destinationValue := s.getOperandValue(instruction.Operands.destination)
	sourceValue := s.getOperandValue(instruction.Operands.source)
	var finalValue uint16

	// Computes the final value, based on the op type
	switch instruction.Op {
	case OpMov:
		finalValue = sourceValue
	case OpAdd:
		finalValue = destinationValue + sourceValue
	}

	// Updates simulated memory. Destination can be a register or memory.
	switch destinationOperand := instruction.Operands.destination.(type) {
	case RegisterOperand:
		// Identify which part of memory should store result
		register := destinationOperand.Register.RegisterName
		// TODO: Maybe return a struct with info of registername, prevValue and new value.
		prevValue := s.registers[register]

		// If is a byte operand, eg: al, bl, cl, dl, ah, bh, ch, dh
		// then we need to write the appropiate part of the register
		if destinationOperand.Register.Count == 1 {
			// Shift value left based on the offset, lower register have a 0 offset
			// while higher register have an 1 offset.

			// Ensures corresponding parth of the original register is cleared
			// Where we have 1 are the places where we want to write
			var mask uint16 = 0b00000000_11111111
			var leftShift uint16 = uint16(destinationOperand.Register.Offset) * 8
			// Resets the part of the register that will be written
			s.registers[register] = s.registers[register] & ^(mask << leftShift)
			// Write new value there.
			s.registers[register] = s.registers[register] | (finalValue << leftShift)
		} else {
			s.registers[register] = finalValue
		}

		if s.asmPrinter != nil {
			s.asmPrinter.AddComment(fmt.Sprintf("%v: %v -> %v", register.String(), prevValue, s.registers[register]))
		}
	case SegmentRegisterOperand:
		registerName := destinationOperand.SegmentRegister
		// TODO: Maybe return a struct with info of registername, prevValue and new value.
		prevValue := s.segmentRegisters[registerName]
		s.segmentRegisters[registerName] = finalValue
		if s.asmPrinter != nil {
			s.asmPrinter.AddComment(fmt.Sprintf("%v: %v -> %v", registerName.String(), prevValue, s.segmentRegisters[registerName]))
		}
	}

	return nil
}

func (s *Simulator) String() string {
	var out = "Final Registers:\n"
	out += s.showRegisterValue(RegisterA)
	out += s.showRegisterValue(RegisterB)
	out += s.showRegisterValue(RegisterC)
	out += s.showRegisterValue(RegisterD)
	out += s.showRegisterValue(RegisterSP)
	out += s.showRegisterValue(RegisterBP)
	out += s.showRegisterValue(RegisterSI)
	out += s.showRegisterValue(RegisterDI)
	out += s.showSegmentRegisterValue(SegmentRegisterES)
	out += s.showSegmentRegisterValue(SegmentRegisterSS)
	out += s.showSegmentRegisterValue(SegmentRegisterDS)

	return out
}

func (s *Simulator) showRegisterValue(registerName RegisterName) string {
	registerValue := s.registers[registerName]
	return fmt.Sprintf("\t - %v: 0x%04x (%v)\n", registerName.String(), registerValue, registerValue)
}

func (s *Simulator) showSegmentRegisterValue(segmentRegisterName SegmentRegisterName) string {
	registerValue := s.segmentRegisters[segmentRegisterName]
	return fmt.Sprintf("\t - %v: 0x%04x (%v)\n", segmentRegisterName.String(), registerValue, registerValue)
}
