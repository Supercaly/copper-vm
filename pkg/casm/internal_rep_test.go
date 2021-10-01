package casm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBindingByName(t *testing.T) {
	rep := internalRep{
		bindings: []binding{
			{name: "a_label",
				value:    expression(ExpressionKindNumLitInt, int64(1)),
				location: FileLocation{},
				isLabel:  true},
			{name: "a_const",
				value:    expression(ExpressionKindNumLitFloat, 2.3),
				location: FileLocation{},
				isLabel:  false},
		},
	}

	exist, binding := rep.getBindingByName("a_label")
	assert.True(t, exist)
	assert.Equal(t, rep.bindings[0], binding)

	exist, binding = rep.getBindingByName("a_const")
	assert.True(t, exist)
	assert.Equal(t, rep.bindings[1], binding)

	exist, binding = rep.getBindingByName("test")
	assert.False(t, exist)
}

func TestGetBindingIndexByName(t *testing.T) {
	rep := internalRep{
		bindings: []binding{
			{name: "a_label",
				value:    expression(ExpressionKindNumLitInt, int64(1)),
				location: FileLocation{},
				isLabel:  true},
			{name: "a_const",
				value:    expression(ExpressionKindNumLitFloat, 2.3),
				location: FileLocation{},
				isLabel:  false},
		},
	}

	index := rep.getBindingIndexByName("a_label")
	assert.Equal(t, 0, index)

	index = rep.getBindingIndexByName("a_const")
	assert.Equal(t, 1, index)

	index = rep.getBindingIndexByName("test")
	assert.Equal(t, -1, index)
}

func TestBindLabel(t *testing.T) {
	rep := internalRep{
		bindings: []binding{
			{name: "a_label",
				value:    expression(ExpressionKindNumLitInt, int64(1)),
				location: FileLocation{},
				isLabel:  true},
		},
	}

	func() {
		defer func() { recover() }()
		rep.bindLabel(LabelIR{"a_label"}, 1, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	rep.bindLabel(LabelIR{"new_label"}, 2, FileLocation{})
	exist, b := rep.getBindingByName("new_label")
	assert.True(t, exist)

	label := binding{
		name:          "new_label",
		evaluatedWord: wordInstAddr(2),
		location:      FileLocation{},
		isLabel:       true,
		status:        bindingEvaluated,
	}
	assert.Equal(t, label, b)
}

func TestBindConst(t *testing.T) {
	tests := []struct {
		constIR      ConstIR
		binding      binding
		memoryLength int
		hasError     bool
	}{
		{constIR: ConstIR{Name: "a_const", Value: expression(ExpressionKindNumLitInt, int64(1))}, hasError: true},
		{constIR: ConstIR{Name: "new_const", Value: expression(ExpressionKindNumLitInt, int64(2))}, hasError: false, binding: binding{
			name:     "new_const",
			value:    expression(ExpressionKindNumLitInt, int64(2)),
			location: FileLocation{},
			isLabel:  false},
		},
		{constIR: ConstIR{Name: "str_const", Value: expression(ExpressionKindStringLit, `"test_str"`)}, hasError: false, binding: binding{
			status:        bindingEvaluated,
			name:          "str_const",
			value:         expression(ExpressionKindStringLit, `"test_str"`),
			evaluatedWord: wordMemoryAddr(0),
			evaluatedKind: ExpressionKindStringLit,
			location:      FileLocation{},
			isLabel:       false},
			memoryLength: 11,
		},
	}

	for _, test := range tests {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			rep := internalRep{
				bindings: []binding{
					{name: "a_const",
						value:    Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 1},
						location: FileLocation{},
						isLabel:  false},
				},
			}
			rep.bindConst(test.constIR, FileLocation{})

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			} else {
				exist, binding := rep.getBindingByName(test.constIR.Name)
				assert.True(t, exist, test)
				assert.Equal(t, test.binding, binding, test)
				assert.Equal(t, test.memoryLength, len(rep.memory), test)
			}
		}()
	}
}

func TestBindEntry(t *testing.T) {
	func() {
		rep := internalRep{}
		defer func() { recover() }()
		rep.hasEntry = true
		rep.bindEntry(EntryIR{"main"}, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	rep := internalRep{}
	rep.bindEntry(EntryIR{"entry"}, fileLocation(10, 0))
	assert.True(t, rep.hasEntry)
	assert.Equal(t, "entry", rep.deferredEntryName)
	assert.Equal(t, 10, rep.entryLocation.Row)
}

func TestBindMemory(t *testing.T) {
	rep := internalRep{
		bindings: []binding{
			{name: "mem",
				value:    expression(ExpressionKindNumLitInt, int64(0)),
				location: FileLocation{},
				isLabel:  false},
		},
	}

	func() {
		defer func() { recover() }()
		rep.bindMemory(MemoryIR{Name: "mem", Value: expression(ExpressionKindByteList, []byte{1})}, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	rep.bindMemory(MemoryIR{Name: "new_mem", Value: expression(ExpressionKindByteList, []byte{2, 3})}, FileLocation{})
	exist, b := rep.getBindingByName("new_mem")
	assert.True(t, exist)

	want := binding{
		status:        bindingEvaluated,
		name:          "new_mem",
		evaluatedWord: wordMemoryAddr(0),
		location:      FileLocation{},
		isLabel:       false,
	}
	assert.Equal(t, want, b)
}

var evaluateExpressionsTests = []struct {
	expr     Expression
	res      word
	hasError bool
}{
	{expr: expression(ExpressionKindNumLitInt, int64(10)), res: wordInt(10)},
	{expr: expression(ExpressionKindNumLitFloat, 2.3), res: wordFloat(2.3)},
	{expr: expression(ExpressionKindStringLit, "str"), res: wordMemoryAddr(0)},
	{expr: expression(ExpressionKindBinding, "a_bind"), res: wordInt(3)},
	{expr: expression(ExpressionKindBinding, "different_bind"), res: wordInstAddr(3), hasError: true},
	{expr: expression(ExpressionKindByteList, []byte{1, 2, 3}), hasError: true},
}

func TestEvaluateExpression(t *testing.T) {
	rep := internalRep{}
	rep.bindings = append(rep.bindings, binding{
		name:  "a_bind",
		value: expression(ExpressionKindNumLitInt, int64(3)),
	})

	for _, test := range evaluateExpressionsTests {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			result := rep.evaluateExpression(test.expr, FileLocation{}).Word

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			} else {
				assert.Equal(t, test.res, result, test)
			}
		}()
	}
}

var binaryOpTests = []struct {
	expr     Expression
	res      word
	hasError bool
}{
	// invalid binary op between types and float
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindStringLit, "str"),
		Rhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinding, "bind"),
		Rhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
		Rhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindByteList, []byte{}),
		Rhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
	}), hasError: true},

	// invalid binary op between types and float
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindStringLit, "str"),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 2.1),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinding, "bind"),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 2.1),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 2.1),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindByteList, []byte{}),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 2.0),
	}), hasError: true},

	// invalid binary op between types and string
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindStringLit, "str"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitFloat, 2.1),
		Rhs:  expressionP(ExpressionKindStringLit, "str"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinding, "bind"),
		Rhs:  expressionP(ExpressionKindStringLit, "str"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
		Rhs:  expressionP(ExpressionKindStringLit, "str"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindByteList, []byte{}),
		Rhs:  expressionP(ExpressionKindStringLit, "str"),
	}), hasError: true},

	// invalid binary op between types and binding
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindBinding, "bind"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitFloat, 2.1),
		Rhs:  expressionP(ExpressionKindBinding, "bind"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindStringLit, "str"),
		Rhs:  expressionP(ExpressionKindBinding, "bind"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinding, "bind"),
		Rhs:  expressionP(ExpressionKindBinding, "bind"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
		Rhs:  expressionP(ExpressionKindBinding, "bind"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindByteList, []byte{}),
		Rhs:  expressionP(ExpressionKindBinding, "bind"),
	}), hasError: true},

	// invalid binary op between types and binop
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitFloat, 2.1),
		Rhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindStringLit, "str"),
		Rhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinding, "bind"),
		Rhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
		Rhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindByteList, []byte{}),
		Rhs:  expressionP(ExpressionKindBinaryOp, BinaryOp{}),
	}), hasError: true},
	// invalid binary op between types and byte list
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindByteList, []byte{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitFloat, 2.1),
		Rhs:  expressionP(ExpressionKindByteList, []byte{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindStringLit, "str"),
		Rhs:  expressionP(ExpressionKindByteList, []byte{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindBinding, "bind"),
		Rhs:  expressionP(ExpressionKindByteList, []byte{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindByteList, []byte{}),
		Rhs:  expressionP(ExpressionKindByteList, []byte{}),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindByteList, []byte{}),
		Rhs:  expressionP(ExpressionKindByteList, []byte{}),
	}), hasError: true},

	// invalid binary op between strings
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindMinus,
		Lhs:  expressionP(ExpressionKindStringLit, "str1"),
		Rhs:  expressionP(ExpressionKindStringLit, "str1"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindTimes,
		Lhs:  expressionP(ExpressionKindStringLit, "str1"),
		Rhs:  expressionP(ExpressionKindStringLit, "str1"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindDivide,
		Lhs:  expressionP(ExpressionKindStringLit, "str1"),
		Rhs:  expressionP(ExpressionKindStringLit, "str1"),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindModulo,
		Lhs:  expressionP(ExpressionKindStringLit, "str1"),
		Rhs:  expressionP(ExpressionKindStringLit, "str1"),
	}), hasError: true},

	// valid binary op
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindStringLit, "str1"),
		Rhs:  expressionP(ExpressionKindStringLit, "str2"),
	}), res: wordMemoryAddr(10)},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 5.2),
	}), res: wordFloat(7.2)},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindMinus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 5.2),
	}), res: wordFloat(-3.2)},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindTimes,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 5.3),
	}), res: wordFloat(10.6)},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindDivide,
		Lhs:  expressionP(ExpressionKindNumLitFloat, 5.0),
		Rhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
	}), res: wordFloat(2.5)},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindDivide,
		Lhs:  expressionP(ExpressionKindNumLitFloat, 5.0),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 0.0),
	}), hasError: true},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindModulo,
		Rhs:  expressionP(ExpressionKindNumLitFloat, 5.0),
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
	}), hasError: true},
}

func TestEvaluateBinaryOp(t *testing.T) {
	for _, test := range binaryOpTests {
		rep := internalRep{}
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			word := rep.evaluateBinaryOp(test.expr, FileLocation{}).Word

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			} else {
				assert.Equal(t, test.res, word, test)
			}
		}()
	}
}

func TestEvaluateBinding(t *testing.T) {
	rep := internalRep{
		bindings: []binding{
			{name: "a_bind", value: expression(ExpressionKindNumLitInt, int64(5))},
			{name: "cycl1", value: expression(ExpressionKindBinding, "cycl2")},
			{name: "cycl2", value: expression(ExpressionKindBinding, "cycl1")},
		},
	}

	word := rep.evaluateBinding(rep.bindings[0], FileLocation{}).Word
	assert.Equal(t, wordInt(5), word)

	word = rep.evaluateBinding(rep.bindings[0], FileLocation{}).Word
	assert.Equal(t, wordInt(5), word)

	func() {
		defer func() { recover() }()
		rep.evaluateBinding(rep.bindings[1], FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	func() {
		defer func() { recover() }()
		rep.evaluateBinding(binding{name: "bind"}, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()
}

func TestPushStringToMemory(t *testing.T) {
	rep := internalRep{}

	strAddr := rep.pushStringToMemory("string1")
	assert.Equal(t, 0, strAddr)
	assert.Equal(t, 8, rep.stringLengths[0])

	strAddr = rep.pushStringToMemory("a different string")
	assert.Equal(t, 8, strAddr)
	assert.Equal(t, 19, rep.stringLengths[8])
}

func TestGetStringByAddress(t *testing.T) {
	rep := internalRep{}
	l1 := rep.pushStringToMemory("string1")
	s1 := rep.getStringByAddress(l1)
	assert.NotEmpty(t, s1)
	assert.Equal(t, "string1", s1)
	assert.Len(t, rep.memory, 8)

	s2 := rep.getStringByAddress(1)
	assert.Empty(t, s2)
}
