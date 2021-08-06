package coppervm

import "fmt"

type Word struct {
	AsU64 uint64
	AsI64 int64
	AsF64 float64
}

func WordU64(u64 uint64) Word {
	return Word{
		AsU64: u64,
		AsI64: int64(u64),
		AsF64: float64(u64),
	}
}

func WordI64(i64 int64) Word {
	return Word{
		AsU64: uint64(i64),
		AsI64: i64,
		AsF64: float64(i64),
	}
}

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

type WordTypeRep int

const (
	WordTypeRepU64 WordTypeRep = iota
	WordTypeRepI64
	WordTypeRepF64
)

func AddWord(a Word, b Word, rep WordTypeRep) (out Word) {
	switch rep {
	case WordTypeRepU64:
		out = WordU64(a.AsU64 + b.AsU64)
	case WordTypeRepI64:
		out = WordI64(a.AsI64 + b.AsI64)
	case WordTypeRepF64:
		out = WordF64(a.AsF64 + b.AsF64)
	}
	return out
}

func SubWord(a Word, b Word, rep WordTypeRep) (out Word) {
	switch rep {
	case WordTypeRepU64:
		out = WordU64(a.AsU64 - b.AsU64)
	case WordTypeRepI64:
		out = WordI64(a.AsI64 - b.AsI64)
	case WordTypeRepF64:
		out = WordF64(a.AsF64 - b.AsF64)
	}
	return out
}

func MulWord(a Word, b Word, rep WordTypeRep) (out Word) {
	switch rep {
	case WordTypeRepU64:
		out = WordU64(a.AsU64 * b.AsU64)
	case WordTypeRepI64:
		out = WordI64(a.AsI64 * b.AsI64)
	case WordTypeRepF64:
		out = WordF64(a.AsF64 * b.AsF64)
	}
	return out
}

func DivWord(a Word, b Word, rep WordTypeRep) (out Word) {
	switch rep {
	case WordTypeRepU64:
		out = WordU64(a.AsU64 / b.AsU64)
	case WordTypeRepI64:
		out = WordI64(a.AsI64 / b.AsI64)
	case WordTypeRepF64:
		out = WordF64(a.AsF64 / b.AsF64)
	}
	return out
}

func ModWord(a Word, b Word, rep WordTypeRep) (out Word) {
	switch rep {
	case WordTypeRepU64:
		out = WordU64(a.AsU64 % b.AsU64)
	case WordTypeRepI64:
		out = WordI64(a.AsI64 % b.AsI64)
	}
	return out
}
