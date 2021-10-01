package casm

import (
	"fmt"

	"github.com/Supercaly/coppervm/internal"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type internalRep struct {
	bindings            []binding
	deferredOperands    []deferredOperand
	deferredExpressions map[int]Expression

	hasEntry          bool
	entry             int
	entryLocation     FileLocation
	deferredEntryName string

	stringLengths map[int]int

	program []instruction
	memory  []byte
}

// Do the first pass in the parsing process.
func (rep *internalRep) firstPass(irs []IR) {
	for _, ir := range irs {
		switch ir.Kind {
		case IRKindLabel:
			rep.bindLabel(ir.AsLabel, len(rep.program), ir.Location)
		case IRKindInstruction:
			inst := ir.AsInstruction
			_, instDef := getInstructionByName(inst.Name)

			if instDef.hasOperand {
				if isExpressionBinding(inst.Operand) {
					rep.pushDeferredOperandFromExpression(inst.Operand, len(rep.program), ir.Location)
					if rep.deferredExpressions == nil {
						rep.deferredExpressions = make(map[int]Expression)
					}
					rep.deferredExpressions[len(rep.program)] = inst.Operand
				} else {
					instDef.operand = rep.evaluateExpression(inst.Operand, ir.Location).Word
				}
			}
			rep.program = append(rep.program, instDef)
		case IRKindEntry:
			rep.bindEntry(ir.AsEntry, ir.Location)
		case IRKindConst:
			rep.bindConst(ir.AsConst, ir.Location)
		case IRKindMemory:
			rep.bindMemory(ir.AsMemory, ir.Location)
		}
	}
}

// Push all deferred operands of an expression.
// Note: There could be more than one operand because of the binary operations.
func (rep *internalRep) pushDeferredOperandFromExpression(expr Expression, address int, location FileLocation) {
	switch expr.Kind {
	case ExpressionKindNumLitInt,
		ExpressionKindNumLitFloat,
		ExpressionKindStringLit,
		ExpressionKindByteList:
	case ExpressionKindBinaryOp:
		rep.pushDeferredOperandFromExpression(*expr.AsBinaryOp.Lhs, address, location)
		rep.pushDeferredOperandFromExpression(*expr.AsBinaryOp.Rhs, address, location)
	case ExpressionKindBinding:
		rep.deferredOperands = append(rep.deferredOperands, deferredOperand{
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
func (rep *internalRep) secondPass() {
	for _, deferredOp := range rep.deferredOperands {
		exist, binding := rep.getBindingByName(deferredOp.Name)
		if !exist {
			panic(fmt.Sprintf("%s: unknown binding '%s'",
				deferredOp.Location,
				deferredOp.Name))
		}
		rep.evaluateBinding(binding, deferredOp.Location)
		var expr Expression
		var ok bool
		if expr, ok = rep.deferredExpressions[deferredOp.Address]; !ok {
			panic(fmt.Sprintf("%s: cannot find deferred expression at address '%d'", deferredOp.Location, deferredOp.Address))
		}
		rep.program[deferredOp.Address].operand = rep.evaluateExpression(expr, deferredOp.Location).Word
	}

	// Print all the bindings
	if internal.DebugPrintEnabled() {
		internal.DebugPrint("[INFO]: bindings:\n")
		for _, b := range rep.bindings {
			internal.DebugPrint("  %s\n", b)
		}
	}

	// Resolve entry point
	if rep.hasEntry && rep.deferredEntryName != "" {
		exist, binding := rep.getBindingByName(rep.deferredEntryName)
		if !exist {
			panic(fmt.Sprintf("%s: unknown binding '%s'",
				rep.entryLocation,
				rep.deferredEntryName))
		}

		if binding.value.Kind != ExpressionKindNumLitInt {
			panic(fmt.Sprintf("%s: only label names can be set as entry point",
				rep.entryLocation))
		}
		entry := rep.evaluateBinding(binding, rep.entryLocation).Word
		rep.entry = int(entry.asInstAddr)
	}

	// Check if at least one halt instruction exists
	hasHalt := false
	for _, inst := range rep.program {
		if inst.kind == coppervm.InstHalt {
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
func (rep *internalRep) getBindingByName(name string) (bool, binding) {
	for _, b := range rep.bindings {
		if b.name == name {
			return true, b
		}
	}
	return false, binding{}
}

// Returns the index of a binding by it's name.
// If the binding doesn't exist -1 is returned.
func (rep *internalRep) getBindingIndexByName(name string) int {
	for idx, b := range rep.bindings {
		if b.name == name {
			return idx
		}
	}
	return -1
}

// Binds a label.
func (rep *internalRep) bindLabel(label LabelIR, address int, location FileLocation) {
	exist, b := rep.getBindingByName(label.Name)
	if exist {
		panic(fmt.Sprintf("%s: label name '%s' is already bound at location '%s'",
			location,
			label.Name,
			b.location))
	}

	rep.bindings = append(rep.bindings, binding{
		status:        bindingEvaluated,
		name:          label.Name,
		evaluatedWord: wordInstAddr(int64(address)),
		evaluatedKind: ExpressionKindNumLitInt,
		location:      location,
		isLabel:       true,
	})
}

// Binds a constant.
func (rep *internalRep) bindConst(constIR ConstIR, location FileLocation) {
	exist, b := rep.getBindingByName(constIR.Name)
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
		baseAddr := rep.pushStringToMemory(constIR.Value.AsStringLit)
		newBinding.evaluatedWord = wordMemoryAddr(int64(baseAddr))
		newBinding.status = bindingEvaluated
		newBinding.evaluatedKind = ExpressionKindStringLit
	}

	rep.bindings = append(rep.bindings, newBinding)
}

// Binds an entry point.
func (rep *internalRep) bindEntry(entry EntryIR, location FileLocation) {
	if rep.hasEntry {
		panic(fmt.Sprintf("%s: entry point is already set to '%s'",
			location,
			rep.entryLocation))
	}

	rep.deferredEntryName = entry.Name
	rep.hasEntry = true
	rep.entryLocation = location
}

// Binds a memory definition.
func (rep *internalRep) bindMemory(memory MemoryIR, location FileLocation) {
	exist, b := rep.getBindingByName(memory.Name)
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
	memAddr := len(rep.memory)
	rep.memory = append(rep.memory, memory.Value.AsByteList...)

	rep.bindings = append(rep.bindings, binding{
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
func (rep *internalRep) evaluateBinding(binding binding, location FileLocation) (ret evalResult) {
	switch binding.status {
	case bindingUnevaluated:
		idx := rep.getBindingIndexByName(binding.name)
		if idx == -1 {
			panic(fmt.Sprintf("%s: cannot find index binding %s", location, binding.name))
		}
		rep.bindings[idx].status = bindingEvaluating
		ret = rep.evaluateExpression(binding.value, location)
		rep.bindings[idx].status = bindingEvaluated
		rep.bindings[idx].evaluatedWord = ret.Word
		rep.bindings[idx].evaluatedKind = ret.Type
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
func (rep *internalRep) evaluateExpression(expr Expression, location FileLocation) (ret evalResult) {
	switch expr.Kind {
	case ExpressionKindBinding:
		exist, binding := rep.getBindingByName(expr.AsBinding)
		if !exist {
			panic(fmt.Sprintf("%s: cannot find binding '%s'", location, expr.AsBinding))
		}
		ret = rep.evaluateBinding(binding, location)
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
		strBase := rep.pushStringToMemory(expr.AsStringLit)
		ret = evalResult{
			wordMemoryAddr(int64(strBase)),
			ExpressionKindStringLit,
		}
	case ExpressionKindBinaryOp:
		ret = rep.evaluateBinaryOp(expr, location)
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
func (rep *internalRep) evaluateBinaryOp(binop Expression, location FileLocation) (result evalResult) {
	lhs_result := rep.evaluateExpression(*binop.AsBinaryOp.Lhs, location)
	rhs_result := rep.evaluateExpression(*binop.AsBinaryOp.Rhs, location)

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
		leftStr := rep.getStringByAddress(int(lhs_result.Word.asMemoryAddr))
		rightStr := rep.getStringByAddress(int(rhs_result.Word.asMemoryAddr))
		result = evalResult{
			wordMemoryAddr(int64(rep.pushStringToMemory(leftStr + rightStr))),
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
func (rep *internalRep) pushStringToMemory(str string) int {
	strBase := len(rep.memory)
	byteStr := []byte(str)
	byteStr = append(byteStr, 0)
	rep.memory = append(rep.memory, byteStr...)

	if rep.stringLengths == nil {
		rep.stringLengths = make(map[int]int)
	}
	rep.stringLengths[strBase] = len(byteStr)
	return strBase
}

// Returns a string from memory at given address without
// null termination.
// If the string doesn't exist an empty string is returned.
func (rep *internalRep) getStringByAddress(addr int) string {
	strLen := rep.stringLengths[addr]
	if strLen == 0 {
		return ""
	}
	strBytes := rep.memory[addr : addr+strLen-1]
	return string(strBytes[:])
}
