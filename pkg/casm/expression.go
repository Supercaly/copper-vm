package casm

import (
	"errors"
	"fmt"
	"strconv"
)

type ExpressionKind int

const (
	ExpressionKindNumLitInt ExpressionKind = iota
	ExpressionKindNumLitFloat
	ExpressionKindBinding
)

func (kind ExpressionKind) String() string {
	return [...]string{
		"ExpressionKindNumLitInt",
		"ExpressionKindNumLitFloat",
		"ExpressionKindBinding",
	}[kind]
}

type Expression struct {
	Kind          ExpressionKind
	AsNumLitInt   int64
	AsNumLitFloat float64
	AsBinding     string
}

// Parse an expression from a source string.
// The string is first tokenized and then is parsed to extract
// an expression.
// Returns an error if something went wrong.
func ParseExprFromString(source string) (Expression, error) {
	tokens, err := Tokenize(source)
	if err != nil {
		return Expression{}, err
	}
	return parseExprPrimary(tokens)
}

// Parse a primary expression form a list of tokens.
// Returns an error if something went wrong.
func parseExprPrimary(tokens []Token) (result Expression, err error) {
	if len(tokens) == 0 {
		return Expression{}, errors.New("trying to parse empty expression")
	}
	switch tokens[0].Kind {
	case TokenKindNumLit:
		// Try integer
		intNumber, err := strconv.ParseInt(tokens[0].Text, 10, 64)
		if err != nil {
			// Try floating point
			floatNumber, err := strconv.ParseFloat(tokens[0].Text, 64)
			if err != nil {
				return Expression{},
					fmt.Errorf("error parsing number literal '%s'",
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
	case TokenKindMinus:
		result, err = parseExprPrimary(tokens[1:])
		if result.Kind == ExpressionKindNumLitInt {
			result.AsNumLitInt = -result.AsNumLitInt
		} else if result.Kind == ExpressionKindNumLitFloat {
			result.AsNumLitFloat = -result.AsNumLitFloat
		}
	}
	return result, err
}

// Parse a byte list from a source string.
// The string is first tokenized and then is parsed to extract
// the data.
// Returns an error if something went wrong.
func ParseByteListFromString(source string) (out []byte, err error) {
	tokens, err := Tokenize(source)
	if err != nil {
		return []byte{}, err
	}
	return parseByteArrayFromTokens(tokens)
}

// Parse a byte list from some tokens.
// Returns a byte array or an error.
func parseByteArrayFromTokens(tokens []Token) (out []byte, err error) {
	if len(tokens) == 0 {
		return []byte{}, nil
	}

	if tokens[0].Kind == TokenKindComma {
		return []byte{}, fmt.Errorf("misplaced comma inside list")
	}

	expr, err := parseExprPrimary([]Token{tokens[0]})
	if err != nil {
		return []byte{}, err
	}
	if expr.Kind != ExpressionKindNumLitInt {
		return []byte{}, fmt.Errorf("unsupported value inside byte array")
	}
	out = append(out, byte(expr.AsNumLitInt))

	if len(tokens) > 1 && tokens[1].Kind != TokenKindComma {
		return []byte{},
			fmt.Errorf("array values must be comma separated")
	}

	if len(tokens) > 2 {
		next, err := parseByteArrayFromTokens(tokens[2:])
		if err != nil {
			return []byte{}, err
		}
		out = append(out, next...)
	}

	return out, nil
}
