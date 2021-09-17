package casm

import (
	"testing"

	"github.com/Supercaly/coppervm/pkg/coppervm"
	"github.com/stretchr/testify/assert"
)

func TestGetBindingByName(t *testing.T) {
	cgen := copperGenerator{
		Bindings: []binding{
			{Name: "a_label",
				Value:    expression(ExpressionKindNumLitInt, int64(1)),
				Location: FileLocation{},
				IsLabel:  true},
			{Name: "a_const",
				Value:    expression(ExpressionKindNumLitFloat, 2.3),
				Location: FileLocation{},
				IsLabel:  false},
		},
	}

	exist, binding := cgen.getBindingByName("a_label")
	assert.True(t, exist)
	assert.Equal(t, cgen.Bindings[0], binding)

	exist, binding = cgen.getBindingByName("a_const")
	assert.True(t, exist)
	assert.Equal(t, cgen.Bindings[1], binding)

	exist, binding = cgen.getBindingByName("test")
	assert.False(t, exist)
}

func TestGetBindingIndexByName(t *testing.T) {
	cgen := copperGenerator{
		Bindings: []binding{
			{Name: "a_label",
				Value:    expression(ExpressionKindNumLitInt, int64(1)),
				Location: FileLocation{},
				IsLabel:  true},
			{Name: "a_const",
				Value:    expression(ExpressionKindNumLitFloat, 2.3),
				Location: FileLocation{},
				IsLabel:  false},
		},
	}

	index := cgen.getBindingIndexByName("a_label")
	assert.Equal(t, 0, index)

	index = cgen.getBindingIndexByName("a_const")
	assert.Equal(t, 1, index)

	index = cgen.getBindingIndexByName("test")
	assert.Equal(t, -1, index)
}

func TestBindLabel(t *testing.T) {
	cgen := copperGenerator{
		Bindings: []binding{
			{Name: "a_label",
				Value:    expression(ExpressionKindNumLitInt, int64(1)),
				Location: FileLocation{},
				IsLabel:  true},
		},
	}

	func() {
		defer func() { recover() }()
		cgen.bindLabel(LabelIR{"a_label"}, 1, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	cgen.bindLabel(LabelIR{"new_label"}, 2, FileLocation{})
	exist, b := cgen.getBindingByName("new_label")
	assert.True(t, exist)

	label := binding{
		Name:          "new_label",
		EvaluatedWord: coppervm.WordU64(2),
		Location:      FileLocation{},
		IsLabel:       true,
		Status:        bindingEvaluated,
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
			Name:     "new_const",
			Value:    expression(ExpressionKindNumLitInt, int64(2)),
			Location: FileLocation{},
			IsLabel:  false},
		},
		{constIR: ConstIR{Name: "str_const", Value: expression(ExpressionKindStringLit, `"test_str"`)}, hasError: false, binding: binding{
			Status:        bindingEvaluated,
			Name:          "str_const",
			Value:         expression(ExpressionKindStringLit, `"test_str"`),
			EvaluatedWord: coppervm.WordU64(0),
			Location:      FileLocation{},
			IsLabel:       false},
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

			cgen := copperGenerator{
				Bindings: []binding{
					{Name: "a_const",
						Value:    Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 1},
						Location: FileLocation{},
						IsLabel:  false},
				},
			}
			cgen.bindConst(test.constIR, FileLocation{})

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			} else {
				exist, binding := cgen.getBindingByName(test.constIR.Name)
				assert.True(t, exist, test)
				assert.Equal(t, test.binding, binding, test)
				assert.Equal(t, test.memoryLength, len(cgen.Memory), test)
			}
		}()
	}
}

func TestBindEntry(t *testing.T) {
	func() {
		cgen := copperGenerator{}
		defer func() { recover() }()
		cgen.HasEntry = true
		cgen.bindEntry(EntryIR{"main"}, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	cgen := copperGenerator{}
	cgen.bindEntry(EntryIR{"entry"}, FileLocation{FileName: "", Location: 10})
	assert.True(t, cgen.HasEntry)
	assert.Equal(t, "entry", cgen.DeferredEntryName)
	assert.Equal(t, 10, cgen.EntryLocation.Location)
}

func TestBindMemory(t *testing.T) {
	cgen := copperGenerator{
		Bindings: []binding{
			{Name: "mem",
				Value:    expression(ExpressionKindNumLitInt, int64(0)),
				Location: FileLocation{},
				IsLabel:  false},
		},
	}

	func() {
		defer func() { recover() }()
		cgen.bindMemory(MemoryIR{Name: "mem", Value: expression(ExpressionKindByteList, []byte{1})}, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	cgen.bindMemory(MemoryIR{Name: "new_mem", Value: expression(ExpressionKindByteList, []byte{2, 3})}, FileLocation{})
	exist, b := cgen.getBindingByName("new_mem")
	assert.True(t, exist)

	want := binding{
		Status:        bindingEvaluated,
		Name:          "new_mem",
		EvaluatedWord: coppervm.WordU64(0),
		Location:      FileLocation{},
		IsLabel:       false,
	}
	assert.Equal(t, want, b)
}

var evaluateExpressionsTests = []struct {
	expr     Expression
	res      coppervm.Word
	hasError bool
}{
	{expr: expression(ExpressionKindNumLitInt, int64(10)), res: coppervm.WordU64(10)},
	{expr: expression(ExpressionKindNumLitFloat, 2.3), res: coppervm.WordF64(2.3)},
	{expr: expression(ExpressionKindStringLit, "str"), res: coppervm.WordI64(0)},
	{expr: expression(ExpressionKindBinding, "a_bind"), res: coppervm.WordI64(3)},
	{expr: expression(ExpressionKindBinding, "different_bind"), res: coppervm.WordI64(3), hasError: true},
	{expr: expression(ExpressionKindByteList, []byte{1, 2, 3}), hasError: true},
}

func TestEvaluateExpression(t *testing.T) {
	cgen := copperGenerator{}
	cgen.Bindings = append(cgen.Bindings, binding{
		Name:  "a_bind",
		Value: expression(ExpressionKindNumLitInt, int64(3)),
	})

	for _, test := range evaluateExpressionsTests {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			result := cgen.evaluateExpression(test.expr, FileLocation{}).Word

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
	res      coppervm.Word
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
	}), res: coppervm.WordI64(10)},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 5.2),
	}), res: coppervm.WordF64(7.2)},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindMinus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 5.2),
	}), res: coppervm.WordF64(-3.2)},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindTimes,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindNumLitFloat, 5.3),
	}), res: coppervm.WordF64(10.6)},
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindDivide,
		Lhs:  expressionP(ExpressionKindNumLitFloat, 5.0),
		Rhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
	}), res: coppervm.WordF64(2.5)},
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
		cgen := copperGenerator{}
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			word := cgen.evaluateBinaryOp(test.expr, FileLocation{}).Word

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			} else {
				assert.Equal(t, test.res, word, test)
			}
		}()
	}
}

func TestEvaluateBinding(t *testing.T) {
	cgen := copperGenerator{
		Bindings: []binding{
			{Name: "a_bind", Value: expression(ExpressionKindNumLitInt, int64(5))},
			{Name: "cycl1", Value: expression(ExpressionKindBinding, "cycl2")},
			{Name: "cycl2", Value: expression(ExpressionKindBinding, "cycl1")},
		},
	}

	word := cgen.evaluateBinding(cgen.Bindings[0], FileLocation{}).Word
	assert.Equal(t, coppervm.WordU64(5), word)

	word = cgen.evaluateBinding(cgen.Bindings[0], FileLocation{}).Word
	assert.Equal(t, coppervm.WordU64(5), word)

	func() {
		defer func() { recover() }()
		cgen.evaluateBinding(cgen.Bindings[1], FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	func() {
		defer func() { recover() }()
		cgen.evaluateBinding(binding{Name: "bind"}, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()
}

func TestPushStringToMemory(t *testing.T) {
	cgen := copperGenerator{}

	strAddr := cgen.pushStringToMemory("string1")
	assert.Equal(t, 0, strAddr)
	assert.Equal(t, 8, cgen.StringLengths[0])

	strAddr = cgen.pushStringToMemory("a different string")
	assert.Equal(t, 8, strAddr)
	assert.Equal(t, 19, cgen.StringLengths[8])
}

func TestGetStringByAddress(t *testing.T) {
	cgen := copperGenerator{}
	l1 := cgen.pushStringToMemory("string1")
	s1 := cgen.getStringByAddress(l1)
	assert.NotEmpty(t, s1)
	assert.Equal(t, "string1", s1)
	assert.Len(t, cgen.Memory, 8)

	s2 := cgen.getStringByAddress(1)
	assert.Empty(t, s2)
}
