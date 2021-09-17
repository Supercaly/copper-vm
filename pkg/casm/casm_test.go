package casm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveProgramToFile(t *testing.T) {
	tests := []struct {
		casm     Casm
		hasError bool
	}{
		{Casm{
			OutputFile: "testdata/test.notcopper",
			Target:     BuildTargetCopper,
		}, true},
		{Casm{
			OutputFile: "testdata/test.copper",
			Target:     BuildTargetCopper,
		}, false},
		{Casm{
			OutputFile: "testdata/test.notasm",
			Target:     BuildTargetX86_64,
		}, false},
		{Casm{
			OutputFile: "testdata/test.asm",
			Target:     BuildTargetX86_64,
		}, false},
	}

	for _, test := range tests {
		err := test.casm.SaveProgramToFile()

		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
		}
	}
}

func TestTranslateSourceFile(t *testing.T) {
	tests := []struct {
		path     string
		hasError bool
	}{
		{"testdata/test.notcasm", true},
		{"testdata/test1.casm", true},
		{"testdata/test.casm", false},
	}
	casm := Casm{}

	for _, test := range tests {
		err := casm.TranslateSourceFile(test.path)

		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
		}
	}
}

// Wrapper function to create an IR.
func ir(kind IRKind, value interface{}, loc FileLocation) (ir IR) {
	ir.Location = loc
	ir.Kind = kind
	switch kind {
	case IRKindLabel:
		ir.AsLabel = value.(LabelIR)
	case IRKindInstruction:
		ir.AsInstruction = value.(InstructionIR)
	case IRKindEntry:
		ir.AsEntry = value.(EntryIR)
	case IRKindConst:
		ir.AsConst = value.(ConstIR)
	case IRKindMemory:
		ir.AsMemory = value.(MemoryIR)
	}
	return ir
}

var testSources = []struct {
	in       string
	out      []IR
	hasError bool
}{
	{"main:\n", []IR{ir(IRKindLabel, LabelIR{"main"}, FileLocation{"test_file", 1})}, false},
	{"push 1\n", []IR{
		ir(IRKindInstruction, InstructionIR{
			Name:       "push",
			HasOperand: true,
			Operand:    expression(ExpressionKindNumLitInt, int64(1)),
		}, FileLocation{"test_file", 1}),
	}, false},
	{"%const N 1\n", []IR{ir(IRKindConst, ConstIR{
		"N",
		expression(ExpressionKindNumLitInt, int64(1)),
	}, FileLocation{"test_file", 1})}, false},
	{":", []IR{}, true},
	{"wrong\n", []IR{}, true},
	{"push \n", []IR{}, true},
	{"%dir 0\n", []IR{}, true},
	{"push N\n", []IR{
		ir(IRKindInstruction, InstructionIR{
			HasOperand: true,
			Name:       "push",
			Operand:    expression(ExpressionKindBinding, "N"),
		}, FileLocation{"test_file", 1}),
	}, false},
	{"%include abc", []IR{}, true},
}

func TestTranslateIR(t *testing.T) {
	for _, test := range testSources {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			lines, err := Linize(test.in, "test_file")
			if err != nil {
				panic(err)
			}

			ctx := Casm{}
			irs := ctx.TranslateIR(lines)

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			}
			assert.Equal(t, test.out, irs, test)
		}()
	}
}
