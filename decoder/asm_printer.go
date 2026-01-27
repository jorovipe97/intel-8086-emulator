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

	fmt.Fprint(p.stringsBuilder, printOperand(instruction.Operands.destination, instruction.Flags))
	fmt.Fprint(p.stringsBuilder, separator)
	fmt.Fprint(p.stringsBuilder, printOperand(instruction.Operands.source, instruction.Flags))
	fmt.Fprint(p.stringsBuilder, "\n")
}

func printOperand(operand Operand, instructionFlags InstructionFlag) string {
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
		fmt.Println("Immediate Operand")
		fmt.Printf("%08b - %08b", instructionFlags, InstructionFlagWide)
		var byteWordSpecifier string = "byte"
		if instructionFlags&InstructionFlagWide == InstructionFlagWide {
			byteWordSpecifier = "word"
		}
		return fmt.Sprintf("%v %v", byteWordSpecifier, specificOperand.Value)
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
