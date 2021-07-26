package casm

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenKind int

const (
	TokenKindNumLit TokenKind = iota
	TokenKindSymbol
	TokenKindMinus
	TokenKindComma
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
		case '-':
			source = source[1:]
			out = append(out, Token{
				Kind: TokenKindMinus,
				Text: "-",
			})
		case ',':
			source = source[1:]
			out = append(out, Token{
				Kind: TokenKindComma,
				Text: ",",
			})
		default:
			if isDigit(rune(source[0])) {
				// Tokenize a number
				number, rest := SplitWhile(source, isDigit)
				source = rest
				out = append(out, Token{
					Kind: TokenKindNumLit,
					Text: number,
				})
			} else if isAlpha(rune(source[0])) {
				// Tokenize a symbol
				symbol, rest := SplitWhile(source, isAlpha)
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

func isAlpha(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

func isDigit(r rune) bool {
	return unicode.IsNumber(r) || r == '.'
}
