package coppervm

import "fmt"

type Word struct {
	AsI64 int64
	AsF64 float64
}

func WordI64(i64 int64) Word {
	return Word{
		AsI64: i64,
	}
}

func WordF64(f64 float64) Word {
	return Word{
		AsF64: f64,
	}
}

func (word Word) String() string {
	return fmt.Sprintf("i64: %d, f64: %f",
		word.AsI64,
		word.AsF64)
}

type typeRep int

const (
	typeRepI64 typeRep = iota
	typeRepF64
)

func addWord(a Word, b Word, rep typeRep) (out Word) {
	switch rep {
	case typeRepI64:
		out = WordI64(a.AsI64 + b.AsI64)
	case typeRepF64:
		out = WordF64(a.AsF64 + b.AsF64)
	}
	return out
}

func subWord(a Word, b Word, rep typeRep) (out Word) {
	switch rep {
	case typeRepI64:
		out = WordI64(a.AsI64 - b.AsI64)
	case typeRepF64:
		out = WordF64(a.AsF64 - b.AsF64)
	}
	return out
}

func mulWord(a Word, b Word, rep typeRep) (out Word) {
	switch rep {
	case typeRepI64:
		out = WordI64(a.AsI64 * b.AsI64)
	case typeRepF64:
		out = WordF64(a.AsF64 * b.AsF64)
	}
	return out
}
