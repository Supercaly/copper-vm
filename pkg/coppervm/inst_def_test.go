package coppervm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var instDefTests = []struct {
	name  string
	exist bool
}{
	{name: "noop", exist: true},
	{name: "push", exist: true},
	{name: "swap", exist: true},
	{name: "dup", exist: true},
	{name: "drop", exist: true},
	{name: "add", exist: true},
	{name: "sub", exist: true},
	{name: "mul", exist: true},
	{name: "imul", exist: true},
	{name: "div", exist: true},
	{name: "idiv", exist: true},
	{name: "mod", exist: true},
	{name: "imod", exist: true},
	{name: "fadd", exist: true},
	{name: "fsub", exist: true},
	{name: "fmul", exist: true},
	{name: "fdiv", exist: true},
	{name: "and", exist: true},
	{name: "or", exist: true},
	{name: "xor", exist: true},
	{name: "shl", exist: true},
	{name: "shr", exist: true},
	{name: "not", exist: true},
	{name: "cmp", exist: true},
	{name: "icmp", exist: true},
	{name: "fcmp", exist: true},
	{name: "jmp", exist: true},
	{name: "jz", exist: true},
	{name: "jnz", exist: true},
	{name: "jg", exist: true},
	{name: "jl", exist: true},
	{name: "jge", exist: true},
	{name: "jle", exist: true},
	{name: "call", exist: true},
	{name: "ret", exist: true},
	{name: "read", exist: true},
	{name: "iread", exist: true},
	{name: "fread", exist: true},
	{name: "write", exist: true},
	{name: "iwrite", exist: true},
	{name: "fwrite", exist: true},
	{name: "syscall", exist: true},
	{name: "print", exist: true},
	{name: "halt", exist: true},
	{name: "abc", exist: false},
}

func TestGetInstDefByName(t *testing.T) {
	for _, test := range instDefTests {
		exist, _ := GetInstDefByName(test.name)

		if test.exist {
			assert.True(t, exist, test)
		} else {
			assert.False(t, exist, test)
		}
	}
}
