package casm

import "github.com/Supercaly/coppervm/pkg/coppervm"

var InstDefs = [coppervm.InstCount]Instruction{
	{
		Kind:       coppervm.InstNoop,
		HasOperand: false,
		Name:       "noop",
	},
	{
		Kind:       coppervm.InstPush,
		HasOperand: true,
		Name:       "push",
	},
	{
		Kind:       coppervm.InstSwap,
		HasOperand: true,
		Name:       "swap",
	},
	{
		Kind:       coppervm.InstDup,
		HasOperand: false,
		Name:       "dup",
	},
	{
		Kind:       coppervm.InstOver,
		HasOperand: true,
		Name:       "over",
	},
	{
		Kind:       coppervm.InstDrop,
		HasOperand: false,
		Name:       "drop",
	},
	{
		Kind:       coppervm.InstAddInt,
		HasOperand: false,
		Name:       "add",
	},
	{
		Kind:       coppervm.InstSubInt,
		HasOperand: false,
		Name:       "sub",
	},
	{
		Kind:       coppervm.InstMulInt,
		HasOperand: false,
		Name:       "mul",
	},
	{
		Kind:       coppervm.InstMulIntSigned,
		HasOperand: false,
		Name:       "imul",
	},
	{
		Kind:       coppervm.InstDivInt,
		HasOperand: false,
		Name:       "div",
	},
	{
		Kind:       coppervm.InstDivIntSigned,
		HasOperand: false,
		Name:       "idiv",
	},
	{
		Kind:       coppervm.InstModInt,
		HasOperand: false,
		Name:       "mod",
	},
	{
		Kind:       coppervm.InstModIntSigned,
		HasOperand: false,
		Name:       "imod",
	},
	{
		Kind:       coppervm.InstAddFloat,
		HasOperand: false,
		Name:       "fadd",
	},
	{
		Kind:       coppervm.InstSubFloat,
		HasOperand: false,
		Name:       "fsub",
	},
	{
		Kind:       coppervm.InstMulFloat,
		HasOperand: false,
		Name:       "fmul",
	},
	{
		Kind:       coppervm.InstDivFloat,
		HasOperand: false,
		Name:       "fdiv",
	},
	{
		Kind:       coppervm.InstAnd,
		HasOperand: false,
		Name:       "and",
	},
	{
		Kind:       coppervm.InstOr,
		HasOperand: false,
		Name:       "or",
	},
	{
		Kind:       coppervm.InstXor,
		HasOperand: false,
		Name:       "xor",
	},
	{
		Kind:       coppervm.InstShiftLeft,
		HasOperand: false,
		Name:       "shl",
	},
	{
		Kind:       coppervm.InstShiftRight,
		HasOperand: false,
		Name:       "shr",
	},
	{
		Kind:       coppervm.InstNot,
		HasOperand: false,
		Name:       "not",
	},
	{
		Kind:       coppervm.InstCmp,
		HasOperand: false,
		Name:       "cmp",
	},
	{
		Kind:       coppervm.InstCmpSigned,
		HasOperand: false,
		Name:       "icmp",
	},
	{
		Kind:       coppervm.InstCmpFloat,
		HasOperand: false,
		Name:       "fcmp",
	},
	{
		Kind:       coppervm.InstJmp,
		HasOperand: true,
		Name:       "jmp",
	},
	{
		Kind:       coppervm.InstJmpZero,
		HasOperand: true,
		Name:       "jz",
	},
	{
		Kind:       coppervm.InstJmpNotZero,
		HasOperand: true,
		Name:       "jnz",
	},
	{
		Kind:       coppervm.InstJmpGreater,
		HasOperand: true,
		Name:       "jg",
	},
	{
		Kind:       coppervm.InstJmpLess,
		HasOperand: true,
		Name:       "jl",
	},
	{
		Kind:       coppervm.InstJmpGreaterEqual,
		HasOperand: true,
		Name:       "jge",
	},
	{
		Kind:       coppervm.InstJmpLessEqual,
		HasOperand: true,
		Name:       "jle",
	},

	{
		Kind:       coppervm.InstFunCall,
		HasOperand: true,
		Name:       "call",
	},
	{
		Kind:       coppervm.InstFunReturn,
		HasOperand: false,
		Name:       "ret",
	},
	{
		Kind:       coppervm.InstMemRead,
		HasOperand: false,
		Name:       "read",
	},
	{
		Kind:       coppervm.InstMemReadInt,
		HasOperand: false,
		Name:       "iread",
	},
	{
		Kind:       coppervm.InstMemReadFloat,
		HasOperand: false,
		Name:       "fread",
	},
	{
		Kind:       coppervm.InstMemWrite,
		HasOperand: false,
		Name:       "write",
	}, {
		Kind:       coppervm.InstMemWriteInt,
		HasOperand: false,
		Name:       "iwrite",
	},
	{
		Kind:       coppervm.InstMemWriteFloat,
		HasOperand: false,
		Name:       "fwrite",
	},
	{
		Kind:       coppervm.InstSyscall,
		HasOperand: true,
		Name:       "syscall",
	},
	{
		Kind:       coppervm.InstPrint,
		HasOperand: false,
		Name:       "print",
	},
	{
		Kind:       coppervm.InstHalt,
		HasOperand: false,
		Name:       "halt",
	},
}

type Instruction struct {
	Kind       coppervm.InstKind
	HasOperand bool
	Name       string
	Operand    word
}

// Return an instruction definition by it's string
// representation.
// This function return true if the instruction exist,
// false otherwise.
func GetInstructionByName(name string) (bool, Instruction) {
	for _, inst := range InstDefs {
		if inst.Name == name {
			return true, inst
		}
	}
	return false, Instruction{}
}
