package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	deco "github.com/jorovipe97/performance-aware-homework/decoder"
)

// 8086 nice tutorial: https://yassinebridi.github.io/asm-docs/
// 8086 cool simulator: https://yjdoc2.github.io/8086-emulator-web/compile
func main() {
	if len(os.Args) < 2 {
		log.Fatal("The filename arg is required.")
	}

	fileName := os.Args[1]
	var shouldSimulate bool
	if len(os.Args) == 3 {
		shouldSimulate = os.Args[2] == "--simulate"
	}

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
	memory := deco.NewMemory(data)
	decoder := deco.NewDecoder(memory)
	simulator := deco.NewSimulator(memory, asmPrinter)

	asmPrinter.AddComment(fileName)
	asmPrinter.Bits16Header()

	var disasmErr error
	// Dis-assembler loop, we iterate over the whole memory
	// ignoring jumps to get all the instructions in the binary
	for {
		if !memory.HasMoreInstructions() {
			break
		}

		var instr deco.Instruction
		instr, disasmErr = decoder.NextInstruction()
		if disasmErr != nil {
			asmPrinter.AddComment(fmt.Sprintf("ERROR: %v", err))
			break
		}

		if instr.Op != deco.OpNone {
			asmPrinter.AddInstruction(instr)
		} else {
			disasmErr = errors.New("unrecognized binary in instruction stream")
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

	// If disassembly failed, then do not simulate
	if disasmErr != nil {
		fmt.Println(disasmErr)
		return
	}

	// Skip simulation if --simulate flag is not passed.
	if !shouldSimulate {
		return
	}

	// Simulator loop, we execute the instructions, including jumps
	// Reset memory position so we start executing instructions from the beginning
	memory.ResetAbsolutePosition()
	for {
		instr, err := decoder.NextInstruction()
		if err != nil {
			asmPrinter.AddComment(fmt.Sprintf("ERROR: %v", err))
			break
		}

		if instr.Op != deco.OpNone {
			err := simulator.ExecInstruction(instr)
			if err != nil {
				asmPrinter.AddComment("ERROR: Error simulating instruction.")
			}
		} else {
			break
		}
	}

	fmt.Println(simulator.String())
}
