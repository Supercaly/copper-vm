package coppervm

import "fmt"

type CoppervmError struct {
	Kind        CoppervmErrorKind
	CurrentIp   InstAddr
	CurrentInst InstDef
}

func ErrorOk(vm *Coppervm) *CoppervmError {
	return &CoppervmError{Kind: ErrorKindOk}
}

func ErrorIllegalInstAccess(vm *Coppervm) *CoppervmError {
	return &CoppervmError{
		Kind:        ErrorKindIllegalInstAccess,
		CurrentIp:   vm.Ip,
		CurrentInst: vm.Program[vm.Ip],
	}
}

func ErrorStackOverflow(vm *Coppervm) *CoppervmError {
	return &CoppervmError{
		Kind:        ErrorKindStackOverflow,
		CurrentIp:   vm.Ip,
		CurrentInst: vm.Program[vm.Ip],
	}
}

func ErrorStackUnderflow(vm *Coppervm) *CoppervmError {
	return &CoppervmError{
		Kind:        ErrorKindStackUnderflow,
		CurrentIp:   vm.Ip,
		CurrentInst: vm.Program[vm.Ip],
	}
}

func ErrorDivideByZero(vm *Coppervm) *CoppervmError {
	return &CoppervmError{
		Kind:        ErrorKindDivideByZero,
		CurrentIp:   vm.Ip,
		CurrentInst: vm.Program[vm.Ip],
	}
}

func ErrorIllegalMemoryAccess(vm *Coppervm) *CoppervmError {
	return &CoppervmError{
		Kind:        ErrorKindIllegalMemoryAccess,
		CurrentIp:   vm.Ip,
		CurrentInst: vm.Program[vm.Ip],
	}
}

func ErrorInvalidInstruction(vm *Coppervm) *CoppervmError {
	return &CoppervmError{
		Kind:        ErrorKindInvalidInstruction,
		CurrentIp:   vm.Ip,
		CurrentInst: vm.Program[vm.Ip],
	}
}

func (err CoppervmError) String() string {
	return fmt.Sprintf("'%s' executing instruction '%s' at ip '%d'",
		err.Kind,
		err.CurrentInst,
		err.CurrentIp)
}

type CoppervmErrorKind int

const (
	ErrorKindOk CoppervmErrorKind = iota
	ErrorKindIllegalInstAccess
	ErrorKindStackOverflow
	ErrorKindStackUnderflow
	ErrorKindDivideByZero
	ErrorKindIllegalMemoryAccess
	ErrorKindInvalidInstruction
)

func (err CoppervmErrorKind) String() string {
	return [...]string{
		"ErrorOk",
		"ErrorIllegalInstAccess",
		"ErrorStackOverflow",
		"ErrorStackUnderflow",
		"ErrorDivideByZero",
		"ErrorIllegalMemoryAccess",
		"ErrorKindInvalidInstruction",
	}[err]
}
