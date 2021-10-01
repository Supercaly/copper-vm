package coppervm

import (
	"fmt"
)

type InstKind int

const (
	// TODO(#9): Add more instructions
	InstNoop InstKind = iota

	// Basic instructions
	InstPush
	InstSwap
	InstDup
	InstOver
	InstDrop
	InstHalt

	// Integer arithmetics
	InstAddInt
	InstSubInt
	InstMulInt
	InstMulIntSigned
	InstDivInt
	InstDivIntSigned
	InstModInt
	InstModIntSigned

	// Floating point arithmetics
	InstAddFloat
	InstSubFloat
	InstMulFloat
	InstDivFloat

	// Boolean operations
	InstAnd
	InstOr
	InstXor
	InstNot
	InstShiftLeft
	InstShiftRight

	// Flow control
	InstCmp
	InstCmpSigned
	InstCmpFloat
	InstJmp
	InstJmpZero
	InstJmpNotZero
	InstJmpGreater
	InstJmpGreaterEqual
	InstJmpLess
	InstJmpLessEqual

	// Functions
	InstFunCall
	InstFunReturn

	// Memory access
	InstMemRead
	InstMemReadInt
	InstMemReadFloat
	InstMemWrite
	InstMemWriteInt
	InstMemWriteFloat

	// Syscall
	InstSyscall

	InstPrint

	InstCount
)

type InstDef struct {
	Kind       InstKind
	HasOperand bool
	Name       string
	Operand    Word
}

func (inst InstDef) String() (out string) {
	out += fmt.Sprint(inst.Name)
	if inst.HasOperand {
		out += fmt.Sprintf(" (%s)", inst.Operand)
	}
	return out
}
