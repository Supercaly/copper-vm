package casm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type x86_64Generator struct {
	s string
}

func (x86_64Gen x86_64Generator) generateProgram(program []IR) {
	var entry string

	out := strings.Builder{}
	out.WriteString("section .text\n")
	out.WriteString("global _start\n")
	for _, inst := range program {
		switch inst.Kind {
		case IRKindLabel:
			out.WriteString(fmt.Sprintf("%s:\n", inst.AsLabel.Name))
		case IRKindInstruction:
			_, instDef := coppervm.GetInstDefByName(inst.AsInstruction.Name)
			switch instDef.Kind {
			// Basic instructions
			case coppervm.InstPush:
				out.WriteString("  ; -- push --\n")
				out.WriteString(fmt.Sprintf("  push %s\n", expressionToString(inst.AsInstruction.Operand)))
			case coppervm.InstSwap:
				out.WriteString("  ; -- swap --\n")
			case coppervm.InstDup:
				out.WriteString("  ; -- dup --\n")
				out.WriteString("  pop rax\n")
				out.WriteString("  push rax\n")
				out.WriteString("  push rax\n")
			case coppervm.InstDrop:
				out.WriteString("  ; -- drop --\n")
				out.WriteString("  pop rax\n")
			case coppervm.InstHalt:
				out.WriteString("  ; -- halt --\n")
				out.WriteString("  mov rax, 0x3c\n")
				out.WriteString("  mov rdi, 0x0\n")
				out.WriteString("  syscall\n")

			// Integer arithmetics
			case coppervm.InstAddInt:
				out.WriteString("  ; -- add --\n")
				out.WriteString("  pop rbx\n")
				out.WriteString("  pop rax\n")
				out.WriteString("  add rax, rbx\n")
				out.WriteString("  push rax\n")
			case coppervm.InstSubInt:
			case coppervm.InstMulInt:
			case coppervm.InstMulIntSigned:
			case coppervm.InstDivInt:
			case coppervm.InstDivIntSigned:
			case coppervm.InstModInt:
			case coppervm.InstModIntSigned:

			// Floating point arithmetics
			case coppervm.InstAddFloat:
			case coppervm.InstSubFloat:
			case coppervm.InstMulFloat:
			case coppervm.InstDivFloat:

			// Boolean operations
			case coppervm.InstAnd:
			case coppervm.InstOr:
			case coppervm.InstXor:
			case coppervm.InstNot:
			case coppervm.InstShiftLeft:
			case coppervm.InstShiftRight:

			// Flow control
			case coppervm.InstCmp:
			case coppervm.InstCmpSigned:
			case coppervm.InstCmpFloat:
			case coppervm.InstJmp:
			case coppervm.InstJmpZero:
			case coppervm.InstJmpNotZero:
			case coppervm.InstJmpGreater:
			case coppervm.InstJmpGreaterEqual:
			case coppervm.InstJmpLess:
			case coppervm.InstJmpLessEqual:

				// Functions
			case coppervm.InstFunCall:
			case coppervm.InstFunReturn:

				// Memory access
			case coppervm.InstMemRead:
			case coppervm.InstMemReadInt:
			case coppervm.InstMemReadFloat:
			case coppervm.InstMemWrite:
			case coppervm.InstMemWriteInt:
			case coppervm.InstMemWriteFloat:

			// Syscall
			case coppervm.InstSyscall:
			case coppervm.InstPrint:
			case coppervm.InstCount:
			}
		case IRKindEntry:
			entry = inst.AsEntry.Name
		case IRKindConst:
			//out.WriteString(fmt.Sprintf("%s:\n", inst.AsConst))
		case IRKindMemory:
			//out.WriteString(fmt.Sprintf("%s:\n", inst.AsLabel.Name))
		}
	}

	// set entry
	out.WriteString("_start:\n")
	out.WriteString(fmt.Sprintf("  jmp %s\n", entry))

	// set data

	x86_64Gen.s = out.String()
}

func (x86_64Gen x86_64Generator) saveProgram() string {
	return x86_64Gen.s
}

func expressionToString(expr Expression) (ret string) {
	switch expr.Kind {
	case ExpressionKindNumLitInt:
		ret = strconv.Itoa(int(expr.AsNumLitInt))
	case ExpressionKindNumLitFloat:
	case ExpressionKindStringLit:
	case ExpressionKindBinaryOp:
	case ExpressionKindBinding:
	case ExpressionKindByteList:
	}
	return ret
}
