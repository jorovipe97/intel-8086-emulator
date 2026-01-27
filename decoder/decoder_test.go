package decoder

import (
	"strings"
	"testing"
)

type testCase struct {
	InputBinaryStream []byte
	OutpuAsmString    string
}

var cases = [...]testCase{
	// Register to register
	{
		InputBinaryStream: []byte{0b10001001, 0b11011110},
		OutpuAsmString:    "mov si, bx",
	},
	// Immediate 8-bits
	{
		InputBinaryStream: []byte{0b10110001, 0b00001100},
		OutpuAsmString:    "mov cl, byte 12",
	},
	{
		InputBinaryStream: []byte{0b10110101, 0b11110100},
		OutpuAsmString:    "mov ch, byte -12",
	},
	// 16-bit immediate-to-register
	{
		InputBinaryStream: []byte{0b10111010, 0b10010100, 0b11110000},
		OutpuAsmString:    "mov dx, word -3948",
	},
	// Source address calculation
	{
		InputBinaryStream: []byte{0b10001010, 0b00000000},
		OutpuAsmString:    "mov al, [bx+si]",
	},
	{
		InputBinaryStream: []byte{0b10001011, 0b01010110, 0b00000000},
		OutpuAsmString:    "mov dx, [bp]",
	},
	// Source address calculation plus 8-bit displacement
	{
		InputBinaryStream: []byte{0b10001010, 0b01100000, 0b00000100},
		OutpuAsmString:    "mov ah, [bx+si+4]",
	},
	// Source address calculation plus 16-bit displacement
	{
		InputBinaryStream: []byte{0b10001010, 0b10000000, 0b10000111, 0b00010011},
		OutpuAsmString:    "mov al, [bx+si+4999]",
	},
	// ; Dest address calculation
	{
		InputBinaryStream: []byte{0b10001001, 0b00001001},
		OutpuAsmString:    "mov [bx+di], cx",
	},
	{
		InputBinaryStream: []byte{0b10001000, 0b00001010},
		OutpuAsmString:    "mov [bp+si], cl",
	},
	{
		InputBinaryStream: []byte{0b10001000, 0b01101110, 0b00000000},
		OutpuAsmString:    "mov [bp], ch",
	},
	// Signed displacements
	{
		InputBinaryStream: []byte{0b10001011, 0b01000001, 0b11011011},
		OutpuAsmString:    "mov ax, [bx+di-37]",
	},
	{
		InputBinaryStream: []byte{0b10001001, 0b10001100, 0b11010100, 0b11111110},
		OutpuAsmString:    "mov [si-300], cx",
	},
	{
		InputBinaryStream: []byte{0b10001011, 0b01010111, 0b11100000},
		OutpuAsmString:    "mov dx, [bx-32]",
	},
	// Explicit sizes
	{
		InputBinaryStream: []byte{0b11000110, 0b00000011, 0b00000111},
		OutpuAsmString:    "mov [bp+di], byte 7",
	},
	{
		InputBinaryStream: []byte{0b11000111, 0b10000101, 0b10000101, 0b00000011, 0b01011011, 0b00000001},
		OutpuAsmString:    "mov [di+901], word 347",
	},
	// Direct address
	{
		InputBinaryStream: []byte{0b10001011, 0b00101110, 0b00000101, 0b00000000},
		OutpuAsmString:    "mov bp, [+5]",
	},
	{
		InputBinaryStream: []byte{0b10001011, 0b00011110, 0b10000010, 0b00001101},
		OutpuAsmString:    "mov bx, [+3458]",
	},
	// Memory-to-accumulator test
	{
		InputBinaryStream: []byte{0b10100001, 0b11111011, 0b00001001},
		OutpuAsmString:    "mov ax, [+2555]",
	},
	{
		InputBinaryStream: []byte{0b10100001, 0b00010000, 0b00000000},
		OutpuAsmString:    "mov ax, [+16]",
	},
	// Accumulator-to-memory test
	{
		InputBinaryStream: []byte{0b10100011, 0b11111010, 0b00001001},
		OutpuAsmString:    "mov [+2554], ax",
	},
	{
		InputBinaryStream: []byte{0b10100011, 0b00001111, 0b00000000},
		OutpuAsmString:    "mov [+15], ax",
	},
}

func TestDisasembling(t *testing.T) {
	for _, singleCase := range cases {
		deco := Decoder{
			Data: singleCase.InputBinaryStream,
		}
		disAsm := NewAsmPrinter()

		instr, err := deco.NextInstruction()
		if err != nil {
			t.Fatal("Could not decode instruction")
		}

		disAsm.AddInstruction(instr)
		out := strings.Trim(disAsm.String(), "\n")

		if out != singleCase.OutpuAsmString {
			t.Errorf("Expected: '%v', Got: '%v'", singleCase.OutpuAsmString, out)
		}
	}
}
