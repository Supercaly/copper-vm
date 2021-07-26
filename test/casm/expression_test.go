package casm

import (
	"testing"

	"coppervm.com/coppervm/pkg/casm"
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
		{"1.2.3", casm.Expression{}, true},
		{"1+", casm.Expression{}, true},
	}

	for _, test := range tests {
		expr, err := casm.ParseExprFromString(test.in)

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if expr != test.out {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, expr)
		}
	}
}

func TestParseByteListFromString(t *testing.T) {
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
		{"1 2 3", []byte{}, true},
		{"1,,2", []byte{}, true},
		{",1", []byte{}, true},
		{",", []byte{}, true},
	}

	for _, test := range tests {
		expr, err := casm.ParseByteListFromString(test.in)

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
