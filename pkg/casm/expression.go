package casm

import (
	"log"
	"strconv"
)

type ExpressionKind int

const (
	ExpressionKindNumLitInt ExpressionKind = iota
	ExpressionKindNumLitFloat
	ExpressionKindBinding
)

type Expression struct {
	Kind          ExpressionKind
	AsNumLitInt   int64
	AsNumLitFloat float64
	AsBinding     string
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
		// Try integer
		intNumber, err := strconv.ParseInt(tokens[0].Text, 10, 64)
		if err != nil {
			// Try floating point
			floatNumber, err := strconv.ParseFloat(tokens[0].Text, 64)
			if err != nil {
				log.Fatalf("%s: [ERROR]: Error parsing number literal '%s'",
					location,
					tokens[0].Text)
			}
			result.Kind = ExpressionKindNumLitFloat
			result.AsNumLitFloat = floatNumber
		} else {
			result.Kind = ExpressionKindNumLitInt
			result.AsNumLitInt = intNumber
		}
	case TokenKindSymbol:
		result.Kind = ExpressionKindBinding
		result.AsBinding = tokens[0].Text
	}
	return result
}
