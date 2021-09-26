package casm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveProgramToFile(t *testing.T) {
	intRep := internalRep{}
	tests := []struct {
		casm     Casm
		hasError bool
	}{
		{Casm{
			OutputFile:  "testdata/test.notcopper",
			Target:      BuildTargetCopper,
			internalRep: &intRep,
			copperGen:   copperGenerator{rep: &intRep},
		}, true},
		{Casm{
			OutputFile:  "testdata/test.copper",
			Target:      BuildTargetCopper,
			internalRep: &intRep,
			copperGen:   copperGenerator{rep: &intRep},
		}, false},
		{Casm{
			OutputFile:  "testdata/test.notasm",
			Target:      BuildTargetX86_64,
			internalRep: &intRep,
			copperGen:   copperGenerator{rep: &intRep},
		}, true},
		{Casm{
			OutputFile:  "testdata/test.asm",
			Target:      BuildTargetX86_64,
			internalRep: &intRep,
			copperGen:   copperGenerator{rep: &intRep},
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
	{"main:\n", []IR{ir(IRKindLabel, LabelIR{"main"}, fileLocation(0, 0))}, false},
	{"push 1\n", []IR{
		ir(IRKindInstruction, InstructionIR{
			Name:       "push",
			HasOperand: true,
			Operand:    expression(ExpressionKindNumLitInt, int64(1)),
		}, fileLocation(0, 0)),
	}, false},
	{"%const N 1\n", []IR{ir(IRKindConst, ConstIR{
		"N",
		expression(ExpressionKindNumLitInt, int64(1)),
	}, fileLocation(0, 0))}, false},
	{":", []IR{}, true},
	{"wrong\n", []IR{}, true},
	{"push \n", []IR{}, true},
	{"%dir 0\n", []IR{}, true},
	{"push N\n", []IR{
		ir(IRKindInstruction, InstructionIR{
			HasOperand: true,
			Name:       "push",
			Operand:    expression(ExpressionKindBinding, "N"),
		}, fileLocation(0, 0)),
	}, false},
	{"%include abc", []IR{}, true},
}

func TestTranslateIR(t *testing.T) {
	for _, test := range testSources {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, fmt.Sprintf("unexpected error %s", r), test)
				}
			}()

			ctx := Casm{}
			tokens := tokenize(test.in, "")
			irs := ctx.translateTokensToIR(&tokens)

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			}
			assert.Equal(t, test.out, irs, test)
		}()
	}
}
