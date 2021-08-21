package casm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

		if test.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.out, tok)
		}
	}
}

type AsciiTestSuite struct {
	suite.Suite
	runes  [128]rune
	result [128]bool
}

func (s *AsciiTestSuite) SetupSuite() {
	for i := 0; i < 127; i++ {
		s.runes[i] = rune(i)
	}
}

func (s *AsciiTestSuite) SetupTest() {
	for i := 0; i < 127; i++ {
		s.result[i] = false
	}
}

func (s *AsciiTestSuite) TestIsDigit() {
	s.result[46] = true
	s.result[48] = true
	s.result[49] = true
	s.result[50] = true
	s.result[51] = true
	s.result[52] = true
	s.result[53] = true
	s.result[54] = true
	s.result[55] = true
	s.result[56] = true
	s.result[57] = true

	for i := 0; i < len(s.runes); i++ {
		assert.Equal(s.T(), s.result[i], isDigit(s.runes[i]))
	}
}

func (s *AsciiTestSuite) TestIsAlpha() {
	s.result[48] = true
	s.result[49] = true
	s.result[50] = true
	s.result[51] = true
	s.result[52] = true
	s.result[53] = true
	s.result[54] = true
	s.result[55] = true
	s.result[56] = true
	s.result[57] = true
	s.result[65] = true
	s.result[66] = true
	s.result[67] = true
	s.result[68] = true
	s.result[69] = true
	s.result[70] = true
	s.result[71] = true
	s.result[72] = true
	s.result[73] = true
	s.result[74] = true
	s.result[75] = true
	s.result[76] = true
	s.result[77] = true
	s.result[78] = true
	s.result[79] = true
	s.result[80] = true
	s.result[81] = true
	s.result[82] = true
	s.result[83] = true
	s.result[84] = true
	s.result[85] = true
	s.result[86] = true
	s.result[87] = true
	s.result[88] = true
	s.result[89] = true
	s.result[90] = true
	s.result[95] = true
	s.result[97] = true
	s.result[98] = true
	s.result[99] = true
	s.result[100] = true
	s.result[101] = true
	s.result[102] = true
	s.result[103] = true
	s.result[104] = true
	s.result[105] = true
	s.result[106] = true
	s.result[107] = true
	s.result[108] = true
	s.result[109] = true
	s.result[110] = true
	s.result[111] = true
	s.result[112] = true
	s.result[113] = true
	s.result[114] = true
	s.result[115] = true
	s.result[116] = true
	s.result[117] = true
	s.result[118] = true
	s.result[119] = true
	s.result[120] = true
	s.result[121] = true
	s.result[122] = true

	for i := 0; i < len(s.runes); i++ {
		assert.Equal(s.T(), s.result[i], isAlpha(s.runes[i]))
	}
}

func (s *AsciiTestSuite) TestIsHex() {
	s.result[48] = true
	s.result[49] = true
	s.result[50] = true
	s.result[51] = true
	s.result[52] = true
	s.result[53] = true
	s.result[54] = true
	s.result[55] = true
	s.result[56] = true
	s.result[57] = true
	s.result[65] = true
	s.result[66] = true
	s.result[67] = true
	s.result[68] = true
	s.result[69] = true
	s.result[70] = true
	s.result[97] = true
	s.result[98] = true
	s.result[99] = true
	s.result[100] = true
	s.result[101] = true
	s.result[102] = true
	s.result[120] = true

	for i := 0; i < len(s.runes); i++ {
		assert.Equal(s.T(), s.result[i], isHex(s.runes[i]))
	}
}

func (s *AsciiTestSuite) TestIsNumber() {
	s.result[46] = true
	s.result[48] = true
	s.result[49] = true
	s.result[50] = true
	s.result[51] = true
	s.result[52] = true
	s.result[53] = true
	s.result[54] = true
	s.result[55] = true
	s.result[56] = true
	s.result[57] = true
	s.result[65] = true
	s.result[66] = true
	s.result[67] = true
	s.result[68] = true
	s.result[69] = true
	s.result[70] = true
	s.result[97] = true
	s.result[98] = true
	s.result[99] = true
	s.result[100] = true
	s.result[101] = true
	s.result[102] = true
	s.result[120] = true

	for i := 0; i < len(s.runes); i++ {
		assert.Equal(s.T(), s.result[i], isNumber(s.runes[i]))
	}
}

func TestAsciiSuite(t *testing.T) {
	suite.Run(t, new(AsciiTestSuite))
}
