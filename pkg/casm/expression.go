package casm

import "strconv"

type ExpressionKind int

const (
	ExpressionKindNumLit ExpressionKind = iota
)

type Expression struct {
	Kind     ExpressionKind
	AsNumLit int
}

func ParseExprFromString(source string, location FileLocation) Expression {
	tokens := Tokenize(source, location)
	return parseExprPrimary(tokens)
}

func parseExprPrimary(tokens []Token) (result Expression) {
	switch tokens[0].Kind {
	case TokenKindNumLit:
		result.Kind = ExpressionKindNumLit
		result.AsNumLit, _ = strconv.Atoi(tokens[0].Text)
	case TokenKindSymbol:
	}
	return result
}
