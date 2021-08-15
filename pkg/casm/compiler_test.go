package casm

import (
	"reflect"
	"testing"

	"github.com/Supercaly/coppervm/pkg/coppervm"
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
			if !instArrayEquals(ctx.Program, test.out) {
				t.Errorf("expected '%#v' but got '%#v'", test.out, ctx.Program)
			}
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
	if !e {
		t.Errorf("cannot find binging with name 'a_label'")
	}
	if b != casm.Bindings[0] {
		t.Errorf("expecting %#v but got %#v", casm.Bindings[0], b)
	}

	e, b = casm.getBindingByName("a_const")
	if !e {
		t.Errorf("cannot find binging with name 'a_const'")
	}
	if b != casm.Bindings[1] {
		t.Errorf("expecting %#v but got %#v", casm.Bindings[1], b)
	}

	e, b = casm.getBindingByName("test")
	if e {
		t.Errorf("no binding with name 'test' should exist")
	}
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
	if !e {
		t.Error("the new_label was not bound correctly")
	}

	lab := Binding{
		Name: "new_label",
		Value: Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: 2,
		},
		Location: FileLocation{},
		IsLabel:  true,
	}
	if !reflect.DeepEqual(b, lab) {
		t.Errorf("expecting %#v but got %#v", lab, b)
	}
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
	if !e {
		t.Error("the new_const was not bound correctly")
	}
	c := Binding{
		Name: "new_const",
		Value: Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: 2,
		},
		Location: FileLocation{},
		IsLabel:  false,
	}
	if !reflect.DeepEqual(b, c) {
		t.Errorf("expecting %#v but got %#v", c, b)
	}

	func() {
		defer func() {
			recover()
		}()
		casm.bindConst(DirectiveLine{Name: "const", Block: "new_const2"},
			FileLocation{})
		t.Error("Expecting an error")
	}()
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
	if !casm.HasEntry ||
		casm.DeferredEntryName != "entry" ||
		casm.EntryLocation.Location != 10 {
		t.Error("entry location not set")
	}
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
	if !e {
		t.Error("the new_mem was not bound correctly")
	}
	c := Binding{
		Name: "new_mem",
		Value: Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: 0,
		},
		Location: FileLocation{},
		IsLabel:  false,
	}
	if !reflect.DeepEqual(b, c) {
		t.Errorf("expecting %#v but got %#v", c, b)
	}

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

	w := casm.evaluateExpression(Expression{
		Kind:        ExpressionKindNumLitInt,
		AsNumLitInt: 10,
	}, FileLocation{})
	if w != coppervm.WordU64(10) {
		t.Errorf("expecting %#v but got %#v", coppervm.WordI64(10), w)
	}

	w2 := casm.evaluateExpression(Expression{
		Kind:          ExpressionKindNumLitFloat,
		AsNumLitFloat: 2.3,
	}, FileLocation{})
	if w2 != coppervm.WordF64(2.3) {
		t.Errorf("expecting %#v but got %#v", coppervm.WordF64(2.3), w2)
	}

	w3 := casm.evaluateExpression(Expression{
		Kind:        ExpressionKindStringLit,
		AsStringLit: "str",
	}, FileLocation{})
	if w3 != coppervm.WordI64(0) {
		t.Errorf("expecting %#v but got %#v", coppervm.WordF64(0), w3)
	}

	w4 := casm.evaluateExpression(Expression{
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
	}, FileLocation{})
	if w4 != coppervm.WordI64(7) {
		t.Errorf("expecting %#v but got %#v", coppervm.WordF64(7), w4)
	}

	w5 := casm.evaluateExpression(Expression{
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
	}, FileLocation{})
	if w5 != coppervm.WordI64(-3) {
		t.Errorf("expecting %#v but got %#v", coppervm.WordF64(-3), w5)
	}

	w6 := casm.evaluateExpression(Expression{
		Kind: ExpressionKindBinaryOp,
		AsBinaryOp: BinaryOp{
			Kind: BinaryOpKindTimes,
			Lhs: &Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: 2,
			},
			Rhs: &Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: 5,
			},
		},
	}, FileLocation{})
	if w6 != coppervm.WordI64(10) {
		t.Errorf("expecting %#v but got %#v", coppervm.WordF64(10), w6)
	}

	func() {
		defer func() { recover() }()
		casm.evaluateExpression(Expression{
			Kind:      ExpressionKindBinding,
			AsBinding: "bind",
		}, FileLocation{})
		t.Error("expecting an error")
	}()

	casm.Bindings = append(casm.Bindings, Binding{
		Name:  "a_bind",
		Value: Expression{Kind: ExpressionKindNumLitInt, AsNumLitInt: 3},
	})
	w7 := casm.evaluateExpression(Expression{
		Kind:      ExpressionKindBinding,
		AsBinding: "a_bind",
	}, FileLocation{})
	if w7 != coppervm.WordI64(3) {
		t.Errorf("expecting %#v but got %#v", coppervm.WordF64(3), w7)
	}
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

	w := casm.evaluateBinding(&casm.Bindings[0], FileLocation{})
	if w != coppervm.WordU64(5) {
		t.Errorf("expecting %#v but got %#v", coppervm.WordU64(5), w)
	}

	// w2 := casm.evaluateBinding(&casm.Bindings[1], FileLocation{})
	// if w2 != coppervm.WordU64(5) {
	// 	t.Errorf("expecting %#v but got %#v", coppervm.WordU64(5), w2)
	// }
}
