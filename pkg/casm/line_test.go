package casm

import (
	"testing"
)

func TestLinize(t *testing.T) {
	filepath := ""
	tests := []struct {
		in       string
		out      []Line
		hasError bool
	}{
		{"label:\ninst op\ninst op\n",
			[]Line{
				{
					Kind:     LineKindLabel,
					AsLabel:  LabelLine{Name: "label"},
					Location: FileLocation{FileName: "", Location: 1},
				},
				{
					Kind:          LineKindInstruction,
					AsInstruction: InstructionLine{Name: "inst", Operand: "op"},
					Location:      FileLocation{FileName: "", Location: 2},
				},
				{
					Kind:          LineKindInstruction,
					AsInstruction: InstructionLine{Name: "inst", Operand: "op"},
					Location:      FileLocation{FileName: "", Location: 3},
				},
			},
			false,
		},
		{"label:\ninst op\n%directive name\n",
			[]Line{
				{
					Kind:     LineKindLabel,
					AsLabel:  LabelLine{Name: "label"},
					Location: FileLocation{FileName: "", Location: 1},
				},
				{
					Kind:          LineKindInstruction,
					AsInstruction: InstructionLine{Name: "inst", Operand: "op"},
					Location:      FileLocation{FileName: "", Location: 2},
				},
				{
					Kind:        LineKindDirective,
					AsDirective: DirectiveLine{Name: "directive", Block: "name"},
					Location:    FileLocation{FileName: "", Location: 3},
				},
			},
			false,
		}, {"%directive\n", []Line{}, true},
	}

	for _, test := range tests {
		lines, err := Linize(test.in, filepath)

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if !lineArrayEquals(lines, test.out) {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, lines)
		}
	}
}

func lineArrayEquals(a []Line, b []Line) bool {
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
