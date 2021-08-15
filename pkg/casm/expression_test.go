package casm

import (
	"reflect"
	"testing"
)

func TestTokenIsOperator(t *testing.T) {
	tests := []struct {
		in  Token
		out bool
	}{
		{Token{Kind: TokenKindNumLit, Text: "1"}, false},
		{Token{Kind: TokenKindStringLit, Text: ""}, false},
		{Token{Kind: TokenKindSymbol, Text: "_a"}, false},
		{Token{Kind: TokenKindPlus, Text: "+"}, true},
		{Token{Kind: TokenKindMinus, Text: "-"}, true},
		{Token{Kind: TokenKindAsterisk, Text: "*"}, true},
		{Token{Kind: TokenKindComma, Text: ","}, false},
		{Token{Kind: TokenKindOpenParen, Text: "("}, false},
		{Token{Kind: TokenKindCloseParen, Text: ")"}, false},
	}

	for _, test := range tests {
		if tokenIsOperator(test.in) != test.out {
			t.Errorf("Expecting %#v %t but got %t", test.in, test.out, !test.out)
		}
	}
}

func TestTokenAsBinaryOpKind(t *testing.T) {
	tests := []struct {
		in       Token
		out      BinaryOpKind
		hasError bool
	}{
		{Token{Kind: TokenKindNumLit, Text: "1"}, 0, true},
		{Token{Kind: TokenKindStringLit, Text: ""}, 0, true},
		{Token{Kind: TokenKindSymbol, Text: "_a"}, 0, true},
		{Token{Kind: TokenKindPlus, Text: "+"}, BinaryOpKindPlus, false},
		{Token{Kind: TokenKindMinus, Text: "-"}, BinaryOpKindMinus, false},
		{Token{Kind: TokenKindAsterisk, Text: "*"}, BinaryOpKindTimes, false},
		{Token{Kind: TokenKindComma, Text: ","}, 0, true},
		{Token{Kind: TokenKindOpenParen, Text: "("}, 0, true},
		{Token{Kind: TokenKindCloseParen, Text: ")"}, 0, true},
	}

	for _, test := range tests {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					t.Error(r)
				}
			}()
			res := tokenAsBinaryOpKind(test.in)
			if res != test.out {
				t.Errorf("Expecting %#v %s but got %s", test.in, test.out, res)
			}
		}()
	}
}

func TestComputeOpWithSameType(t *testing.T) {
	tests := []struct {
		left     Expression
		right    Expression
		op       BinaryOpKind
		out      Expression
		hasError bool
	}{
		{
			left: Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: int64(1),
			},
			right: Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: int64(1),
			},
			op: BinaryOpKindPlus,
			out: Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: int64(2),
			},
			hasError: false,
		},
		{
			left: Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: int64(1),
			},
			right: Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: int64(1),
			},
			op: BinaryOpKindMinus,
			out: Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: int64(0),
			},
			hasError: false,
		},
		{
			left: Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: int64(2),
			},
			right: Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: int64(2),
			},
			op: BinaryOpKindTimes,
			out: Expression{
				Kind:        ExpressionKindNumLitInt,
				AsNumLitInt: int64(4),
			},
			hasError: false,
		},
		{
			left: Expression{
				Kind:          ExpressionKindNumLitFloat,
				AsNumLitFloat: float64(1.0),
			},
			right: Expression{
				Kind:          ExpressionKindNumLitFloat,
				AsNumLitFloat: float64(1.0),
			},
			op: BinaryOpKindPlus,
			out: Expression{
				Kind:          ExpressionKindNumLitFloat,
				AsNumLitFloat: float64(2.0),
			},
			hasError: false,
		},
		{
			left: Expression{
				Kind:          ExpressionKindNumLitFloat,
				AsNumLitFloat: float64(1.0),
			},
			right: Expression{
				Kind:          ExpressionKindNumLitFloat,
				AsNumLitFloat: float64(1.0),
			},
			op: BinaryOpKindMinus,
			out: Expression{
				Kind:          ExpressionKindNumLitFloat,
				AsNumLitFloat: float64(0.0),
			},
			hasError: false,
		},
		{
			left: Expression{
				Kind:          ExpressionKindNumLitFloat,
				AsNumLitFloat: float64(1.0),
			},
			right: Expression{
				Kind:          ExpressionKindNumLitFloat,
				AsNumLitFloat: float64(2.0),
			},
			op: BinaryOpKindTimes,
			out: Expression{
				Kind:          ExpressionKindNumLitFloat,
				AsNumLitFloat: float64(2.0),
			},
			hasError: false,
		},
		{
			left: Expression{
				Kind:        ExpressionKindStringLit,
				AsStringLit: "first",
			},
			right: Expression{
				Kind:        ExpressionKindStringLit,
				AsStringLit: "second",
			},
			op: BinaryOpKindPlus,
			out: Expression{
				Kind:        ExpressionKindStringLit,
				AsStringLit: "firstsecond",
			},
			hasError: false,
		},
		{
			left: Expression{
				Kind:        ExpressionKindStringLit,
				AsStringLit: "first",
			},
			right: Expression{
				Kind:        ExpressionKindStringLit,
				AsStringLit: "second",
			},
			op:       BinaryOpKindMinus,
			out:      Expression{},
			hasError: true,
		},
		{
			left: Expression{
				Kind:        ExpressionKindStringLit,
				AsStringLit: "first",
			},
			right: Expression{
				Kind:        ExpressionKindStringLit,
				AsStringLit: "second",
			},
			op:       BinaryOpKindTimes,
			out:      Expression{},
			hasError: true,
		},
		{
			left: Expression{
				Kind: ExpressionKindBinaryOp,
			},
			right: Expression{
				Kind: ExpressionKindBinaryOp,
			},
			op:       BinaryOpKindPlus,
			out:      Expression{},
			hasError: true,
		},
		{
			left: Expression{
				Kind: ExpressionKindBinaryOp,
			},
			right: Expression{
				Kind: ExpressionKindBinaryOp,
			},
			op:       BinaryOpKindMinus,
			out:      Expression{},
			hasError: true,
		},
		{
			left: Expression{
				Kind: ExpressionKindBinaryOp,
			},
			right: Expression{
				Kind: ExpressionKindBinaryOp,
			},
			op:       BinaryOpKindTimes,
			out:      Expression{},
			hasError: true,
		},
		{
			left: Expression{
				Kind: ExpressionKindBinding,
			},
			right: Expression{
				Kind: ExpressionKindBinding,
			},
			op:       BinaryOpKindPlus,
			out:      Expression{},
			hasError: true,
		},
		{
			left: Expression{
				Kind: ExpressionKindBinding,
			},
			right: Expression{
				Kind: ExpressionKindBinding,
			},
			op:       BinaryOpKindMinus,
			out:      Expression{},
			hasError: true,
		},
		{
			left: Expression{
				Kind: ExpressionKindBinding,
			},
			right: Expression{
				Kind: ExpressionKindBinding,
			},
			op:       BinaryOpKindTimes,
			out:      Expression{},
			hasError: true,
		},
	}

	for i, test := range tests {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					t.Error(r)
				}
			}()
			res := computeOpWithSameType(test.left, test.right, test.op)
			if !expressionEquals(res, test.out) {
				t.Errorf("test %d: expected '%#v' but got '%#v'", i, test.out, res)
			}
		}()
	}
}

func TestParseExprFromString(t *testing.T) {
	tests := []struct {
		in       string
		out      Expression
		hasError bool
	}{
		{"1", Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: 1,
		}, false},
		{"2.0", Expression{
			Kind:          ExpressionKindNumLitFloat,
			AsNumLitFloat: 2.0,
		}, false},
		{"3.14", Expression{
			Kind:          ExpressionKindNumLitFloat,
			AsNumLitFloat: 3.14,
		}, false},
		{"-2", Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: -2,
		}, false},
		{"-2.5", Expression{
			Kind:          ExpressionKindNumLitFloat,
			AsNumLitFloat: -2.5,
		}, false},
		{"test", Expression{
			Kind:      ExpressionKindBinding,
			AsBinding: "test",
		}, false},
		{"\"a string\"", Expression{
			Kind:        ExpressionKindStringLit,
			AsStringLit: "a string",
		}, false},
		{"\"an escaped\\nstring\"", Expression{
			Kind:        ExpressionKindStringLit,
			AsStringLit: "an escaped\nstring",
		}, false},
		{"0xFF", Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: 255,
		}, false},
		{"2+3*4+5", Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: 19,
		}, false},
		{"1.2+2.3", Expression{
			Kind:          ExpressionKindNumLitFloat,
			AsNumLitFloat: 3.5,
		}, false},
		{"\"first\"+\"second\"", Expression{
			Kind:        ExpressionKindStringLit,
			AsStringLit: "firstsecond",
		}, false},
		{"\"first\"-\"second\"", Expression{}, true},
		{"\"first\"*\"second\"", Expression{}, true},
		{"1+test", Expression{
			Kind: ExpressionKindBinaryOp,
			AsBinaryOp: BinaryOp{
				Kind: BinaryOpKindPlus,
				Lhs: &Expression{
					Kind:        ExpressionKindNumLitInt,
					AsNumLitInt: 1,
				},
				Rhs: &Expression{
					Kind:      ExpressionKindBinding,
					AsBinding: "test",
				},
			},
		}, false},
		{"2*1+test", Expression{
			Kind: ExpressionKindBinaryOp,
			AsBinaryOp: BinaryOp{
				Kind: BinaryOpKindPlus,
				Lhs: &Expression{
					Kind:        ExpressionKindNumLitInt,
					AsNumLitInt: 2,
				},
				Rhs: &Expression{
					Kind:      ExpressionKindBinding,
					AsBinding: "test",
				},
			},
		}, false},
		{"2.1+1+test+\"str\"", Expression{
			Kind: ExpressionKindBinaryOp,
			AsBinaryOp: BinaryOp{
				Kind: BinaryOpKindPlus,
				Lhs: &Expression{
					Kind: ExpressionKindBinaryOp,
					AsBinaryOp: BinaryOp{
						Kind: BinaryOpKindPlus,
						Lhs: &Expression{
							Kind: ExpressionKindBinaryOp,
							AsBinaryOp: BinaryOp{
								Kind: BinaryOpKindPlus,
								Lhs: &Expression{
									Kind:          ExpressionKindNumLitFloat,
									AsNumLitFloat: 2.1,
								},
								Rhs: &Expression{
									Kind:        ExpressionKindNumLitInt,
									AsNumLitInt: 1,
								},
							},
						},
						Rhs: &Expression{
							Kind:      ExpressionKindBinding,
							AsBinding: "test",
						},
					},
				},
				Rhs: &Expression{
					Kind:        ExpressionKindStringLit,
					AsStringLit: "str",
				},
			},
		}, false},
		{"-2*3", Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: -6,
		}, false},
		{"(1+2)*(1+2)", Expression{
			Kind:        ExpressionKindNumLitInt,
			AsNumLitInt: 9,
		}, false},
		{"0xG", Expression{}, true},
		{"1.2.3", Expression{}, true},
		{"1$", Expression{}, true},
		{"(1", Expression{}, true},
	}

	for _, test := range tests {
		expr, err := ParseExprFromString(test.in)

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
		expr, err := ParseByteArrayFromString(test.in)

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if err == nil && !reflect.DeepEqual(expr, test.out) {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, expr)
		}
	}
}

func expressionEquals(a Expression, b Expression) bool {
	if a.Kind != b.Kind {
		return false
	}
	if a.Kind == ExpressionKindBinaryOp {
		return a.AsBinaryOp.Kind == b.AsBinaryOp.Kind &&
			expressionEquals(*a.AsBinaryOp.Lhs, *b.AsBinaryOp.Lhs) &&
			expressionEquals(*a.AsBinaryOp.Rhs, *b.AsBinaryOp.Rhs)
	}
	return a == b
}
