package decoder

import (
	"fmt"
	"strings"
)

// Register/memory to/from register
// [1 0 0 0 1 0 d w]
// [mod(2 bits) reg(3 bits) rm(3 bits)]
// [Displacement Low (8 bits)]
// [Displacement Hight (8 bits)]
func decodeMovRegisterMemoryToFromRegister(instruction []byte) string {
	var builder strings.Builder // Zero value is ready to use

	// The bit 8 of first byte determine the w field:
	// when 0, instruction operates on byte data
	// when 1, instructions operate on word data
	w := instruction[0]&0b1 == 1

	// when 0, instruction source is specified in ref field.
	// when 1, instruction destination is specified in reg field
	d := (instruction[0]>>1)&0b1 == 1

	modField := (instruction[1] >> 6) & 0b11

	// 2. Decode the source registry (when bit 7 of first byte is 0, reg is the source)
	// Destination is in in second byte.
	regField := (instruction[1] >> 3) & 0b0000_0111

	// 3. Decode the destination registry.
	rmField := instruction[1] & 0b0000_0111

	builder.WriteString("mov ")

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

// Register/memory to/from register
// [1 1 0 0 0 1 1 w]
// [mod(2 bits) 0 0 0 rm(3 bits)]
// [Displacement Low (8 bits)]
// [Displacement Hight (8 bits)]
// [Data]
// [Data(if w = 1)]
func decodeMovImmediateToRegisterMemory(instruction []byte) string {
	var builder strings.Builder // Zero value is ready to use

	// The bit 8 of first byte determine the w field:
	// when 0, instruction operates on byte data
	// when 1, instructions operate on word data
	w := instruction[0]&0b1 == 1

	modField := (instruction[1] >> 6) & 0b11

	// 2. Decode the source registry (when bit 7 of first byte is 0, reg is the source)
	// Destination is in in second byte.
	// regField := (instruction[1] >> 3) & 0b0000_0111

	// 3. Decode the destination registry.
	rmField := instruction[1] & 0b0000_0111

	builder.WriteString("mov ")

	switch modField {
	case 0b00:
		// Memory only, no displacement follows
		// except when rmField = 110
		if rmField == 0b110 {
			builder.WriteString("direct address 2")
		} else {
			builder.WriteString(
				fmt.Sprintf("[%v]", effectiveAddressCalculation(rmField)),
			)
		}

		if w {
			data := uint16(instruction[2]) | uint16(instruction[3])<<8
			builder.WriteString(
				fmt.Sprintf(", word %v", data),
			)
		} else {
			data := instruction[2]
			builder.WriteString(
				fmt.Sprintf(", byte %v", data),
			)
		}
	case 0b01:
		// Memory mode, 8-bit displacement follows.
		displacement := int8(instruction[2])
		builder.WriteString(
			fmt.Sprintf("[%v + %v]", effectiveAddressCalculation(rmField), displacement),
		)

		if w {
			data := uint16(instruction[3]) | uint16(instruction[4])<<8
			builder.WriteString(
				fmt.Sprintf(", word %v", data),
			)
		} else {
			data := instruction[3]
			builder.WriteString(
				fmt.Sprintf(", byte %v", data),
			)
		}
	case 0b10:
		// Memory mode, 16-bit displacement follows
		displacement := uint16(instruction[2]) | uint16(instruction[3])<<8
		builder.WriteString(
			fmt.Sprintf("[%v + %v]", effectiveAddressCalculation(rmField), displacement),
		)

		if w {
			data := uint16(instruction[4]) | uint16(instruction[5])<<8
			builder.WriteString(
				fmt.Sprintf(", word %v", data),
			)
		} else {
			data := instruction[4]
			builder.WriteString(
				fmt.Sprintf(", byte %v", data),
			)
		}
	}

	return builder.String()
}

// Immediate to register.
// [1 0 1 1 w reg(3 bits)]
// [data(8 bits)]
// [data(8 bits - if w = 1)]
func decodeMovImmediateToRegister(instruction []byte) string {
	var builder strings.Builder // Zero value is ready to use

	// The bit 8 of first byte determine the isWord field:
	// when 0, instruction operates on byte data
	// when 1, instructions operate on word data
	isWord := (instruction[0]>>3)&0b1 == 1
	// 2. Decode the source registry (when bit 7 of first byte is 0, reg is the source)
	// Destination is in in second byte.
	regField := instruction[0] & 0b0000_0111

	var data uint16 = uint16(instruction[1])
	// 16-bit immediate-to-register
	if isWord {
		data = data | (uint16(instruction[2]) << 8)
	}

	builder.WriteString("mov ")
	builder.WriteString(
		fmt.Sprintf(
			"%v, %v",
			byteToRegisterString(isWord, regField),
			data,
		),
	)

	return builder.String()
}

// [1 0 1 0 0 0 0 w]
// [address low]
// [address high]
func decodeMovMemoryToAccumulator(instruction []byte) string {
	var builder strings.Builder // Zero value is ready to use

	// The bit 8 of first byte determine the isWord field:
	// when 0, instruction operates on byte data
	// when 1, instructions operate on word data
	isWord := instruction[0]&0b1 == 1

	builder.WriteString("mov ax, ")
	if isWord {
		address := uint16(instruction[1]) | uint16(instruction[2])<<8
		builder.WriteString(
			fmt.Sprintf("[%v]", address),
		)
	} else {
		address := instruction[1]
		builder.WriteString(
			fmt.Sprintf("[%v]", address),
		)
	}

	return builder.String()
}

// [1 0 1 0 0 0 1 w]
// [address low]
// [address high]
func decodeMovAccumulatorToMemory(instruction []byte) string {
	var builder strings.Builder // Zero value is ready to use

	// The bit 8 of first byte determine the isWord field:
	// when 0, instruction operates on byte data
	// when 1, instructions operate on word data
	isWord := instruction[0]&0b1 == 1

	builder.WriteString("mov ")
	if isWord {
		address := uint16(instruction[1]) | uint16(instruction[2])<<8
		builder.WriteString(
			fmt.Sprintf("[%v]", address),
		)
	} else {
		address := instruction[1]
		builder.WriteString(
			fmt.Sprintf("[%v]", address),
		)
	}
	builder.WriteString(", ax")

	return builder.String()
}
