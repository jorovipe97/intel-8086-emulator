package decoder

import (
	"fmt"
	"strings"
)

type AsmPrinter struct {
	stringsBuilder *strings.Builder
}

func NewAsmPrinter() *AsmPrinter {
	return &AsmPrinter{
		stringsBuilder: &strings.Builder{},
	}
}

func (p *AsmPrinter) AddComment(comment string) {
	fmt.Fprintf(p.stringsBuilder, "; %v\n", comment)
}

// Tells assembler we intent to run assembly for old 8086 architecture.
func (p *AsmPrinter) Bits16Header() {
	p.stringsBuilder.WriteString("bits 16\n")
}

func (p *AsmPrinter) AddInstruction(instruction Instruction) {
	fmt.Fprintf(p.stringsBuilder, "%v ", instruction.Mnemonic)

	var separator string
	if instruction.Operands.destination != nil && instruction.Operands.source != nil {
		separator = ", "
	}

	// When the destination is a register, the assembler infers the size from the register name:
	// MOV AX, [BX]      ; AX is 16-bit → word operation (no specifier needed)
	// MOV AL, [BX]      ; AL is 8-bit → byte operation (no specifier needed)
	// But when the destination is memory, there's no way to know the size:
	// MOV [BX], 5       ; Is this 8-bit or 16-bit? Assembler can't tell!
	// So you MUST specify:
	// MOV byte [BX], 5   ; Store 5 as 8-bit
	// MOV word [BX], 5   ; Store 5 as 16-bit
	//
	// If destination is not register or a insutrction pointer increment do not add size of destination.
	switch instruction.Operands.destination.(type) {
	case RegisterOperand, InstructionPointerIncrementOperand:
		fmt.Println("Instruction is a RegisterOperand or InstructionPointerIncrementOperand, ")
	default:
		fmt.Println("Not Register Operand nor Instruction Pointer Increment Operand...")
		fmt.Printf("%08b - %08b", instruction.InstructionExtras, InstructionFlagWide)
		if instruction.InstructionExtras&InstructionFlagWide == InstructionFlagWide {
			fmt.Fprint(p.stringsBuilder, "word ")
		} else {
			fmt.Fprint(p.stringsBuilder, "byte ")
		}
	}

	fmt.Fprint(p.stringsBuilder, printOperand(instruction.Operands.destination, instruction.Size))
	fmt.Fprint(p.stringsBuilder, separator)
	fmt.Fprint(p.stringsBuilder, printOperand(instruction.Operands.source, instruction.Size))
	fmt.Fprint(p.stringsBuilder, "\n")
}

func printOperand(operand Operand, instructionSize int) string {
	fmt.Println("Printing Operand...")

	if operand == nil {
		fmt.Println("Operand is nil")
		return ""
	}

	switch specificOperand := operand.(type) {
	case RegisterOperand:
		fmt.Println("Register Operand")
		return getRegName(specificOperand.Register)
	case MemoryOperand:
		fmt.Println("Memory Operand")
		return getEffectiveAddressExpression(specificOperand)
	case ImmediateOperand:
		return fmt.Sprintf("%v", specificOperand.Value)
	case SegmentRegisterOperand:
		return specificOperand.SegmentRegister.String()
	case InstructionPointerIncrementOperand:
		// In NASM, the $ symbol is a special token that represents the current
		// address of the line being assembled. It is essentially a "you are here"
		// marker for the assembler.
		//
		// We need to do this because if we used the offset value directly, nasm will interpret it
		// as an absolute address.
		//
		// So we convert the relative offset to an absolute address so that when nasm assembles the generated assembly
		// it returns a correct binary. This is just needed for assembler, instruction jumps are always relative according to cpu.
		// If you look at the raw binary you will see the actual relative jump.
		return fmt.Sprintf("$+%v+%v", instructionSize, specificOperand.Increment)
	}

	fmt.Println("Unknown operand type")
	return ""
}

var registersMappings = [...][3]string{
	{"", "", ""}, // RegisterNone
	{"al", "ah", "ax"},
	{"cl", "ch", "cx"},
	{"dl", "dh", "dx"},
	{"bl", "bh", "bx"},
	{"sp", "sp", "sp"},
	{"bp", "bp", "bp"},
	{"si", "si", "si"},
	{"di", "di", "di"},
	{"", "", ""}, // RegisterCount
}

func getRegName(register RegisterInfo) string {
	fmt.Printf("regName: %v, offset: %v, count: %v", register.RegisterName, register.Offset, register.Count)
	var tableColumn int
	if register.Count == 2 {
		tableColumn = 2 // Use third column
	} else {
		// & 0b1 is a defensive progrraming thecnique,
		// in case offset is neither 0 or 1.
		tableColumn = register.Offset & 0b1
	}

	return registersMappings[register.RegisterName][tableColumn]
}

func getEffectiveAddressExpression(memoryOperand MemoryOperand) string {
	var out string
	var separator string

	// Tracks if the expression had register terms, there are cases where effective getEffectiveAddressExpression
	// is a direct displacement.
	var hadTerms bool

	// TODO: Handle wide, byte cases.
	out += "["
	for _, term := range memoryOperand.Terms {
		if term.RegisterName != RegisterNone {
			out += separator
			out += getRegName(term)
			separator = "+"
			hadTerms = true
		}
	}

	// Print the displacement if we had no register terms OR if the displacement is non-zero
	if !hadTerms || memoryOperand.Displacement != 0 {
		out += fmt.Sprintf("%+d", memoryOperand.Displacement)
	}

	out += "]"

	return out
}

func (p *AsmPrinter) String() string {
	return p.stringsBuilder.String()
}
