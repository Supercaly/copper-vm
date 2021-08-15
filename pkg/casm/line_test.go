package casm

import (
	"reflect"
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
		} else if err == nil && !reflect.DeepEqual(lines, test.out) {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, lines)
		}
	}
}

func TestLineFromString(t *testing.T) {
	tests := []struct {
		in       string
		out      Line
		hasError bool
	}{
		{"label:", Line{
			Kind:     LineKindLabel,
			AsLabel:  LabelLine{Name: "label"},
			Location: FileLocation{},
		}, false},
		{"inst op", Line{
			Kind:          LineKindInstruction,
			AsInstruction: InstructionLine{Name: "inst", Operand: "op"},
			Location:      FileLocation{},
		}, false},
		{"%directive name", Line{
			Kind:        LineKindDirective,
			AsDirective: DirectiveLine{Name: "directive", Block: "name"},
			Location:    FileLocation{},
		}, false},
		{"%directive", Line{}, true},
	}

	for _, test := range tests {
		line, err := lineFromString(test.in, FileLocation{})

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if err == nil && line != test.out {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, line)
		}
	}
}
