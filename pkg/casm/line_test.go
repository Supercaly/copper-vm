package casm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Wrapper function to create a Line.
func line(kind LineKind, value interface{}, location int) (line Line) {
	line.Kind = kind
	switch kind {
	case LineKindDirective:
		line.AsDirective = value.(DirectiveLine)
	case LineKindInstruction:
		line.AsInstruction = value.(InstructionLine)
	case LineKindLabel:
		line.AsLabel = value.(LabelLine)
	}
	line.Location = FileLocation{"", location}
	return line
}

var testLines = []struct {
	in       string
	out      []Line
	hasError bool
}{
	{"label:\ninst op\ninst op\n",
		[]Line{
			line(LineKindLabel, LabelLine{Name: "label"}, 1),
			line(LineKindInstruction, InstructionLine{Name: "inst", Operand: "op"}, 2),
			line(LineKindInstruction, InstructionLine{Name: "inst", Operand: "op"}, 3),
		},
		false,
	},
	{"label:\ninst op\n%directive name\n",
		[]Line{
			line(LineKindLabel, LabelLine{Name: "label"}, 1),
			line(LineKindInstruction, InstructionLine{Name: "inst", Operand: "op"}, 2),
			line(LineKindDirective, DirectiveLine{Name: "directive", Block: "name"}, 3),
		},
		false,
	}, {"%directive\n", []Line{}, true},
}

func TestLinize(t *testing.T) {
	for _, test := range testLines {
		lines, err := Linize(test.in, "")

		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
			assert.Equal(t, test.out, lines, test)
		}
	}
}

var testSingleLine = []struct {
	in       string
	out      Line
	hasError bool
}{
	{"label:", line(LineKindLabel, LabelLine{Name: "label"}, 0), false},
	{"inst op", line(LineKindInstruction, InstructionLine{Name: "inst", Operand: "op"}, 0), false},
	{"%directive name", line(LineKindDirective, DirectiveLine{Name: "directive", Block: "name"}, 0), false},
	{"%directive", Line{}, true},
}

func TestLineFromString(t *testing.T) {
	for _, test := range testSingleLine {
		line, err := lineFromString(test.in, FileLocation{})

		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
			assert.Equal(t, test.out, line, test)
		}
	}
}
