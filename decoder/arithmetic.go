package decoder

import (
	"fmt"
	"strings"
)

// Register/memory to/from register
// [0 0 0 0 0 0 d w]
// [mod(2 bits) reg(3 bits) rm(3 bits)]
// [Displacement Low (8 bits)]
// [Displacement Hight (8 bits)]
func decodeAddRegMemoryWithRegisterToEither(instruction []byte) string {
	var builder strings.Builder // Zero value is ready to use

	// The bit 8 of first byte determine the w field:
	// when 0, instruction operates on byte data
	// when 1, instructions operate on word data
	w := instruction[0]&0b1 == 1

	// when 0, instruction source is specified in reg field.
	// when 1, instruction destination is specified in reg field
	d := (instruction[0]>>1)&0b1 == 1

	modField := (instruction[1] >> 6) & 0b11

	// 2. Decode the source registry (when bit 7 of first byte is 0, reg is the source)
	// Destination is in in second byte.
	regField := (instruction[1] >> 3) & 0b0000_0111

	// 3. Decode the destination registry.
	rmField := instruction[1] & 0b0000_0111

	builder.WriteString("add ")

	switch modField {
	case 0b00:
		// Memory only, no displacement follows
		// except when rmField = 110

		// Use reg field as the destination.
		if d {
			builder.WriteString(fmt.Sprintf("%v, ", byteToRegisterString(w, regField)))
		}

		if rmField == 0b110 {
			displacement := int16(instruction[2]) | int16(instruction[3])<<8
			builder.WriteString(
				fmt.Sprintf("[%v]", displacement),
			)
		} else {
			builder.WriteString(
				fmt.Sprintf("[%v]", effectiveAddressCalculation(rmField)),
			)
		}

		// use reg field as the source
		if !d {
			builder.WriteString(fmt.Sprintf(", %v", byteToRegisterString(w, regField)))
		}
	case 0b01:
		// Memory mode, 8-bit displacement follows.

		// Use reg field as the destination.
		if d {
			builder.WriteString(fmt.Sprintf("%v, ", byteToRegisterString(w, regField)))
		}

		displacement := int8(instruction[2])
		builder.WriteString(
			fmt.Sprintf("[%v + %v]", effectiveAddressCalculation(rmField), displacement),
		)

		// use reg field as the source
		if !d {
			builder.WriteString(fmt.Sprintf(", %v", byteToRegisterString(w, regField)))
		}
	case 0b10:
		// Memory mode, 16-bit displacement follows
		// Use reg field as the destination.
		if d {
			builder.WriteString(fmt.Sprintf("%v, ", byteToRegisterString(w, regField)))
		}

		displacement := int16(instruction[2]) | int16(instruction[3])<<8
		builder.WriteString(
			fmt.Sprintf("[%v + %v]", effectiveAddressCalculation(rmField), displacement),
		)

		// use reg field as the source
		if !d {
			builder.WriteString(fmt.Sprintf(", %v", byteToRegisterString(w, regField)))
		}
	case 0b11:
		// Register mode, no displacement
		builder.WriteString(
			fmt.Sprintf(
				"%v, %v",
				byteToRegisterString(w, rmField),
				byteToRegisterString(w, regField),
			),
		)
	}

	return builder.String()
}

// Immediate to register/memory
// [1 0 0 0 0 0 s w]
// [mod(2 bits) 0 0 0 rm(3 bits)]
// [Displacement Low (8 bits)]
// [Displacement Hight (8 bits)]
// [Data]
// [Data(if w = 1)]
func decodeAddImmediateToRegisterMemory(instruction []byte) string {
	var builder strings.Builder // Zero value is ready to use

	// The bit 8 of first byte determine the w field:
	// when 0, instruction operates on byte data
	// when 1, instructions operate on word data
	w := instruction[0]&0b1 == 1

	// Sign extend 8-bit immediate data to 16 bits if w=1
	// If s==1 and w==1 then, data is a single byte (not two)
	s := (instruction[0]>>1)&0b1 == 1

	// Wheter data field is a byte (8 bits) or a word (16 bits)
	var dataIsByte bool
	if s && w {
		dataIsByte = true
	} else if w {
		dataIsByte = false
	} else {
		dataIsByte = true
	}

	modField := (instruction[1] >> 6) & 0b11

	// 3. Decode the destination registry.
	rmField := instruction[1] & 0b0000_0111

	builder.WriteString("add ")

	switch modField {
	case 0b00:
		// Memory only, no displacement follows
		// except when rmField = 110
		if rmField == 0b110 {
			builder.WriteString("direct address 2")
		} else {
			if dataIsByte {
				builder.WriteString("byte ")
			} else {
				builder.WriteString("word ")
			}
			builder.WriteString(
				fmt.Sprintf("[%v]", effectiveAddressCalculation(rmField)),
			)
		}

		if dataIsByte {
			data := instruction[2]
			builder.WriteString(
				fmt.Sprintf(", %v", data),
			)
		} else {
			data := uint16(instruction[2]) | uint16(instruction[3])<<8
			builder.WriteString(
				fmt.Sprintf(", %v", data),
			)
		}
	case 0b01:
		// Memory mode, 8-bit displacement follows.
		displacement := int8(instruction[2])

		if dataIsByte {
			builder.WriteString("byte ")
		} else {
			builder.WriteString("word ")
		}

		builder.WriteString(
			fmt.Sprintf("[%v + %v]", effectiveAddressCalculation(rmField), displacement),
		)

		if dataIsByte {
			data := instruction[3]
			builder.WriteString(
				fmt.Sprintf(", %v", data),
			)
		} else {
			data := uint16(instruction[3]) | uint16(instruction[4])<<8
			builder.WriteString(
				fmt.Sprintf(", %v", data),
			)
		}
	case 0b10:
		// Memory mode, 16-bit displacement follows
		displacement := uint16(instruction[2]) | uint16(instruction[3])<<8
		if dataIsByte {
			builder.WriteString("byte ")
		} else {
			builder.WriteString("word ")
		}

		builder.WriteString(
			fmt.Sprintf("[%v + %v]", effectiveAddressCalculation(rmField), displacement),
		)

		// If should sign extend 8 bytes to 16 bytes, then, data is a single byte.
		if dataIsByte {
			data := instruction[4]
			builder.WriteString(
				fmt.Sprintf(", %v", data),
			)
		} else {
			data := uint16(instruction[4]) | uint16(instruction[5])<<8
			builder.WriteString(
				fmt.Sprintf(", %v", data),
			)
		}
	case 0b11:
		// Register mode (no displacement)
		builder.WriteString(
			byteToRegisterString(w, rmField),
		)

		// If should sign extend 8 bytes to 16 bytes, then, data is a single byte.
		if dataIsByte {
			data := uint16(instruction[2])
			builder.WriteString(
				fmt.Sprintf(", %v", data),
			)
		} else {
			data := int16(instruction[2]) | int16(instruction[3])<<8
			builder.WriteString(
				fmt.Sprintf(", %v", data),
			)
		}
	}

	return builder.String()
}

// [0 0 0 0 0 1 0 w]
// [data low]
// [data high if w = 1]
func decodeAddImmediateToAccumulator(instruction []byte) string {
	var builder strings.Builder // Zero value is ready to use

	// The bit 8 of first byte determine the isWord field:
	// when 0, instruction operates on byte data
	// when 1, instructions operate on word data
	isWord := instruction[0]&0b1 == 1

	builder.WriteString("add ")
	if isWord {
		data := int16(instruction[1]) | int16(instruction[2])<<8
		builder.WriteString(
			fmt.Sprintf("ax, %v", data),
		)
	} else {
		data := int8(instruction[1])
		builder.WriteString(
			fmt.Sprintf("al, %v", data),
		)
	}

	return builder.String()
}
