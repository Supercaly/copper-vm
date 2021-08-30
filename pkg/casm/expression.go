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
	}

	return out
}

// Parse an expression from a source string.
// The string is first tokenized and then is parsed to extract
// an expression.
// Returns an error if something went wrong.
func ParseExprFromString(source string) (out Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	tokens, err := Tokenize(source)
	if err != nil {
		panic(err)
	}
	out = parseExprBinaryOp(&tokens, 0)
	return out, err
}

// Parse an expression as a binary operation using the precedence climbing algorithm.
// The implementation is inspired by this:
// - "https://www.engr.mun.ca/~theo/Misc/exp_parsing.htm"
// - "https://en.wikipedia.org/wiki/Operator-precedence_parser"
func parseExprBinaryOp(tokens *[]Token, precedence int) (result Expression) {
	result = parseExprPrimary(tokens)

	for len(*tokens) > 1 && tokenIsOperator((*tokens)[0]) &&
		binOpPrecedenceMap[tokenAsBinaryOpKind((*tokens)[0])] >= precedence {
		op := tokenAsBinaryOpKind((*tokens)[0])
		*tokens = (*tokens)[1:]
		rhs := parseExprBinaryOp(tokens, binOpPrecedenceMap[op]+1)

		// left and right have the same type and are not bindings
		// so we can already compute the operation ad return the result.
		if result.Kind == rhs.Kind && result.Kind != ExpressionKindBinding {
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
func parseExprPrimary(tokens *[]Token) (result Expression) {
	if len(*tokens) == 0 {
		panic("trying to parse empty expression")
	}

	switch (*tokens)[0].Kind {
	case TokenKindNumLit:
		numberStr := (*tokens)[0].Text
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
		*tokens = (*tokens)[1:]
	case TokenKindStringLit:
		result.Kind = ExpressionKindStringLit
		result.AsStringLit = (*tokens)[0].Text
		*tokens = (*tokens)[1:]
	case TokenKindCharLit:
		charStr := (*tokens)[0].Text
		char, _, t, err := strconv.UnquoteChar(charStr+`\'`, '\'')
		if err != nil {
			panic(fmt.Sprintf("error parsing character literal '%s'", charStr))
		}
		if t != "\\'" {
			panic("unsupported multi-character character literals")
		}
		result.Kind = ExpressionKindNumLitInt
		result.AsNumLitInt = int64(char)
	case TokenKindSymbol:
		result.Kind = ExpressionKindBinding
		result.AsBinding = (*tokens)[0].Text
		*tokens = (*tokens)[1:]
	case TokenKindMinus:
		*tokens = (*tokens)[1:]
		result = parseExprBinaryOp(tokens, 3)
		if result.Kind == ExpressionKindNumLitInt {
			result.AsNumLitInt = -result.AsNumLitInt
		} else if result.Kind == ExpressionKindNumLitFloat {
			result.AsNumLitFloat = -result.AsNumLitFloat
		}
	case TokenKindOpenParen:
		*tokens = (*tokens)[1:]
		result = parseExprBinaryOp(tokens, 0)
		if len(*tokens) == 0 || (*tokens)[0].Kind != TokenKindCloseParen {
			panic("cannot find matching closing parenthesis")
		}
		*tokens = (*tokens)[1:]
	}
	return result
}

// Parse a byte list from a source string.
// The string is first tokenized and then is parsed to extract
// the data.
// Returns an error if something went wrong.
func ParseByteArrayFromString(source string) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	tokens, err := Tokenize(source)
	if err != nil {
		panic(err)
	}
	out = parseByteArrayFromTokens(&tokens)
	return out, err
}

// Parse a byte list from some tokens.
// Returns a byte array or an error.
func parseByteArrayFromTokens(tokens *[]Token) (out []byte) {
	if len(*tokens) == 0 {
		return []byte{}
	}

	if (*tokens)[0].Kind == TokenKindComma {
		panic("misplaced comma inside list")
	}

	expr := parseExprBinaryOp(tokens, 0)

	if expr.Kind == ExpressionKindNumLitInt {
		out = append(out, byte(expr.AsNumLitInt))
	} else if expr.Kind == ExpressionKindStringLit {
		out = append(out, []byte(expr.AsStringLit)...)
	} else {
		panic(fmt.Sprintf("unsupported value of type '%s' inside byte array", expr.Kind))
	}

	if len(*tokens) != 0 {
		if (*tokens)[0].Kind != TokenKindComma {
			panic("array values must be comma separated")
		}
		*tokens = (*tokens)[1:]
		next := parseByteArrayFromTokens(tokens)
		out = append(out, next...)
	}

	return out
}
