package coppervm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	CoppervmDebug         bool  = true
	CoppervmStackCapacity int64 = 1024
)

type Coppervm struct {
	Stack     [CoppervmStackCapacity]Word
	StackSize int64
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
	// Basic instructions
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
		if currentInst.Operand.AsI64 >= vm.StackSize {
			return ErrorStackUnderflow
		}
		a := vm.StackSize - 1
		b := vm.StackSize - 1 - currentInst.Operand.AsI64
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
	case InstHalt:
		vm.Halt = true
	// Integer arithmetics
	case InstAddInt:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = addWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepU64)
		vm.StackSize--
		vm.Ip++
	case InstSubInt:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = subWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepU64)
		vm.StackSize--
		vm.Ip++
	case InstMulInt:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = mulWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepU64)
		vm.StackSize--
		vm.Ip++
	case InstMulIntSigned:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = mulWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepI64)
		vm.StackSize--
		vm.Ip++
	// Floating point arithmetics
	case InstAddFloat:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = addWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepF64)
		vm.StackSize--
		vm.Ip++
	case InstSubFloat:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = subWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepF64)
		vm.StackSize--
		vm.Ip++
	case InstMulFloat:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		vm.Stack[vm.StackSize-2] = mulWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepF64)
		vm.StackSize--
		vm.Ip++
	// Flow control
	case InstJmp:
		vm.Ip = uint(currentInst.Operand.AsI64)
	case InstJmpNotZero:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsI64 != 0 {
			vm.Ip = uint(currentInst.Operand.AsI64)
		} else {
			vm.Ip++
		}
		vm.StackSize--
	case InstPrint:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		fmt.Printf("[write]: %s\n", vm.Stack[vm.StackSize-1])
		vm.Ip++
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
		for i := int64(0); i < vm.StackSize; i++ {
			fmt.Printf("  %s\n", vm.Stack[i])
		}
	} else {
		fmt.Printf("  [empty]\n")
	}
}
