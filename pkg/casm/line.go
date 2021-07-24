package casm

import (
	"fmt"
	"strings"
)

const (
	CasmCommentSymbol rune = ';'
	CasmLabelSymbol   rune = ':'
	CasmPPSymbol      rune = '%'
)

type LineKind int

const (
	LineKindLabel LineKind = iota
	LineKindInstruction
	LineKindDirective
)

func (kind LineKind) String() string {
	return [...]string{
		"LineKindLabel",
		"LineKindInstruction",
		"LineKindDirective",
	}[kind]
}

type LabelLine struct {
	Name string
}

type InstructionLine struct {
	Name    string
	Operand string
}

type DirectiveLine struct {
	Name  string
	Block string
}

type Line struct {
	Kind          LineKind
	AsLabel       LabelLine
	AsInstruction InstructionLine
	AsDirective   DirectiveLine
	Location      FileLocation
}

// Convert a source string to a list of Lines
func Linize(source string, fileName string) (out []Line, err error) {
	lines := strings.Split(source, "\n")
	for location, lineStr := range lines {
		lineStr = strings.TrimSpace(lineStr)

		if lineStr == "" || lineStr[0] == byte(CasmCommentSymbol) {
			continue
		}
		lineStr, _ = SplitByDelim(lineStr, CasmCommentSymbol)

		line, err := lineFromString(lineStr, FileLocation{
			FileName: fileName,
			Location: location + 1,
		})
		if err != nil {
			return []Line{}, err
		}
		out = append(out, line)
	}
	return out, err
}

// Parse a line from string
// Return a Line from a string
func lineFromString(line string, location FileLocation) (out Line, err error) {
	if line[0] == byte(CasmPPSymbol) {
		// Parse a directive line
		name, block := SplitByDelim(line, ' ')
		name = name[1:]
		name = strings.TrimSpace(name)
		block = strings.TrimSpace(block)

		// Fail if no block is declared
		if block == "" {
			return Line{},
				fmt.Errorf("%s: [ERROR]: trying to declare a directive without a block",
					location)
		}

		out = Line{
			Kind: LineKindDirective,
			AsDirective: DirectiveLine{
				Name:  name,
				Block: block,
			},
			Location: location,
		}
	} else if line[len(line)-1] == byte(CasmLabelSymbol) {
		// Parse a label line
		name := strings.TrimSpace(line[:len(line)-1])
		out = Line{
			Kind: LineKindLabel,
			AsLabel: LabelLine{
				Name: name,
			},
			Location: location,
		}
	} else {
		// Parse an instruction line
		name, operand := SplitByDelim(line, ' ')
		name = strings.TrimSpace(name)
		operand = strings.TrimSpace(operand)

		out = Line{
			Kind: LineKindInstruction,
			AsInstruction: InstructionLine{
				Name:    name,
				Operand: operand,
			},
			Location: location,
		}
	}

	return out, err
}
