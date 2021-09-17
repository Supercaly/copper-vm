package casm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestTokensEmpty(t *testing.T) {
	assert := assert.New(t)
	tok := tokens{}
	assert.Len(tok, 0)
	assert.True(tok.Empty())

	tok = tokens{
		newToken(tokenKindAsterisk, "", fileLocation(0, 0)),
		newToken(tokenKindPlus, "", fileLocation(0, 0)),
		newToken(tokenKindMinus, "", fileLocation(0, 0)),
	}
	assert.Len(tok, 3)
	assert.False(tok.Empty())
}

func TestTokensFirst(t *testing.T) {
	assert := assert.New(t)
	tok := tokens{
		newToken(tokenKindAsterisk, "", fileLocation(0, 0)),
		newToken(tokenKindPlus, "", fileLocation(0, 0)),
		newToken(tokenKindMinus, "", fileLocation(0, 0)),
	}
	token := tok.First()
	assert.Equal(tok[0], token)

	func() {
		defer func() { recover() }()
		tok := tokens{}
		tok.First()
		assert.Fail("expecting an error")
	}()
}

func TestTokensPop(t *testing.T) {
	assert := assert.New(t)
	tok := tokens{
		newToken(tokenKindAsterisk, "", fileLocation(0, 0)),
		newToken(tokenKindPlus, "", fileLocation(0, 0)),
		newToken(tokenKindMinus, "", fileLocation(0, 0)),
	}
	tokenBeforePop := tok.First()
	token := tok.Pop()
	assert.Equal(tokenBeforePop, token)
	assert.Len(tok, 2)
	assert.True(!tok.Empty())

	func() {
		defer func() { recover() }()
		tok := tokens{}
		tok.Pop()
		assert.Fail("expecting an error")
	}()
}

func TestExpectTokenKind(t *testing.T) {
	tests := []struct {
		in       tokens
		expected tokenKind
		hasError bool
	}{
		{tokens{
			newToken(tokenKindAsterisk, "", fileLocation(0, 0)),
			newToken(tokenKindPlus, "", fileLocation(0, 0)),
			newToken(tokenKindMinus, "", fileLocation(0, 0)),
		}, tokenKindAsterisk, false},
		{tokens{
			newToken(tokenKindPlus, "", fileLocation(0, 0)),
		}, tokenKindPlus, false},
		{tokens{
			newToken(tokenKindPlus, "", fileLocation(0, 0)),
		}, tokenKindSymbol, true},
		{tokens{}, tokenKindSymbol, true},
	}

	for _, test := range tests {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			test.in.expectTokenKind(test.expected)
			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			}
		}()
	}
}

// Wrapper function to create a FileLocation.
func fileLocation(row int, col int) FileLocation {
	return FileLocation{FileName: "", Col: col, Row: row}
}

// Wrapper function to create a Token.
func newToken(kind tokenKind, text string, location FileLocation) token {
	return token{kind, text, location}
}

var testTokens = []struct {
	in       string
	out      tokens
	hasError bool
}{
	{"1", tokens{newToken(tokenKindNumLit, "1", fileLocation(0, 0))}, false},
	{"1.2", tokens{newToken(tokenKindNumLit, "1.2", fileLocation(0, 0))}, false},
	{".2", tokens{newToken(tokenKindNumLit, ".2", fileLocation(0, 0))}, false},
	{"test", tokens{newToken(tokenKindSymbol, "test", fileLocation(0, 0))}, false},
	{"-5", tokens{
		newToken(tokenKindMinus, "", fileLocation(0, 0)),
		newToken(tokenKindNumLit, "5", fileLocation(0, 1)),
	}, false},
	{"test12", tokens{newToken(tokenKindSymbol, "test12", fileLocation(0, 0))}, false},
	{"12test", tokens{
		newToken(tokenKindNumLit, "12", fileLocation(0, 0)),
		newToken(tokenKindSymbol, "test", fileLocation(0, 2)),
	}, false},
	{"test_case", tokens{newToken(tokenKindSymbol, "test_case", fileLocation(0, 0))}, false},
	{"_test", tokens{newToken(tokenKindSymbol, "_test", fileLocation(0, 0))}, false},
	{"1,2,3", tokens{
		newToken(tokenKindNumLit, "1", fileLocation(0, 0)),
		newToken(tokenKindComma, "", fileLocation(0, 1)),
		newToken(tokenKindNumLit, "2", fileLocation(0, 2)),
		newToken(tokenKindComma, "", fileLocation(0, 3)),
		newToken(tokenKindNumLit, "3", fileLocation(0, 4)),
	}, false},
	{`"string"`, tokens{newToken(tokenKindStringLit, "string", fileLocation(0, 0))}, false},
	{`"string`, tokens{}, true},
	{`'a'`, tokens{newToken(tokenKindCharLit, "a", fileLocation(0, 0))}, false},
	{`'a`, tokens{}, true},
	{"0x5CFF", tokens{newToken(tokenKindNumLit, "0x5CFF", fileLocation(0, 0))}, false},
	{"0X5CFF", tokens{newToken(tokenKindNumLit, "0X5CFF", fileLocation(0, 0))}, false},
	{"0b110011", tokens{newToken(tokenKindNumLit, "0b110011", fileLocation(0, 0))}, false},
	{"0B110011", tokens{newToken(tokenKindNumLit, "0B110011", fileLocation(0, 0))}, false},
	{"+-*/%", tokens{
		newToken(tokenKindPlus, "", fileLocation(0, 0)),
		newToken(tokenKindMinus, "", fileLocation(0, 1)),
		newToken(tokenKindAsterisk, "", fileLocation(0, 2)),
		newToken(tokenKindSlash, "", fileLocation(0, 3)),
		newToken(tokenKindPercent, "", fileLocation(0, 4)),
	}, false},
	{"()", tokens{
		newToken(tokenKindOpenParen, "", fileLocation(0, 0)),
		newToken(tokenKindCloseParen, "", fileLocation(0, 1)),
	}, false},
	{"[]", tokens{
		newToken(tokenKindOpenBracket, "", fileLocation(0, 0)),
		newToken(tokenKindCloseBracket, "", fileLocation(0, 1)),
	}, false},
	{"\n", tokens{newToken(tokenKindNewLine, "", fileLocation(0, 0))}, false},
	{":", tokens{newToken(tokenKindColon, "", fileLocation(0, 0))}, false},
	{"$", tokens{}, true},
}

func TestTokenize(t *testing.T) {
	for _, test := range testTokens {
		func() {
			defer func() {
				r := recover()
				if r != nil && !test.hasError {
					assert.Fail(t, "unexpected error", test)
				}
			}()

			tok := tokenize(test.in, "")

			if test.hasError {
				assert.Fail(t, "expecting an error", test)
			}
			assert.Equal(t, test.out, tok, test)
		}()
	}
}

func TestTokenizationDeDuplication(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "unexpected error")
		}
	}()

	tok := tokenize("a\n\n\n\nb", "")
	assert.Equal(t, tokens{
		newToken(tokenKindSymbol, "a", fileLocation(0, 0)),
		newToken(tokenKindNewLine, "", fileLocation(0, 1)),
		newToken(tokenKindSymbol, "b", fileLocation(4, 0)),
	}, tok)
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
