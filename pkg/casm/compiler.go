package casm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Supercaly/coppervm/internal"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

const (
	CasmDebug           bool = false
	CasmMaxIncludeLevel int  = 10
)

type Casm struct {
	InputFile  string
	OutputFile string

	AddDebugSymbols bool

	Bindings         []Binding
	DeferredOperands []DeferredOperand

	Program []coppervm.InstDef

	HasEntry          bool
	Entry             int
	EntryLocation     FileLocation
	DeferredEntryName string

	Memory []byte

	IncludeLevel int
	IncludePaths []string
}

// Save a copper vm program to binary file.
func (casm *Casm) SaveProgramToFile() error {
	var dbSymbols coppervm.DebugSymbols
	// Append debug symbols
	if casm.AddDebugSymbols {
		for _, b := range casm.Bindings {
			if b.IsLabel {
				dbSymbols = append(dbSymbols, coppervm.DebugSymbol{
					Name:    b.Name,
					Address: coppervm.InstAddr(b.Value.AsNumLitInt),
				})
			}
		}
	}

	meta := coppervm.FileMeta(casm.Entry, casm.Program, casm.Memory, dbSymbols)
	metaJson, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("error writign program to file %s", err)
	}

	fileErr := ioutil.WriteFile(casm.OutputFile, []byte(metaJson), os.ModePerm)
	if fileErr != nil {
		return fmt.Errorf("error saving file '%s': %s", casm.OutputFile, fileErr)
	}
	fmt.Printf("[INFO]: Program saved to '%s'\n", casm.OutputFile)

	return nil
}

// Translate a copper assembly file to copper vm's binary.
// Given a file path this function will read it and generate
// the correct program in-memory.
// Use TranslateSource is you already have a source string.
// Use SaveProgramToFile to save the program to binary file.
func (casm *Casm) TranslateSourceFile(filePath string) error {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %s", filePath, err)
	}
	if err := casm.TranslateSource(string(bytes), filePath); err != nil {
		return err
	}
	return nil
}

// Translate a copper assembly file to copper vm's binary.
// Given a source string this function will read it and parse it
// as assembly code for the vm generating the correct program
// in-memory.
// Use TranslateSourceFile if you want to parse a file.
// Use SaveProgramToFile to save the program to binary file.
func (casm *Casm) TranslateSource(source string, filePath string) error {
	// Linize the source
	lines, err := Linize(source, filePath)
	if err != nil {
		return err
	}

	// First Pass
	for _, line := range lines {
		switch line.Kind {
		case LineKindLabel:
			if line.AsLabel.Name == "" {
				return fmt.Errorf("%s: empty labels are not supported", line.Location)
			}

			err := casm.bindLabel(line.AsLabel.Name, len(casm.Program), line.Location)
			if err != nil {
				return fmt.Errorf("%s: %s", line.Location, err)
			}
		case LineKindInstruction:
			exist, instDef := coppervm.GetInstDefByName(line.AsInstruction.Name)
			if !exist {
				return fmt.Errorf("%s: unknown instruction '%s'",
					line.Location,
					line.AsInstruction.Name)
			}

			if instDef.HasOperand {
				operand, err := ParseExprFromString(line.AsInstruction.Operand)
				if err != nil {
					return fmt.Errorf("%s: %s", line.Location, err)
				}
				if operand.Kind == ExpressionKindBinding {
					if CasmDebug {
						fmt.Println("[INFO]: push deferred operand" + operand.AsBinding)
					}
					casm.DeferredOperands = append(casm.DeferredOperands,
						DeferredOperand{
							Name:     operand.AsBinding,
							Address:  len(casm.Program),
							Location: line.Location,
						})
				} else {
					instDef.Operand, err = casm.evaluateExpression(operand)
					if err != nil {
						return fmt.Errorf("%s: %s", line.Location, err)
					}
				}
			}
			casm.Program = append(casm.Program, instDef)
		case LineKindDirective:
			switch line.AsDirective.Name {
			case "entry":
				err := casm.bindEntry(line.AsDirective.Block, line.Location)
				if err != nil {
					return fmt.Errorf("%s: %s", line.Location, err)
				}
			case "const":
				err := casm.bindConst(line.AsDirective, line.Location)
				if err != nil {
					return fmt.Errorf("%s: %s", line.Location, err)
				}
			case "memory":
				err := casm.bindMemory(line.AsDirective, line.Location)
				if err != nil {
					return fmt.Errorf("%s: %s", line.Location, err)
				}
			case "include":
				err := casm.translateInclude(line.AsDirective, line.Location)
				if err != nil {
					return fmt.Errorf("%s: %s", line.Location, err)
				}
			default:
				return fmt.Errorf("%s: unknown directive '%s'", line.Location, line.AsDirective.Name)
			}
		}
	}

	// Second Pass
	for _, deferredOp := range casm.DeferredOperands {
		if CasmDebug {
			fmt.Printf("[INFO]: resolve deferred operand '%s' at '%d'\n", deferredOp.Name, deferredOp.Address)
		}

		exist, binding := casm.getBindingByName(deferredOp.Name)
		if !exist {
			return fmt.Errorf("%s: unknown binding '%s'",
				deferredOp.Location,
				deferredOp.Name)
		}
		casm.Program[deferredOp.Address].Operand, err = casm.evaluateBinding(&binding)
		if err != nil {
			return fmt.Errorf("%s: %s", deferredOp.Location, err)
		}
	}

	// Print all the bindings
	if CasmDebug {
		for _, b := range casm.Bindings {
			fmt.Printf("[INFO]: binding: %s (%d %d %s) %s\n",
				b.Name,
				b.Value.Kind,
				b.Value.AsNumLitInt,
				b.Value.AsBinding,
				b.Location)
		}
	}

	// Resolve entry point
	if casm.HasEntry && casm.DeferredEntryName != "" {
		exist, binding := casm.getBindingByName(casm.DeferredEntryName)
		if !exist {
			return fmt.Errorf("%s: unknown binding '%s'",
				casm.EntryLocation,
				casm.DeferredEntryName)
		}

		if binding.Value.Kind != ExpressionKindNumLitInt {
			return fmt.Errorf("%s: only label names can be set as entry point",
				casm.EntryLocation)
		}
		entry, err := casm.evaluateBinding(&binding)
		if err != nil {
			return fmt.Errorf("%s: %s", casm.EntryLocation, err)
		}
		casm.Entry = int(entry.AsI64)
	}

	return nil
}

// Returns a binding by its name.
// If the binding exist the first return parameter will be true,
// otherwise it'll be null.
func (casm *Casm) getBindingByName(name string) (bool, Binding) {
	for _, b := range casm.Bindings {
		if b.Name == name {
			return true, b
		}
	}
	return false, Binding{}
}

// Binds a label.
func (casm *Casm) bindLabel(name string, address int, location FileLocation) error {
	exist, binding := casm.getBindingByName(name)
	if exist {
		return fmt.Errorf("label name '%s' is already bound at location '%s'",
			name,
			binding.Location)
	}

	casm.Bindings = append(casm.Bindings, Binding{
		Name: name,
		Value: Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: int64(address),
		},
		Location: location,
		IsLabel:  true,
	})
	return nil
}

// Binds a constant.
func (casm *Casm) bindConst(directive DirectiveLine, location FileLocation) error {
	name, block := internal.SplitByDelim(directive.Block, ' ')
	name = strings.TrimSpace(name)
	block = strings.TrimSpace(block)

	exist, binding := casm.getBindingByName(name)
	if exist {
		return fmt.Errorf("constant name '%s' is already bound at location '%s'",
			name,
			binding.Location)
	}

	value, err := ParseExprFromString(block)
	if err != nil {
		return err
	}

	casm.Bindings = append(casm.Bindings, Binding{
		Name:     name,
		Value:    value,
		Location: location,
		IsLabel:  false,
	})
	return nil
}

// Binds an entry point.
func (casm *Casm) bindEntry(name string, location FileLocation) error {
	if casm.HasEntry {
		return fmt.Errorf("entry point is already set to '%s'",
			casm.EntryLocation)
	}

	casm.DeferredEntryName = name
	casm.HasEntry = true
	casm.EntryLocation = location

	return nil
}

// Binds a memory definition
func (casm *Casm) bindMemory(directive DirectiveLine, location FileLocation) error {
	name, block := internal.SplitByDelim(directive.Block, ' ')
	name = strings.TrimSpace(name)
	block = strings.TrimSpace(block)

	exist, binding := casm.getBindingByName(name)
	if exist {
		return fmt.Errorf("memory name '%s' is already bound at location '%s'",
			name,
			binding.Location)
	}

	chunk, err := ParseByteArrayFromString(block)
	if err != nil {
		return err
	}
	memAddr := len(casm.Memory)
	casm.Memory = append(casm.Memory, chunk...)

	casm.Bindings = append(casm.Bindings, Binding{
		Name: name,
		Value: Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: int64(memAddr),
		},
		Location: location,
		IsLabel:  false,
	})
	return nil
}

// Translate include directive
func (casm *Casm) translateInclude(directive DirectiveLine, location FileLocation) error {
	exist, resolvedPath := casm.resolveIncludePath(directive.Block)
	if !exist {
		return fmt.Errorf("cannot resolve include file '%s'", directive.Block)
	}

	if casm.IncludeLevel >= CasmMaxIncludeLevel {
		return fmt.Errorf("maximum include level reached")
	}

	casm.IncludeLevel++
	if err := casm.TranslateSourceFile(resolvedPath); err != nil {
		return err
	}
	casm.IncludeLevel--
	return nil
}

// Resolve an include path from the list of includes.
func (casm *Casm) resolveIncludePath(path string) (exist bool, resolved string) {
	// Check the path itself
	_, err := os.Stat(path)
	if err == nil {
		return true, path
	}

	// Check the include paths
	for _, includePath := range casm.IncludePaths {
		resolved = filepath.Join(includePath, path)
		_, err := os.Stat(resolved)
		if err == nil {
			return true, resolved
		}
	}
	return false, ""
}

// Evaluate a binding to extract a word.
func (casm *Casm) evaluateBinding(binding *Binding) (value coppervm.Word, err error) {
	// TODO(#6): Cyclic bindings cause a stack overflow
	return casm.evaluateExpression(binding.Value)
}

// Evaluate an expression to extract a word.
func (casm *Casm) evaluateExpression(expr Expression) (ret coppervm.Word, err error) {
	switch expr.Kind {
	case ExpressionKindBinding:
		exist, binding := casm.getBindingByName(expr.AsBinding)
		if !exist {
			return coppervm.Word{}, fmt.Errorf("cannot find binding '%s'",
				expr.AsBinding)
		}
		ret, err = casm.evaluateBinding(&binding)
		if err != nil {
			return coppervm.Word{}, err
		}
	case ExpressionKindNumLitInt:
		ret = coppervm.WordI64(expr.AsNumLitInt)
	case ExpressionKindNumLitFloat:
		ret = coppervm.WordF64(expr.AsNumLitFloat)
	case ExpressionKindStringLit:
		strBase := len(casm.Memory)
		byteStr := []byte(expr.AsStringLit)
		byteStr = append(byteStr, 0)
		casm.Memory = append(casm.Memory, byteStr...)
		ret = coppervm.WordU64(uint64(strBase))
	case ExpressionKindBinaryOp:
		lhs, err := casm.evaluateExpression(*expr.AsBinaryOp.Lhs)
		if err != nil {
			return coppervm.Word{}, err
		}
		rhs, err := casm.evaluateExpression(*expr.AsBinaryOp.Rhs)
		if err != nil {
			return coppervm.Word{}, err
		}
		switch expr.AsBinaryOp.Kind {
		case BinaryOpKindPlus:
			ret = coppervm.AddWord(lhs, rhs)
		case BinaryOpKindMinus:
			ret = coppervm.SubWord(lhs, rhs)
		case BinaryOpKindTimes:
			ret = coppervm.MulWord(lhs, rhs)
		}
	}
	return ret, nil
}
