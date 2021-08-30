package casm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenIsOperator(t *testing.T) {
	tests := []struct {
		in  Token
		out bool
	}{
		{token(TokenKindNumLit, "1"), false},
		{token(TokenKindStringLit, ""), false},
		{token(TokenKindSymbol, "_a"), false},
		{token(TokenKindPlus, "+"), true},
		{token(TokenKindMinus, "-"), true},
		{token(TokenKindAsterisk, "*"), true},
		{token(TokenKindSlash, "/"), true},
		{token(TokenKindPercent, "%"), true},
		{token(TokenKindComma, ","), false},
		{token(TokenKindOpenParen, "("), false},
		{token(TokenKindCloseParen, ")"), false},
	}

	for _, test := range tests {
		assert.Equal(t, test.out, tokenIsOperator(test.in), test)
	}
}

func TestTokenAsBinaryOpKind(t *testing.T) {
	tests := []struct {
		in       Token
		out      BinaryOpKind
		hasError bool
	}{
		{token(TokenKindNumLit, "1"), 0, true},
		{token(TokenKindStringLit, ""), 0, true},
		{token(TokenKindSymbol, "_a"), 0, true},
		{token(TokenKindPlus, "+"), BinaryOpKindPlus, false},
		{token(TokenKindMinus, "-"), BinaryOpKindMinus, false},
		{token(TokenKindAsterisk, "*"), BinaryOpKindTimes, false},
		{token(TokenKindSlash, "/"), BinaryOpKindDivide, false},
		{token(TokenKindPercent, "%"), BinaryOpKindModulo, false},
		{token(TokenKindComma, ","), 0, true},
		{token(TokenKindOpenParen, "("), 0, true},
		{token(TokenKindCloseParen, ")"), 0, true},
	}

	for _, test := range tests {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()
			res := tokenAsBinaryOpKind(test.in)

			assert.Equal(t, test.out, res, test)
			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			}
		}()
	}
}

// Wrapper function to create an Expression.
func expression(kind ExpressionKind, value interface{}) (expr Expression) {
	expr.Kind = kind
	switch kind {
	case ExpressionKindNumLitInt:
		expr.AsNumLitInt = value.(int64)
	case ExpressionKindNumLitFloat:
		expr.AsNumLitFloat = value.(float64)
	case ExpressionKindStringLit:
		expr.AsStringLit = value.(string)
	case ExpressionKindBinaryOp:
		expr.AsBinaryOp = value.(BinaryOp)
	case ExpressionKindBinding:
		expr.AsBinding = value.(string)
	}
	return expr
}

// Wrapper function to create an Expression pointer.
func expressionP(kind ExpressionKind, value interface{}) *Expression {
	expr := &Expression{}
	(*expr).Kind = kind
	switch kind {
	case ExpressionKindNumLitInt:
		(*expr).AsNumLitInt = value.(int64)
	case ExpressionKindNumLitFloat:
		(*expr).AsNumLitFloat = value.(float64)
	case ExpressionKindStringLit:
		(*expr).AsStringLit = value.(string)
	case ExpressionKindBinaryOp:
		(*expr).AsBinaryOp = value.(BinaryOp)
	case ExpressionKindBinding:
		(*expr).AsBinding = value.(string)
	}
	return expr
}

var testOpWithSameType = []struct {
	left     Expression
	right    Expression
	op       BinaryOpKind
	out      Expression
	hasError bool
}{
	// integer ops
	{
		left:     expression(ExpressionKindNumLitInt, int64(1)),
		right:    expression(ExpressionKindNumLitInt, int64(1)),
		op:       BinaryOpKindPlus,
		out:      expression(ExpressionKindNumLitInt, int64(2)),
		hasError: false,
	},
	{
		left:     expression(ExpressionKindNumLitInt, int64(1)),
		right:    expression(ExpressionKindNumLitInt, int64(1)),
		op:       BinaryOpKindMinus,
		out:      expression(ExpressionKindNumLitInt, int64(0)),
		hasError: false,
	},
	{
		left:     expression(ExpressionKindNumLitInt, int64(2)),
		right:    expression(ExpressionKindNumLitInt, int64(2)),
		op:       BinaryOpKindTimes,
		out:      expression(ExpressionKindNumLitInt, int64(4)),
		hasError: false,
	},
	{
		left:     expression(ExpressionKindNumLitInt, int64(4)),
		right:    expression(ExpressionKindNumLitInt, int64(2)),
		op:       BinaryOpKindDivide,
		out:      expression(ExpressionKindNumLitInt, int64(2)),
		hasError: false,
	},
	{
		left:     expression(ExpressionKindNumLitInt, int64(4)),
		right:    expression(ExpressionKindNumLitInt, int64(0)),
		op:       BinaryOpKindDivide,
		hasError: true,
	},
	{
		left:     expression(ExpressionKindNumLitInt, int64(5)),
		right:    expression(ExpressionKindNumLitInt, int64(2)),
		op:       BinaryOpKindModulo,
		out:      expression(ExpressionKindNumLitInt, int64(1)),
		hasError: false,
	},
	// float ops
	{
		left:     expression(ExpressionKindNumLitFloat, float64(1.0)),
		right:    expression(ExpressionKindNumLitFloat, float64(1.0)),
		op:       BinaryOpKindPlus,
		out:      expression(ExpressionKindNumLitFloat, float64(2.0)),
		hasError: false,
	},
	{
		left:     expression(ExpressionKindNumLitFloat, float64(1.0)),
		right:    expression(ExpressionKindNumLitFloat, float64(1.0)),
		op:       BinaryOpKindMinus,
		out:      expression(ExpressionKindNumLitFloat, float64(0.0)),
		hasError: false,
	},
	{
		left:     expression(ExpressionKindNumLitFloat, float64(1.0)),
		right:    expression(ExpressionKindNumLitFloat, float64(2.0)),
		op:       BinaryOpKindTimes,
		out:      expression(ExpressionKindNumLitFloat, float64(2.0)),
		hasError: false,
	},
	{
		left:     expression(ExpressionKindNumLitFloat, float64(1.0)),
		right:    expression(ExpressionKindNumLitFloat, float64(2.0)),
		op:       BinaryOpKindDivide,
		out:      expression(ExpressionKindNumLitFloat, float64(0.5)),
		hasError: false,
	},
	{
		left:     expression(ExpressionKindNumLitFloat, float64(1.0)),
		right:    expression(ExpressionKindNumLitFloat, float64(0.0)),
		op:       BinaryOpKindDivide,
		hasError: true,
	},
	{
		left:     expression(ExpressionKindNumLitFloat, float64(1.0)),
		right:    expression(ExpressionKindNumLitFloat, float64(2.0)),
		op:       BinaryOpKindModulo,
		hasError: true,
	},
	// string ops
	{
		left:     expression(ExpressionKindStringLit, "first"),
		right:    expression(ExpressionKindStringLit, "second"),
		op:       BinaryOpKindPlus,
		out:      expression(ExpressionKindStringLit, "firstsecond"),
		hasError: false,
	},
	{
		left:     expression(ExpressionKindStringLit, "first"),
		right:    expression(ExpressionKindStringLit, "second"),
		op:       BinaryOpKindMinus,
		hasError: true,
	},
	{
		left:     expression(ExpressionKindStringLit, "first"),
		right:    expression(ExpressionKindStringLit, "second"),
		op:       BinaryOpKindTimes,
		hasError: true,
	},
	{
		left:     expression(ExpressionKindStringLit, "first"),
		right:    expression(ExpressionKindStringLit, "second"),
		op:       BinaryOpKindDivide,
		hasError: true,
	},
	{
		left:     expression(ExpressionKindStringLit, "first"),
		right:    expression(ExpressionKindStringLit, "second"),
		op:       BinaryOpKindModulo,
		hasError: true,
	},
	// binop ops
	{
		left:     Expression{Kind: ExpressionKindBinaryOp},
		right:    Expression{Kind: ExpressionKindBinaryOp},
		op:       BinaryOpKindPlus,
		hasError: true,
	},
	{
		left:     Expression{Kind: ExpressionKindBinaryOp},
		right:    Expression{Kind: ExpressionKindBinaryOp},
		op:       BinaryOpKindMinus,
		hasError: true,
	},
	{
		left:     Expression{Kind: ExpressionKindBinaryOp},
		right:    Expression{Kind: ExpressionKindBinaryOp},
		op:       BinaryOpKindTimes,
		hasError: true,
	},
	{
		left:     Expression{Kind: ExpressionKindBinaryOp},
		right:    Expression{Kind: ExpressionKindBinaryOp},
		op:       BinaryOpKindDivide,
		hasError: true,
	},
	{
		left:     Expression{Kind: ExpressionKindBinaryOp},
		right:    Expression{Kind: ExpressionKindBinaryOp},
		op:       BinaryOpKindModulo,
		hasError: true,
	},
	// bindings ops
	{
		left:     Expression{Kind: ExpressionKindBinding},
		right:    Expression{Kind: ExpressionKindBinding},
		op:       BinaryOpKindPlus,
		hasError: true,
	},
	{
		left:     Expression{Kind: ExpressionKindBinding},
		right:    Expression{Kind: ExpressionKindBinding},
		op:       BinaryOpKindMinus,
		hasError: true,
	},
	{
		left:     Expression{Kind: ExpressionKindBinding},
		right:    Expression{Kind: ExpressionKindBinding},
		op:       BinaryOpKindTimes,
		hasError: true,
	},
	{
		left:     Expression{Kind: ExpressionKindBinding},
		right:    Expression{Kind: ExpressionKindBinding},
		op:       BinaryOpKindDivide,
		hasError: true,
	},
	{
		left:     Expression{Kind: ExpressionKindBinding},
		right:    Expression{Kind: ExpressionKindBinding},
		op:       BinaryOpKindModulo,
		hasError: true,
	},
}

func TestComputeOpWithSameType(t *testing.T) {
	for _, test := range testOpWithSameType {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()
			res := computeOpWithSameType(test.left, test.right, test.op)

			assert.Equal(t, test.out, res, test)
			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			}
		}()
	}
}

var testExpressions = []struct {
	in       string
	out      Expression
	hasError bool
}{
	{"1", expression(ExpressionKindNumLitInt, int64(1)), false},
	{"2.0", expression(ExpressionKindNumLitFloat, 2.0), false},
	{"3.14", expression(ExpressionKindNumLitFloat, 3.14), false},
	{"-2", expression(ExpressionKindNumLitInt, int64(-2)), false},
	{"-2.5", expression(ExpressionKindNumLitFloat, -2.5), false},
	{"test", expression(ExpressionKindBinding, "test"), false},
	{`"a string"`, expression(ExpressionKindStringLit, "a string"), false},
	{`"an escaped\nstring"`, expression(ExpressionKindStringLit, "an escaped\nstring"), false},
	{`'a'`, expression(ExpressionKindNumLitInt, int64('a')), false},
	{`'\r'`, expression(ExpressionKindNumLitInt, int64('\r')), false},
	{"'abc'", Expression{}, true},
	{"0xFF", expression(ExpressionKindNumLitInt, int64(255)), false},
	{"0XFF", expression(ExpressionKindNumLitInt, int64(255)), false},
	{"0b0101", expression(ExpressionKindNumLitInt, int64(5)), false},
	{"0B0101", expression(ExpressionKindNumLitInt, int64(5)), false},
	{"2+3*4+5", expression(ExpressionKindNumLitInt, int64(19)), false},
	{"1.2+2.3", expression(ExpressionKindNumLitFloat, 3.5), false},
	{`"first"+"second"`, expression(ExpressionKindStringLit, "firstsecond"), false},
	{`"first"-"second"`, Expression{}, true},
	{`"first"*"second"`, Expression{}, true},
	{"1+test", expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(1)),
		Rhs:  expressionP(ExpressionKindBinding, "test"),
	}), false},
	{"2*1+test", expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
		Rhs:  expressionP(ExpressionKindBinding, "test"),
	}), false},
	{"2.1+1+test+\"str\"", expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindPlus,
		Lhs: expressionP(ExpressionKindBinaryOp, BinaryOp{
			Kind: BinaryOpKindPlus,
			Lhs: expressionP(ExpressionKindBinaryOp, BinaryOp{
				Kind: BinaryOpKindPlus,
				Lhs:  expressionP(ExpressionKindNumLitFloat, 2.1),
				Rhs:  expressionP(ExpressionKindNumLitInt, int64(1)),
			}),
			Rhs: expressionP(ExpressionKindBinding, "test"),
		}),
		Rhs: expressionP(ExpressionKindStringLit, "str"),
	}), false},
	{"-2*3", expression(ExpressionKindNumLitInt, int64(-6)), false},
	{"(1+2)*(1+2)", expression(ExpressionKindNumLitInt, int64(9)), false},
	{"1.0/2", expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindDivide,
		Lhs:  expressionP(ExpressionKindNumLitFloat, 1.0),
		Rhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
	}), false},
	{"5.2%2", expression(ExpressionKindBinaryOp, BinaryOp{
		Kind: BinaryOpKindModulo,
		Lhs:  expressionP(ExpressionKindNumLitFloat, 5.2),
		Rhs:  expressionP(ExpressionKindNumLitInt, int64(2)),
	}), false},
	{"0xG", Expression{}, true},
	{"0x", Expression{}, true},
	{"0b", Expression{}, true},
	{"1.2.3", Expression{}, true},
	{"1$", Expression{}, true},
	{"(1", Expression{}, true},
}

func TestParseExprFromString(t *testing.T) {
	for _, test := range testExpressions {
		expr, err := ParseExprFromString(test.in)

		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
			assert.Condition(t, func() (success bool) { return expressionEquals(expr, test.out) }, test)
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
		{"1,\"test\"", []byte{1, 't', 'e', 's', 't'}, false},
		{"1, 0xf", []byte{1, 0xf}, false},
		{"1 2 3", []byte{}, true},
		{"1,,2", []byte{}, true},
		{",1", []byte{}, true},
		{",", []byte{}, true},
		{"1.2,test", []byte{}, true},
	}

	for _, test := range tests {
		expr, err := ParseByteArrayFromString(test.in)

		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
			assert.Equal(t, test.out, expr, test)
		}
	}
}

// Compares two Expressions.
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
