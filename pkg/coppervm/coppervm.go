package coppervm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	CoppervmDebug         bool = true
	CoppervmStackCapacity int  = 1024
)

type Coppervm struct {
	Stack     [CoppervmStackCapacity]int
	StackSize int
	Program   []InstDef
	Ip        uint
	Halt      bool
}

// Load program's binary to vm from file.
func (vm *Coppervm) LoadProgramFromFile(filePath string) {
	content, fileErr := ioutil.ReadFile(filePath)
	if fileErr != nil {
		log.Fatalf("[ERROR]: Error reading file '%s': %s",
			filePath,
			fileErr)
	}

	var meta CoppervmFileMeta
	if err := json.Unmarshal(content, &meta); err != nil {
		log.Fatalf("[ERROR]: Error reading content of file '%s': %s",
			filePath,
			err)
	}

	vm.Halt = false
	vm.Ip = uint(meta.Entry)
	vm.Program = meta.Program
}

// Executes all the program of the vm.
// Return a CoppervmError if something went wrong or ErrorOk.
func (vm *Coppervm) ExecuteProgram(limit int) CoppervmError {
	for limit != 0 && !vm.Halt {
		if err := vm.ExecuteInstruction(); err != ErrorOk {
			return err
		}
		limit--
	}
	return ErrorOk
}

// Executes a single instruction of the program where the
// current ip points and then increments the ip.
// Return a CoppervmError if something went wrong or ErrorOk.
func (vm *Coppervm) ExecuteInstruction() CoppervmError {
	if vm.Ip >= uint(len(vm.Program)) {
		return ErrorIllegalInstAccess
	}

	currentInst := vm.Program[vm.Ip]
	switch currentInst.Kind {
	case InstNoop:
		vm.Ip++
	case InstPush:
		if vm.StackSize >= CoppervmStackCapacity {
			return ErrorStackOverflow
		}
		vm.Stack[vm.StackSize] = currentInst.Operand
		vm.StackSize++
		vm.Ip++
	case InstSwap:
		if currentInst.Operand >= vm.StackSize {
			return ErrorStackUnderflow
		}
		a := vm.StackSize - 1
		b := vm.StackSize - 1 - currentInst.Operand
		tmp := vm.Stack[a]
		vm.Stack[a] = vm.Stack[b]
		vm.Stack[b] = tmp
		vm.Ip++
	case InstDup:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		if vm.StackSize >= CoppervmStackCapacity {
			return ErrorStackOverflow
		}
		newVal := vm.Stack[vm.StackSize-1]
		vm.Stack[vm.StackSize] = newVal
		vm.StackSize++
		vm.Ip++
	case InstAddInt:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = vm.Stack[vm.StackSize-2] + vm.Stack[vm.StackSize-1]
		vm.StackSize--
		vm.Ip++
	case InstSubInt:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = vm.Stack[vm.StackSize-2] - vm.Stack[vm.StackSize-1]
		vm.StackSize--
		vm.Ip++
	case InstMulInt:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = vm.Stack[vm.StackSize-2] * vm.Stack[vm.StackSize-1]
		vm.StackSize--
		vm.Ip++
	case InstAddFloat:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = vm.Stack[vm.StackSize-2] + vm.Stack[vm.StackSize-1]
		vm.StackSize--
		vm.Ip++
	case InstSubFloat:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = vm.Stack[vm.StackSize-2] - vm.Stack[vm.StackSize-1]
		vm.StackSize--
		vm.Ip++
	case InstMulFloat:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = vm.Stack[vm.StackSize-2] * vm.Stack[vm.StackSize-1]
		vm.StackSize--
		vm.Ip++
	case InstJmp:
		vm.Ip = uint(currentInst.Operand)
	case InstJmpNotZero:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1] != 0 {
			vm.Ip = uint(currentInst.Operand)
		} else {
			vm.Ip++
		}
		vm.StackSize--
	case InstPrint:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		fmt.Printf("[write]: %d\n", vm.Stack[vm.StackSize-1])
		vm.Ip++
	case InstHalt:
		vm.Halt = true
	case InstCount:
		fallthrough
	default:
		log.Fatalf("Invalid instruction %s", currentInst.Name)
	}

	if CoppervmDebug {
		vm.dumpStack()
	}

	return ErrorOk
}

// Prints the stack content to standard output.
func (vm *Coppervm) dumpStack() {
	fmt.Printf("Stack:\n")
	if vm.StackSize > 0 {
		for i := 0; i < vm.StackSize; i++ {
			fmt.Printf("  int: %d\n", vm.Stack[i])
		}
	} else {
		fmt.Printf("  [empty]\n")
	}
}
