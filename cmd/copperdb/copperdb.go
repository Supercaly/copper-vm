package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Supercaly/coppervm/pkg/copperdb"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

func main() {
	input := "examples/bin/123.vm"
	vm := coppervm.Coppervm{}
	vm.LoadProgramFromFile(input)
	vm.Halt = true
	db := copperdb.Copperdb{
		InputFile: input,
		Vm:        &vm,
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("(coppervm) ")
		str, err := reader.ReadString('\n')
		str = strings.TrimSuffix(str, "\n")
		str = strings.TrimSpace(str)

		if err != nil {
			log.Fatalf("Something went wrong with the debugger: %s", err)
		}
		db.ExecuteInputString(str)
	}
}
