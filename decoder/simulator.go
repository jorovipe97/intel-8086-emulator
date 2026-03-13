package decoder

import (
	"fmt"
	"math/bits"
)

type Simulator struct {
	// First item is unnuced, registers start from index 1 to 8
	registers        [9]uint16
	segmentRegisters [4]uint16
	flags            int
	memory           *Memory
}

func NewSimulator(memory *Memory) *Simulator {
	return &Simulator{
		memory: memory,
	}
}

func (s *Simulator) ExecInstruction(instruction Instruction) error {
	destinationValue := s.getOperandValue(instruction.Operands.destination)
	sourceValue := s.getOperandValue(instruction.Operands.source)
	var finalValue uint16

	// Computes the final value, based on the op type
	switch instruction.Op {
	case OpMov:
		finalValue = sourceValue
	case OpAdd:
		finalValue = destinationValue + sourceValue
	case OpSub, OpCmp:
		finalValue = destinationValue - sourceValue
	}

	// Compute flags.
	if instruction.AffectedFlags&FlagZF == FlagZF {
		// When 0, value is not zero
		zfValue := 0
		if finalValue == 0 {
			zfValue = 1
		}

		s.setFlagValue(FlagZF, zfValue)
	}

	if instruction.AffectedFlags&FlagSF == FlagSF {
		// When 0, value is positive
		sfValue := 0
		// TODO: Handle byte and word cases
		if finalValue&(1<<15) != 0 {
			sfValue = 1
		}
		s.setFlagValue(FlagSF, sfValue)
	}

	if instruction.AffectedFlags&FlagPF == FlagPF {
		// Is 0 when the count is odd.
		pfValue := 0

		if bits.OnesCount8(uint8(finalValue))&1 == 0 {
			// Is 1 when the count is even
			pfValue = 1
		}

		s.setFlagValue(FlagPF, pfValue)
	}

	// if instruction.AffectedFlags&FlagAF == FlagAF {
	// 	// Is 0 when the count is odd.
	// 	afValue := 0

	// 	if bits.OnesCount8(uint8(finalValue))&1 == 0 {
	// 		// Is 1 when the count is even
	// 		afValue = 1
	// 	}

	// 	s.setFlagValue(FlagAF, afValue)
	// }

	if instruction.AffectedFlags&FlagCF == FlagCF {
		// https://www.youtube.com/watch?v=F20rPdjGI8k
		cfValue := 0

		// Overflow calculation, depends on the operation
		switch instruction.Op {
		case OpAdd:
			// If the final value is lower than one of the operands
			// then is because there happened an overflow.
			if finalValue < destinationValue {
				cfValue = 1
			}
		case OpSub, OpCmp:
			// Checks if value crossed below 0
			if destinationValue < sourceValue {
				cfValue = 1
			}
		}

		s.setFlagValue(FlagCF, cfValue)
	}

	if instruction.AffectedFlags&FlagOF == FlagOF {
		ofValue := 0

		// Overflow calculation, depends on the operation
		switch instruction.Op {
		case OpAdd:
			if instruction.InstructionExtras&InstructionFlagWide == InstructionFlagWide {
				// If operands are positive but result is negative, an overflow happened.
				positiveOverflow := int16(destinationValue) >= 0 && int16(sourceValue) >= 0 && int16(finalValue) < 0

				// If operands are negative but result is positive or 0, an overflow happened
				negativeOverflow := int16(destinationValue) <= 0 && int16(sourceValue) <= 0 && int16(finalValue) >= 0

				if positiveOverflow || negativeOverflow {
					ofValue = 1
				}
			} else {
				// If operands are positive but result is negative, an overflow happened.
				positiveOverflow := int8(destinationValue) >= 0 && int8(sourceValue) >= 0 && int8(finalValue) < 0

				// If operands are negative but result is positive or 0, an overflow happened
				negativeOverflow := int8(destinationValue) <= 0 && int8(sourceValue) <= 0 && int8(finalValue) >= 0

				if positiveOverflow || negativeOverflow {
					ofValue = 1
				}
			}
		case OpSub, OpCmp:
			// For sub, an overflow happens if operands have different sign
			// and the result have a sign different to the first operand
			// then an overflow happened
			if instruction.InstructionExtras&InstructionFlagWide == InstructionFlagWide {
				// For example if:
				// 1000 0010 ^
				// 0000 1010
				// ---------
				// 1000 1000 (Note the most significate bit is 1, then a negative result indicates signs are different)
				// We cast to int16, so the result give us a negative number when msb is 1.
				areOperandsSignsDifferent := int16(destinationValue)^int16(sourceValue) < 0

				// Did result sign changed from sign of first operand?
				didResultChangedSign := int16(destinationValue)^int16(finalValue) < 0

				if areOperandsSignsDifferent && didResultChangedSign {
					ofValue = 1
				}
			} else {
				areOperandsSignsDifferent := int8(destinationValue)^int8(sourceValue) < 0
				// Did result sign changed from sign of first operand?
				didResultChangedSign := int8(destinationValue)^int8(finalValue) < 0

				if areOperandsSignsDifferent && didResultChangedSign {
					ofValue = 1
				}
			}
		}

		s.setFlagValue(FlagOF, ofValue)
	}

	// Check if instruction is a cmp, This instructions does not writes to destination
	// operand, just affects flags, this instruction is usually used to control the program
	// execution flow.
	if instruction.Op == OpCmp {
		return nil
	}

	// Updates simulated memory. Destination can be a register or memory.
	switch destinationOperand := instruction.Operands.destination.(type) {
	case RegisterOperand:
		// Identify which part of memory should store result
		register := destinationOperand.Register.RegisterName

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
			s.registers[register] = s.registers[register] | ((finalValue & mask) << leftShift)
		} else {
			s.registers[register] = finalValue
		}
	case SegmentRegisterOperand:
		registerName := destinationOperand.SegmentRegister
		s.segmentRegisters[registerName] = finalValue
	case InstructionPointerIncrementOperand:
		switch instruction.Op {
		case OpJNZ:
			if s.getFlagValue(FlagZF) == 0 {
				s.memory.IncrementPosition(destinationOperand.Increment)
			}
		case OpJZ:
			if s.getFlagValue(FlagZF) == 1 {
				s.memory.IncrementPosition(destinationOperand.Increment)
			}
		case OpJP:
			if s.getFlagValue(FlagPF) == 1 {
				s.memory.IncrementPosition(destinationOperand.Increment)
			}
		}
	}

	// TODO: Implement IP register and interaction with memory
	// Maybe move where we incrmente absolute position to this function?. But then decoder wont work without simulator, bad design?

	return nil
}

func (s *Simulator) getOperandValue(operand Operand) uint16 {
	switch specificOperand := operand.(type) {
	case ImmediateOperand:
		return uint16(specificOperand.Value)
	case RegisterOperand:
		// If is a byte operand, eg: al, bl, cl, dl, ah, bh, ch, dh
		// then we need to write the appropiate part of the register
		if specificOperand.Register.Count == 1 {
			// This is used to remove the part of the register we are not interested in.
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

func (s *Simulator) setFlagValue(flag Flag, value int) {
	if flag == FlagNone {
		// No flag should be affected.
		return
	}

	if value != 0 && value != 1 {
		// Value can only be a 1 or 0. Maybe use a bool parameter?
		return
	}

	// We substract 1 to the flag because enum values start at 1,
	// to let the 0 value as the FlagNone
	// - compute flag offset
	flagOffset := int(flag - 1)
	// - reset flag position
	s.flags = s.flags & ^(1 << flagOffset)
	// - sets new value into flag position
	s.flags = s.flags | (value << flagOffset)
}

func (s *Simulator) getFlagValue(flag Flag) int {
	if flag == FlagNone {
		return 0
	}

	// We substract 1 to the flag because enum values start at 1,
	// to let the 0 value as the FlagNone
	flagOffset := flag - 1
	return (s.flags & (1 << int(flagOffset))) >> int(flagOffset)
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
	out += s.showIpRegister()

	out += "\n\nFlags:"
	if s.getFlagValue(FlagCF) == 1 {
		out += "\t - CF\n"
	}
	if s.getFlagValue(FlagPF) == 1 {
		out += "\t - PF\n"
	}
	if s.getFlagValue(FlagAF) == 1 {
		out += "\t - AF\n"
	}
	if s.getFlagValue(FlagZF) == 1 {
		out += "\t - ZF\n"
	}
	if s.getFlagValue(FlagSF) == 1 {
		out += "\t - SF\n"
	}
	if s.getFlagValue(FlagOF) == 1 {
		out += "\t - OF\n"
	}

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

func (s *Simulator) showIpRegister() string {
	ipRegister := s.memory.GetIPRegister()
	return fmt.Sprintf("\t - ip: 0x%04x (%v)", ipRegister, ipRegister)
}
