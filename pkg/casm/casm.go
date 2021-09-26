package casm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

	copperGen copperGenerator
	x86_64Gen x86_64Generator

	includeLevel int
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

	// Tokenize the source
	tokens := tokenize(source, filePath)

	// Create intermediate representation
	irs := casm.translateTokensToIR(&tokens)

	// Generate the program depending on the build target
	switch casm.Target {
	case BuildTargetCopper:
		casm.copperGen.generateProgram(irs)
	case BuildTargetX86_64:
		casm.x86_64Gen.generateProgram(irs)
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
		programSource = casm.copperGen.saveProgram(casm.AddDebugSymbols)
		if filepath.Ext(casm.OutputFile) != coppervm.CoppervmFileExtention {
			panic(fmt.Errorf("file '%s' is not a valid %s file", casm.OutputFile, coppervm.CoppervmFileExtention))
		}
	case BuildTargetX86_64:
		programSource = casm.x86_64Gen.saveProgram()
	}

	// save program to file
	if err := ioutil.WriteFile(casm.OutputFile, []byte(programSource), os.ModePerm); err != nil {
		panic(fmt.Errorf("error saving file '%s': %s", casm.OutputFile, err))
	}

	fmt.Printf("[INFO]: Program saved to '%s'\n", casm.OutputFile)
	return err
}

// Convert tokens to intermediate representation.
func (casm *Casm) translateTokensToIR(tokens *tokens) (out []IR) {
	for !tokens.Empty() {
		switch tokens.First().Kind {
		case tokenKindSymbol:
			symbol := tokens.Pop()

			if !tokens.Empty() && tokens.First().Kind == tokenKindColon {
				// Label definition
				tokens.Pop()
				out = append(out, IR{
					Kind:     IRKindLabel,
					AsLabel:  LabelIR{symbol.Text},
					Location: symbol.Location,
				})
			} else {
				// Intruction definition
				exist, instDef := GetInstructionByName(symbol.Text)
				if !exist {
					panic(fmt.Sprintf("%s: unknown instruction '%s'",
						symbol.Location,
						symbol.Text))
				}

				var operand Expression
				if instDef.HasOperand {
					operand = parseExprFromTokens(tokens)
				}
				out = append(out, IR{
					Kind: IRKindInstruction,
					AsInstruction: InstructionIR{
						Name:       instDef.Name,
						Operand:    operand,
						HasOperand: instDef.HasOperand,
					},
					Location: symbol.Location,
				})
			}
			if len(*tokens) != 0 {
				tokens.expectTokenKind(tokenKindNewLine)
				tokens.Pop()
			}
		case tokenKindPercent:
			directive := tokens.Pop()
			tokens.expectTokenKind(tokenKindSymbol)

			directiveName := tokens.Pop().Text
			switch directiveName {
			case "entry":
				tokens.expectTokenKindMsg(tokenKindSymbol, "no name given to entry directive")
				name := tokens.Pop()
				out = append(out, IR{
					Kind: IRKindEntry,
					AsEntry: EntryIR{
						Name: name.Text,
					},
					Location: directive.Location,
				})
			case "const":
				tokens.expectTokenKindMsg(tokenKindSymbol, "no name given to const directive")
				name := tokens.Pop()
				expr := parseExprFromTokens(tokens)

				out = append(out, IR{
					Kind: IRKindConst,
					AsConst: ConstIR{
						Name:  name.Text,
						Value: expr,
					},
					Location: directive.Location,
				})
			case "memory":
				tokens.expectTokenKindMsg(tokenKindSymbol, "no name given to memory directive")
				name := tokens.Pop()
				expr := parseExprFromTokens(tokens)

				out = append(out, IR{
					Kind: IRKindMemory,
					AsMemory: MemoryIR{
						Name:  name.Text,
						Value: expr,
					},
					Location: directive.Location,
				})
			case "include":
				tokens.expectTokenKindMsg(tokenKindStringLit, "no path given to include directive")
				includePath := tokens.Pop()
				out = append(out, casm.translateInclude(includePath.Text, includePath.Location)...)
			default:
				panic(fmt.Sprintf("%s: unknown directive '%s'", directive.Location, directiveName))
			}
			if len(*tokens) != 0 {
				tokens.expectTokenKind(tokenKindNewLine)
				tokens.Pop()
			}
		case tokenKindColon:
			panic(fmt.Sprintf("%s: empty labels are not supported", tokens.First().Location))
		case tokenKindNewLine:
			tokens.Pop()
		default:
			panic(fmt.Sprintf("%s: unsupported line start '%s'", tokens.First().Location, tokens.First().Kind))
		}
	}

	return out
}

// Translate include directive.
func (casm *Casm) translateInclude(path string, location FileLocation) (out []IR) {
	exist, resolvedPath := casm.resolveIncludePath(path)
	if !exist {
		panic(fmt.Sprintf("%s: cannot resolve include file '%s'", location, path))
	}

	if casm.includeLevel >= CasmMaxIncludeLevel {
		panic("maximum include level reached")
	}

	// Generate IR from included file
	casm.includeLevel++
	includeSource := readSourceFile(resolvedPath)
	tokens := tokenize(includeSource, resolvedPath)
	out = casm.translateTokensToIR(&tokens)
	casm.includeLevel--

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
