package decoder

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Details of the mov operation are after page 256 of 8086 user's manual

// Reads an array of binary instructions and iterates over each instruction.
type Decoder struct {
	Data                []byte
	DisAsmStringBuilder strings.Builder
	pos                 int
}

func (deco *Decoder) HasNext() bool {
	return (deco.pos + 1) < len(deco.Data)
}

// Decodes and returns the next instruction in the binary array of instructions.
// This moves the position at the end of the instruction.
func (deco *Decoder) NextInstruction() (Instruction, error) {
	if !deco.HasNext() {
		return Instruction{}, io.EOF
	}

	var result Instruction
	var err error
	// Try to decode the instruction, Compare current instruction
	// with all entries in the instruction table.
	for _, candidateInstruction := range instructionsTable {
		result, err = deco.tryDecode(candidateInstruction)
		if err != nil {
			return Instruction{}, err
		}
		if result.Op != OpNone {
			break
		}
	}

	fmt.Printf("- result.size: %v\n", result.Size)
	if result.Size >= MaxInstructionByteCount {
		return Instruction{}, errors.New("instruction is too long")
	}

	// Move current position the amount of bytes of the parsed instruction
	deco.pos += result.Size
	fmt.Printf("---- deco.pos: %v | result.size: %v\n", deco.pos, result.Size)

	return result, nil
}

// Tracks the internal position
type internalPos struct {
	pos int
}

func (deco *Decoder) tryDecode(candidateInstruction InstructionEncoding) (Instruction, error) {
	// This is an optimized map, we use this to track the values of
	// each bit field in the instructino
	bitsParts := make([]int, BitsCount)

	// Track if the instruction has an specific bit field.
	has := make([]bool, BitsCount)

	// Tracks if the instruction can be decoded as the candidateInstruction.
	var valid bool = true

	// The resulting variable.
	var result Instruction

	// The actual bits data pending for procesing.
	var bitsPending uint8
	// The number of bits pending for processing.
	var bitsPendingCount uint8

	// We capture the starting position at which we are going to try decode
	// the candidate instruction
	// Remember this is a try decode we cannot keep moving current deco.pos on each tryDecode
	// all tryies should use the same byte.
	internalPosition := &internalPos{
		pos: deco.pos,
	}

	for _, testBits := range candidateInstruction.bits {
		fmt.Printf("Testing with: %v\n", testBits.Usage)

		// If was not valid, then break
		if !valid {
			break
		}

		// Due to the nature of go o zero initialize everything. Firts bit
		// that is not manually set, will have this to the zero value (BitsEnd)
		// This is a neat way that make us not to specify the terminator explicitly
		if testBits.Usage == BitsEnd {
			break
		}

		// This ussually is 0, however for implied fields, it will have an existing value
		readBits := testBits.Value
		// Mutate the readBits variable accordingly.
		if testBits.BitCount != 0 {
			if bitsPendingCount == 0 {
				// Get next byte
				bitsPendingCount = 8
				// remember this is a try decode we cannot keep moving current position on each tryDecode
				// all tryies should use the same byte.
				bitsPending = deco.Data[internalPosition.pos]
				internalPosition.pos++
			}

			// NOTE(casey): If this assert fires, it means we have an error in our table,
			// since there are no 8086 instructions that have bit values straddling a
			// byte boundary.
			if testBits.BitCount > bitsPendingCount {
				return Instruction{}, errors.New("instructions table has an error")
			}

			bitsPendingCount -= testBits.BitCount
			readBits = bitsPending
			readBits = readBits >> bitsPendingCount

			// Let's say BitsPending = 0b10001101 and we want to extract 6 bits, then 2 bits are pending:
			// First extraction (6 bits):
			//
			// BitsPending     = 10001101
			// BitsPendingCount = 8 - 6 = 2
			//
			// ReadBits = 10001101 >> 2 = 00100011
			// Mask = ~(0xff << 6) = ~(11000000) = 00111111
			//
			// ReadBits & Mask = 00100011 & 00111111 = 00100011 ✓ (extracted top 6 bits)
			readBits = readBits & ^(0xff << testBits.BitCount)
		}

		// Either check if literal bits is equal, or save the fields.
		if testBits.Usage == BitsLiteral {
			// All instructions start here, if readBits is not the opcode of the candidateInstruction
			// then is invalid, this byte does not represents the candidate instruction.
			valid = valid && (readBits == testBits.Value)
			fmt.Printf("- Literal - readBits: %b - testBits: %b\n", readBits, testBits.Value)
		} else {
			fmt.Printf("- Field - %v: %b\n", testBits.Usage, readBits)
			bitsParts[testBits.Usage] = int(readBits)
			has[testBits.Usage] = true
		}
	}

	fmt.Printf("- Is Valid? %v\n\n", valid)
	if valid {
		// If code entered here, then it already read all the bytes in the candidateInstruction

		d := bitsParts[BitsD]
		w := bitsParts[BitsW]
		mod := bitsParts[BitsMod]
		rm := bitsParts[BitsRm]
		s := bitsParts[BitsS]

		hasDirectAddress := mod == 0b00 && rm == 0b110
		dispIsW := mod == 0b10 || hasDirectAddress
		dataIsW := bitsParts[BitsWMakesDataWide] == 0b1 && s == 0b0 && w == 0b1

		// Warning, the order calling of parseDispValue and parseDataValue
		// is important, because those function update the internalPosition.
		bitsParts[BitsDisp] = deco.parseDispValue(internalPosition, mod, dispIsW)
		bitsParts[BitsData] = deco.parseDataValue(internalPosition, has[BitsData], dataIsW)

		result.Op = candidateInstruction.op
		// How many bytes did move the internal cursor to decode the full instruction.
		result.Size = internalPosition.pos - deco.pos
		fmt.Printf("- result.size: %v\n", result.Size)
		result.RawBits = deco.Data[deco.pos:internalPosition.pos]
		result.Parts = candidateInstruction.bits
		result.Mnemonic = candidateInstruction.mnemonic

		if has[BitsW] && bitsParts[BitsW] == 0b1 {
			result.Flags |= InstructionFlagWide
		}

		// Instruction source is specified in REG field
		var regOperand Operand
		var modOperand Operand

		if has[BitsReg] {
			regOperand = getRegOperand(bitsParts[BitsReg], w)
		}

		if has[BitsMod] {
			if mod == 0b11 {
				// If MOD==0b11 (register-to-register mode), then
				// R/M identifies the second register operand.
				modOperand = getRegOperand(rm, w)
			} else {
				terms0Table := [8]RegisterName{
					RegisterB,
					RegisterB,
					RegisterBP,
					RegisterBP,
					RegisterSI,
					RegisterDI,
					RegisterBP,
					RegisterB,
				}
				terms1Table := [8]RegisterName{
					RegisterSI,
					RegisterDI,
					RegisterSI,
					RegisterDI,
				}

				term0 := terms0Table[rm&0b111]
				term1 := terms1Table[rm&0b111]

				if mod == 0b00 && rm == 0b110 {
					// On this case effective address calculation simpy uses
					// a direct address.
					term0 = RegisterNone
					term1 = RegisterNone
				}
				modOperand = MemoryOperand{
					Terms: [2]RegisterInfo{
						{
							RegisterName: term0,
							Count:        2,
							Offset:       0,
						},
						{
							RegisterName: term1,
							Count:        2,
							Offset:       0,
						},
					},
					Displacement: bitsParts[BitsDisp],
				}
			}
		}

		if has[BitsD] {
			fmt.Println("- has bits d, using it....")
			switch d {
			case 0:
				// Instruction source is specified in REG field.
				result.Operands.destination = modOperand
				result.Operands.source = regOperand
			case 1:
				// Instruction destination is specified in REG field.
				result.Operands.destination = regOperand
				result.Operands.source = modOperand
			}
		} else if has[BitsReg] {
			fmt.Println("- it does not have bits d....")

			// For example, Mov Immediate To Register mode does not have
			// a d flag. In that case reg is always the destination
			// eg: mov cl, 12
			result.Operands.destination = regOperand
			if !has[BitsData] {
				result.Operands.source = modOperand
			}
		} else if !has[BitsReg] {
			// Some operations don have reg field, eg: the mov immediate to register/memory
			// eg: mov [bp + di], byte 7
			result.Operands.destination = modOperand
		}

		// NOTE(casey): Because there are some strange opcodes that do things like have an immediate as
		// a _destination_ ("out", for example), I define immediates and other "additional operands" to
		// go in "whatever slot was not used by the reg and mod fields".
		if has[BitsData] {
			if result.Operands.source == nil {
				fmt.Println("- Source is nil")
				result.Operands.source = ImmediateOperand{
					Value: int16(bitsParts[BitsData]),
				}
			} else if result.Operands.destination == nil {
				fmt.Println("- destination is nil")
				result.Operands.destination = ImmediateOperand{
					Value: int16(bitsParts[BitsData]),
				}
			}
		}
	}

	fmt.Printf("- result.size: %v\n", result.Size)

	// If instruction was not valid, it will return an zero initialized Instruction struct.
	// whose op is None, this will make parent code to try with the next candidate instruction.
	return result, nil
}

func (deco *Decoder) parseDispValue(currentPosition *internalPos, mod int, dispIsW bool) int {
	// TODO: Handle sign extension....
	var displacement int
	if dispIsW {
		// Memory mode, 16-bit displacement follows
		// Or mod == was 0b00 and rm == 0b110.
		disp0 := int(deco.Data[currentPosition.pos])
		disp1 := int(deco.Data[currentPosition.pos+1])
		// Perform cast to sign 16 so we get symbol correctly.
		// Then cast to int so go compiler performs a sign extension.
		displacement = int(int16((disp1 << 8) | disp0))
		currentPosition.pos += 2
	} else if mod == 0b01 {
		// Memory mode, 8-bit displacement follows
		// Perform cast to sign 8 so we get symbol correctly.
		// Then cast to int so go compiler performs a sign extension.
		displacement = int(int8(deco.Data[currentPosition.pos]))
		currentPosition.pos += 1
	}

	return displacement
}

func (deco *Decoder) parseDataValue(currentPosition *internalPos, hasData, dataIsW bool) int {
	var data int

	if hasData {
		if dataIsW {
			data0 := int(deco.Data[currentPosition.pos])
			data1 := int(deco.Data[currentPosition.pos+1])
			// Perform cast to sign 16 so we get symbol correctly.
			// Then cast to int so go compiler performs a sign extension.
			data = int(int16((data1 << 8) | data0))
			currentPosition.pos += 2
		} else {
			// Perform cast to sign 8 so we get symbol correctly.
			// Then cast to int so go compiler performs a sign extension.
			data = int(int8(deco.Data[currentPosition.pos]))
			currentPosition.pos += 1
		}
	}

	return data
}
