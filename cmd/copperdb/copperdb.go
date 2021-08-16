package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Supercaly/coppervm/pkg/copperdb"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

func usage(stream io.Writer, program string) {
	fmt.Fprintf(stream, "Usage: %s <input.copper>\n", program)
}

func main() {
	if len(os.Args) != 2 {
		usage(os.Stderr, os.Args[0])
		log.Fatalf("[ERROR]: input was not provided\n")
	}

	inputFilePath := os.Args[1]

	vm := coppervm.Coppervm{}
	meta, err := vm.LoadProgramFromFile(inputFilePath)
	if err != nil {
		log.Fatalf("[ERROR]: %s", err)
	}
	vm.Halt = true
	db := copperdb.Copperdb{
		InputFile:    inputFilePath,
		Vm:           &vm,
		DebugSymbols: meta.DebugSymbols,
	}
	db.StartProgramDebug()
}
