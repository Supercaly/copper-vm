package casm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestTokensEmpty(t *testing.T) {
	assert := assert.New(t)
	tokens := Tokens{}
	assert.Len(tokens, 0)
	assert.True(tokens.Empty())

	tokens = Tokens{
		token(TokenKindAsterisk, ""),
		token(TokenKindPlus, ""),
		token(TokenKindMinus, ""),
	}
	assert.Len(tokens, 3)
	assert.False(tokens.Empty())
}

func TestTokensFirst(t *testing.T) {
	assert := assert.New(t)
	tokens := Tokens{
		token(TokenKindAsterisk, ""),
		token(TokenKindPlus, ""),
		token(TokenKindMinus, ""),
	}
	token := tokens.First()
	assert.Equal(tokens[0], token)

	func() {
		defer func() { recover() }()
		tokens := Tokens{}
		tokens.First()
		assert.Fail("expecting an error")
	}()
}

func TestTokensPop(t *testing.T) {
	assert := assert.New(t)
	tokens := Tokens{
		token(TokenKindAsterisk, ""),
		token(TokenKindPlus, ""),
		token(TokenKindMinus, ""),
	}
	tokenBeforePop := tokens.First()
	token := tokens.Pop()
	assert.Equal(tokenBeforePop, token)
	assert.Len(tokens, 2)
	assert.True(!tokens.Empty())

	func() {
		defer func() { recover() }()
		tokens := Tokens{}
		tokens.Pop()
		assert.Fail("expecting an error")
	}()
}

// Wrapper function to create a Token.
func token(kind TokenKind, text string) Token {
	return Token{Kind: kind, Text: text}
}

var testTokens = []struct {
	in       string
	out      Tokens
	hasError bool
}{
	{"1", Tokens{token(TokenKindNumLit, "1")}, false},
	{"1.2", Tokens{token(TokenKindNumLit, "1.2")}, false},
	{".2", Tokens{token(TokenKindNumLit, ".2")}, false},
	{"test", Tokens{token(TokenKindSymbol, "test")}, false},
	{"-5", Tokens{
		token(TokenKindMinus, "-"),
		token(TokenKindNumLit, "5"),
	}, false},
	{"test12", Tokens{token(TokenKindSymbol, "test12")}, false},
	{"12test", Tokens{
		token(TokenKindNumLit, "12"),
		token(TokenKindSymbol, "test"),
	}, false},
	{"test_case", Tokens{token(TokenKindSymbol, "test_case")}, false},
	{"_test", Tokens{token(TokenKindSymbol, "_test")}, false},
	{"1,2,3", Tokens{
		token(TokenKindNumLit, "1"),
		token(TokenKindComma, ","),
		token(TokenKindNumLit, "2"),
		token(TokenKindComma, ","),
		token(TokenKindNumLit, "3"),
	}, false},
	{`"string"`, Tokens{token(TokenKindStringLit, "string")}, false},
	{`"string`, Tokens{}, true},
	{`'a'`, Tokens{token(TokenKindCharLit, "a")}, false},
	{`'a`, Tokens{}, true},
	{"0x5CFF", Tokens{token(TokenKindNumLit, "0x5CFF")}, false},
	{"0X5CFF", Tokens{token(TokenKindNumLit, "0X5CFF")}, false},
	{"0b110011", Tokens{token(TokenKindNumLit, "0b110011")}, false},
	{"0B110011", Tokens{token(TokenKindNumLit, "0B110011")}, false},
	{"+-*/%", Tokens{
		token(TokenKindPlus, "+"),
		token(TokenKindMinus, "-"),
		token(TokenKindAsterisk, "*"),
		token(TokenKindSlash, "/"),
		token(TokenKindPercent, "%"),
	}, false},
	{"()", Tokens{
		token(TokenKindOpenParen, "("),
		token(TokenKindCloseParen, ")"),
	}, false},
	{"[]", Tokens{
		token(TokenKindOpenBracket, "["),
		token(TokenKindCloseBracket, "]"),
	}, false},
	{"$", Tokens{}, true},
}

func TestTokenize(t *testing.T) {
	for _, test := range testTokens {
		tok, err := Tokenize(test.in)

		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
			assert.Equal(t, test.out, tok, test)
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
	s.result['0'] = true
	s.result['1'] = true
	s.result['2'] = true
	s.result['3'] = true
	s.result['4'] = true
	s.result['5'] = true
	s.result['6'] = true
	s.result['7'] = true
	s.result['8'] = true
	s.result['9'] = true
	s.result['.'] = true

	for i := 0; i < len(s.runes); i++ {
		assert.Equal(s.T(), s.result[i], isDigit(s.runes[i]), "rune: "+string(s.runes[i]))
	}
}

func (s *AsciiTestSuite) TestIsAlpha() {
	s.result['_'] = true
	s.result['0'] = true
	s.result['1'] = true
	s.result['2'] = true
	s.result['3'] = true
	s.result['4'] = true
	s.result['5'] = true
	s.result['6'] = true
	s.result['7'] = true
	s.result['8'] = true
	s.result['9'] = true
	s.result['A'] = true
	s.result['B'] = true
	s.result['C'] = true
	s.result['D'] = true
	s.result['E'] = true
	s.result['F'] = true
	s.result['G'] = true
	s.result['H'] = true
	s.result['I'] = true
	s.result['J'] = true
	s.result['K'] = true
	s.result['L'] = true
	s.result['M'] = true
	s.result['N'] = true
	s.result['O'] = true
	s.result['P'] = true
	s.result['Q'] = true
	s.result['R'] = true
	s.result['S'] = true
	s.result['T'] = true
	s.result['U'] = true
	s.result['V'] = true
	s.result['W'] = true
	s.result['X'] = true
	s.result['Y'] = true
	s.result['Z'] = true
	s.result['a'] = true
	s.result['b'] = true
	s.result['c'] = true
	s.result['d'] = true
	s.result['e'] = true
	s.result['f'] = true
	s.result['g'] = true
	s.result['h'] = true
	s.result['i'] = true
	s.result['j'] = true
	s.result['k'] = true
	s.result['l'] = true
	s.result['m'] = true
	s.result['n'] = true
	s.result['o'] = true
	s.result['p'] = true
	s.result['q'] = true
	s.result['r'] = true
	s.result['s'] = true
	s.result['t'] = true
	s.result['u'] = true
	s.result['v'] = true
	s.result['w'] = true
	s.result['x'] = true
	s.result['y'] = true
	s.result['z'] = true

	for i := 0; i < len(s.runes); i++ {
		assert.Equal(s.T(), s.result[i], isAlpha(s.runes[i]), "rune: "+string(s.runes[i]))
	}
}

func (s *AsciiTestSuite) TestIsHex() {
	s.result['0'] = true
	s.result['1'] = true
	s.result['2'] = true
	s.result['3'] = true
	s.result['4'] = true
	s.result['5'] = true
	s.result['6'] = true
	s.result['7'] = true
	s.result['8'] = true
	s.result['9'] = true
	s.result['A'] = true
	s.result['B'] = true
	s.result['C'] = true
	s.result['D'] = true
	s.result['E'] = true
	s.result['F'] = true
	s.result['a'] = true
	s.result['b'] = true
	s.result['c'] = true
	s.result['d'] = true
	s.result['e'] = true
	s.result['f'] = true
	s.result['X'] = true
	s.result['x'] = true

	for i := 0; i < len(s.runes); i++ {
		assert.Equal(s.T(), s.result[i], isHex(s.runes[i]), "rune: "+string(s.runes[i]))
	}
}

func (s *AsciiTestSuite) TestIsNumber() {
	s.result['.'] = true
	s.result['0'] = true
	s.result['1'] = true
	s.result['2'] = true
	s.result['3'] = true
	s.result['4'] = true
	s.result['5'] = true
	s.result['6'] = true
	s.result['7'] = true
	s.result['8'] = true
	s.result['9'] = true
	s.result['A'] = true
	s.result['B'] = true
	s.result['C'] = true
	s.result['D'] = true
	s.result['E'] = true
	s.result['F'] = true
	s.result['a'] = true
	s.result['b'] = true
	s.result['c'] = true
	s.result['d'] = true
	s.result['e'] = true
	s.result['f'] = true
	s.result['X'] = true
	s.result['x'] = true

	for i := 0; i < len(s.runes); i++ {
		assert.Equal(s.T(), s.result[i], isNumber(s.runes[i]), "rune: "+string(s.runes[i]))
	}
}

func TestAsciiSuite(t *testing.T) {
	suite.Run(t, new(AsciiTestSuite))
}
