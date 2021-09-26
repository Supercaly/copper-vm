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

	rep *internalRep

	labels map[int]string
}

func (gen *x86_64Generator) generateProgram() {
	gen.labels = make(map[int]string)
	for _, b := range gen.rep.bindings {
		if b.isLabel {
			gen.labels[int(b.evaluatedWord.asInstAddr)] = b.name
		}
	}

	writeLine(&gen.textSection, "section .text")
	writeLine(&gen.dataSection, "section .data")
	writeLine(&gen.bssSection, "section .bss")

	writeLine(&gen.textSection, "global _start")
	for idx, inst := range gen.rep.program {
		if label, ok := gen.labels[idx]; ok {
			writeLine(&gen.textSection, fmt.Sprintf("%s:", label))
		}
		gen.translateInstruction(inst)
	}

	// set entry
	writeLine(&gen.textSection, "_start:")
	writeLine(&gen.textSection, fmt.Sprintf("  jmp %s", gen.rep.deferredEntryName))

	// Write static memory
	var memStr string
	for i, b := range gen.rep.memory {
		memStr += fmt.Sprintf("0x%x", b)
		if i != len(gen.rep.memory)-1 {
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

func (gen *x86_64Generator) translateInstruction(inst instruction) {
	switch inst.kind {
	// Basic instructions
	case coppervm.InstPush:
		writeLine(&gen.textSection, "  ; -- push --")
		writeLine(&gen.textSection, fmt.Sprintf("  mov rax, %s", gen.wordToString(inst.operand)))
		writeLine(&gen.textSection, "  push rax")
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
		writeLine(&gen.textSection, fmt.Sprintf("  jmp %s", gen.wordToString(inst.operand)))
	case coppervm.InstJmpZero:
		writeLine(&gen.textSection, "  ; -- jz --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jz %s", gen.wordToString(inst.operand)))
	case coppervm.InstJmpNotZero:
		writeLine(&gen.textSection, "  ; -- jnz --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jnz %s", gen.wordToString(inst.operand)))
	case coppervm.InstJmpGreater:
		writeLine(&gen.textSection, "  ; -- jg --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jg %s", gen.wordToString(inst.operand)))
	case coppervm.InstJmpGreaterEqual:
		writeLine(&gen.textSection, "  ; -- jge --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jge %s", gen.wordToString(inst.operand)))
	case coppervm.InstJmpLess:
		writeLine(&gen.textSection, "  ; -- jl --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jl %s", gen.wordToString(inst.operand)))
	case coppervm.InstJmpLessEqual:
		writeLine(&gen.textSection, "  ; -- jle --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jle %s", gen.wordToString(inst.operand)))

		// Functions
	case coppervm.InstFunCall:
		writeLine(&gen.textSection, "  ; -- call --")
		writeLine(&gen.textSection, fmt.Sprintf("  call %s", gen.wordToString(inst.operand)))
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
		switch inst.operand.asInt {
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
		panic(fmt.Sprintf("unknown instruction %s", inst.name))
	}
}

func (gen *x86_64Generator) wordToString(w word) (ret string) {
	switch w.kind {
	case wordKindInt:
		ret = fmt.Sprint(w.asInt)
	case wordKindFloat:
		ret = fmt.Sprint(w.asFloat)
	case wordKindInstAddr:
		ret = fmt.Sprint(w.asInstAddr)
	case wordKindMemoryAddr:
		ret = fmt.Sprint(w.asMemoryAddr)
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
