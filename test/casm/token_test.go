package casm

import (
	"testing"

	"coppervm.com/coppervm/pkg/casm"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		in       string
		out      []casm.Token
		hasError bool
	}{
		{"1", []casm.Token{
			{Kind: casm.TokenKindNumLit, Text: "1"},
		}, false},
		{"1.2", []casm.Token{
			{Kind: casm.TokenKindNumLit, Text: "1.2"},
		}, false},
		{".2", []casm.Token{
			{Kind: casm.TokenKindNumLit, Text: ".2"},
		}, false},
		{"test", []casm.Token{
			{Kind: casm.TokenKindSymbol, Text: "test"},
		}, false},
		{"-5", []casm.Token{
			{Kind: casm.TokenKindMinus, Text: "-"},
			{Kind: casm.TokenKindNumLit, Text: "5"},
		}, false},
		{"test12", []casm.Token{
			{Kind: casm.TokenKindSymbol, Text: "test12"},
		}, false},
		{"12test", []casm.Token{
			{Kind: casm.TokenKindNumLit, Text: "12"},
			{Kind: casm.TokenKindSymbol, Text: "test"},
		}, false},
		{"test_case", []casm.Token{
			{Kind: casm.TokenKindSymbol, Text: "test_case"},
		}, false},
		{"_test", []casm.Token{
			{Kind: casm.TokenKindSymbol, Text: "_test"},
		}, false},
		{"+", []casm.Token{}, true},
	}

	for _, test := range tests {
		tok, err := casm.Tokenize(test.in)

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if !tokenArrayEquals(tok, test.out) {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, tok)
		}
	}
}

func tokenArrayEquals(a []casm.Token, b []casm.Token) bool {
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}