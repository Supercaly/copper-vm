package casm

import (
	"testing"

	"github.com/Supercaly/coppervm/pkg/casm"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

func TestTranslateSource(t *testing.T) {
	tests := []struct {
		in       string
		out      []coppervm.InstDef
		hasError bool
	}{
		{"main:\n", []coppervm.InstDef{}, false},
		{"push 1\n", []coppervm.InstDef{
			{
				Kind:       coppervm.InstPush,
				HasOperand: true,
				Name:       "push",
				Operand:    coppervm.WordU64(1),
			},
		}, false},
		{"%const N 1\n", []coppervm.InstDef{}, false},
		{":\n", []coppervm.InstDef{}, true},
		{"wrong\n", []coppervm.InstDef{}, true},
		{"push \n", []coppervm.InstDef{}, true},
		{"%dir 0\n", []coppervm.InstDef{}, true},
		{"push N\n", []coppervm.InstDef{
			{
				Kind:       coppervm.InstPush,
				HasOperand: true,
				Name:       "push",
				Operand:    coppervm.WordU64(0),
			},
		}, true},
		{"%entry main\n%const main 2.0", []coppervm.InstDef{}, true},
		{"%entry main\n%entry main2", []coppervm.InstDef{}, true},
		{"%include abc", []coppervm.InstDef{}, true},
	}

	for _, test := range tests {
		ctx := casm.Casm{}
		err := ctx.TranslateSource(test.in, "test_file")

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if !instArrayEquals(ctx.Program, test.out) {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, ctx.Program)
		}
	}
}

func instArrayEquals(a []coppervm.InstDef, b []coppervm.InstDef) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}
