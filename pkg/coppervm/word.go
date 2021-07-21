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
