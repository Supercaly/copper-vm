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
func (casm *Casm) TranslateSourceFile(filePath string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("error reading file '%s': %s", filePath, err))
	}

	casm.translateSource(string(bytes), filePath)
	return err
}

// Translate a copper assembly file to copper vm's binary.
// Given a source string this function will read it and parse it
// as assembly code for the vm generating the correct program
// in-memory.
// Use TranslateSourceFile if you want to parse a file.
// Use SaveProgramToFile to save the program to binary file.
func (casm *Casm) translateSource(source string, filePath string) {
	// Linize the source
	lines, err := Linize(source, filePath)
	if err != nil {
		panic(err)
	}

	// First Pass
	for _, line := range lines {
		switch line.Kind {
		case LineKindLabel:
			if line.AsLabel.Name == "" {
				panic(fmt.Sprintf("%s: empty labels are not supported", line.Location))
			}

			casm.bindLabel(line.AsLabel.Name, len(casm.Program), line.Location)
		case LineKindInstruction:
			exist, instDef := coppervm.GetInstDefByName(line.AsInstruction.Name)
			if !exist {
				panic(fmt.Sprintf("%s: unknown instruction '%s'",
					line.Location,
					line.AsInstruction.Name))
			}

			if instDef.HasOperand {
				operand, err := ParseExprFromString(line.AsInstruction.Operand)
				if err != nil {
					panic(fmt.Sprintf("%s: %s", line.Location, err))
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
					instDef.Operand = casm.evaluateExpression(operand, line.Location)
				}
			}
			casm.Program = append(casm.Program, instDef)
		case LineKindDirective:
			switch line.AsDirective.Name {
			case "entry":
				casm.bindEntry(line.AsDirective.Block, line.Location)
			case "const":
				casm.bindConst(line.AsDirective, line.Location)
			case "memory":
				casm.bindMemory(line.AsDirective, line.Location)
			case "include":
				casm.translateInclude(line.AsDirective, line.Location)
			default:
				panic(fmt.Sprintf("%s: unknown directive '%s'", line.Location, line.AsDirective.Name))
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
			panic(fmt.Sprintf("%s: unknown binding '%s'",
				deferredOp.Location,
				deferredOp.Name))
		}
		casm.Program[deferredOp.Address].Operand = casm.evaluateBinding(&binding, deferredOp.Location)
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
			panic(fmt.Sprintf("%s: unknown binding '%s'",
				casm.EntryLocation,
				casm.DeferredEntryName))
		}

		if binding.Value.Kind != ExpressionKindNumLitInt {
			panic(fmt.Sprintf("%s: only label names can be set as entry point",
				casm.EntryLocation))
		}
		entry := casm.evaluateBinding(&binding, casm.EntryLocation)
		casm.Entry = int(entry.AsI64)
	}
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
func (casm *Casm) bindLabel(name string, address int, location FileLocation) {
	exist, binding := casm.getBindingByName(name)
	if exist {
		panic(fmt.Sprintf("%s: label name '%s' is already bound at location '%s'",
			location,
			name,
			binding.Location))
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
}

// Binds a constant.
func (casm *Casm) bindConst(directive DirectiveLine, location FileLocation) {
	name, block := internal.SplitByDelim(directive.Block, ' ')
	name = strings.TrimSpace(name)
	block = strings.TrimSpace(block)

	exist, binding := casm.getBindingByName(name)
	if exist {
		panic(fmt.Sprintf("%s: constant name '%s' is already bound at location '%s'",
			location,
			name,
			binding.Location))
	}

	value, err := ParseExprFromString(block)
	if err != nil {
		panic(fmt.Sprintf("%s: %s", location, err))
	}

	casm.Bindings = append(casm.Bindings, Binding{
		Name:     name,
		Value:    value,
		Location: location,
		IsLabel:  false,
	})
}

// Binds an entry point.
func (casm *Casm) bindEntry(name string, location FileLocation) {
	if casm.HasEntry {
		panic(fmt.Sprintf("%s: entry point is already set to '%s'",
			location,
			casm.EntryLocation))
	}

	casm.DeferredEntryName = name
	casm.HasEntry = true
	casm.EntryLocation = location
}

// Binds a memory definition
func (casm *Casm) bindMemory(directive DirectiveLine, location FileLocation) {
	name, block := internal.SplitByDelim(directive.Block, ' ')
	name = strings.TrimSpace(name)
	block = strings.TrimSpace(block)

	exist, binding := casm.getBindingByName(name)
	if exist {
		panic(fmt.Sprintf("%s: memory name '%s' is already bound at location '%s'",
			location,
			name,
			binding.Location))
	}

	chunk, err := ParseByteArrayFromString(block)
	if err != nil {
		panic(fmt.Sprintf("%s: %s", location, err))
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
}

// Translate include directive
func (casm *Casm) translateInclude(directive DirectiveLine, location FileLocation) {
	exist, resolvedPath := casm.resolveIncludePath(directive.Block)
	if !exist {
		panic(fmt.Sprintf("%s: cannot resolve include file '%s'", location, directive.Block))
	}

	if casm.IncludeLevel >= CasmMaxIncludeLevel {
		panic("maximum include level reached")
	}

	casm.IncludeLevel++
	if err := casm.TranslateSourceFile(resolvedPath); err != nil {
		panic(err)
	}
	casm.IncludeLevel--
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
func (casm *Casm) evaluateBinding(binding *Binding, location FileLocation) coppervm.Word {
	// TODO(#6): Cyclic bindings cause a stack overflow
	return casm.evaluateExpression(binding.Value, location)
}

// Evaluate an expression to extract a word.
func (casm *Casm) evaluateExpression(expr Expression, location FileLocation) (ret coppervm.Word) {
	switch expr.Kind {
	case ExpressionKindBinding:
		exist, binding := casm.getBindingByName(expr.AsBinding)
		if !exist {
			panic(fmt.Sprintf("%s: cannot find binding '%s'", location, expr.AsBinding))
		}
		ret = casm.evaluateBinding(&binding, location)
	case ExpressionKindNumLitInt:
		ret = coppervm.WordI64(expr.AsNumLitInt)
	case ExpressionKindNumLitFloat:
		ret = coppervm.WordF64(expr.AsNumLitFloat)
	case ExpressionKindStringLit:
		// TODO: Create method push string to memory
		strBase := len(casm.Memory)
		byteStr := []byte(expr.AsStringLit)
		byteStr = append(byteStr, 0)
		casm.Memory = append(casm.Memory, byteStr...)
		ret = coppervm.WordU64(uint64(strBase))
	case ExpressionKindBinaryOp:
		lhs := casm.evaluateExpression(*expr.AsBinaryOp.Lhs, location)
		rhs := casm.evaluateExpression(*expr.AsBinaryOp.Rhs, location)
		switch expr.AsBinaryOp.Kind {
		case BinaryOpKindPlus:
			ret = coppervm.AddWord(lhs, rhs)
		case BinaryOpKindMinus:
			ret = coppervm.SubWord(lhs, rhs)
		case BinaryOpKindTimes:
			ret = coppervm.MulWord(lhs, rhs)
		}
	}
	return ret
}
