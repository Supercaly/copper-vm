package casm

import (
	"testing"

	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type TestHolder struct {
	in       string
	out      []coppervm.InstDef
	hasError bool
}

func TestTranslateSource(t *testing.T) {
	tests := []TestHolder{
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
		{":", []coppervm.InstDef{}, true},
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
		runTest(t, test)
	}
}

func runTest(t *testing.T, test TestHolder) {
	defer func() {
		r := recover()
		if r != nil && !test.hasError {
			t.Error(r)
		}
	}()
	ctx := Casm{}
	ctx.translateSource(test.in, "test_file")

	if test.hasError {
		t.Error("expecting an error")
	}
	if !instArrayEquals(ctx.Program, test.out) {
		t.Errorf("expected '%#v' but got '%#v'", test.out, ctx.Program)
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
