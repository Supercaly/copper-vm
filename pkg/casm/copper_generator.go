package casm

import (
	"encoding/json"
	"fmt"

	"github.com/Supercaly/coppervm/internal"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type copperGenerator struct {
	bindings            []binding
	deferredOperands    []deferredOperand
	deferredExpressions map[int]Expression

	program []Instruction

	hasEntry          bool
	entry             int
	entryLocation     FileLocation
	deferredEntryName string

	memory []byte

	stringLengths map[int]int
}

// Generate a string file for the program.
func (cg *copperGenerator) saveProgram(addDebugSymbols bool) string {
	var dbSymbols coppervm.DebugSymbols
	// Append debug symbols
	if addDebugSymbols {
		for _, b := range cg.bindings {
			if b.isLabel {
				dbSymbols = append(dbSymbols, coppervm.DebugSymbol{
					Name:    b.name,
					Address: coppervm.InstAddr(b.evaluatedWord.asInstAddr),
				})
			}
		}
	}

	var prog []coppervm.InstDef
	for _, i := range cg.program {
		prog = append(prog, coppervm.InstDef{
			Kind:       i.Kind,
			HasOperand: i.HasOperand,
			Name:       i.Name,
			Operand:    i.Operand.toCoppervmWord(),
		})
	}
	meta := coppervm.FileMeta(cg.entry, prog, cg.memory, dbSymbols)
	metaJson, err := json.Marshal(meta)
	if err != nil {
		panic(fmt.Errorf("error writing program to file %s", err))
	}

	return string(metaJson)
}

// Generate a coppervm program from ir.
func (cg *copperGenerator) generateProgram(irs []IR) {
	// First pass
	cg.firstPass(irs)

	// Second pass
	cg.secondPass()
}

// Do the first pass in the parsing process.
func (cgen *copperGenerator) firstPass(irs []IR) {
	for _, ir := range irs {
		switch ir.Kind {
		case IRKindLabel:
			cgen.bindLabel(ir.AsLabel, len(cgen.program), ir.Location)
		case IRKindInstruction:
			inst := ir.AsInstruction
			_, instDef := GetInstructionByName(inst.Name)

			if instDef.HasOperand {
				if isExpressionBinding(inst.Operand) {
					cgen.pushDeferredOperandFromExpression(inst.Operand, len(cgen.program), ir.Location)
					if cgen.deferredExpressions == nil {
						cgen.deferredExpressions = make(map[int]Expression)
					}
					cgen.deferredExpressions[len(cgen.program)] = inst.Operand
				} else {
					instDef.Operand = cgen.evaluateExpression(inst.Operand, ir.Location).Word
				}
			}
			cgen.program = append(cgen.program, instDef)
		case IRKindEntry:
			cgen.bindEntry(ir.AsEntry, ir.Location)
		case IRKindConst:
			cgen.bindConst(ir.AsConst, ir.Location)
		case IRKindMemory:
			cgen.bindMemory(ir.AsMemory, ir.Location)
		}
	}
}

// Push all deferred operands of an expression.
// Note: There could be more than one operand because of the binary operations.
func (cgen *copperGenerator) pushDeferredOperandFromExpression(expr Expression, address int, location FileLocation) {
	switch expr.Kind {
	case ExpressionKindNumLitInt,
		ExpressionKindNumLitFloat,
		ExpressionKindStringLit,
		ExpressionKindByteList:
	case ExpressionKindBinaryOp:
		cgen.pushDeferredOperandFromExpression(*expr.AsBinaryOp.Lhs, address, location)
		cgen.pushDeferredOperandFromExpression(*expr.AsBinaryOp.Rhs, address, location)
	case ExpressionKindBinding:
		cgen.deferredOperands = append(cgen.deferredOperands, deferredOperand{
			Name:     expr.AsBinding,
			Address:  address,
			Location: location,
		})
	}
}

// Returns true if an expression is a binding, false otherwise.
// This code will check even if a binary operation contains a
// binding as his operand.
func isExpressionBinding(expr Expression) (ret bool) {
	switch expr.Kind {
	case ExpressionKindNumLitInt:
		ret = false
	case ExpressionKindNumLitFloat:
		ret = false
	case ExpressionKindStringLit:
		ret = false
	case ExpressionKindBinaryOp:
		ret = isExpressionBinding(*expr.AsBinaryOp.Lhs) ||
			isExpressionBinding(*expr.AsBinaryOp.Rhs)
	case ExpressionKindBinding:
		ret = true
	case ExpressionKindByteList:
		ret = false
	}
	return ret
}

// Do the second pass in the parsing process.
func (cgen *copperGenerator) secondPass() {
	for _, deferredOp := range cgen.deferredOperands {
		exist, binding := cgen.getBindingByName(deferredOp.Name)
		if !exist {
			panic(fmt.Sprintf("%s: unknown binding '%s'",
				deferredOp.Location,
				deferredOp.Name))
		}
		cgen.evaluateBinding(binding, deferredOp.Location)
		var expr Expression
		var ok bool
		if expr, ok = cgen.deferredExpressions[deferredOp.Address]; !ok {
			panic(fmt.Sprintf("%s: cannot find deferred expression at address '%d'", deferredOp.Location, deferredOp.Address))
		}
		cgen.program[deferredOp.Address].Operand = cgen.evaluateExpression(expr, deferredOp.Location).Word
	}

	// Print all the bindings
	if internal.DebugPrintEnabled() {
		internal.DebugPrint("[INFO]: bindings:\n")
		for _, b := range cgen.bindings {
			internal.DebugPrint("  %s\n", b)
		}
	}

	// Resolve entry point
	if cgen.hasEntry && cgen.deferredEntryName != "" {
		exist, binding := cgen.getBindingByName(cgen.deferredEntryName)
		if !exist {
			panic(fmt.Sprintf("%s: unknown binding '%s'",
				cgen.entryLocation,
				cgen.deferredEntryName))
		}

		if binding.value.Kind != ExpressionKindNumLitInt {
			panic(fmt.Sprintf("%s: only label names can be set as entry point",
				cgen.entryLocation))
		}
		entry := cgen.evaluateBinding(binding, cgen.entryLocation).Word
		cgen.entry = int(entry.asInt)
	}

	// Check if at least one halt instruction exists
	hasHalt := false
	for _, inst := range cgen.program {
		if inst.Kind == coppervm.InstHalt {
			hasHalt = true
		}
	}
	if !hasHalt {
		fmt.Printf("[WARN]: no 'halt' instruction found in the program! This program could not work as intended.\n")
	}
}

// Returns a binding by its name.
// If the binding exist the first return parameter will be true,
// otherwise it'll be null.
func (cgen *copperGenerator) getBindingByName(name string) (bool, binding) {
	for _, b := range cgen.bindings {
		if b.name == name {
			return true, b
		}
	}
	return false, binding{}
}

// Returns the index of a binding by it's name.
// If the binding doesn't exist -1 is returned.
func (cgen *copperGenerator) getBindingIndexByName(name string) int {
	for idx, b := range cgen.bindings {
		if b.name == name {
			return idx
		}
	}
	return -1
}

// Binds a label.
func (cgen *copperGenerator) bindLabel(label LabelIR, address int, location FileLocation) {
	exist, b := cgen.getBindingByName(label.Name)
	if exist {
		panic(fmt.Sprintf("%s: label name '%s' is already bound at location '%s'",
			location,
			label.Name,
			b.location))
	}

	cgen.bindings = append(cgen.bindings, binding{
		status:        bindingEvaluated,
		name:          label.Name,
		evaluatedWord: wordInstAddr(int64(address)),
		evaluatedKind: ExpressionKindNumLitInt,
		location:      location,
		isLabel:       true,
	})
}

// Binds a constant.
func (cgen *copperGenerator) bindConst(constIR ConstIR, location FileLocation) {
	exist, b := cgen.getBindingByName(constIR.Name)
	if exist {
		panic(fmt.Sprintf("%s: constant name '%s' is already bound at location '%s'",
			location,
			constIR.Name,
			b.location))
	}

	newBinding := binding{
		status:   bindingUnevaluated,
		name:     constIR.Name,
		value:    constIR.Value,
		location: location,
		isLabel:  false,
	}

	// If it's a const string push it in memory and bind his base address
	if constIR.Value.Kind == ExpressionKindStringLit {
		baseAddr := cgen.pushStringToMemory(constIR.Value.AsStringLit)
		newBinding.evaluatedWord = wordMemoryAddr(int64(baseAddr))
		newBinding.status = bindingEvaluated
		newBinding.evaluatedKind = ExpressionKindStringLit
	}

	cgen.bindings = append(cgen.bindings, newBinding)
}

// Binds an entry point.
func (cgen *copperGenerator) bindEntry(entry EntryIR, location FileLocation) {
	if cgen.hasEntry {
		panic(fmt.Sprintf("%s: entry point is already set to '%s'",
			location,
			cgen.entryLocation))
	}

	cgen.deferredEntryName = entry.Name
	cgen.hasEntry = true
	cgen.entryLocation = location
}

// Binds a memory definition.
func (cgen *copperGenerator) bindMemory(memory MemoryIR, location FileLocation) {
	exist, b := cgen.getBindingByName(memory.Name)
	if exist {
		panic(fmt.Sprintf("%s: memory name '%s' is already bound at location '%s'",
			location,
			memory.Name,
			b.location))
	}

	if memory.Value.Kind != ExpressionKindByteList {
		panic(fmt.Sprintf("%s: expected '%s' but got '%s'",
			location, ExpressionKindByteList, memory.Value.Kind))
	}
	memAddr := len(cgen.memory)
	cgen.memory = append(cgen.memory, memory.Value.AsByteList...)

	cgen.bindings = append(cgen.bindings, binding{
		status:        bindingEvaluated,
		name:          memory.Name,
		evaluatedWord: wordMemoryAddr(int64(memAddr)),
		evaluatedKind: ExpressionKindNumLitInt,
		location:      location,
		isLabel:       false,
	})
}

// Represent the result of an expression evaluation.
type evalResult struct {
	Word word
	Type ExpressionKind
}

// Evaluate a binding to extract am eval result.
func (cgen *copperGenerator) evaluateBinding(binding binding, location FileLocation) (ret evalResult) {
	switch binding.status {
	case bindingUnevaluated:
		idx := cgen.getBindingIndexByName(binding.name)
		if idx == -1 {
			panic(fmt.Sprintf("%s: cannot find index binding %s", location, binding.name))
		}
		cgen.bindings[idx].status = bindingEvaluating
		ret = cgen.evaluateExpression(binding.value, location)
		cgen.bindings[idx].status = bindingEvaluated
		cgen.bindings[idx].evaluatedWord = ret.Word
		cgen.bindings[idx].evaluatedKind = ret.Type
	case bindingEvaluating:
		panic(fmt.Sprintf("%s: cycling binding definition detected", location))
	case bindingEvaluated:
		ret = evalResult{
			binding.evaluatedWord,
			binding.evaluatedKind,
		}
	}
	internal.DebugPrint("[INFO]: evaluated binding with result %s\n", ret)
	return ret
}

// Evaluate an expression to extract an eval result.
func (cgen *copperGenerator) evaluateExpression(expr Expression, location FileLocation) (ret evalResult) {
	switch expr.Kind {
	case ExpressionKindBinding:
		exist, binding := cgen.getBindingByName(expr.AsBinding)
		if !exist {
			panic(fmt.Sprintf("%s: cannot find binding '%s'", location, expr.AsBinding))
		}
		ret = cgen.evaluateBinding(binding, location)
	case ExpressionKindNumLitInt:
		ret = evalResult{
			wordInt(expr.AsNumLitInt),
			ExpressionKindNumLitInt,
		}
	case ExpressionKindNumLitFloat:
		ret = evalResult{
			wordFloat(expr.AsNumLitFloat),
			ExpressionKindNumLitFloat,
		}
	case ExpressionKindStringLit:
		strBase := cgen.pushStringToMemory(expr.AsStringLit)
		ret = evalResult{
			wordMemoryAddr(int64(strBase)),
			ExpressionKindStringLit,
		}
	case ExpressionKindBinaryOp:
		ret = cgen.evaluateBinaryOp(expr, location)
	case ExpressionKindByteList:
		panic(fmt.Sprintf("%s: cannot use byte lists as operands, only supported use is in memory directives", location))
	}
	internal.DebugPrint("[INFO]: evaluated expression with result %s\n", ret)
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
var binaryOpEvaluationMap = [6][6]ExpressionKind{
	{ExpressionKindNumLitInt, ExpressionKindNumLitFloat, -1, -1, -1, -1},
	{ExpressionKindNumLitFloat, ExpressionKindNumLitFloat, -1, -1, -1, -1},
	{-1, -1, ExpressionKindStringLit, -1, -1, -1},
	{-1, -1, -1, -1, -1, -1},
	{-1, -1, -1, -1, -1, -1},
	{-1, -1, -1, -1, -1, -1},
}

// Evaluate a binary op expression to extract an eval result.
func (cgen *copperGenerator) evaluateBinaryOp(binop Expression, location FileLocation) (result evalResult) {
	lhs_result := cgen.evaluateExpression(*binop.AsBinaryOp.Lhs, location)
	rhs_result := cgen.evaluateExpression(*binop.AsBinaryOp.Rhs, location)

	resultType := binaryOpEvaluationMap[lhs_result.Type][rhs_result.Type]
	if resultType == -1 {
		panic(fmt.Sprintf("%s: unsupported binary operation between types '%s' and '%s'",
			location,
			lhs_result.Type,
			rhs_result.Type))
	}

	// At this point the only permitted operations are between
	// int-int int-float float-int float-float string-string
	// so we can reduce the next checks.

	if resultType == ExpressionKindStringLit {
		// The op is string-string
		if binop.AsBinaryOp.Kind != BinaryOpKindPlus {
			panic(fmt.Sprintf("%s: unsupported operations ['-', '*', '/', '%%'] between string literals",
				location))
		}
		leftStr := cgen.getStringByAddress(int(lhs_result.Word.asMemoryAddr))
		rightStr := cgen.getStringByAddress(int(rhs_result.Word.asMemoryAddr))
		result = evalResult{
			wordMemoryAddr(int64(cgen.pushStringToMemory(leftStr + rightStr))),
			ExpressionKindStringLit,
		}
	} else {
		// The only ops at this point are int-float float-int.
		// int-int and float-float are removed because in Expression we precompute
		// the operations with same type
		switch binop.AsBinaryOp.Kind {
		case BinaryOpKindPlus:
			result = evalResult{addWord(lhs_result.Word, rhs_result.Word), resultType}
		case BinaryOpKindMinus:
			result = evalResult{subWord(lhs_result.Word, rhs_result.Word), resultType}
		case BinaryOpKindTimes:
			result = evalResult{mulWord(lhs_result.Word, rhs_result.Word), resultType}
		case BinaryOpKindDivide:
			result = evalResult{divWord(lhs_result.Word, rhs_result.Word), resultType}
		case BinaryOpKindModulo:
			// Since the only pos are int-float and float-int allways panic
			panic(fmt.Sprintf("%s: unsupported '%%' operation between floating point literals", location))
		}
	}
	return result
}

// Push a string to memory and return the base address.
func (cgen *copperGenerator) pushStringToMemory(str string) int {
	strBase := len(cgen.memory)
	byteStr := []byte(str)
	byteStr = append(byteStr, 0)
	cgen.memory = append(cgen.memory, byteStr...)

	if cgen.stringLengths == nil {
		cgen.stringLengths = make(map[int]int)
	}
	cgen.stringLengths[strBase] = len(byteStr)
	return strBase
}

// Returns a string from memory at given address without
// null termination.
// If the string doesn't exist an empty string is returned.
func (cgen *copperGenerator) getStringByAddress(addr int) string {
	strLen := cgen.stringLengths[addr]
	if strLen == 0 {
		return ""
	}
	strBytes := cgen.memory[addr : addr+strLen-1]
	return string(strBytes[:])
}
