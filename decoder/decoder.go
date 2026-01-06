package decoder

import (
	"errors"
	"fmt"
	"io"
)

// Details of the mov operation are after page 256 of 8086 user's manual

// Reads an array of binary instructions and iterates over each instruction.
type Decoder struct {
	Data []byte
	pos  int
}

func (d *Decoder) HasNext() bool {
	return (d.pos + 1) < len(d.Data)
}

func (d *Decoder) Next() (Opcode, []byte, error) {
	if !d.HasNext() {
		return 0, nil, io.EOF
	}

	// We pass in the next two bytes, to try to analyze the opcode.
	// Creates an slice that contains the bytes of the instruction
	inst := d.Data[d.pos : d.pos+2] // end of slice range is exlusive
	opcode, bytesToRead, error := d.analyzeOpCode(inst)
	fmt.Printf("Instruction Analysis: %08b\n", inst)

	if error != nil {
		return 0, nil, error
	}

	fullInstruction := d.Data[d.pos : d.pos+bytesToRead]
	fmt.Printf("Full Instruction: %08b\n", fullInstruction)

	d.pos += bytesToRead
	return opcode, fullInstruction, nil
}

// Returns the opcode name and the lenght of bytes to read for this opcode.
func (d *Decoder) analyzeOpCode(instruction []byte) (Opcode, int, error) {
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
	} else if firstByte>>2 == byte(AddRegMemoryWithRegisterToEither) {
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

		return AddRegMemoryWithRegisterToEither, bytesToRead, nil
	} else if firstByte>>2 == byte(AddImmediateToRegisterMemory) {
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
			// If bot are false, then data is a byte.
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
		return AddImmediateToRegisterMemory, bytesToRead, nil
	} else if firstByte>>1 == byte(AddImmediateToAccumulator) {
		// Two bytes for op encoding, and first data.
		var bytesToRead int = 2
		// Instruction operates on word data.
		wField := firstByte&0b1 == 1

		if wField {
			// Extra data field.
			bytesToRead += 1
		}

		return AddImmediateToAccumulator, bytesToRead, nil
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

func (d *Decoder) AsmString(opcode Opcode, instruction []byte) string {
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
	case AddRegMemoryWithRegisterToEither:
		return decodeAddRegMemoryWithRegisterToEither(instruction)
	case AddImmediateToRegisterMemory:
		return decodeAddImmediateToRegisterMemory(instruction)
	case AddImmediateToAccumulator:
		return decodeAddImmediateToAccumulator(instruction)
	}

	return ""
}
