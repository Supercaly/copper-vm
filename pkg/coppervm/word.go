package coppervm

import "fmt"

type TypeRepresentation int

const (
	TypeU64 TypeRepresentation = iota
	TypeI64
	TypeF64
)

type Word struct {
	AsU64 uint64
	AsI64 int64
	AsF64 float64
}

// Create a Word from a u64 value.
func WordU64(u64 uint64) Word {
	return Word{
		AsU64: u64,
		AsI64: int64(u64),
		AsF64: float64(u64),
	}
}

// Create a Word from a i64 value.
func WordI64(i64 int64) Word {
	return Word{
		AsU64: uint64(i64),
		AsI64: i64,
		AsF64: float64(i64),
	}
}

// Create a Word from a f64 value.
func WordF64(f64 float64) Word {
	return Word{
		AsU64: uint64(f64),
		AsI64: int64(f64),
		AsF64: f64,
	}
}

func (word Word) String() string {
	return fmt.Sprintf("u64: %d, i64: %d, f64: %f",
		word.AsU64,
		word.AsI64,
		word.AsF64)
}

// Returns the sum of two Words.
func AddWord(a Word, b Word, t TypeRepresentation) (out Word) {
	switch t {
	case TypeU64:
		out = WordU64(a.AsU64 + b.AsU64)
	case TypeI64:
		out = WordI64(a.AsI64 + b.AsI64)
	case TypeF64:
		out = WordF64(a.AsF64 + b.AsF64)
	}
	return out
}

// Returns the difference of two Words.
func SubWord(a Word, b Word, t TypeRepresentation) (out Word) {
	switch t {
	case TypeU64:
		out = WordU64(a.AsU64 - b.AsU64)
	case TypeI64:
		out = WordI64(a.AsI64 - b.AsI64)
	case TypeF64:
		out = WordF64(a.AsF64 - b.AsF64)
	}
	return out
}

// Returns the product of two Words.
func MulWord(a Word, b Word, t TypeRepresentation) (out Word) {
	switch t {
	case TypeU64:
		out = WordU64(a.AsU64 * b.AsU64)
	case TypeI64:
		out = WordI64(a.AsI64 * b.AsI64)
	case TypeF64:
		out = WordF64(a.AsF64 * b.AsF64)
	}
	return out
}

// Returns the division of two Words.
func DivWord(a Word, b Word, t TypeRepresentation) (out Word) {
	switch t {
	case TypeU64:
		out = WordU64(a.AsU64 / b.AsU64)
	case TypeI64:
		out = WordI64(a.AsI64 / b.AsI64)
	case TypeF64:
		out = WordF64(a.AsF64 / b.AsF64)
	}
	return out
}

// Returns the modulo of two Words.
// Note: % operator don't support floats, so
// calling this will set the float value to 0.
func ModWord(a Word, b Word, t TypeRepresentation) (out Word) {
	switch t {
	case TypeU64:
		out = WordU64(a.AsU64 % b.AsU64)
	case TypeI64:
		out = WordI64(a.AsI64 % b.AsI64)
	case TypeF64:
		panic("unsupported modulo for type f64")
	}
	return out
}
