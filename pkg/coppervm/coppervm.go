package coppervm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type InstKind int

const (
	InstNoop InstKind = iota
	InstPush
	InstAdd
	InstHalt
	InstCount
)

type InstDef struct {
	Kind       InstKind
	HasOperand bool
	Name       string
	// TODO(#5): Use Word as operand type
	Operand int
}

var InstDefs = [InstCount]InstDef{
	{
		Kind:       InstNoop,
		HasOperand: false,
		Name:       "noop",
	},
	{
		Kind:       InstPush,
		HasOperand: true,
		Name:       "push",
	},
	{
		Kind:       InstAdd,
		HasOperand: false,
		Name:       "add",
	},
	{
		Kind:       InstHalt,
		HasOperand: false,
		Name:       "halt",
	},
}

func GetInstDefByName(name string) (bool, InstDef) {
	for _, inst := range InstDefs {
		if inst.Name == name {
			return true, inst
		}
	}
	return false, InstDef{}
}

type CoppervmError int

const (
	ErrorOk CoppervmError = iota
)

func (err CoppervmError) String() string {
	return [...]string{
		"ErrorOk",
	}[err]
}

const (
	CoppervmStackCapacity int = 1024
)

type Coppervm struct {
	Stack   [CoppervmStackCapacity]int
	Program []InstDef
	Ip      uint
	Halt    bool
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

func (vm *Coppervm) ExecuteProgram(limit int) CoppervmError {
	for limit != 0 && !vm.Halt {
		if err := vm.ExecuteInstruction(); err != ErrorOk {
			return err
		}
		limit--
	}
	return ErrorOk
}

func (vm *Coppervm) ExecuteInstruction() CoppervmError {
	fmt.Printf("Exec inst at %d\n", vm.Ip)
	vm.Ip++
	return ErrorOk
}

const (
	CoppervmFileVersion int = 1
)

type CoppervmFileMeta struct {
	Version int       `json:"version"`
	Entry   int       `json:"entry_point"`
	Program []InstDef `json:"program"`
}

func FileMeta(entryPoint int, program []InstDef) CoppervmFileMeta {
	return CoppervmFileMeta{
		Version: CoppervmFileVersion,
		Entry:   entryPoint,
		Program: program,
	}
}
