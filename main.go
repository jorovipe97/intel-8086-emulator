package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	deco "github.com/jorovipe97/performance-aware-homework/decoder"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("The filename arg is required.")
	}

	fileName := os.Args[1]

	// Get the working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(wd, "listings", fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%08b\n", data)

	asmPrinter := deco.NewAsmPrinter()
	asmPrinter.AddComment(fileName)
	asmPrinter.Bits16Header()

	decoder := deco.Decoder{
		Data: data,
	}
	for {
		if !decoder.HasNext() {
			break
		}

		instr, err := decoder.NextInstruction()
		if err != nil {
			fmt.Println(err)
			asmPrinter.AddComment(fmt.Sprintf("ERROR: %v", err))
			break
		}

		if instr.Op != deco.OpNone {
			asmPrinter.AddInstruction(instr)
		} else {
			asmPrinter.AddComment("ERROR: Unrecognized binary in instruction stream.")
			break
		}
	}

	// Shows assembly code:
	asmString := asmPrinter.String()
	fmt.Println(asmString)

	// Saves the final assembly into disk.
	newAsmFile := filepath.Join(wd, "result.asm")
	err = os.WriteFile(newAsmFile, []byte(asmString), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
