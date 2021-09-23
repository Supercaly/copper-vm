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
	memory     [1024]byte
	memorySize int
}

func (gen *x86_64Generator) generateProgram(program []IR) {
	writeLine(&gen.textSection, "section .text")
	writeLine(&gen.dataSection, "section .data")
	writeLine(&gen.bssSection, "section .bss")

	writeLine(&gen.textSection, "global _start")
	for _, inst := range program {
		switch inst.Kind {
		case IRKindLabel:
			writeLine(&gen.textSection, inst.AsLabel.Name+":")
		case IRKindInstruction:
			gen.translateInstruction(inst.AsInstruction)
		case IRKindEntry:
			gen.entryName = inst.AsEntry.Name
		case IRKindConst:
			writeLine(&gen.dataSection, fmt.Sprintf("  %s: db %s", inst.AsConst.Name, gen.expressionToString(inst.AsConst.Value)))
		case IRKindMemory:
			addr := gen.memorySize
			bytes := inst.AsMemory.Value.AsByteList
			for i := 0; i < len(bytes); i++ {
				gen.memory[addr+i] = bytes[i]
			}
			gen.memorySize += len(bytes)
			writeLine(&gen.dataSection, fmt.Sprintf("  %s: db %d", inst.AsMemory.Name, addr))
		}
	}

	// set entry
	writeLine(&gen.textSection, "_start:")
	writeLine(&gen.textSection, fmt.Sprintf("  jmp %s", gen.entryName))

	// Write static memory
	var memStr string
	for i, b := range gen.memory {
		memStr += fmt.Sprintf("0x%x", b)
		if i != len(gen.memory)-1 {
			memStr += ","
		}
	}
	writeLine(&gen.dataSection, fmt.Sprintf("  mem: db %s", memStr))
}

func (gen *x86_64Generator) saveProgram() string {
	return gen.bssSection.String() + "\n" +
		gen.dataSection.String() + "\n" +
		gen.textSection.String()
}

func (gen *x86_64Generator) translateInstruction(inst InstructionIR) {
	_, instDef := coppervm.GetInstDefByName(inst.Name)
	switch instDef.Kind {
	// Basic instructions
	case coppervm.InstPush:
		writeLine(&gen.textSection, "  ; -- push --")
		writeLine(&gen.textSection, fmt.Sprintf("  push %s", gen.expressionToString(inst.Operand)))
	case coppervm.InstSwap:
		writeLine(&gen.textSection, "  ; -- swap --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  push rbx")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstDup:
		writeLine(&gen.textSection, "  ; -- dup --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  push rax")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstDrop:
		writeLine(&gen.textSection, "  ; -- drop --")
		writeLine(&gen.textSection, "  pop rax")
	case coppervm.InstHalt:
		writeLine(&gen.textSection, "  ; -- halt --")
		writeLine(&gen.textSection, "  mov rax, 0x3c")
		writeLine(&gen.textSection, "  mov rdi, 0x0")
		writeLine(&gen.textSection, "  syscall")

	// Integer arithmetics
	case coppervm.InstAddInt:
		binaryOpToNative(&gen.textSection, "add", "add")
	case coppervm.InstSubInt:
		binaryOpToNative(&gen.textSection, "sub", "sub")
	case coppervm.InstMulInt:
		binaryOpToNative(&gen.textSection, "mul", "mul")
	case coppervm.InstMulIntSigned:
		binaryOpToNative(&gen.textSection, "imul", "imul")
	case coppervm.InstDivInt:
		writeLine(&gen.textSection, "  ; -- div --")
		writeLine(&gen.textSection, "  mov rdx, 0")
		writeLine(&gen.textSection, "  pop rcx")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  div rcx")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstDivIntSigned:
		writeLine(&gen.textSection, "  ; -- idiv --")
		writeLine(&gen.textSection, "  mov rdx, 0")
		writeLine(&gen.textSection, "  pop rcx")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  idiv rcx")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstModInt:
		writeLine(&gen.textSection, "  ; -- mod --")
		writeLine(&gen.textSection, "  mov rdx, 0")
		writeLine(&gen.textSection, "  pop rcx")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  div rcx")
		writeLine(&gen.textSection, "  push rdx")
	case coppervm.InstModIntSigned:
		writeLine(&gen.textSection, "  ; -- imod --")
		writeLine(&gen.textSection, "  mov rdx, 0")
		writeLine(&gen.textSection, "  pop rcx")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  idiv rcx")
		writeLine(&gen.textSection, "  push rdx")

	// Floating point arithmetics
	case coppervm.InstAddFloat:
		binaryOpToNative(&gen.textSection, "fadd", "fadd")
	case coppervm.InstSubFloat:
		binaryOpToNative(&gen.textSection, "fsub", "fsub")
	case coppervm.InstMulFloat:
		binaryOpToNative(&gen.textSection, "fmul", "fmul")
	case coppervm.InstDivFloat:

	// Boolean operations
	case coppervm.InstAnd:
		binaryOpToNative(&gen.textSection, "and", "and")
	case coppervm.InstOr:
		binaryOpToNative(&gen.textSection, "or", "or")
	case coppervm.InstXor:
		binaryOpToNative(&gen.textSection, "xor", "xor")
	case coppervm.InstNot:
		writeLine(&gen.textSection, "  ; -- not --")
		writeLine(&gen.textSection, "  ; pop rax")
		writeLine(&gen.textSection, "  ; not rax")
		writeLine(&gen.textSection, "  ; push rax")
	case coppervm.InstShiftLeft:
		binaryOpToNative(&gen.textSection, "shl", "shl")
	case coppervm.InstShiftRight:
		binaryOpToNative(&gen.textSection, "shr", "shr")

	// Flow control
	case coppervm.InstCmp:
		writeLine(&gen.textSection, "  ; -- cmp --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  sub rbx, rax")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstCmpSigned:
		writeLine(&gen.textSection, "  ; -- icmp --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  sub rbx, rax")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstCmpFloat:
		writeLine(&gen.textSection, "  ; -- fcmp --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  fsub rbx, rax")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstJmp:
		writeLine(&gen.textSection, "  ; -- jmp --")
		writeLine(&gen.textSection, fmt.Sprintf("  jmp %s", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpZero:
		writeLine(&gen.textSection, "  ; -- jz --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jz %s", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpNotZero:
		writeLine(&gen.textSection, "  ; -- jnz --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jnz %s", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpGreater:
		writeLine(&gen.textSection, "  ; -- jg --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jg %s", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpGreaterEqual:
		writeLine(&gen.textSection, "  ; -- jge --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jge %s", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpLess:
		writeLine(&gen.textSection, "  ; -- jl --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jl %s", gen.expressionToString(inst.Operand)))
	case coppervm.InstJmpLessEqual:
		writeLine(&gen.textSection, "  ; -- jle --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jle %s", gen.expressionToString(inst.Operand)))

		// Functions
	case coppervm.InstFunCall:
		writeLine(&gen.textSection, "  ; -- call --")
		writeLine(&gen.textSection, fmt.Sprintf("  call %s", gen.expressionToString(inst.Operand)))
	case coppervm.InstFunReturn:
		writeLine(&gen.textSection, "  ; -- ret --")
		writeLine(&gen.textSection, "  ret")

		// Memory access
	case coppervm.InstMemRead:
		writeLine(&gen.textSection, "  ; -- read --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  mov rbx, [mem+rax]")
		writeLine(&gen.textSection, "  push rbx")
	case coppervm.InstMemReadInt:
		writeLine(&gen.textSection, "  ; -- iread --")
		writeLine(&gen.textSection, "  ; operation not supported")
	case coppervm.InstMemReadFloat:
		writeLine(&gen.textSection, "  ; -- fread --")
		writeLine(&gen.textSection, "  ; operation not supported")
	case coppervm.InstMemWrite:
		writeLine(&gen.textSection, "  ; -- write --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  mov [mem+rax], rbx")
	case coppervm.InstMemWriteInt:
		writeLine(&gen.textSection, "  ; -- iwrite --")
		writeLine(&gen.textSection, "  ; operation not supported")
	case coppervm.InstMemWriteFloat:
		writeLine(&gen.textSection, "  ; -- fwrite --")
		writeLine(&gen.textSection, "  ; operation not supported")

	// Syscall
	case coppervm.InstSyscall:
		writeLine(&gen.textSection, "  ; -- syscall --")
		switch inst.Operand.AsNumLitInt {
		case 0:
			writeLine(&gen.textSection, "  pop rdx")
			writeLine(&gen.textSection, "  pop rsi")
			writeLine(&gen.textSection, "  pop rdi")
			writeLine(&gen.textSection, "  mov rax, 0x0")
		case 1:
			writeLine(&gen.textSection, "  pop rdx")
			writeLine(&gen.textSection, "  pop rsi")
			writeLine(&gen.textSection, "  pop rdi")
			writeLine(&gen.textSection, "  mov rax, 0x1")
		case 2:
			writeLine(&gen.textSection, "  mov rdx, 0x2")
			writeLine(&gen.textSection, "  mov rsi, 0x2")
			writeLine(&gen.textSection, "  pop rdi")
			writeLine(&gen.textSection, "  mov rax, 0x2")
		case 3:
			writeLine(&gen.textSection, "  pop rdi")
			writeLine(&gen.textSection, "  mov rax, 0x3")
		case 4:
			writeLine(&gen.textSection, "  pop rdx")
			writeLine(&gen.textSection, "  pop rsi")
			writeLine(&gen.textSection, "  pop rdi")
			writeLine(&gen.textSection, "  mov rax, 0x8")
		case 5:
			writeLine(&gen.textSection, "  pop rdi")
			writeLine(&gen.textSection, "  mov rax, 0x3c")
		}
		writeLine(&gen.textSection, "  syscall")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstPrint:
		writeLine(&gen.textSection, "  ; -- print --")
		writeLine(&gen.textSection, "  ; operation not supported")
	default:
		panic(fmt.Sprintf("unknown instruction %s", instDef.Name))
	}
}

func (gen *x86_64Generator) expressionToString(expr Expression) (ret string) {
	switch expr.Kind {
	case ExpressionKindNumLitInt:
		ret = fmt.Sprint(expr.AsNumLitInt)
	case ExpressionKindNumLitFloat:
		ret = fmt.Sprint(expr.AsNumLitFloat)
	case ExpressionKindStringLit:
		addr := gen.memorySize
		byteStr := []byte(expr.AsStringLit)
		for i := 0; i < len(byteStr); i++ {
			gen.memory[addr+i] = byteStr[i]
		}
		gen.memorySize += len(byteStr)
		ret = fmt.Sprint(addr)
	case ExpressionKindBinaryOp:
	case ExpressionKindBinding:
		ret = expr.AsBinding
	case ExpressionKindByteList:
	}
	return ret
}

func binaryOpToNative(builder *strings.Builder, instName string, asmName string) {
	writeLine(builder, fmt.Sprintf("  ; -- %s --", instName))
	writeLine(builder, "  pop rbx")
	writeLine(builder, "  pop rax")
	writeLine(builder, fmt.Sprintf("  %s rax, rbx", asmName))
	writeLine(builder, "  push rax")
}

func writeLine(builder *strings.Builder, line string) {
	builder.WriteString(line)
	builder.WriteRune('\n')
}
