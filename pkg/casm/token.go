package casm

import (
	"log"
	"strings"
	"unicode"
)

type TokenKind int

const (
	TokenKindNumLit TokenKind = iota
	TokenKindSymbol
	TokenKindMinus
)

type Token struct {
	Kind TokenKind
	Text string
}

// Tokenize a source string.
// Returns a list of tokens from a string
func Tokenize(source string, location FileLocation) (out []Token) {
	for source != "" {
		source = strings.TrimSpace(source)
		switch source[0] {
		case '-':
			source = source[1:]
			out = append(out, Token{
				Kind: TokenKindMinus,
				Text: "-",
			})
		default:
			if isAlpha(rune(source[0])) {
				// Tokenize a symbol
				symbol, rest := SplitWhile(source, isAlpha)
				source = rest
				out = append(out, Token{
					Kind: TokenKindSymbol,
					Text: symbol,
				})
			} else if isDigit(rune(source[0])) {
				// Tokenize a number
				number, rest := SplitWhile(source, isDigit)
				source = rest
				out = append(out, Token{
					Kind: TokenKindNumLit,
					Text: number,
				})
			} else {
				log.Fatalf("%s: [ERROR]: Unknown token starting with '%s'",
					location,
					string(source[0]))
			}
		}
	}
	return out
}

func isAlpha(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isDigit(r rune) bool {
	return unicode.IsNumber(r) || r == '.'
}
