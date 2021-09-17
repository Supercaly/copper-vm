package casm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Supercaly/coppervm/internal"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

const (
	CasmMaxIncludeLevel int    = 10
	CasmFileExtention   string = ".casm"
)

type Casm struct {
	InputFile  string
	OutputFile string

	Target BuildTarget

	CopperGen copperGenerator
	X86_64Gen x86_64Generator

	IncludeLevel int
	IncludePaths []string

	AddDebugSymbols bool
}

// Translate a copper assembly file to copper vm's binary.
// Given a file path this function will read it and generate
// the correct program in-memory.
// Use TranslateSource is you already have a source string.
// Use SaveProgramToFile to save the program to binary file.
func (casm *Casm) TranslateSourceFile(filePath string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	source := readSourceFile(filePath)

	internal.DebugPrint("[INFO]: Building program '%s'\n", filePath)

	// Linize the source
	lines, err := Linize(source, filePath)
	if err != nil {
		panic(err)
	}

	// Create intermediate representation
	irs := casm.TranslateIR(lines)

	// Generate the program depending on the build target
	switch casm.Target {
	case BuildTargetCopper:
		casm.CopperGen.generateProgram(irs)
	case BuildTargetX86_64:
		casm.X86_64Gen.generateProgram(irs)
	}

	internal.DebugPrint("[INFO]: Built program '%s'\n", filePath)
	return err
}

// Save a copper vm program to binary file.
func (casm *Casm) SaveProgramToFile() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	// generate the correct source program
	var programSource string
	switch casm.Target {
	case BuildTargetCopper:
		programSource = casm.CopperGen.saveProgram(casm.AddDebugSymbols)
		if filepath.Ext(casm.OutputFile) != coppervm.CoppervmFileExtention {
			panic(fmt.Errorf("file '%s' is not a valid %s file", casm.OutputFile, coppervm.CoppervmFileExtention))
		}
	case BuildTargetX86_64:
		programSource = casm.X86_64Gen.saveProgram()
	}

	// save program to file
	if err := ioutil.WriteFile(casm.OutputFile, []byte(programSource), os.ModePerm); err != nil {
		panic(fmt.Errorf("error saving file '%s': %s", casm.OutputFile, err))
	}

	fmt.Printf("[INFO]: Program saved to '%s'\n", casm.OutputFile)
	return err
}

// Convert lines to intermediate representation.
func (casm *Casm) TranslateIR(lines []Line) (out []IR) {
	for _, line := range lines {
		switch line.Kind {
		case LineKindLabel:
			if line.AsLabel.Name == "" {
				panic(fmt.Sprintf("%s: empty labels are not supported", line.Location))
			}

			out = append(out, IR{
				Kind:     IRKindLabel,
				AsLabel:  LabelIR{line.AsLabel.Name},
				Location: line.Location,
			})
		case LineKindInstruction:
			exist, instDef := coppervm.GetInstDefByName(line.AsInstruction.Name)
			if !exist {
				panic(fmt.Sprintf("%s: unknown instruction '%s'",
					line.Location,
					line.AsInstruction.Name))
			}

			var operand Expression
			if instDef.HasOperand {
				var err error
				operand, err = ParseExprFromString(line.AsInstruction.Operand)
				if err != nil {
					panic(fmt.Sprintf("%s: %s", line.Location, err))
				}
			}
			out = append(out, IR{
				Kind: IRKindInstruction,
				AsInstruction: InstructionIR{
					Name:       instDef.Name,
					Operand:    operand,
					HasOperand: instDef.HasOperand,
				},
				Location: line.Location,
			})
		case LineKindDirective:
			switch line.AsDirective.Name {
			case "entry":
				out = append(out, IR{
					Kind: IRKindEntry,
					AsEntry: EntryIR{
						Name: line.AsDirective.Block,
					},
					Location: line.Location,
				})
			case "const":
				name, block := internal.SplitByDelim(line.AsDirective.Block, ' ')
				name = strings.TrimSpace(name)
				block = strings.TrimSpace(block)
				expr, err := ParseExprFromString(block)
				if err != nil {
					panic(fmt.Sprintf("%s: %s", line.Location, err))
				}

				out = append(out, IR{
					Kind: IRKindConst,
					AsConst: ConstIR{
						Name:  name,
						Value: expr,
					},
					Location: line.Location,
				})
			case "memory":
				name, block := internal.SplitByDelim(line.AsDirective.Block, ' ')
				name = strings.TrimSpace(name)
				block = strings.TrimSpace(block)
				expr, err := ParseExprFromString(block)
				if err != nil {
					panic(fmt.Sprintf("%s: %s", line.Location, err))
				}

				out = append(out, IR{
					Kind: IRKindMemory,
					AsMemory: MemoryIR{
						Name:  name,
						Value: expr,
					},
					Location: line.Location,
				})
			case "include":
				out = append(out, casm.translateInclude(line.AsDirective, line.Location)...)
			default:
				panic(fmt.Sprintf("%s: unknown directive '%s'", line.Location, line.AsDirective.Name))
			}
		}
	}

	return out
}

// Translate include directive.
func (casm *Casm) translateInclude(directive DirectiveLine, location FileLocation) (out []IR) {
	exist, resolvedPath := casm.resolveIncludePath(directive.Block)
	if !exist {
		panic(fmt.Sprintf("%s: cannot resolve include file '%s'", location, directive.Block))
	}

	if casm.IncludeLevel >= CasmMaxIncludeLevel {
		panic("maximum include level reached")
	}

	// Generate IR from included file
	casm.IncludeLevel++
	includeSource := readSourceFile(resolvedPath)
	lines, err := Linize(includeSource, resolvedPath)
	if err != nil {
		panic(err)
	}
	out = casm.TranslateIR(lines)
	casm.IncludeLevel--

	return out
}

// Resolve an include path from the list of includes.
func (casm *Casm) resolveIncludePath(path string) (exist bool, resolved string) {
	// Check the include paths
	for _, includePath := range casm.IncludePaths {
		resolved = filepath.Join(includePath, path)
		internal.DebugPrint("[INFO]: search for '%s' in %s\n", path, includePath)
		_, err := os.Stat(resolved)
		if err == nil {
			return true, resolved
		}
	}
	return false, ""
}

// Returns the source from a given .casm file path.
// This function will panic if something goes wrong.
func readSourceFile(filePath string) string {
	if filepath.Ext(filePath) != CasmFileExtention {
		panic(fmt.Sprintf("file '%s' is not a valid %s file", filePath, CasmFileExtention))
	}
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("error reading file '%s': %s", filePath, err))
	}
	return string(bytes)
}
