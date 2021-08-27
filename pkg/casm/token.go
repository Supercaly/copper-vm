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
	TokenKindSymbol
	TokenKindPlus
	TokenKindMinus
	TokenKindAsterisk
	TokenKindComma
	TokenKindOpenParen
	TokenKindCloseParen
)

type Token struct {
	Kind TokenKind
	Text string
}

// Tokenize a source string.
// Returns a list of tokens from a string or an error
// if something went wrong.
func Tokenize(source string) (out []Token, err error) {
	for source != "" {
		source = strings.TrimSpace(source)
		switch source[0] {
		case '+':
			source = source[1:]
			out = append(out, Token{
				Kind: TokenKindPlus,
				Text: "+",
			})
		case '-':
			source = source[1:]
			out = append(out, Token{
				Kind: TokenKindMinus,
				Text: "-",
			})
		case '*':
			source = source[1:]
			out = append(out, Token{
				Kind: TokenKindAsterisk,
				Text: "*",
			})
		case ',':
			source = source[1:]
			out = append(out, Token{
				Kind: TokenKindComma,
				Text: ",",
			})
		case '"':
			source = source[1:]
			if strings.Contains(source, "\"") {
				str, rest := internal.SplitByDelim(source, '"')
				source = rest[1:]
				unquotedStr, err := strconv.Unquote(`"` + str + `"`)
				if err != nil {
					return []Token{},
						fmt.Errorf("error tokenizing literal string '%s'", str)
				}
				out = append(out, Token{
					Kind: TokenKindStringLit,
					Text: unquotedStr,
				})
			} else {
				return []Token{}, fmt.Errorf("could not find closing \"")
			}
		case '(':
			source = source[1:]
			out = append(out, Token{
				Kind: TokenKindOpenParen,
				Text: "(",
			})
		case ')':
			source = source[1:]
			out = append(out, Token{
				Kind: TokenKindCloseParen,
				Text: ")",
			})
		default:
			if isDigit(rune(source[0])) {
				// Tokenize a number
				number, rest := internal.SplitWhile(source, isNumber)
				source = rest
				out = append(out, Token{
					Kind: TokenKindNumLit,
					Text: number,
				})
			} else if isAlpha(rune(source[0])) {
				// Tokenize a symbol
				symbol, rest := internal.SplitWhile(source, isAlpha)
				source = rest
				out = append(out, Token{
					Kind: TokenKindSymbol,
					Text: symbol,
				})
			} else {
				return []Token{},
					fmt.Errorf("unknown token starting with '%s'", string(source[0]))
			}
		}
	}
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
