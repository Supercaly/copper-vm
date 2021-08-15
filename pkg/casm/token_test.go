package casm

import (
	"reflect"
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
		} else if err == nil && !reflect.DeepEqual(tok, test.out) {
			t.Errorf("Expected '%#v' but got '%#v'", test.out, tok)
		}
	}
}

type asciiTestHolder struct {
	in  rune
	out bool
}

func getAsciArray() (out []asciiTestHolder) {
	for i := 32; i < 127; i++ {
		out = append(out, asciiTestHolder{
			in:  rune(i),
			out: false,
		})
	}
	return out
}

func TestIsDigit(t *testing.T) {
	tests := getAsciArray()
	for i := 48; i < 58; i++ {
		tests[i-32].out = true
	}
	tests[46-32].out = true

	for _, test := range tests {
		if isDigit(test.in) != test.out {
			t.Errorf("Expecting %s %t but got %t", string(test.in), test.out, !test.out)
		}
	}
}

func TestIsAlpha(t *testing.T) {
	tests := getAsciArray()
	for i := 48; i < 58; i++ {
		tests[i-32].out = true
	}
	for i := 65; i < 91; i++ {
		tests[i-32].out = true
	}
	for i := 97; i < 123; i++ {
		tests[i-32].out = true
	}
	tests[95-32].out = true

	for _, test := range tests {
		if isAlpha(test.in) != test.out {
			t.Errorf("Expecting %s %t but got %t", string(test.in), test.out, !test.out)
		}
	}
}

func TestIsHex(t *testing.T) {
	tests := getAsciArray()
	for i := 48; i < 58; i++ {
		tests[i-32].out = true
	}
	for i := 65; i < 71; i++ {
		tests[i-32].out = true
	}
	for i := 97; i < 103; i++ {
		tests[i-32].out = true
	}
	tests[120-32].out = true

	for _, test := range tests {
		if isHex(test.in) != test.out {
			t.Errorf("Expecting %s %t but got %t", string(test.in), test.out, !test.out)
		}
	}
}
func TestIsNumber(t *testing.T) {
	tests := getAsciArray()
	for i := 48; i < 58; i++ {
		tests[i-32].out = true
	}
	for i := 65; i < 71; i++ {
		tests[i-32].out = true
	}
	for i := 97; i < 103; i++ {
		tests[i-32].out = true
	}
	tests[120-32].out = true
	tests[46-32].out = true

	for _, test := range tests {
		if isNumber(test.in) != test.out {
			t.Errorf("Expecting %s %t but got %t", string(test.in), test.out, !test.out)
		}
	}
}
