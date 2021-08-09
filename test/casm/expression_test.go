package casm

import (
	"testing"

	"github.com/Supercaly/coppervm/pkg/casm"
)

func TestParseExprFromString(t *testing.T) {
	tests := []struct {
		in       string
		out      casm.Expression
		hasError bool
	}{
		{"1", casm.Expression{
			Kind:        casm.ExpressionKindNumLitInt,
			AsNumLitInt: 1,
		}, false},
		{"2.0", casm.Expression{
			Kind:          casm.ExpressionKindNumLitFloat,
			AsNumLitFloat: 2.0,
		}, false},
		{"3.14", casm.Expression{
			Kind:          casm.ExpressionKindNumLitFloat,
			AsNumLitFloat: 3.14,
		}, false},
		{"-2", casm.Expression{
			Kind:        casm.ExpressionKindNumLitInt,
			AsNumLitInt: -2,
		}, false},
		{"-2.5", casm.Expression{
			Kind:          casm.ExpressionKindNumLitFloat,
			AsNumLitFloat: -2.5,
		}, false},
		{"test", casm.Expression{
			Kind:      casm.ExpressionKindBinding,
			AsBinding: "test",
		}, false},
		{"\"a string\"", casm.Expression{
			Kind:        casm.ExpressionKindStringLit,
			AsStringLit: "a string",
		}, false},
		{"\"an escaped\\nstring\"", casm.Expression{
			Kind:        casm.ExpressionKindStringLit,
			AsStringLit: "an escaped\nstring",
		}, false},
		{"0xFF", casm.Expression{
			Kind:        casm.ExpressionKindNumLitInt,
			AsNumLitInt: 255,
		}, false},
		{"2+3*4+5", casm.Expression{
			Kind:        casm.ExpressionKindNumLitInt,
			AsNumLitInt: 19,
		}, false},
		{"1.2+2.3", casm.Expression{
			Kind:          casm.ExpressionKindNumLitFloat,
			AsNumLitFloat: 3.5,
		}, false},
		{"\"first\"+\"second\"", casm.Expression{
			Kind:        casm.ExpressionKindStringLit,
			AsStringLit: "firstsecond",
		}, false},
		{"\"first\"-\"second\"", casm.Expression{}, true},
		{"\"first\"*\"second\"", casm.Expression{}, true},
		{"1+test", casm.Expression{
			Kind: casm.ExpressionKindBinaryOp,
			AsBinaryOp: casm.BinaryOp{
				Kind: casm.BinaryOpKindPlus,
				Lhs: &casm.Expression{
					Kind:        casm.ExpressionKindNumLitInt,
					AsNumLitInt: 1,
				},
				Rhs: &casm.Expression{
					Kind:      casm.ExpressionKindBinding,
					AsBinding: "test",
				},
			},
		}, false},
		{"2*1+test", casm.Expression{
			Kind: casm.ExpressionKindBinaryOp,
			AsBinaryOp: casm.BinaryOp{
				Kind: casm.BinaryOpKindPlus,
				Lhs: &casm.Expression{
					Kind:        casm.ExpressionKindNumLitInt,
					AsNumLitInt: 2,
				},
				Rhs: &casm.Expression{
					Kind:      casm.ExpressionKindBinding,
					AsBinding: "test",
				},
			},
		}, false},
		{"2.1+1+test+\"str\"", casm.Expression{
			Kind: casm.ExpressionKindBinaryOp,
			AsBinaryOp: casm.BinaryOp{
				Kind: casm.BinaryOpKindPlus,
				Lhs: &casm.Expression{
					Kind: casm.ExpressionKindBinaryOp,
					AsBinaryOp: casm.BinaryOp{
						Kind: casm.BinaryOpKindPlus,
						Lhs: &casm.Expression{
							Kind: casm.ExpressionKindBinaryOp,
							AsBinaryOp: casm.BinaryOp{
								Kind: casm.BinaryOpKindPlus,
								Lhs: &casm.Expression{
									Kind:          casm.ExpressionKindNumLitFloat,
									AsNumLitFloat: 2.1,
								},
								Rhs: &casm.Expression{
									Kind:        casm.ExpressionKindNumLitInt,
									AsNumLitInt: 1,
								},
							},
						},
						Rhs: &casm.Expression{
							Kind:      casm.ExpressionKindBinding,
							AsBinding: "test",
						},
					},
				},
				Rhs: &casm.Expression{
					Kind:        casm.ExpressionKindStringLit,
					AsStringLit: "str",
				},
			},
		}, false},
		{"-2*3", casm.Expression{
			Kind:        casm.ExpressionKindNumLitInt,
			AsNumLitInt: -6,
		}, false},
		{"(1+2)*(1+2)", casm.Expression{
			Kind:        casm.ExpressionKindNumLitInt,
			AsNumLitInt: 9,
		}, false},
		{"0xG", casm.Expression{}, true},
		{"1.2.3", casm.Expression{}, true},
		{"1$", casm.Expression{}, true},
	}

	for _, test := range tests {
		expr, err := casm.ParseExprFromString(test.in)

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if !expressionEquals(expr, test.out) {
			t.Errorf("Expected '%s' but got '%s'", test.out, expr)
		}
	}
}

func TestParseByteArrayFromString(t *testing.T) {
	tests := []struct {
		in       string
		out      []byte
		hasError bool
	}{
		{"1, 2, 3, 4", []byte{1, 2, 3, 4}, false},
		{"1, 2, 3, 4,", []byte{1, 2, 3, 4}, false},
		{"1", []byte{1}, false},
		{"1,", []byte{1}, false},
		{"", []byte{}, false},
		{"1,\"test\"", []byte{1, 0x74, 0x65, 0x73, 0x74}, false},
		{"1, 0xf", []byte{1, 0xf}, false},
		{"1 2 3", []byte{}, true},
		{"1,,2", []byte{}, true},
		{",1", []byte{}, true},
		{",", []byte{}, true},
		{"1.2,test", []byte{}, true},
	}

	for _, test := range tests {
		expr, err := casm.ParseByteArrayFromString(test.in)

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if !byteArrayEquals(expr, test.out) {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, expr)
		}
	}
}

func byteArrayEquals(a []byte, b []byte) bool {
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

func expressionEquals(a casm.Expression, b casm.Expression) bool {
	if a.Kind != b.Kind {
		return false
	}
	if a.Kind == casm.ExpressionKindBinaryOp {
		return a.AsBinaryOp.Kind == b.AsBinaryOp.Kind &&
			expressionEquals(*a.AsBinaryOp.Lhs, *b.AsBinaryOp.Lhs) &&
			expressionEquals(*a.AsBinaryOp.Rhs, *b.AsBinaryOp.Rhs)
	}
	return a == b
}
