package main

import (
	"log"

	"coppervm.com/coppervm/pkg/coppervm"
)

func main() {
	vm := coppervm.Coppervm{}
	vm.LoadProgramFromFile("examples/bin/123.vm")
	if err := vm.ExecuteProgram(10); err != coppervm.ErrorOk {
		log.Fatalf("[ERROR]: %s", err)
	}
}
