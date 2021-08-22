package casm

import (
	"testing"

	"github.com/Supercaly/coppervm/pkg/coppervm"
	"github.com/stretchr/testify/assert"
)

func TestTranslateSource(t *testing.T) {
	tests := []struct {
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

	for _, test := range tests {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					t.Error(r)
				}
			}()

			ctx := Casm{}
			ctx.translateSource(test.in, "test_file")

			if test.hasError {
				t.Error("expecting an error")
			}
			assert.Condition(t, func() (success bool) { return instArrayEquals(test.out, ctx.Program) })
		}()
	}
}

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
				Value:    Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 1},
				Location: FileLocation{},
				IsLabel:  true},
			{Name: "a_const",
				Value:    Expression{Kind: ExpressionKindNumLitFloat, AsNumLitFloat: 2.3},
				Location: FileLocation{},
				IsLabel:  false},
		},
	}

	e, b := casm.getBindingByName("a_label")
	assert.True(t, e)
	assert.Equal(t, casm.Bindings[0], b)

	e, b = casm.getBindingByName("a_const")
	assert.True(t, e)
	assert.Equal(t, casm.Bindings[1], b)

	e, b = casm.getBindingByName("test")
	assert.False(t, e)
}

func TestGetBindingIndexByName(t *testing.T) {
	casm := Casm{
		Bindings: []Binding{
			{Name: "a_label",
				Value:    Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 1},
				Location: FileLocation{},
				IsLabel:  true},
			{Name: "a_const",
				Value:    Expression{Kind: ExpressionKindNumLitFloat, AsNumLitFloat: 2.3},
				Location: FileLocation{},
				IsLabel:  false},
		},
	}

	i := casm.getBindingIndexByName("a_label")
	assert.Equal(t, 0, i)

	i = casm.getBindingIndexByName("a_const")
	assert.Equal(t, 1, i)

	i = casm.getBindingIndexByName("test")
	assert.Equal(t, -1, i)
}

func TestBindLabel(t *testing.T) {
	casm := Casm{
		Bindings: []Binding{
			{Name: "a_label",
				Value:    Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 1},
				Location: FileLocation{},
				IsLabel:  true},
		},
	}

	func() {
		defer func() { recover() }()
		casm.bindLabel("a_label", 1, FileLocation{})
		t.Error("Expecting an error")
	}()

	casm.bindLabel("new_label", 2, FileLocation{})
	e, b := casm.getBindingByName("new_label")
	assert.True(t, e)

	lab := Binding{
		Name:          "new_label",
		EvaluatedWord: coppervm.WordU64(2),
		Location:      FileLocation{},
		IsLabel:       true,
		Status:        BindingEvaluated,
	}
	assert.Equal(t, lab, b)
}

func TestBindConst(t *testing.T) {
	casm := Casm{
		Bindings: []Binding{
			{Name: "a_const",
				Value:    Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 1},
				Location: FileLocation{},
				IsLabel:  false},
		},
	}

	func() {
		defer func() { recover() }()
		casm.bindConst(DirectiveLine{Name: "const", Block: "a_const 1"},
			FileLocation{})
		t.Error("Expecting an error")
	}()

	casm.bindConst(DirectiveLine{Name: "const", Block: "new_const 2"},
		FileLocation{})
	e, b := casm.getBindingByName("new_const")
	assert.True(t, e)

	c := Binding{
		Name: "new_const",
		Value: Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: 2,
		},
		Location: FileLocation{},
		IsLabel:  false,
	}
	assert.Equal(t, c, b)

	func() {
		defer func() {
			recover()
		}()
		casm.bindConst(DirectiveLine{Name: "const", Block: "new_const2"},
			FileLocation{})
		t.Error("Expecting an error")
	}()

	casm.bindConst(DirectiveLine{Name: "const", Block: "str_const \"test_str\""}, FileLocation{})
	e, b = casm.getBindingByName("str_const")
	assert.True(t, e)
	assert.Equal(t, Binding{
		Status: BindingEvaluated,
		Name:   "str_const",
		Value: Expression{
			Kind:        ExpressionKindStringLit,
			AsStringLit: "test_str",
		},
		EvaluatedWord: coppervm.WordU64(0),
		Location:      FileLocation{},
		IsLabel:       false,
	}, b)
	assert.Equal(t, 9, len(casm.Memory))
}

func TestBindEntry(t *testing.T) {
	func() {
		casm := Casm{}
		defer func() { recover() }()
		casm.HasEntry = true
		casm.bindEntry("main", FileLocation{})
		t.Error("expecting an error")
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
				Value:    Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 0},
				Location: FileLocation{},
				IsLabel:  false},
		},
	}

	func() {
		defer func() { recover() }()
		casm.bindMemory(DirectiveLine{Name: "memory", Block: "mem 1"},
			FileLocation{})
		t.Error("Expecting an error")
	}()

	casm.bindMemory(DirectiveLine{Name: "memory", Block: "new_mem 2,3"},
		FileLocation{})
	e, b := casm.getBindingByName("new_mem")
	assert.True(t, e)

	c := Binding{
		Status:        BindingEvaluated,
		Name:          "new_mem",
		EvaluatedWord: coppervm.WordU64(0),
		Location:      FileLocation{},
		IsLabel:       false,
	}
	assert.Equal(t, c, b)

	func() {
		defer func() {
			recover()
		}()
		casm.bindMemory(DirectiveLine{Name: "memory", Block: "new_mem2 ,"},
			FileLocation{})
		t.Error("Expecting an error")
	}()
}

func TestEvaluateExpression(t *testing.T) {
	casm := Casm{}
	tests := []struct {
		expr Expression
		res  coppervm.Word
	}{
		{Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: 10,
		}, coppervm.WordU64(10)},
		{Expression{
			Kind:          ExpressionKindNumLitFloat,
			AsNumLitFloat: 2.3,
		}, coppervm.WordF64(2.3)},
		{Expression{
			Kind:        ExpressionKindStringLit,
			AsStringLit: "str",
		}, coppervm.WordI64(0)},
		{Expression{
			Kind: ExpressionKindBinaryOp,
			AsBinaryOp: BinaryOp{
				Kind: BinaryOpKindPlus,
				Lhs: &Expression{
					Kind:        ExpressionKindNumLitInt,
					AsNumLitInt: 2,
				},
				Rhs: &Expression{
					Kind:        ExpressionKindNumLitInt,
					AsNumLitInt: 5,
				},
			},
		}, coppervm.WordI64(7)},
		{Expression{
			Kind: ExpressionKindBinaryOp,
			AsBinaryOp: BinaryOp{
				Kind: BinaryOpKindMinus,
				Lhs: &Expression{
					Kind:        ExpressionKindNumLitInt,
					AsNumLitInt: 2,
				},
				Rhs: &Expression{
					Kind:        ExpressionKindNumLitInt,
					AsNumLitInt: 5,
				},
			},
		}, coppervm.WordI64(-3)},
		{Expression{
			Kind: ExpressionKindBinaryOp,
			AsBinaryOp: BinaryOp{
				Kind: BinaryOpKindTimes,
				Lhs: &Expression{
					Kind:        ExpressionKindNumLitInt,
					AsNumLitInt: 2,
				},
				Rhs: &Expression{
					Kind:          ExpressionKindNumLitFloat,
					AsNumLitFloat: 5.3,
				},
			},
		}, coppervm.WordF64(10.6)},
	}

	for _, test := range tests {
		assert.Equal(t, test.res, casm.evaluateExpression(test.expr, FileLocation{}).Word)
	}

	casm.Bindings = append(casm.Bindings, Binding{
		Name:  "a_bind",
		Value: Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 3},
	})
	w := casm.evaluateExpression(Expression{
		Kind:      ExpressionKindBinding,
		AsBinding: "a_bind",
	}, FileLocation{}).Word
	assert.Equal(t, coppervm.WordI64(3), w)

	func() {
		defer func() { recover() }()
		casm.evaluateExpression(Expression{
			Kind:      ExpressionKindBinding,
			AsBinding: "bind",
		}, FileLocation{})
		t.Error("expecting an error")
	}()
}

func TestEvaluateBinding(t *testing.T) {
	casm := Casm{
		Bindings: []Binding{
			{
				Name: "a_bind",
				Value: Expression{
					Kind:        ExpressionKindNumLitInt,
					AsNumLitInt: 5,
				},
			},
			{
				Name: "cycl1",
				Value: Expression{
					Kind:      ExpressionKindBinding,
					AsBinding: "cycl2",
				},
			},
			{
				Name: "cycl2",
				Value: Expression{
					Kind:      ExpressionKindBinding,
					AsBinding: "cycl1",
				},
			},
		},
	}

	w := casm.evaluateBinding(casm.Bindings[0], FileLocation{}).Word
	assert.Equal(t, coppervm.WordU64(5), w)

	w2 := casm.evaluateBinding(casm.Bindings[0], FileLocation{}).Word
	assert.Equal(t, coppervm.WordU64(5), w2)

	func() {
		defer func() { recover() }()
		w3 := casm.evaluateBinding(casm.Bindings[1], FileLocation{}).Word
		if w3 != coppervm.WordU64(5) {
			t.Errorf("expecting %#v but got %#v", coppervm.WordU64(5), w3)
		}
		t.Error("expecting an error")
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
