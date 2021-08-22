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

	StringLengths map[int]int
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
						fmt.Println("[INFO]: push deferred operand " + operand.AsBinding)
					}
					casm.DeferredOperands = append(casm.DeferredOperands,
						DeferredOperand{
							Name:     operand.AsBinding,
							Address:  len(casm.Program),
							Location: line.Location,
						})
				} else {
					instDef.Operand = casm.evaluateExpression(operand, line.Location).Word
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
			fmt.Printf("[INFO]: resolve deferred operand '%s' at '%s'\n",
				deferredOp.Name,
				deferredOp.Location)
		}

		exist, binding := casm.getBindingByName(deferredOp.Name)
		if !exist {
			panic(fmt.Sprintf("%s: unknown binding '%s'",
				deferredOp.Location,
				deferredOp.Name))
		}
		casm.Program[deferredOp.Address].Operand = casm.evaluateBinding(binding,
			deferredOp.Location).Word
	}

	// Print all the bindings
	if CasmDebug {
		for _, b := range casm.Bindings {
			fmt.Printf("[INFO]: binding: %s \n", b)
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
		entry := casm.evaluateBinding(binding, casm.EntryLocation).Word
		casm.Entry = int(entry.AsI64)
	}

	// Check if halt instruction exist only in the main program
	// skip this when translating includes
	if casm.IncludeLevel == 0 {
		hasHalt := false
		for _, inst := range casm.Program {
			if inst.Kind == coppervm.InstHalt {
				hasHalt = true
			}
		}
		if !hasHalt {
			fmt.Printf("[WARN]: no 'halt' instruction found in the program! This program could not work as intended.\n")
		}
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

// Returns the index of a binding by it's name.
// If the binding doesn't exist -1 is returned.
func (casm *Casm) getBindingIndexByName(name string) int {
	for idx, b := range casm.Bindings {
		if b.Name == name {
			return idx
		}
	}
	return -1
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
		Status:        BindingEvaluated,
		Name:          name,
		EvaluatedWord: coppervm.WordU64(uint64(address)),
		Location:      location,
		IsLabel:       true,
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

	newBinding := Binding{
		Status:   BindingUnevaluated,
		Name:     name,
		Value:    value,
		Location: location,
		IsLabel:  false,
	}

	// If it's a const string push it in memory and bind his base address
	if value.Kind == ExpressionKindStringLit {
		baseAddr := casm.pushStringToMemory(value.AsStringLit)
		newBinding.EvaluatedWord = coppervm.WordU64(uint64(baseAddr))
		newBinding.Status = BindingEvaluated
	}

	casm.Bindings = append(casm.Bindings, newBinding)
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
		Status:        BindingEvaluated,
		Name:          name,
		EvaluatedWord: coppervm.WordU64(uint64(memAddr)),
		Location:      location,
		IsLabel:       false,
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

// Represent the result of an expression evaluation.
type EvalResult struct {
	Word coppervm.Word
	Type ExpressionKind
}

// Evaluate a binding to extract am eval result.
func (casm *Casm) evaluateBinding(binding Binding, location FileLocation) (ret EvalResult) {
	switch binding.Status {
	case BindingUnevaluated:
		idx := casm.getBindingIndexByName(binding.Name)
		if idx == -1 {
			panic(fmt.Sprintf("%s: cannot find index binding %s", location, binding.Name))
		}
		casm.Bindings[idx].Status = BindingEvaluating
		ret = casm.evaluateExpression(binding.Value, location)
		casm.Bindings[idx].Status = BindingEvaluated
		casm.Bindings[idx].EvaluatedWord = ret.Word
	case BindingEvaluating:
		panic(fmt.Sprintf("%s: cycling binding definition detected", location))
	case BindingEvaluated:
		ret = EvalResult{
			binding.EvaluatedWord,
			binding.Value.Kind,
		}
	}
	return ret
}

// Evaluate an expression to extract an eval result.
func (casm *Casm) evaluateExpression(expr Expression, location FileLocation) (ret EvalResult) {
	switch expr.Kind {
	case ExpressionKindBinding:
		exist, binding := casm.getBindingByName(expr.AsBinding)
		if !exist {
			panic(fmt.Sprintf("%s: cannot find binding '%s'", location, expr.AsBinding))
		}
		ret = casm.evaluateBinding(binding, location)
	case ExpressionKindNumLitInt:
		ret = EvalResult{
			coppervm.WordI64(expr.AsNumLitInt),
			ExpressionKindNumLitInt,
		}
	case ExpressionKindNumLitFloat:
		ret = EvalResult{
			coppervm.WordF64(expr.AsNumLitFloat),
			ExpressionKindNumLitFloat,
		}
	case ExpressionKindStringLit:
		strBase := casm.pushStringToMemory(expr.AsStringLit)
		ret = EvalResult{
			coppervm.WordU64(uint64(strBase)),
			ExpressionKindStringLit,
		}
	case ExpressionKindBinaryOp:
		ret = casm.evaluateBinaryOp(expr, location)
	}
	return ret
}

// Map of types of binary operation sides to
// the result type.
// The unsupported operations between types are
// marked as -1 following this table:
//
//[
//   // i  f  s  b  o
//     [i, f, -, -, -], //i
//     [f, f, -, -, -], //f
//     [-, -, s, -, -], //s
//     [-, -, -, -, -], //o
//     [-, -, -, -, -], //b
// ]
var binaryOpEvaluationMap = [5][5]ExpressionKind{
	{ExpressionKindNumLitInt, ExpressionKindNumLitFloat, -1, -1, -1},
	{ExpressionKindNumLitFloat, ExpressionKindNumLitFloat, -1, -1, -1},
	{-1, -1, ExpressionKindStringLit, -1, -1},
	{-1, -1, -1, -1, -1},
	{-1, -1, -1, -1, -1},
}

// Map an ExpressionKind to a TypeRepresentation.
var exprKindToTypeRepMap = map[ExpressionKind]coppervm.TypeRepresentation{
	ExpressionKindNumLitInt:   coppervm.TypeI64,
	ExpressionKindNumLitFloat: coppervm.TypeF64,
	ExpressionKindStringLit:   coppervm.TypeU64,
}

// Evaluate a binary op expression to extract an eval result.
func (casm *Casm) evaluateBinaryOp(binop Expression, location FileLocation) (ret EvalResult) {
	lhs_result := casm.evaluateExpression(*binop.AsBinaryOp.Lhs, location)
	rhs_result := casm.evaluateExpression(*binop.AsBinaryOp.Rhs, location)

	opType := binaryOpEvaluationMap[lhs_result.Type][rhs_result.Type]
	if opType == -1 {
		panic(fmt.Sprintf("%s: unsupported binary operation between types '%s' and '%s'",
			location,
			lhs_result.Type,
			rhs_result.Type))
	}
	typeRep := exprKindToTypeRepMap[opType]

	if opType == ExpressionKindStringLit {
		if binop.AsBinaryOp.Kind == BinaryOpKindMinus ||
			binop.AsBinaryOp.Kind == BinaryOpKindTimes {
			panic(fmt.Sprintf("%s: unsupported operations ['-', '*'] between string literals",
				location))
		}
		leftStr := casm.getStringByAddress(int(lhs_result.Word.AsU64))
		rightStr := casm.getStringByAddress(int(rhs_result.Word.AsU64))
		ret = EvalResult{
			coppervm.WordU64(uint64(casm.pushStringToMemory(leftStr + rightStr))),
			ExpressionKindStringLit,
		}
	} else {
		switch binop.AsBinaryOp.Kind {
		case BinaryOpKindPlus:
			ret = EvalResult{coppervm.AddWord(lhs_result.Word, rhs_result.Word, typeRep), opType}
		case BinaryOpKindMinus:
			ret = EvalResult{coppervm.SubWord(lhs_result.Word, rhs_result.Word, typeRep), opType}
		case BinaryOpKindTimes:
			ret = EvalResult{coppervm.MulWord(lhs_result.Word, rhs_result.Word, typeRep), opType}
		}
	}
	return ret
}

// Push a string to casm memory and return the base address.
func (casm *Casm) pushStringToMemory(str string) int {
	strBase := len(casm.Memory)
	byteStr := []byte(str)
	byteStr = append(byteStr, 0)
	casm.Memory = append(casm.Memory, byteStr...)

	if casm.StringLengths == nil {
		casm.StringLengths = make(map[int]int)
	}
	casm.StringLengths[strBase] = len(byteStr)
	return strBase
}

// Returns a string from memory at given address without
// null termination.
// If the string doesn't exist an empty string is returned.
func (casm *Casm) getStringByAddress(addr int) string {
	strLen := casm.StringLengths[addr]
	if strLen == 0 {
		return ""
	}
	strBytes := casm.Memory[addr : addr+strLen-1]
	return string(strBytes[:])
}
