package coppervm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	CoppervmDebug          bool  = false
	CoppervmStackCapacity  int64 = 1024
	CoppervmMemoryCapacity int64 = 1024
)

type InstAddr uint64

type Coppervm struct {
	// VM Stack
	Stack     [CoppervmStackCapacity]Word
	StackSize int64

	// VM Program
	Program []InstDef
	Ip      InstAddr

	// VM Memory
	Memory     [CoppervmMemoryCapacity]byte
	MemorySize int64

	// Is the VM halted?
	Halt bool
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
	vm.Ip = InstAddr(meta.Entry)
	vm.Program = meta.Program

	if len(meta.Memory) > int(CoppervmMemoryCapacity) {
		log.Fatalf("[ERROR]: Memory exceed the maximum memory capacity!")
	}
	for i := 0; i < len(meta.Memory); i++ {
		vm.Memory[i] = meta.Memory[i]
	}
	vm.MemorySize = int64(len(meta.Memory))
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
	if vm.Ip >= InstAddr(len(vm.Program)) {
		return ErrorIllegalInstAccess
	}

	currentInst := vm.Program[vm.Ip]
	switch currentInst.Kind {
	// Basic instructions
	case InstNoop:
		vm.Ip++
	case InstPush:
		if err := vm.pushStack(currentInst.Operand); err != ErrorOk {
			return err
		}
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
		newVal := vm.Stack[vm.StackSize-1]
		if err := vm.pushStack(newVal); err != ErrorOk {
			return err
		}
		vm.Ip++
	case InstDrop:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		vm.StackSize--
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
	case InstDivInt:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsU64 == 0 {
			return ErrorDivideByZero
		}
		vm.Stack[vm.StackSize-2] = divWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepU64)
		vm.StackSize--
		vm.Ip++
	case InstDivIntSigned:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsI64 == 0 {
			return ErrorDivideByZero
		}
		vm.Stack[vm.StackSize-2] = divWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepI64)
		vm.StackSize--
		vm.Ip++
	case InstModInt:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsU64 == 0 {
			return ErrorDivideByZero
		}
		vm.Stack[vm.StackSize-2] = modWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepU64)
		vm.StackSize--
		vm.Ip++
	case InstModIntSigned:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsI64 == 0 {
			return ErrorDivideByZero
		}
		vm.Stack[vm.StackSize-2] = modWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepI64)
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
	case InstDivFloat:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsF64 == 0 {
			return ErrorDivideByZero
		}
		vm.Stack[vm.StackSize-2] = divWord(vm.Stack[vm.StackSize-2], vm.Stack[vm.StackSize-1], typeRepF64)
		vm.StackSize--
		vm.Ip++
	// Flow control
	case InstCmp:
		if vm.StackSize < 2 {
			return ErrorStackUnderflow
		}
		a := vm.Stack[vm.StackSize-2]
		b := vm.Stack[vm.StackSize-1]
		var res Word
		if a.AsI64 == b.AsI64 {
			res = WordI64(0)
		} else if a.AsI64 > b.AsI64 {
			res = WordI64(1)
		} else {
			res = WordI64(-1)
		}
		vm.Stack[vm.StackSize-2] = res
		vm.StackSize--
		vm.Ip++
	case InstJmp:
		vm.Ip = InstAddr(currentInst.Operand.AsI64)
	case InstJmpZero:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsI64 == 0 {
			vm.Ip = InstAddr(currentInst.Operand.AsI64)
		} else {
			vm.Ip++
		}
		vm.StackSize--
	case InstJmpNotZero:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsI64 != 0 {
			vm.Ip = InstAddr(currentInst.Operand.AsI64)
		} else {
			vm.Ip++
		}
		vm.StackSize--
	case InstJmpGreater:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsI64 > 0 {
			vm.Ip = InstAddr(currentInst.Operand.AsI64)
		} else {
			vm.Ip++
		}
		vm.StackSize--
	case InstJmpLess:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsI64 < 0 {
			vm.Ip = InstAddr(currentInst.Operand.AsI64)
		} else {
			vm.Ip++
		}
		vm.StackSize--
	case InstJmpGreaterEqual:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsI64 >= 0 {
			vm.Ip = InstAddr(currentInst.Operand.AsI64)
		} else {
			vm.Ip++
		}
		vm.StackSize--
	case InstJmpLessEqual:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		if vm.Stack[vm.StackSize-1].AsI64 <= 0 {
			vm.Ip = InstAddr(currentInst.Operand.AsI64)
		} else {
			vm.Ip++
		}
		vm.StackSize--
	// Functions
	case InstFunCall:
		if err := vm.pushStack(WordU64(uint64(vm.Ip + 1))); err != ErrorOk {
			return err
		}
		vm.Ip = InstAddr(currentInst.Operand.AsU64)
	case InstFunReturn:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		retAdds := vm.Stack[vm.StackSize-1]
		vm.StackSize--
		vm.Ip = InstAddr(retAdds.AsU64)
	// Memory Access
	case InstMemRead:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		addr := vm.Stack[vm.StackSize-1].AsU64
		if addr >= uint64(CoppervmMemoryCapacity) {
			return ErrorIllegalMemoryAccess
		}
		vm.Stack[vm.StackSize-1] = WordU64(uint64(vm.Memory[addr]))
		vm.Ip++
	case InstPrint:
		if vm.StackSize < 1 {
			return ErrorStackUnderflow
		}
		fmt.Printf("%s", string(rune(vm.Stack[vm.StackSize-1].AsU64)))
		vm.StackSize--
		vm.Ip++
	case InstCount:
		fallthrough
	default:
		log.Fatalf("Invalid instruction %s", currentInst.Name)
	}

	// Print stack on debug
	if CoppervmDebug {
		vm.dumpStack()
	}

	return ErrorOk
}

// Push a Word to the stack.
// Return a ErrorStackOverflow if the stack overflows, or
// ErrorOk otherwise.
func (vm *Coppervm) pushStack(w Word) CoppervmError {
	if vm.StackSize >= CoppervmStackCapacity {
		return ErrorStackOverflow
	}
	vm.Stack[vm.StackSize] = w
	vm.StackSize++
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
