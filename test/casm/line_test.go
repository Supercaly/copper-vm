package casm

import (
	"testing"

	"github.com/Supercaly/coppervm/pkg/casm"
)

func TestLinize(t *testing.T) {
	filepath := ""
	tests := []struct {
		in       string
		out      []casm.Line
		hasError bool
	}{
		{"label:\ninst op\ninst op\n",
			[]casm.Line{
				{
					Kind:     casm.LineKindLabel,
					AsLabel:  casm.LabelLine{Name: "label"},
					Location: casm.FileLocation{FileName: "", Location: 1},
				},
				{
					Kind:          casm.LineKindInstruction,
					AsInstruction: casm.InstructionLine{Name: "inst", Operand: "op"},
					Location:      casm.FileLocation{FileName: "", Location: 2},
				},
				{
					Kind:          casm.LineKindInstruction,
					AsInstruction: casm.InstructionLine{Name: "inst", Operand: "op"},
					Location:      casm.FileLocation{FileName: "", Location: 3},
				},
			},
			false,
		},
		{"label:\ninst op\n%directive name\n",
			[]casm.Line{
				{
					Kind:     casm.LineKindLabel,
					AsLabel:  casm.LabelLine{Name: "label"},
					Location: casm.FileLocation{FileName: "", Location: 1},
				},
				{
					Kind:          casm.LineKindInstruction,
					AsInstruction: casm.InstructionLine{Name: "inst", Operand: "op"},
					Location:      casm.FileLocation{FileName: "", Location: 2},
				},
				{
					Kind:        casm.LineKindDirective,
					AsDirective: casm.DirectiveLine{Name: "directive", Block: "name"},
					Location:    casm.FileLocation{FileName: "", Location: 3},
				},
			},
			false,
		}, {"%directive\n", []casm.Line{}, true},
	}

	for _, test := range tests {
		lines, err := casm.Linize(test.in, filepath)

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if !lineArrayEquals(lines, test.out) {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, lines)
		}
	}
}

func lineArrayEquals(a []casm.Line, b []casm.Line) bool {
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
