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
	ExpressionKindByteList
)

func (kind ExpressionKind) String() string {
	return [...]string{
		"ExpressionKindNumLitInt",
		"ExpressionKindNumLitFloat",
		"ExpressionKindStringLit",
		"ExpressionKindBinaryOp",
		"ExpressionKindBinding",
		"ExpressionKindByteList",
	}[kind]
}

type Expression struct {
	Kind          ExpressionKind
	AsNumLitInt   int64
	AsNumLitFloat float64
	AsStringLit   string
	AsBinaryOp    BinaryOp
	AsBinding     string
	AsByteList    []byte
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
	case ExpressionKindByteList:
		out += fmt.Sprintf("AsByteList: %s", expr.AsByteList)
	}
	out += "}"
	return out
}

type BinaryOpKind int

const (
	BinaryOpKindPlus BinaryOpKind = iota
	BinaryOpKindMinus
	BinaryOpKindTimes
	BinaryOpKindDivide
	BinaryOpKindModulo
)

func (kind BinaryOpKind) String() string {
	return [...]string{
		"BinaryOpKindPlus",
		"BinaryOpKindMinus",
		"BinaryOpKindTimes",
		"BinaryOpKindDivide",
		"BinaryOpKindModulo",
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
	case TokenKindPlus,
		TokenKindMinus,
		TokenKindAsterisk,
		TokenKindSlash,
		TokenKindPercent:
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
	case TokenKindSlash:
		out = BinaryOpKindDivide
	case TokenKindPercent:
		out = BinaryOpKindModulo
	default:
		panic(fmt.Sprintf("token %#v is not a binary operatator", token))
	}
	return out
}

// Computes an operation between two expression that have the same type and return it.
// Note: Before calling this method make sure that lhs and rhs have the same type.
// Some operation are not supported, so this method can call panic.
func computeOpWithSameType(lhs Expression, rhs Expression, op BinaryOpKind) (out Expression) {
	if lhs.Kind != rhs.Kind {
		panic("lhs and rhs must have the same type")
	}

	out.Kind = lhs.Kind
	switch lhs.Kind {
	case ExpressionKindNumLitInt:
		switch op {
		case BinaryOpKindPlus:
			out.AsNumLitInt = lhs.AsNumLitInt + rhs.AsNumLitInt
		case BinaryOpKindMinus:
			out.AsNumLitInt = lhs.AsNumLitInt - rhs.AsNumLitInt
		case BinaryOpKindTimes:
			out.AsNumLitInt = lhs.AsNumLitInt * rhs.AsNumLitInt
		case BinaryOpKindDivide:
			if rhs.AsNumLitInt == 0 {
				panic("divide by zero")
			}
			out.AsNumLitInt = lhs.AsNumLitInt / rhs.AsNumLitInt
		case BinaryOpKindModulo:
			out.AsNumLitInt = lhs.AsNumLitInt % rhs.AsNumLitInt
		}
	case ExpressionKindNumLitFloat:
		switch op {
		case BinaryOpKindPlus:
			out.AsNumLitFloat = lhs.AsNumLitFloat + rhs.AsNumLitFloat
		case BinaryOpKindMinus:
			out.AsNumLitFloat = lhs.AsNumLitFloat - rhs.AsNumLitFloat
		case BinaryOpKindTimes:
			out.AsNumLitFloat = lhs.AsNumLitFloat * rhs.AsNumLitFloat
		case BinaryOpKindDivide:
			if rhs.AsNumLitFloat == 0 {
				panic("divide by zero")
			}
			out.AsNumLitFloat = lhs.AsNumLitFloat / rhs.AsNumLitFloat
		case BinaryOpKindModulo:
			panic("unsupported operation '%' between floating point literals")
		}
	case ExpressionKindStringLit:
		switch op {
		case BinaryOpKindPlus:
			out.AsStringLit = lhs.AsStringLit + rhs.AsStringLit
		case BinaryOpKindMinus:
			panic("unsupported operation '-' between string literals")
		case BinaryOpKindTimes:
			panic("unsupported operation '*' between string literals")
		case BinaryOpKindDivide:
			panic("unsupported operation '/' between string literals")
		case BinaryOpKindModulo:
			panic("unsupported operation '%' between string literals")
		}
	case ExpressionKindBinaryOp:
		panic("at this point binary op is unreachable! Something really went wrong WTF")
	case ExpressionKindBinding:
		panic("at this point binding is unreachable! Something really went wrong WTF")
	case ExpressionKindByteList:
		panic("at this point byte list is unreachable! Something really went wrong WTF")
	}

	return out
}

// Parse an expression from a list of tokens.
// This method will panic if something went wrong.
func parseExprFromTokens(tokens *Tokens) Expression {
	return parseExprBinaryOp(tokens, 0)
}

// Parse an expression as a binary operation using the precedence climbing algorithm.
// The implementation is inspired by this:
// - "https://www.engr.mun.ca/~theo/Misc/exp_parsing.htm"
// - "https://en.wikipedia.org/wiki/Operator-precedence_parser"
func parseExprBinaryOp(tokens *Tokens, precedence int) (result Expression) {
	result = parseExprPrimary(tokens)

	for len(*tokens) > 1 && tokenIsOperator((*tokens)[0]) &&
		binOpPrecedenceMap[tokenAsBinaryOpKind(tokens.First())] >= precedence {
		op := tokenAsBinaryOpKind(tokens.First())
		tokens.Pop()
		rhs := parseExprBinaryOp(tokens, binOpPrecedenceMap[op]+1)

		// left and right have the same type and are not bindings
		// so we can already compute the operation ad return the result.
		if result.Kind == rhs.Kind &&
			result.Kind != ExpressionKindBinding &&
			result.Kind != ExpressionKindByteList {
			result = computeOpWithSameType(result, rhs, op)
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
	return result
}

// Parse a primary expression form a list of tokens.
// Returns an error if something went wrong.
func parseExprPrimary(tokens *Tokens) (result Expression) {
	if tokens.Empty() {
		panic("trying to parse empty expression")
	}

	switch tokens.First().Kind {
	case TokenKindNumLit:
		numberStr := tokens.First().Text
		if strings.HasPrefix(numberStr, "0x") || strings.HasPrefix(numberStr, "0X") {
			// Try hexadecimal
			hexNumber, err := strconv.ParseUint(numberStr[2:], 16, 64)
			if err != nil {
				panic(fmt.Sprintf("error parsing hex number literal '%s'",
					numberStr))
			}
			result.Kind = ExpressionKindNumLitInt
			result.AsNumLitInt = int64(hexNumber)
		} else if strings.HasPrefix(numberStr, "0b") || strings.HasPrefix(numberStr, "0B") {
			// Try binary
			binNumber, err := strconv.ParseUint(numberStr[2:], 2, 64)
			if err != nil {
				panic(fmt.Sprintf("error parsing binary number literal '%s'",
					numberStr))
			}
			result.Kind = ExpressionKindNumLitInt
			result.AsNumLitInt = int64(binNumber)
		} else {
			// Try integer
			intNumber, err := strconv.ParseInt(numberStr, 10, 64)
			if err != nil {
				// Try floating point
				floatNumber, err := strconv.ParseFloat(numberStr, 64)
				if err != nil {
					panic(fmt.Sprintf("error parsing number literal '%s'",
						numberStr))
				}
				result.Kind = ExpressionKindNumLitFloat
				result.AsNumLitFloat = floatNumber
			} else {
				result.Kind = ExpressionKindNumLitInt
				result.AsNumLitInt = intNumber
			}
		}
		tokens.Pop()
	case TokenKindStringLit:
		result.Kind = ExpressionKindStringLit
		result.AsStringLit = tokens.First().Text
		tokens.Pop()
	case TokenKindCharLit:
		charStr := tokens.First().Text
		char, _, t, err := strconv.UnquoteChar(charStr+`\'`, '\'')
		if err != nil {
			panic(fmt.Sprintf("error parsing character literal '%s'", charStr))
		}
		if t != "\\'" {
			panic("unsupported multi-character character literals")
		}
		result.Kind = ExpressionKindNumLitInt
		result.AsNumLitInt = int64(char)
		tokens.Pop()
	case TokenKindSymbol:
		result.Kind = ExpressionKindBinding
		result.AsBinding = tokens.First().Text
		tokens.Pop()
	case TokenKindMinus:
		tokens.Pop()
		result = parseExprBinaryOp(tokens, 3)
		if result.Kind == ExpressionKindNumLitInt {
			result.AsNumLitInt = -result.AsNumLitInt
		} else if result.Kind == ExpressionKindNumLitFloat {
			result.AsNumLitFloat = -result.AsNumLitFloat
		}
	case TokenKindOpenParen:
		tokens.Pop()
		result = parseExprBinaryOp(tokens, 0)
		if tokens.Empty() || tokens.First().Kind != TokenKindCloseParen {
			panic("cannot find matching closing parenthesis ')'")
		}
		tokens.Pop()
	case TokenKindOpenBracket:
		tokens.Pop()

		var byteResult []byte
		for !tokens.Empty() && tokens.First().Kind != TokenKindCloseBracket {
			expr := parseExprBinaryOp(tokens, 0)
			if expr.Kind == ExpressionKindNumLitInt {
				byteResult = append(byteResult, byte(expr.AsNumLitInt))
			} else if expr.Kind == ExpressionKindStringLit {
				byteResult = append(byteResult, []byte(expr.AsStringLit)...)
			} else {
				panic(fmt.Sprintf("unsupported value of type '%s' inside byte array", expr.Kind))
			}

			if tokens.Empty() {
				panic("expected ',' or ']'")
			}
			if tokens.First().Kind != TokenKindComma {
				break
			}
			tokens.Pop()
		}

		if tokens.First().Kind != TokenKindCloseBracket {
			panic("cannot find matching closing bracket ']'")
		}
		tokens.Pop()

		result.Kind = ExpressionKindByteList
		result.AsByteList = append(result.AsByteList, byteResult...)
	default:
		panic(fmt.Sprintf("unknown expression starting with token %s", tokens.First().Text))
	}
	return result
}
