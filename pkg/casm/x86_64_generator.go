package casm

import (
	"fmt"
	"strings"

	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type x86_64Generator struct {
	textSection strings.Builder
	dataSection strings.Builder
	bssSection  strings.Builder

	entryName  string
	memory     []byte
	memAddress map[string]int
}

func (gen *x86_64Generator) generateProgram(program []IR) {
	gen.textSection.WriteString("section .text\n")
	gen.dataSection.WriteString("section .data\n")
	gen.bssSection.WriteString("section .bss\n")

	gen.textSection.WriteString("  global _start\n")
	for _, inst := range program {
		switch inst.Kind {
		case IRKindLabel:
			gen.textSection.WriteString(fmt.Sprintf("%s:\n", inst.AsLabel.Name))
		case IRKindInstruction:
			gen.translateIntruction(inst.AsInstruction)
		case IRKindEntry:
			gen.entryName = inst.AsEntry.Name
		case IRKindConst:
			gen.dataSection.WriteString(fmt.Sprintf("  %s: db %s\n",
				inst.AsConst.Name,
				gen.expressionToString(inst.AsConst.Value)))
		case IRKindMemory:
			if inst.AsMemory.Value.Kind != ExpressionKindByteList {
				panic(fmt.Sprintf("expected '%s' but got '%s'",
					ExpressionKindByteList,
					inst.AsMemory.Value.Kind))
			}
			if gen.memAddress == nil {
				gen.memAddress = make(map[string]int)
			}
			gen.memAddress[inst.AsMemory.Name] = len(gen.memory)
			gen.expressionToString(inst.AsMemory.Value)
		}
	}

	// set the memory as array
	gen.dataSection.WriteString("  mem: db ")
	for _, b := range gen.memory {
		gen.dataSection.WriteString(fmt.Sprintf("0x%v,", b))
	}
	gen.dataSection.WriteRune('\n')

	// set the entry
	gen.textSection.WriteString("_start:\n")
	gen.textSection.WriteString(fmt.Sprintf("  jmp %s\n", gen.entryName))
}

func (gen *x86_64Generator) saveProgram() string {
	return gen.bssSection.String() + "\n" + gen.dataSection.String() + "\n" + gen.textSection.String()
}

func (gen *x86_64Generator) translateIntruction(inst InstructionIR) {
	_, instDef := coppervm.GetInstDefByName(inst.Name)
	switch instDef.Kind {
	// Basic instructions
	case coppervm.InstPush:
		gen.textSection.WriteString("  ; -- push --\n")
		str := gen.expressionToString(inst.Operand)
		gen.textSection.WriteString(fmt.Sprintf("  push %s\n", str))
	case coppervm.InstSwap:
		gen.textSection.WriteString("  ; -- swap --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  push rax\n")
		gen.textSection.WriteString("  push rbx\n")
	case coppervm.InstDup:
		gen.textSection.WriteString("  ; -- dup --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  push rax\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstDrop:
		gen.textSection.WriteString("  ; -- drop --\n")
		gen.textSection.WriteString("  pop rax\n")
	case coppervm.InstHalt:
		gen.textSection.WriteString("  ; -- halt --\n")
		gen.textSection.WriteString("  mov rax, 0x3c\n")
		gen.textSection.WriteString("  mov rdi, 0x0\n")
		gen.textSection.WriteString("  syscall\n")

	// Integer arithmetics
	case coppervm.InstAddInt:
		gen.textSection.WriteString("  ; -- add --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  add rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstSubInt:
		gen.textSection.WriteString("  ; -- sub --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  sub rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstMulInt:
		gen.textSection.WriteString("  ; -- mul --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  mul rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstMulIntSigned:
		gen.textSection.WriteString("  ; -- imul --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  imul rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstDivInt:
		gen.textSection.WriteString("  ; -- div --\n")
		gen.textSection.WriteString("  mov rdx, 0\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  div rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstDivIntSigned:
		gen.textSection.WriteString("  ; -- idiv --\n")
		gen.textSection.WriteString("  mov rdx, 0\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  idiv rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstModInt:
		gen.textSection.WriteString("  ; -- mod --\n")
		gen.textSection.WriteString("  mov rdx, 0\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  div rbx\n")
		gen.textSection.WriteString("  push rdx\n")
	case coppervm.InstModIntSigned:
		gen.textSection.WriteString("  ; -- imod --\n")
		gen.textSection.WriteString("  mov rdx, 0\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  idiv rbx\n")
		gen.textSection.WriteString("  push rdx\n")
	// Floating point arithmetics
	case coppervm.InstAddFloat:
		gen.textSection.WriteString("  ; -- fadd --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  fadd rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstSubFloat:
		gen.textSection.WriteString("  ; -- fsub --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  fsub rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstMulFloat:
		gen.textSection.WriteString("  ; -- fmul --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  fmul rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstDivFloat:
		gen.textSection.WriteString("  ; -- fdiv --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  fdiv rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")

	// Boolean operations
	case coppervm.InstAnd:
		gen.textSection.WriteString("  ; -- and --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  and rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstOr:
		gen.textSection.WriteString("  ; -- or --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  or rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstXor:
		gen.textSection.WriteString("  ; -- xor --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  xor rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstNot:
		gen.textSection.WriteString("  ; -- not --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  not rax\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstShiftLeft:
		gen.textSection.WriteString("  ; -- shl --\n")
		gen.textSection.WriteString("  pop rcx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  shl rax, rcx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstShiftRight:
		gen.textSection.WriteString("  ; -- shr --\n")
		gen.textSection.WriteString("  pop rcx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  shr rax, rcx\n")
		gen.textSection.WriteString("  push rax\n")

	// Flow control
	case coppervm.InstCmp:
		gen.textSection.WriteString("  ; -- cmp --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  sub rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstCmpSigned:
		gen.textSection.WriteString("  ; -- icmp --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  isub rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstCmpFloat:
		gen.textSection.WriteString("  ; -- fcmp --\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  fsub rax, rbx\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstJmp:
		gen.textSection.WriteString("  ; -- jmp --\n")
		gen.textSection.WriteString(fmt.Sprintf("  jmp %s\n", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpZero:
		gen.textSection.WriteString("  ; -- jz --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  cmp rax, 0\n")
		gen.textSection.WriteString(fmt.Sprintf("  jz %s\n", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpNotZero:
		gen.textSection.WriteString("  ; -- jnz --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  cmp rax, 0\n")
		gen.textSection.WriteString(fmt.Sprintf("  jnz %s\n", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpGreater:
		gen.textSection.WriteString("  ; -- jg --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  cmp rax, 0\n")
		gen.textSection.WriteString(fmt.Sprintf("  ja %s\n", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpGreaterEqual:
		gen.textSection.WriteString("  ; -- jge --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  cmp rax, 0\n")
		gen.textSection.WriteString(fmt.Sprintf("  jae %s\n", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpLess:
		gen.textSection.WriteString("  ; -- jl --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  cmp rax, 0\n")
		gen.textSection.WriteString(fmt.Sprintf("  jb %s\n", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpLessEqual:
		gen.textSection.WriteString("  ; -- jle --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  cmp rax, 0\n")
		gen.textSection.WriteString(fmt.Sprintf("  jbe %s\n", gen.expressionToString(inst.Operand)))

		// Functions
	case coppervm.InstFunCall:
		gen.textSection.WriteString("  ; -- call --\n")
		gen.textSection.WriteString(fmt.Sprintf("  call %s\n", gen.expressionToString(inst.Operand)))
	case coppervm.InstFunReturn:
		gen.textSection.WriteString("  ; -- ret --\n")
		gen.textSection.WriteString("  ret\n")

		// Memory access
	case coppervm.InstMemRead:
		gen.textSection.WriteString("  ; -- read --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  mov rax, [mem+rax]\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstMemReadInt:
		gen.textSection.WriteString("  ; -- iread --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  mov rax, 0\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstMemReadFloat:
		gen.textSection.WriteString("  ; -- fread --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  mov rax, 0\n")
		gen.textSection.WriteString("  push rax\n")
	case coppervm.InstMemWrite:
		gen.textSection.WriteString("  ; -- write --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  mov [mem+rax], rbx\n")
	case coppervm.InstMemWriteInt:
		gen.textSection.WriteString("  ; -- iwrite --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  mov qword [mem+rax], rbx\n")
	case coppervm.InstMemWriteFloat:
		gen.textSection.WriteString("  ; -- fwrite --\n")
		gen.textSection.WriteString("  pop rax\n")
		gen.textSection.WriteString("  pop rbx\n")
		gen.textSection.WriteString("  mov qword [mem+rax], rbx\n")

	// Syscall
	case coppervm.InstSyscall:
		gen.textSection.WriteString("  ; -- syscall --\n")
		switch inst.Operand.AsNumLitInt {
		case 0:
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  push 0\n")
		case 1:
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  push 0\n")
		case 2:
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  push 0\n")
		case 3:
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  push 0\n")
		case 4:
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  push 0\n")
		case 5:
			gen.textSection.WriteString("  pop rax\n")
			gen.textSection.WriteString("  push 0\n")
		}
	case coppervm.InstPrint:
		gen.textSection.WriteString("  ; -- print --\n")
		gen.textSection.WriteString("  ; unsupported operation\n")
	case coppervm.InstCount:
		panic(fmt.Sprintf("unsupported instruction %s", instDef.Name))
	}
}

func (gen *x86_64Generator) expressionToString(expr Expression) (ret string) {
	switch expr.Kind {
	case ExpressionKindNumLitInt:
		ret = fmt.Sprint(expr.AsNumLitInt)
	case ExpressionKindNumLitFloat:
		ret = fmt.Sprint(expr.AsNumLitFloat)
	case ExpressionKindStringLit:
		ret = fmt.Sprintf("\"%s\"", expr.AsStringLit)
	case ExpressionKindBinaryOp:
		ret = gen.expressionToString(*expr.AsBinaryOp.Lhs)
		switch expr.AsBinaryOp.Kind {
		case BinaryOpKindPlus:
			ret += " + "
		case BinaryOpKindMinus:
			ret += " - "
		case BinaryOpKindTimes:
			ret += " * "
		case BinaryOpKindDivide:
			ret += " / "
		case BinaryOpKindModulo:
			ret += " % "
		}
		ret += gen.expressionToString(*expr.AsBinaryOp.Rhs)
	case ExpressionKindBinding:
		if idx, ok := gen.memAddress[expr.AsBinding]; ok {
			ret = fmt.Sprint(idx)
		} else {
			ret = fmt.Sprint(expr.AsBinding)
		}
	case ExpressionKindByteList:
		gen.memory = append(gen.memory, expr.AsByteList...)
		ret = ""
	}
	return ret
}
