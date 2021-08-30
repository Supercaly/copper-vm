package casm

import (
	"testing"

	"github.com/Supercaly/coppervm/pkg/coppervm"
	"github.com/stretchr/testify/assert"
)

func TestSaveProgramToFile(t *testing.T) {
	tests := []struct {
		casm     Casm
		hasError bool
	}{
		{Casm{OutputFile: "testdata/test.notcopper"}, true},
		{Casm{OutputFile: "testdata/test.copper"}, false},
	}

	for _, test := range tests {
		err := test.casm.SaveProgramToFile()

		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
		}
	}
}

func TestTranslateSourceFile(t *testing.T) {
	tests := []struct {
		path     string
		hasError bool
	}{
		{"testdata/test.notcasm", true},
		{"testdata/test1.casm", true},
		{"testdata/test.casm", false},
	}
	casm := Casm{}

	for _, test := range tests {
		err := casm.TranslateSourceFile(test.path)

		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
		}
	}
}

var testSources = []struct {
	in       string
	out      []coppervm.InstDef
	hasError bool
}{
	{"main:\n", []coppervm.InstDef{}, false},
	{"push 1\n", []coppervm.InstDef{
		{
			Kind:       coppervm.InstPush,
			HasOperand: true,
			Name:       "push",
			Operand:    coppervm.WordU64(1),
		},
	}, false},
	{"%const N 1\n", []coppervm.InstDef{}, false},
	{":", []coppervm.InstDef{}, true},
	{"wrong\n", []coppervm.InstDef{}, true},
	{"push \n", []coppervm.InstDef{}, true},
	{"%dir 0\n", []coppervm.InstDef{}, true},
	{"push N\n", []coppervm.InstDef{
		{
			Kind:       coppervm.InstPush,
			HasOperand: true,
			Name:       "push",
			Operand:    coppervm.WordU64(0),
		},
	}, true},
	{"%entry main\n%const main 2.0", []coppervm.InstDef{}, true},
	{"%entry main\n%entry main2", []coppervm.InstDef{}, true},
	{"%include abc", []coppervm.InstDef{}, true},
}

func TestTranslateSource(t *testing.T) {
	for _, test := range testSources {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			ctx := Casm{}
			ctx.translateSource(test.in, "test_file")

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			}
			assert.Condition(t, func() (success bool) { return instArrayEquals(test.out, ctx.Program) }, test)
		}()
	}
}

// Compares two arrays of InstDef.
func instArrayEquals(a []coppervm.InstDef, b []coppervm.InstDef) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

func TestGetBindingByName(t *testing.T) {
	casm := Casm{
		Bindings: []Binding{
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

	exist, binding := casm.getBindingByName("a_label")
	assert.True(t, exist)
	assert.Equal(t, casm.Bindings[0], binding)

	exist, binding = casm.getBindingByName("a_const")
	assert.True(t, exist)
	assert.Equal(t, casm.Bindings[1], binding)

	exist, binding = casm.getBindingByName("test")
	assert.False(t, exist)
}

func TestGetBindingIndexByName(t *testing.T) {
	casm := Casm{
		Bindings: []Binding{
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

	index := casm.getBindingIndexByName("a_label")
	assert.Equal(t, 0, index)

	index = casm.getBindingIndexByName("a_const")
	assert.Equal(t, 1, index)

	index = casm.getBindingIndexByName("test")
	assert.Equal(t, -1, index)
}

func TestBindLabel(t *testing.T) {
	casm := Casm{
		Bindings: []Binding{
			{Name: "a_label",
				Value:    expression(ExpressionKindNumLitInt, int64(1)),
				Location: FileLocation{},
				IsLabel:  true},
		},
	}

	func() {
		defer func() { recover() }()
		casm.bindLabel("a_label", 1, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	casm.bindLabel("new_label", 2, FileLocation{})
	exist, binding := casm.getBindingByName("new_label")
	assert.True(t, exist)

	label := Binding{
		Name:          "new_label",
		EvaluatedWord: coppervm.WordU64(2),
		Location:      FileLocation{},
		IsLabel:       true,
		Status:        BindingEvaluated,
	}
	assert.Equal(t, label, binding)
}

func TestBindConst(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		binding      Binding
		memoryLength int
		hasError     bool
	}{
		{name: "a_const", value: "1", hasError: true},
		{name: "new_const", value: "2", hasError: false, binding: Binding{
			Name:     "new_const",
			Value:    expression(ExpressionKindNumLitInt, int64(2)),
			Location: FileLocation{},
			IsLabel:  false},
		},
		{name: "new_const2", hasError: true},
		{name: "str_const", value: `"test_str"`, hasError: false, binding: Binding{
			Status:        BindingEvaluated,
			Name:          "str_const",
			Value:         expression(ExpressionKindStringLit, "test_str"),
			EvaluatedWord: coppervm.WordU64(0),
			Location:      FileLocation{},
			IsLabel:       false},
			memoryLength: 9,
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

			casm := Casm{
				Bindings: []Binding{
					{Name: "a_const",
						Value:    Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 1},
						Location: FileLocation{},
						IsLabel:  false},
				},
			}
			casm.bindConst(DirectiveLine{Name: "const", Block: test.name + " " + test.value}, FileLocation{})

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			} else {
				exist, binding := casm.getBindingByName(test.name)
				assert.True(t, exist, test)
				assert.Equal(t, test.binding, binding, test)
				assert.Equal(t, test.memoryLength, len(casm.Memory), test)
			}
		}()
	}
}

func TestBindEntry(t *testing.T) {
	func() {
		casm := Casm{}
		defer func() { recover() }()
		casm.HasEntry = true
		casm.bindEntry("main", FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	casm := Casm{}
	casm.bindEntry("entry", FileLocation{"", 10})
	assert.True(t, casm.HasEntry)
	assert.Equal(t, "entry", casm.DeferredEntryName)
	assert.Equal(t, 10, casm.EntryLocation.Location)
}

func TestBindMemory(t *testing.T) {
	casm := Casm{
		Bindings: []Binding{
			{Name: "mem",
				Value:    expression(ExpressionKindNumLitInt, int64(0)),
				Location: FileLocation{},
				IsLabel:  false},
		},
	}

	func() {
		defer func() { recover() }()
		casm.bindMemory(DirectiveLine{Name: "memory", Block: "mem 1"}, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	casm.bindMemory(DirectiveLine{Name: "memory", Block: "new_mem 2,3"}, FileLocation{})
	exist, binding := casm.getBindingByName("new_mem")
	assert.True(t, exist)

	want := Binding{
		Status:        BindingEvaluated,
		Name:          "new_mem",
		EvaluatedWord: coppervm.WordU64(0),
		Location:      FileLocation{},
		IsLabel:       false,
	}
	assert.Equal(t, want, binding)

	func() {
		defer func() {
			recover()
		}()
		casm.bindMemory(DirectiveLine{Name: "memory", Block: "new_mem2 ,"}, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()
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
}

func TestEvaluateExpression(t *testing.T) {
	casm := Casm{}
	casm.Bindings = append(casm.Bindings, Binding{
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

			result := casm.evaluateExpression(test.expr, FileLocation{}).Word

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

	// valid binary op
	{expr: expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindStringLit, "str1"),
		Rhs:  expressionP(ExpressionKindStringLit, "str2"),
	}), res: coppervm.WordI64(54)},
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
}

func TestEvaluateBinaryOp(t *testing.T) {
	casm := Casm{}
	for _, test := range binaryOpTests {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			word := casm.evaluateBinaryOp(test.expr, FileLocation{}).Word

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			} else {
				assert.Equal(t, test.res, word, test)
			}
		}()
	}
}

func TestEvaluateBinding(t *testing.T) {
	casm := Casm{
		Bindings: []Binding{
			{Name: "a_bind", Value: expression(ExpressionKindNumLitInt, int64(5))},
			{Name: "cycl1", Value: expression(ExpressionKindBinding, "cycl2")},
			{Name: "cycl2", Value: expression(ExpressionKindBinding, "cycl1")},
		},
	}

	word := casm.evaluateBinding(casm.Bindings[0], FileLocation{}).Word
	assert.Equal(t, coppervm.WordU64(5), word)

	word = casm.evaluateBinding(casm.Bindings[0], FileLocation{}).Word
	assert.Equal(t, coppervm.WordU64(5), word)

	func() {
		defer func() { recover() }()
		casm.evaluateBinding(casm.Bindings[1], FileLocation{})
		assert.Fail(t, "expecting an error")
	}()

	func() {
		defer func() { recover() }()
		casm.evaluateBinding(Binding{Name: "bind"}, FileLocation{})
		assert.Fail(t, "expecting an error")
	}()
}

func TestStrings(t *testing.T) {
	casm := Casm{}
	casm.translateSource("%const str \"a string\"\npush str\npush str\npush \"a new string\"\npush \"a\"+str", "")
	assert.Equal(t, uint64(0), casm.Program[0].Operand.AsU64)
	assert.Equal(t, uint64(0), casm.Program[1].Operand.AsU64)
	assert.Equal(t, uint64(9), casm.Program[2].Operand.AsU64)
	assert.Equal(t, uint64(24), casm.Program[3].Operand.AsU64)
	assert.Equal(t, 34, len(casm.Memory))
}

func TestPushStringToMemory(t *testing.T) {
	casm := Casm{}

	strAddr := casm.pushStringToMemory("string1")
	assert.Equal(t, 0, strAddr)
	assert.Equal(t, 8, casm.StringLengths[0])

	strAddr = casm.pushStringToMemory("a different string")
	assert.Equal(t, 8, strAddr)
	assert.Equal(t, 19, casm.StringLengths[8])
}

func TestGetStringByAddress(t *testing.T) {
	casm := Casm{}
	l1 := casm.pushStringToMemory("string1")
	s1 := casm.getStringByAddress(l1)
	assert.NotEmpty(t, s1)
	assert.Equal(t, "string1", s1)
	assert.Len(t, casm.Memory, 8)

	s2 := casm.getStringByAddress(1)
	assert.Empty(t, s2)
}
