package casm

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/Supercaly/coppervm/internal"
)

type TokenKind int

const (
	TokenKindNumLit TokenKind = iota
	TokenKindStringLit
	TokenKindCharLit
	TokenKindSymbol
	TokenKindPlus
	TokenKindMinus
	TokenKindAsterisk
	TokenKindSlash
	TokenKindPercent
	TokenKindComma
	TokenKindOpenParen
	TokenKindCloseParen
	TokenKindOpenBracket
	TokenKindCloseBracket
	TokenKindNewLine
	TokenKindColon
)

type Token struct {
	Kind     TokenKind
	Text     string
	Location FileLocation
}

type Tokens []Token

// Returns the first Token of the Tokens list.
func (t Tokens) First() Token {
	if len(t) == 0 {
		panic("trying to access the elements of an empty tokens list")
	}
	return t[0]
}

// Returns true if the Tokens list if empty, false otherwise.
func (t Tokens) Empty() bool {
	return len(t) == 0
}

// Removes and returns the first element of the Tokens list.
func (t *Tokens) Pop() (out Token) {
	if len(*t) == 0 {
		panic("trying to pop the elements of an empty tokens list")
	}
	out = (*t)[0]
	*t = (*t)[1:]
	return out
}

// Tokenize a source string.
// Returns a list of tokens from a string or an error
// if something went wrong.
func Tokenize(source string, filePath string) (out Tokens, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	location := FileLocation{FileName: filePath}

	// Tokenize the whole source string
	for len(source) != 0 {
		switch source[0] {
		case ' ':
			source = source[1:]
			location.Col++
		case ';':
			var comment string
			comment, source = internal.SplitByDelim(source, '\n')
			location.Col += len(comment)
		case '\n':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindNewLine, Location: location})
			location.Row++
			location.Col = 0
		case '+':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindPlus, Location: location})
			location.Col++
		case '-':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindMinus, Location: location})
			location.Col++
		case '*':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindAsterisk, Location: location})
			location.Col++
		case '/':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindSlash, Location: location})
			location.Col++
		case '%':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindPercent, Location: location})
			location.Col++
		case ',':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindComma, Location: location})
			location.Col++
		case ':':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindColon, Location: location})
			location.Col++
		case '(':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindOpenParen, Location: location})
			location.Col++
		case ')':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindCloseParen, Location: location})
			location.Col++
		case '[':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindOpenBracket, Location: location})
			location.Col++
		case ']':
			source = source[1:]
			out = append(out, Token{Kind: TokenKindCloseBracket, Location: location})
			location.Col++
		case '"':
			source = source[1:]
			if strings.Contains(source, "\"") {
				str, rest := internal.SplitByDelim(source, '"')
				source = rest[1:]
				unquotedStr, err := strconv.Unquote(`"` + str + `"`)
				if err != nil {
					panic(fmt.Sprintf("error tokenizing literal string '%s'", str))
				}
				out = append(out, Token{
					Kind:     TokenKindStringLit,
					Text:     unquotedStr,
					Location: location,
				})
				// TODO: Location in not incremented correctly if there's a new line in the string
				location.Col += len(unquotedStr) + 2
			} else {
				panic("could not find closing \"")
			}
		case '\'':
			source = source[1:]
			if strings.Contains(source, "'") {
				char, rest := internal.SplitByDelim(source, '\'')
				source = rest[1:]
				out = append(out, Token{
					Kind:     TokenKindCharLit,
					Text:     char,
					Location: location,
				})
				// TODO: Location in not incremented correctly if there's a new line in the char
				location.Col += len(char) + 2
			} else {
				panic("could not find closing '")
			}
		default:
			if isDigit(rune(source[0])) {
				// Tokenize a number
				number, rest := internal.SplitWhile(source, isNumber)
				source = rest
				out = append(out, Token{
					Kind:     TokenKindNumLit,
					Text:     number,
					Location: location,
				})
				location.Col += len(number)
			} else if isAlpha(rune(source[0])) {
				// Tokenize a symbol
				symbol, rest := internal.SplitWhile(source, isAlpha)
				source = rest
				out = append(out, Token{
					Kind:     TokenKindSymbol,
					Text:     symbol,
					Location: location,
				})
				location.Col += len(symbol)
			} else {
				panic(fmt.Sprintf("%s: unknown token starting with '%s'", location, string(source[0])))
			}
		}
	}

	// Remove duplicate consecutive new lines
	var newOut []Token
	var lastToken Token
	for _, t := range out {
		if t.Kind != TokenKindNewLine || lastToken.Kind != TokenKindNewLine {
			newOut = append(newOut, t)
		}
		lastToken = t
	}
	out = newOut

	return out, err
}

func isNumber(r rune) bool {
	return isDigit(r) || isHex(r)
}

func isHex(r rune) bool {
	return unicode.In(r, unicode.Hex_Digit) || r == 'x' || r == 'X'
}

func isAlpha(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

func isDigit(r rune) bool {
	return unicode.IsNumber(r) || r == '.'
}
