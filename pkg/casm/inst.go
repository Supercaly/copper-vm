package casm

import "github.com/Supercaly/coppervm/pkg/coppervm"

var instDefs = [coppervm.InstCount]instruction{
	{
		kind:       coppervm.InstNoop,
		hasOperand: false,
		name:       "noop",
	},
	{
		kind:       coppervm.InstPush,
		hasOperand: true,
		name:       "push",
	},
	{
		kind:       coppervm.InstSwap,
		hasOperand: true,
		name:       "swap",
	},
	{
		kind:       coppervm.InstDup,
		hasOperand: false,
		name:       "dup",
	},
	{
		kind:       coppervm.InstOver,
		hasOperand: true,
		name:       "over",
	},
	{
		kind:       coppervm.InstDrop,
		hasOperand: false,
		name:       "drop",
	},
	{
		kind:       coppervm.InstAddInt,
		hasOperand: false,
		name:       "add",
	},
	{
		kind:       coppervm.InstSubInt,
		hasOperand: false,
		name:       "sub",
	},
	{
		kind:       coppervm.InstMulInt,
		hasOperand: false,
		name:       "mul",
	},
	{
		kind:       coppervm.InstMulIntSigned,
		hasOperand: false,
		name:       "imul",
	},
	{
		kind:       coppervm.InstDivInt,
		hasOperand: false,
		name:       "div",
	},
	{
		kind:       coppervm.InstDivIntSigned,
		hasOperand: false,
		name:       "idiv",
	},
	{
		kind:       coppervm.InstModInt,
		hasOperand: false,
		name:       "mod",
	},
	{
		kind:       coppervm.InstModIntSigned,
		hasOperand: false,
		name:       "imod",
	},
	{
		kind:       coppervm.InstAddFloat,
		hasOperand: false,
		name:       "fadd",
	},
	{
		kind:       coppervm.InstSubFloat,
		hasOperand: false,
		name:       "fsub",
	},
	{
		kind:       coppervm.InstMulFloat,
		hasOperand: false,
		name:       "fmul",
	},
	{
		kind:       coppervm.InstDivFloat,
		hasOperand: false,
		name:       "fdiv",
	},
	{
		kind:       coppervm.InstAnd,
		hasOperand: false,
		name:       "and",
	},
	{
		kind:       coppervm.InstOr,
		hasOperand: false,
		name:       "or",
	},
	{
		kind:       coppervm.InstXor,
		hasOperand: false,
		name:       "xor",
	},
	{
		kind:       coppervm.InstShiftLeft,
		hasOperand: false,
		name:       "shl",
	},
	{
		kind:       coppervm.InstShiftRight,
		hasOperand: false,
		name:       "shr",
	},
	{
		kind:       coppervm.InstNot,
		hasOperand: false,
		name:       "not",
	},
	{
		kind:       coppervm.InstCmp,
		hasOperand: false,
		name:       "cmp",
	},
	{
		kind:       coppervm.InstCmpSigned,
		hasOperand: false,
		name:       "icmp",
	},
	{
		kind:       coppervm.InstCmpFloat,
		hasOperand: false,
		name:       "fcmp",
	},
	{
		kind:       coppervm.InstJmp,
		hasOperand: true,
		name:       "jmp",
	},
	{
		kind:       coppervm.InstJmpZero,
		hasOperand: true,
		name:       "jz",
	},
	{
		kind:       coppervm.InstJmpNotZero,
		hasOperand: true,
		name:       "jnz",
	},
	{
		kind:       coppervm.InstJmpGreater,
		hasOperand: true,
		name:       "jg",
	},
	{
		kind:       coppervm.InstJmpLess,
		hasOperand: true,
		name:       "jl",
	},
	{
		kind:       coppervm.InstJmpGreaterEqual,
		hasOperand: true,
		name:       "jge",
	},
	{
		kind:       coppervm.InstJmpLessEqual,
		hasOperand: true,
		name:       "jle",
	},

	{
		kind:       coppervm.InstFunCall,
		hasOperand: true,
		name:       "call",
	},
	{
		kind:       coppervm.InstFunReturn,
		hasOperand: false,
		name:       "ret",
	},
	{
		kind:       coppervm.InstMemRead,
		hasOperand: false,
		name:       "read",
	},
	{
		kind:       coppervm.InstMemReadInt,
		hasOperand: false,
		name:       "iread",
	},
	{
		kind:       coppervm.InstMemReadFloat,
		hasOperand: false,
		name:       "fread",
	},
	{
		kind:       coppervm.InstMemWrite,
		hasOperand: false,
		name:       "write",
	}, {
		kind:       coppervm.InstMemWriteInt,
		hasOperand: false,
		name:       "iwrite",
	},
	{
		kind:       coppervm.InstMemWriteFloat,
		hasOperand: false,
		name:       "fwrite",
	},
	{
		kind:       coppervm.InstSyscall,
		hasOperand: true,
		name:       "syscall",
	},
	{
		kind:       coppervm.InstPrint,
		hasOperand: false,
		name:       "print",
	},
	{
		kind:       coppervm.InstHalt,
		hasOperand: false,
		name:       "halt",
	},
}

type instruction struct {
	kind       coppervm.InstKind
	hasOperand bool
	name       string
	operand    word
}

// Return an instruction definition by it's string
// representation.
// This function return true if the instruction exist,
// false otherwise.
func getInstructionByName(name string) (bool, instruction) {
	for _, inst := range instDefs {
		if inst.name == name {
			return true, inst
		}
	}
	return false, instruction{}
}
