package casm

import (
	"log"
	"strconv"
)

type ExpressionKind int

const (
	ExpressionKindNumLit ExpressionKind = iota
	ExpressionKindBinding
)

type Expression struct {
	Kind      ExpressionKind
	AsNumLit  int
	AsBinding string
}

func ParseExprFromString(source string, location FileLocation) Expression {
	tokens := Tokenize(source, location)
	return parseExprPrimary(tokens, location)
}

func parseExprPrimary(tokens []Token, location FileLocation) (result Expression) {
	if len(tokens) == 0 {
		log.Fatalf("%s: [ERROR]: Trying to parse empty expression!",
			location)
	}
	switch tokens[0].Kind {
	case TokenKindNumLit:
		result.Kind = ExpressionKindNumLit
		var err error = nil
		result.AsNumLit, err = strconv.Atoi(tokens[0].Text)
		if err != nil {
			log.Fatalf("%s: [ERROR]: Error parsing number literal '%s'",
				location,
				tokens[0].Text)
		}
	case TokenKindSymbol:
		result.Kind = ExpressionKindBinding
		result.AsBinding = tokens[0].Text
	}
	return result
}
