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

	hasPrintFn bool
}

func (gen *x86_64Generator) generateProgram() {
	// Populate the labels map
	gen.labels = make(map[int]string)
	for _, b := range gen.rep.bindings {
		if b.isLabel {
			gen.labels[int(b.evaluatedWord.asInstAddr)] = b.name
		}
	}

	// Write the sections
	writeLine(&gen.textSection, "section .text")
	writeLine(&gen.dataSection, "section .data")
	writeLine(&gen.bssSection, "section .bss")

	// Write the _start condition
	writeLine(&gen.textSection, "global _start")
	if !gen.rep.hasEntry {
		writeLine(&gen.textSection, "_start:")
	} else {
		writeLine(&gen.textSection, "_start:")
		writeLine(&gen.textSection, fmt.Sprintf("  jmp %s", gen.rep.deferredEntryName))
	}

	// Write the main program
	for idx, inst := range gen.rep.program {
		if label, ok := gen.labels[idx]; ok {
			writeLine(&gen.textSection, fmt.Sprintf("%s:", label))
		}
		gen.translateInstruction(inst)
	}

	// Write static memory
	var memStr string
	if len(gen.rep.memory) > 0 {
		for i, b := range gen.rep.memory {
			memStr += fmt.Sprintf("0x%x", b)
			if i != len(gen.rep.memory)-1 {
				memStr += ","
			}
		}
	} else {
		memStr = "0x0"
	}
	writeLine(&gen.dataSection, fmt.Sprintf("  mem: db %s", memStr))

	// Append debug print instruction
	if gen.hasPrintFn {
		writeLine(&gen.dataSection, "  print_memory: db 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,10,0")
		writeLine(&gen.dataSection, "  print_memory_size: equ $-print_memory")

		writeLine(&gen.textSection, "")
		writeLine(&gen.textSection, "debug_print:")
		writeLine(&gen.textSection, "  xor rsi, rsi")
		writeLine(&gen.textSection, "  xor rdi, rdi")
		writeLine(&gen.textSection, "  call clean_print_memory")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, "  jge print_positive")
		writeLine(&gen.textSection, "  xor rdx, rdx")
		writeLine(&gen.textSection, "  mov rbx, -1")
		writeLine(&gen.textSection, "  imul rbx")
		writeLine(&gen.textSection, "  mov [print_memory], byte '-'")
		writeLine(&gen.textSection, "print_positive:")
		writeLine(&gen.textSection, "  mov rsi, 18")
		writeLine(&gen.textSection, "  mov rdi, rax")
		writeLine(&gen.textSection, "  call print_int_rec")
		writeLine(&gen.textSection, "  mov rax, 1")
		writeLine(&gen.textSection, "  mov rdi, 1")
		writeLine(&gen.textSection, "  mov rsi, print_memory")
		writeLine(&gen.textSection, "  mov rdx, print_memory_size")
		writeLine(&gen.textSection, "  syscall")
		writeLine(&gen.textSection, "  ret ")
		writeLine(&gen.textSection, "print_int_rec:")
		writeLine(&gen.textSection, "  cmp rdi, 9")
		writeLine(&gen.textSection, "  jg print_int_rec_else")
		writeLine(&gen.textSection, "  mov rax, rdi")
		writeLine(&gen.textSection, "  add rax, 48")
		writeLine(&gen.textSection, "  mov [print_memory+rsi], al")
		writeLine(&gen.textSection, "  ret")
		writeLine(&gen.textSection, "print_int_rec_else:")
		writeLine(&gen.textSection, "  xor rdx, rdx")
		writeLine(&gen.textSection, "  mov rax, rdi")
		writeLine(&gen.textSection, "  mov rbx, 10")
		writeLine(&gen.textSection, "  idiv rbx")
		writeLine(&gen.textSection, "  add rdx, 48")
		writeLine(&gen.textSection, "  mov [print_memory+rsi], dl")
		writeLine(&gen.textSection, "  sub rsi, 1")
		writeLine(&gen.textSection, "  mov rdi, rax")
		writeLine(&gen.textSection, "  call print_int_rec")
		writeLine(&gen.textSection, "  ret")
		writeLine(&gen.textSection, "clean_print_memory:")
		writeLine(&gen.textSection, "  mov rdi, 0")
		writeLine(&gen.textSection, "  mov rbx, 1")
		writeLine(&gen.textSection, "clean_print_memory_loop:")
		writeLine(&gen.textSection, "  mov [print_memory+rdi], byte 0")
		writeLine(&gen.textSection, "  add rdi, rbx")
		writeLine(&gen.textSection, "  cmp rdi, 18")
		writeLine(&gen.textSection, "  jle clean_print_memory_loop")
		writeLine(&gen.textSection, "  ret")
	}
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
		writeLine(&gen.textSection, "  mov rax, [rsp]")
		writeLine(&gen.textSection, fmt.Sprintf("  mov rbx, [rsp+%d]", inst.operand.asInt*8))
		writeLine(&gen.textSection, "  mov [rsp], rbx")
		writeLine(&gen.textSection, fmt.Sprintf("  mov [rsp+%d], rax", inst.operand.asInt*8))
	case coppervm.InstDup:
		writeLine(&gen.textSection, "  ; -- dup --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  push rax")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstOver:
		writeLine(&gen.textSection, "  ; -- over --")
		writeLine(&gen.textSection, fmt.Sprintf("  mov rax, [rsp+%d]", inst.operand.asInt*8))
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
		binpToNative(&gen.textSection, "add", "add")
	case coppervm.InstSubInt:
		binpToNative(&gen.textSection, "sub", "sub")
	case coppervm.InstMulInt:
		mulDivBinpToNative(&gen.textSection, "mul", "mul", "rax")
	case coppervm.InstMulIntSigned:
		mulDivBinpToNative(&gen.textSection, "imul", "imul", "rax")
	case coppervm.InstDivInt:
		mulDivBinpToNative(&gen.textSection, "div", "div", "rax")
	case coppervm.InstDivIntSigned:
		mulDivBinpToNative(&gen.textSection, "idiv", "idiv", "rax")
	case coppervm.InstModInt:
		mulDivBinpToNative(&gen.textSection, "mod", "div", "rdx")
	case coppervm.InstModIntSigned:
		mulDivBinpToNative(&gen.textSection, "imod", "idiv", "rdx")

	// Floating point arithmetics
	case coppervm.InstAddFloat:
		floatBinopToNative(&gen.textSection, "fadd", "addsd")
	case coppervm.InstSubFloat:
		floatBinopToNative(&gen.textSection, "fsub", "subsd")
	case coppervm.InstMulFloat:
		floatBinopToNative(&gen.textSection, "fmul", "mulsd")
	case coppervm.InstDivFloat:
		floatBinopToNative(&gen.textSection, "fdiv", "divsd")

	// Boolean operations
	case coppervm.InstAnd:
		binpToNative(&gen.textSection, "and", "and")
	case coppervm.InstOr:
		binpToNative(&gen.textSection, "or", "or")
	case coppervm.InstXor:
		binpToNative(&gen.textSection, "xor", "xor")
	case coppervm.InstNot:
		writeLine(&gen.textSection, "  ; -- not --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  not rax")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstShiftLeft:
		writeLine(&gen.textSection, "  ; -- shl --")
		writeLine(&gen.textSection, "  pop rcx")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  shl rax, cl")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstShiftRight:
		writeLine(&gen.textSection, "  ; -- shr --")
		writeLine(&gen.textSection, "  pop rcx")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  shr rax, cl")
		writeLine(&gen.textSection, "  push rax")

	// Flow control
	case coppervm.InstCmp:
		writeLine(&gen.textSection, "  ; -- cmp --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  sub rbx, rax")
		writeLine(&gen.textSection, "  push rbx")
	case coppervm.InstCmpSigned:
		writeLine(&gen.textSection, "  ; -- icmp --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  sub rbx, rax")
		writeLine(&gen.textSection, "  push rbx")
	case coppervm.InstCmpFloat:
		writeLine(&gen.textSection, "  ; -- fcmp --")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  movq xmm0, rax")
		writeLine(&gen.textSection, "  movq xmm1, rbx")
		writeLine(&gen.textSection, "  subsd xmm0, xmm1")
		writeLine(&gen.textSection, "  cvtsd2si rax, xmm0")
		writeLine(&gen.textSection, "  push rax")
		writeLine(&gen.textSection, "  push rax")
	case coppervm.InstJmp:
		writeLine(&gen.textSection, "  ; -- jmp --")
		writeLine(&gen.textSection, fmt.Sprintf("  jmp %s", gen.wordToLabel(inst.operand)))
	case coppervm.InstJmpZero:
		writeLine(&gen.textSection, "  ; -- jz --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jz %s", gen.wordToLabel(inst.operand)))
	case coppervm.InstJmpNotZero:
		writeLine(&gen.textSection, "  ; -- jnz --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jnz %s", gen.wordToLabel(inst.operand)))
	case coppervm.InstJmpGreater:
		writeLine(&gen.textSection, "  ; -- jg --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jg %s", gen.wordToLabel(inst.operand)))
	case coppervm.InstJmpGreaterEqual:
		writeLine(&gen.textSection, "  ; -- jge --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jge %s", gen.wordToLabel(inst.operand)))
	case coppervm.InstJmpLess:
		writeLine(&gen.textSection, "  ; -- jl --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jl %s", gen.wordToLabel(inst.operand)))
	case coppervm.InstJmpLessEqual:
		writeLine(&gen.textSection, "  ; -- jle --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  cmp rax, 0")
		writeLine(&gen.textSection, fmt.Sprintf("  jle %s", gen.wordToLabel(inst.operand)))

		// Functions
	case coppervm.InstFunCall:
		writeLine(&gen.textSection, "  ; -- call --")
		writeLine(&gen.textSection, fmt.Sprintf("  call %s", gen.wordToLabel(inst.operand)))
	case coppervm.InstFunReturn:
		writeLine(&gen.textSection, "  ; -- ret --")
		writeLine(&gen.textSection, "  ret")

		// Memory access
	case coppervm.InstMemRead:
		writeLine(&gen.textSection, "  ; -- read --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  xor rbx, rbx")
		writeLine(&gen.textSection, "  mov bl, [mem+rax]")
		writeLine(&gen.textSection, "  push rbx")
	case coppervm.InstMemReadInt:
		writeLine(&gen.textSection, "  ; -- iread --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  xor rbx, rbx")
		writeLine(&gen.textSection, "  mov bl, [mem+rax]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+1]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+2]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+3]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+4]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+5]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+6]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+7]")
		writeLine(&gen.textSection, "  push rbx")
	case coppervm.InstMemReadFloat:
		writeLine(&gen.textSection, "  ; -- fread --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  xor rbx, rbx")
		writeLine(&gen.textSection, "  mov bl, [mem+rax]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+1]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+2]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+3]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+4]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+5]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+6]")
		writeLine(&gen.textSection, "  shl rbx, 8")
		writeLine(&gen.textSection, "  mov bl, [mem+rax+7]")
		writeLine(&gen.textSection, "  push rbx")
	case coppervm.InstMemWrite:
		writeLine(&gen.textSection, "  ; -- write --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  mov [mem+rax], bl")
	case coppervm.InstMemWriteInt:
		writeLine(&gen.textSection, "  ; -- iwrite --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  mov [mem+rax+7], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+6], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+5], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+4], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+3], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+2], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+1], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax], bl")
	case coppervm.InstMemWriteFloat:
		writeLine(&gen.textSection, "  ; -- fwrite --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  pop rbx")
		writeLine(&gen.textSection, "  mov [mem+rax+7], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+6], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+5], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+4], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+3], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+2], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax+1], bl")
		writeLine(&gen.textSection, "  shr rbx, 8")
		writeLine(&gen.textSection, "  mov [mem+rax], bl")

		// Syscall
	case coppervm.InstSyscall:
		writeLine(&gen.textSection, "  ; -- syscall --")
		switch inst.operand.asInt {
		case 0:
			writeLine(&gen.textSection, "  pop rdx")
			writeLine(&gen.textSection, "  pop rsi")
			writeLine(&gen.textSection, "  pop rdi")
			writeLine(&gen.textSection, "  add rsi, mem")
			writeLine(&gen.textSection, "  mov rax, 0x0")
		case 1:
			writeLine(&gen.textSection, "  pop rdx")
			writeLine(&gen.textSection, "  pop rsi")
			writeLine(&gen.textSection, "  add rsi, mem")
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
		gen.hasPrintFn = true
		writeLine(&gen.textSection, "  ; -- print --")
		writeLine(&gen.textSection, "  pop rax")
		writeLine(&gen.textSection, "  call debug_print")
	case coppervm.InstNoop:
		writeLine(&gen.textSection, "  ; -- noop --")
		writeLine(&gen.textSection, "  nop")
	default:
		panic(fmt.Sprintf("unknown instruction %s", inst.name))
	}
}

// Convert a word to an immediate string.
func (gen *x86_64Generator) wordToString(w word) (ret string) {
	switch w.kind {
	case wordKindInt:
		ret = fmt.Sprint(w.asInt)
	case wordKindFloat:
		ret = fmt.Sprintf("__float64__(%f)", w.asFloat)
	case wordKindInstAddr:
		ret = fmt.Sprint(w.asInstAddr)
	case wordKindMemoryAddr:
		ret = fmt.Sprint(w.asMemoryAddr)
	}
	return ret
}

// Convert a word to a label name.
func (gen *x86_64Generator) wordToLabel(w word) (label string) {
	label, exist := gen.labels[int(w.asInstAddr)]
	if !exist {
		panic(fmt.Sprintf("no label for address %d", w.asInstAddr))
	}
	return label
}

// Creates a binary operation with given instruction name and assembly name.
func binpToNative(builder *strings.Builder, instName string, asmName string) {
	writeLine(builder, fmt.Sprintf("  ; -- %s --", instName))
	writeLine(builder, "  pop rbx")
	writeLine(builder, "  pop rax")
	writeLine(builder, fmt.Sprintf("  %s rax, rbx", asmName))
	writeLine(builder, "  push rax")
}

// Creates a binary operation with given instruction name and assembly name and destination reg.
func mulDivBinpToNative(builder *strings.Builder, instName string, asmName string, dstReg string) {
	writeLine(builder, fmt.Sprintf("  ; -- %s --", instName))
	writeLine(builder, "  xor rdx, rdx")
	writeLine(builder, "  pop rbx")
	writeLine(builder, "  pop rax")
	writeLine(builder, fmt.Sprintf("  %s rbx", asmName))
	writeLine(builder, fmt.Sprintf("  push %s", dstReg))
}

// Creates a binary operation with given instruction name and assembly name.
func floatBinopToNative(builder *strings.Builder, instName string, asmName string) {
	writeLine(builder, fmt.Sprintf("  ; -- %s --", instName))
	writeLine(builder, "  pop rbx")
	writeLine(builder, "  movq xmm1, rbx")
	writeLine(builder, "  pop rax")
	writeLine(builder, "  movq xmm0, rax")
	writeLine(builder, fmt.Sprintf("  %s xmm0, xmm1", asmName))
	writeLine(builder, "  movq rax, xmm0")
	writeLine(builder, "  push rax")
}

// Appends a line to given strings.Builder.
func writeLine(builder *strings.Builder, line string) {
	builder.WriteString(line)
	builder.WriteRune('\n')
}
