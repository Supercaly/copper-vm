package casm

import (
	"fmt"
	"strconv"
	"strings"
)

type ExpressionKind int

const (
	ExpressionKindNumLitInt ExpressionKind = iota
	ExpressionKindNumLitFloat
	ExpressionKindStringLit
	ExpressionKindBinaryOp
	ExpressionKindBinding
)

func (kind ExpressionKind) String() string {
	return [...]string{
		"ExpressionKindNumLitInt",
		"ExpressionKindNumLitFloat",
		"ExpressionKindStringLit",
		"ExpressionKindBinaryOp",
		"ExpressionKindBinding",
	}[kind]
}

type Expression struct {
	Kind          ExpressionKind
	AsNumLitInt   int64
	AsNumLitFloat float64
	AsStringLit   string
	AsBinaryOp    BinaryOp
	AsBinding     string
}

func (expr Expression) String() (out string) {
	out += "{"
	out += fmt.Sprintf("Kind: %s, ", expr.Kind)
	switch expr.Kind {
	case ExpressionKindNumLitInt:
		out += fmt.Sprintf("AsNumLitInt: %d", expr.AsNumLitInt)
	case ExpressionKindNumLitFloat:
		out += fmt.Sprintf("AsNumLitFloat: %f", expr.AsNumLitFloat)
	case ExpressionKindStringLit:
		out += fmt.Sprintf("AsStringLit: %s", expr.AsStringLit)
	case ExpressionKindBinaryOp:
		out += fmt.Sprintf("AsBinaryOp: %s", expr.AsBinaryOp)
	case ExpressionKindBinding:
		out += fmt.Sprintf("AsBinding: %s", expr.AsBinding)
	}
	out += "}"
	return out
}

type BinaryOpKind int

const (
	BinaryOpKindPlus BinaryOpKind = iota
	BinaryOpKindMinus
	BinaryOpKindTimes
)

func (kind BinaryOpKind) String() string {
	return [...]string{
		"BinaryOpKindPlus",
		"BinaryOpKindMinus",
		"BinaryOpKindTimes",
	}[kind]
}

type BinaryOp struct {
	Kind BinaryOpKind
	Lhs  *Expression
	Rhs  *Expression
}

func (binop BinaryOp) String() (out string) {
	out += "{"
	out += fmt.Sprintf("Kind: %s, ", binop.Kind)
	if binop.Lhs != nil {
		out += fmt.Sprintf("Lhs: %s, ", *binop.Lhs)
	}
	if binop.Rhs != nil {
		out += fmt.Sprintf("Rhs: %s", *binop.Rhs)
	}
	out += "}"
	return out
}

// Map with the precedence of the binary operations.
var binOpPrecedenceMap = map[BinaryOpKind]int{
	BinaryOpKindPlus:  1,
	BinaryOpKindMinus: 1,
	BinaryOpKindTimes: 2,
}

// Returns true if given token is a binary operator
// false otherwise.
func tokenIsOperator(token Token) (out bool) {
	switch token.Kind {
	case TokenKindPlus:
		out = true
	case TokenKindMinus:
		out = true
	case TokenKindAsterisk:
		out = true
	default:
		out = false
	}
	return out
}

// Returns the correct binary operation kind from the given token.
// NOTE: Before calling this method make sure the token is a binary operator
// calling tokenIsOperator.
func tokenAsBinaryOpKind(token Token) (out BinaryOpKind) {
	switch token.Kind {
	case TokenKindPlus:
		out = BinaryOpKindPlus
	case TokenKindMinus:
		out = BinaryOpKindMinus
	case TokenKindAsterisk:
		out = BinaryOpKindTimes
	}
	return out
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
	return parseExprBinaryOp(&tokens, 0)
}

// Parse an expression as a binary operation using the precedence climbing algorithm.
// The implementation is inspired by this:
// - "https://www.engr.mun.ca/~theo/Misc/exp_parsing.htm"
// - "https://en.wikipedia.org/wiki/Operator-precedence_parser"
func parseExprBinaryOp(tokens *[]Token, precedence int) (result Expression, err error) {
	result, err = parseExprPrimary(tokens)
	if err != nil {
		return Expression{}, err
	}

	for len(*tokens) > 1 && tokenIsOperator((*tokens)[0]) &&
		binOpPrecedenceMap[tokenAsBinaryOpKind((*tokens)[0])] >= precedence {
		op := tokenAsBinaryOpKind((*tokens)[0])
		*tokens = (*tokens)[1:]
		rhs, err := parseExprBinaryOp(tokens, binOpPrecedenceMap[op]+1)
		if err != nil {
			return Expression{}, err
		}

		// left and right have the same type and are not bindings
		// so we can already compute the operation ad return the result.
		if result.Kind == rhs.Kind && result.Kind != ExpressionKindBinding {
			// TODO: move operation computation in it's own function
			switch result.Kind {
			case ExpressionKindNumLitInt:
				switch op {
				case BinaryOpKindPlus:
					result.AsNumLitInt = result.AsNumLitInt + rhs.AsNumLitInt
				case BinaryOpKindMinus:
					result.AsNumLitInt = result.AsNumLitInt - rhs.AsNumLitInt
				case BinaryOpKindTimes:
					result.AsNumLitInt = result.AsNumLitInt * rhs.AsNumLitInt
				}
			case ExpressionKindNumLitFloat:
				switch op {
				case BinaryOpKindPlus:
					result.AsNumLitFloat = result.AsNumLitFloat + rhs.AsNumLitFloat
				case BinaryOpKindMinus:
					result.AsNumLitFloat = result.AsNumLitFloat - rhs.AsNumLitFloat
				case BinaryOpKindTimes:
					result.AsNumLitFloat = result.AsNumLitFloat * rhs.AsNumLitFloat
				}
			case ExpressionKindStringLit:
				switch op {
				case BinaryOpKindPlus:
					result.AsStringLit = result.AsStringLit + rhs.AsStringLit
				case BinaryOpKindMinus:
					return Expression{}, fmt.Errorf("unsupported operation '-' between string literals")
				case BinaryOpKindTimes:
					return Expression{}, fmt.Errorf("unsupported operation '*' between string literals")
				}
			case ExpressionKindBinaryOp:
				return Expression{}, fmt.Errorf("WTF, unreachable")
			}
		} else {
			// we can't compute the operation yet, so we return an expression
			// as a binary operation.
			lhs := result
			result.Kind = ExpressionKindBinaryOp
			result.AsBinaryOp = BinaryOp{
				Kind: op,
				Lhs:  &lhs,
				Rhs:  &rhs,
			}
		}
	}
	return result, nil
}

// Parse a primary expression form a list of tokens.
// Returns an error if something went wrong.
func parseExprPrimary(tokens *[]Token) (result Expression, err error) {
	if len(*tokens) == 0 {
		return Expression{}, fmt.Errorf("trying to parse empty expression")
	}
	switch (*tokens)[0].Kind {
	case TokenKindNumLit:
		// Try hexadecimal
		if strings.HasPrefix((*tokens)[0].Text, "0x") {
			number := (*tokens)[0].Text[2:]
			hexNumber, err := strconv.ParseUint(number, 16, 64)
			if err != nil {
				return Expression{},
					fmt.Errorf("error parsing hex number literal '%s'",
						(*tokens)[0].Text)
			}
			result.Kind = ExpressionKindNumLitInt
			result.AsNumLitInt = int64(hexNumber)
		} else {
			// Try integer
			intNumber, err := strconv.ParseInt((*tokens)[0].Text, 10, 64)
			if err != nil {
				// Try floating point
				floatNumber, err := strconv.ParseFloat((*tokens)[0].Text, 64)
				if err != nil {
					return Expression{},
						fmt.Errorf("error parsing number literal '%s'",
							(*tokens)[0].Text)
				}
				result.Kind = ExpressionKindNumLitFloat
				result.AsNumLitFloat = floatNumber
			} else {
				result.Kind = ExpressionKindNumLitInt
				result.AsNumLitInt = intNumber
			}
		}
		*tokens = (*tokens)[1:]
	case TokenKindStringLit:
		result.Kind = ExpressionKindStringLit
		result.AsStringLit = (*tokens)[0].Text
		*tokens = (*tokens)[1:]
	case TokenKindSymbol:
		result.Kind = ExpressionKindBinding
		result.AsBinding = (*tokens)[0].Text
		*tokens = (*tokens)[1:]
	case TokenKindMinus:
		*tokens = (*tokens)[1:]
		result, err = parseExprBinaryOp(tokens, 3)
		if err != nil {
			return Expression{}, err
		}
		if result.Kind == ExpressionKindNumLitInt {
			result.AsNumLitInt = -result.AsNumLitInt
		} else if result.Kind == ExpressionKindNumLitFloat {
			result.AsNumLitFloat = -result.AsNumLitFloat
		}
	case TokenKindOpenParen:
		*tokens = (*tokens)[1:]
		result, err = parseExprBinaryOp(tokens, 0)
		if err != nil {
			return Expression{}, err
		}
		if len(*tokens) == 0 || (*tokens)[0].Kind != TokenKindCloseParen {
			return Expression{}, fmt.Errorf("cannot find matching closing parenthesis")
		}
		*tokens = (*tokens)[1:]
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

	expr, err := parseExprPrimary(&[]Token{tokens[0]})
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
