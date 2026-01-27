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

// Returns the opcode name and the lenght of bytes to read for this opcode.
func (deco *Decoder) analyzeOpCode(instruction []byte) (Opcode, int, error) {
	firstByte := instruction[0]
	// op code is usually encoded in the first 6 bits of the first byte.
	if firstByte>>2 == byte(MovRegisterMemoryToFromRegister) {
		// Register mode/Memory mode with displacement length
		modField := instruction[1] >> 6
		var bytesToRead int = 0

		switch modField {
		case 0b00:
			// Memory mode, no displacement follows.
			// Except when R/M field = 110, then, 16-bit displacement follwos.
			rmField := instruction[1] & 0b0000_0111
			if rmField == 0b110 {
				bytesToRead = 4
			} else {
				bytesToRead = 2
			}
		case 0b01:
			// Memory mode, 8 bit displacement follows
			bytesToRead = 3 // An additional byte
		case 0b10:
			// Memory mode, 16 bit displacement follows
			bytesToRead = 4 // Two additional bytes.
		case 0b11:
			// Register mode (no displacement)
			bytesToRead = 2
		}

		return MovRegisterMemoryToFromRegister, bytesToRead, nil
	} else if firstByte>>1 == byte(MovMemoryToAccumulator) {
		var bytesToRead int = 2
		wField := firstByte&0b1 == 1
		if wField {
			// additional data byte
			bytesToRead += 1
		}
		return MovMemoryToAccumulator, bytesToRead, nil
	} else if firstByte>>1 == byte(MovAccumulatorToMemory) {
		var bytesToRead int = 2
		wField := firstByte&0b1 == 1
		if wField {
			// additional data byte
			bytesToRead += 1
		}
		return MovAccumulatorToMemory, bytesToRead, nil
	} else if firstByte>>1 == byte(MovImmediateToRegisterMemory) {
		// Two bytes for op encoding, and a data byte
		var bytesToRead int = 3
		wField := firstByte&0b1 == 1
		if wField {
			// additional data byte
			bytesToRead += 1
		}

		// Register mode/Memory mode with displacement length
		modField := instruction[1] >> 6
		switch modField {
		case 0b01:
			// Memory mode, 8 bit displacement follows
			bytesToRead += 1 // An additional byte
		case 0b10:
			// Memory mode, 16 bit displacement follows
			bytesToRead += 2 // Two additional bytes.
		}
		return MovImmediateToRegisterMemory, bytesToRead, nil
	} else if firstByte>>4 == byte(MovImmediateToRegister) {
		var bytesToRead int = 2
		var isWord bool = (firstByte>>3)&0b00001 == 1
		if isWord {
			bytesToRead = 3
		}

		return MovImmediateToRegister, bytesToRead, nil
	} else if firstByte>>2 == byte(AddRegMemoryWithRegisterToEither) ||
		firstByte>>2 == byte(SubRegMemoryWithRegisterToEither) ||
		firstByte>>2 == byte(CmpRegMemoryWithRegisterToEither) {
		// Register mode/Memory mode with displacement length
		modField := instruction[1] >> 6
		var bytesToRead int = 0

		switch modField {
		case 0b00:
			// Memory mode, no displacement follows.
			// Except when R/M field = 110, then, 16-bit displacement follwos.
			rmField := instruction[1] & 0b0000_0111
			if rmField == 0b110 {
				bytesToRead = 4
			} else {
				bytesToRead = 2
			}
		case 0b01:
			// Memory mode, 8 bit displacement follows
			bytesToRead = 3 // An additional byte
		case 0b10:
			// Memory mode, 16 bit displacement follows
			bytesToRead = 4 // Two additional bytes.
		case 0b11:
			// Register mode (no displacement)
			bytesToRead = 2
		}

		var opcode Opcode
		switch firstByte >> 2 {
		case byte(AddRegMemoryWithRegisterToEither):
			opcode = AddRegMemoryWithRegisterToEither
		case byte(SubRegMemoryWithRegisterToEither):
			opcode = SubRegMemoryWithRegisterToEither
		case byte(CmpRegMemoryWithRegisterToEither):
			opcode = CmpRegMemoryWithRegisterToEither
		}

		return opcode, bytesToRead, nil
	} else if firstByte>>2 == byte(ArithmeticImmediateToRegisterMemory) {
		// Two bytes for op encoding
		var bytesToRead int = 2

		// Instruction operates on word data.
		wField := firstByte&0b1 == 1

		// Sign extend 8-bit immediate data to 16 bits if w=1
		// If s==1 and w==1 then, data is a single byte (not two)
		signField := (firstByte>>1)&0b1 == 1

		if wField && signField {
			// additional data byte
			bytesToRead += 1
		} else if wField {
			bytesToRead += 2
		} else {
			// If both are false, then data is a byte.
			bytesToRead += 1
		}

		// Register mode/Memory mode with displacement length
		modField := instruction[1] >> 6
		switch modField {
		case 0b00:
			// Memory mode, no displacement follows.
			// Except when R/M field = 110, then, 16-bit displacement follwos.
			rmField := instruction[1] & 0b0000_0111
			if rmField == 0b110 {
				bytesToRead += 2
			}
		case 0b01:
			// Memory mode, 8 bit displacement follows
			bytesToRead += 1 // An additional byte
		case 0b10:
			// Memory mode, 16 bit displacement follows
			bytesToRead += 2 // Two additional bytes.
		}
		fmt.Printf("Bytes to read: %v\n", bytesToRead)
		return ArithmeticImmediateToRegisterMemory, bytesToRead, nil
	} else if firstByte>>1 == byte(AddImmediateToAccumulator) ||
		firstByte>>1 == byte(SubImmediateToAccumulator) ||
		firstByte>>1 == byte(CmpImmediateToAccumulator) {
		// Two bytes for op encoding, and first data.
		var bytesToRead int = 2
		// Instruction operates on word data.
		wField := firstByte&0b1 == 1

		if wField {
			// Extra data field.
			bytesToRead += 1
		}

		var opcode Opcode
		switch firstByte >> 1 {
		case byte(AddImmediateToAccumulator):
			opcode = AddImmediateToAccumulator
		case byte(SubImmediateToAccumulator):
			opcode = SubImmediateToAccumulator
		case byte(CmpImmediateToAccumulator):
			opcode = CmpImmediateToAccumulator
		}

		return opcode, bytesToRead, nil
	}

	return 0, 0, errors.New("cannot identify instruction")
}

func effectiveAddressCalculation(rmField byte) string {
	switch rmField {
	case 0b00:
		return "bx + si"
	case 0b001:
		return "bx + di"
	case 0b010:
		return "bp + si"
	case 0b011:
		return "bp + di"
	case 0b100:
		return "si"
	case 0b101:
		return "di"
	case 0b110:
		// 16 bits direct address when mod = 00
		return "bp"
	case 0b111:
		return "bx"
	}

	return ""
}

func (deco *Decoder) AsmString(opcode Opcode, instruction []byte) string {
	// op code is usually encoded in the first 6 bits of the first byte.
	switch opcode {
	case MovRegisterMemoryToFromRegister:
		return decodeMovRegisterMemoryToFromRegister(instruction)
	case MovImmediateToRegisterMemory:
		return decodeMovImmediateToRegisterMemory(instruction)
	case MovMemoryToAccumulator:
		return decodeMovMemoryToAccumulator(instruction)
	case MovAccumulatorToMemory:
		return decodeMovAccumulatorToMemory(instruction)
	case MovImmediateToRegister:
		return decodeMovImmediateToRegister(instruction)
	case AddRegMemoryWithRegisterToEither,
		SubRegMemoryWithRegisterToEither,
		CmpRegMemoryWithRegisterToEither:
		return decodeAddRegMemoryWithRegisterToEither(instruction)
	case ArithmeticImmediateToRegisterMemory:
		return decodeAddImmediateToRegisterMemory(instruction)
	case AddImmediateToAccumulator,
		SubImmediateToAccumulator,
		CmpImmediateToAccumulator:
		return decodeAddImmediateToAccumulator(instruction)
	}

	return ""
}
