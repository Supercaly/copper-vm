package casm

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		in       string
		out      []Token
		hasError bool
	}{
		{"1", []Token{
			{Kind: TokenKindNumLit, Text: "1"},
		}, false},
		{"1.2", []Token{
			{Kind: TokenKindNumLit, Text: "1.2"},
		}, false},
		{".2", []Token{
			{Kind: TokenKindNumLit, Text: ".2"},
		}, false},
		{"test", []Token{
			{Kind: TokenKindSymbol, Text: "test"},
		}, false},
		{"-5", []Token{
			{Kind: TokenKindMinus, Text: "-"},
			{Kind: TokenKindNumLit, Text: "5"},
		}, false},
		{"test12", []Token{
			{Kind: TokenKindSymbol, Text: "test12"},
		}, false},
		{"12test", []Token{
			{Kind: TokenKindNumLit, Text: "12"},
			{Kind: TokenKindSymbol, Text: "test"},
		}, false},
		{"test_case", []Token{
			{Kind: TokenKindSymbol, Text: "test_case"},
		}, false},
		{"_test", []Token{
			{Kind: TokenKindSymbol, Text: "_test"},
		}, false},
		{"1,2,3", []Token{
			{Kind: TokenKindNumLit, Text: "1"},
			{Kind: TokenKindComma, Text: ","},
			{Kind: TokenKindNumLit, Text: "2"},
			{Kind: TokenKindComma, Text: ","},
			{Kind: TokenKindNumLit, Text: "3"},
		}, false},
		{"\"string\"", []Token{
			{Kind: TokenKindStringLit, Text: "string"},
		}, false},
		{"\"string", []Token{}, true},
		{"0x5CFF", []Token{
			{Kind: TokenKindNumLit, Text: "0x5CFF"},
		}, false},
		{"+-*", []Token{
			{Kind: TokenKindPlus, Text: "+"},
			{Kind: TokenKindMinus, Text: "-"},
			{Kind: TokenKindAsterisk, Text: "*"},
		}, false},
		{"()", []Token{
			{Kind: TokenKindOpenParen, Text: "("},
			{Kind: TokenKindCloseParen, Text: ")"},
		}, false},
		{"$", []Token{}, true},
	}

	for _, test := range tests {
		tok, err := Tokenize(test.in)

		if err != nil && !test.hasError {
			t.Error(err)
		} else if err == nil && test.hasError {
			t.Errorf("Expecting an error")
		} else if !tokenArrayEquals(tok, test.out) {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, tok)
		}
	}
}

func tokenArrayEquals(a []Token, b []Token) bool {
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
