package coppervm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var instDefTests = []string{
	"noop", "push", "swap", "dup", "over", "drop", "add", "sub", "mul", "imul", "div", "idiv", "mod", "imod",
	"fadd", "fsub", "fmul", "fdiv", "and", "or", "xor", "shl", "shr", "not", "cmp", "icmp", "fcmp",
	"jmp", "jz", "jnz", "jg", "jl", "jge", "jle", "call", "ret", "read", "iread", "fread",
	"write", "iwrite", "fwrite", "syscall", "print", "halt",
}

func TestGetInstDefByName(t *testing.T) {
	for _, test := range instDefTests {
		exist, _ := GetInstDefByName(test)
		assert.True(t, exist, test)
	}
}
