package decoder

import "fmt"

type Simulator struct {
	// First item is unnuced, registers start from index 1 to 8
	registers  [9]int16
	asmPrinter *AsmPrinter
}

func NewSimulator(printer *AsmPrinter) *Simulator {
	return &Simulator{
		asmPrinter: printer,
	}
}

func (s *Simulator) getOperandValue(operand Operand) int16 {
	switch sourceOperand := operand.(type) {
	case ImmediateOperand:
		return sourceOperand.Value
	case RegisterOperand:
		return s.registers[sourceOperand.Register.RegisterName]
	}

	return 0
}

func (s *Simulator) ExecInstruction(instruction Instruction) error {
	fmt.Println("#### Executing Instruction")
	destinationValue := s.getOperandValue(instruction.Operands.destination)
	sourceValue := s.getOperandValue(instruction.Operands.source)
	var finalValue int16

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
		s.registers[register] = finalValue
		s.asmPrinter.AddComment(fmt.Sprintf("%v: %v -> %v", register.String(), prevValue, s.registers[register]))
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
	return out
}

func (s *Simulator) showRegisterValue(registerName RegisterName) string {
	registerValue := s.registers[registerName]
	return fmt.Sprintf("\t - %v: 0x%04x (%v)\n", registerName.String(), registerValue, registerValue)
}
