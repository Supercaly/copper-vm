package casm

import (
	"fmt"

	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type wordKind int

const (
	wordKindInt wordKind = iota
	wordKindFloat
	wordKindInstAddr
	wordKindMemoryAddr
)

type word struct {
	kind         wordKind
	asInt        int64
	asFloat      float64
	asInstAddr   int64
	asMemoryAddr int64
}

func (w word) toCoppervmWord() (ret coppervm.Word) {
	switch w.kind {
	case wordKindInt:
		ret = coppervm.WordI64(w.asInt)
	case wordKindFloat:
		ret = coppervm.WordF64(w.asFloat)
	case wordKindInstAddr:
		ret = coppervm.WordU64(uint64(w.asInstAddr))
	case wordKindMemoryAddr:
		ret = coppervm.WordU64(uint64(w.asMemoryAddr))
	}
	return ret
}

// Create a word from a int value.
func wordInt(i int64) word {
	return word{
		kind:  wordKindInt,
		asInt: i,
	}
}

// Create a word from a float value.
func wordFloat(f float64) word {
	return word{
		kind:    wordKindFloat,
		asFloat: f,
	}
}

// Create a word from a instruction address.
func wordInstAddr(i int64) word {
	return word{
		kind:       wordKindInstAddr,
		asInstAddr: i,
	}
}

// Create a word from a memory address.
func wordMemoryAddr(i int64) word {
	return word{
		kind:         wordKindMemoryAddr,
		asMemoryAddr: i,
	}
}

func (word word) String() (out string) {
	out += fmt.Sprintf("(kind: %d, ", word.kind)
	switch word.kind {
	case wordKindInt:
		out += fmt.Sprint(word.asInt)
	case wordKindFloat:
		out += fmt.Sprint(word.asFloat)
	case wordKindInstAddr:
		out += fmt.Sprint(word.asInstAddr)
	case wordKindMemoryAddr:
		out += fmt.Sprint(word.asMemoryAddr)
	}
	out += ")"
	return out
}

// Returns the sum of two words.
func addWord(a word, b word) (out word) {
	switch a.kind {
	case wordKindInt:
		switch b.kind {
		case wordKindInt:
			out = wordInt(a.asInt + b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asInt) + b.asFloat)
		case wordKindInstAddr:
			out = wordInstAddr(a.asInt + b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordMemoryAddr(a.asInt + b.asMemoryAddr)
		}
	case wordKindFloat:
		switch b.kind {
		case wordKindInt:
			out = wordFloat((a.asFloat) + float64(b.asInt))
		case wordKindFloat:
			out = wordFloat(a.asFloat + b.asFloat)
		case wordKindInstAddr:
			out = wordFloat(a.asFloat + float64(b.asInstAddr))
		case wordKindMemoryAddr:
			out = wordFloat(a.asFloat + float64(b.asMemoryAddr))
		}
	case wordKindInstAddr:
		switch b.kind {
		case wordKindInt:
			out = wordInstAddr(a.asInstAddr + b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asInstAddr) + b.asFloat)
		case wordKindInstAddr:
			out = wordInstAddr(a.asInstAddr + b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordInt(a.asInstAddr + b.asMemoryAddr)
		}
	case wordKindMemoryAddr:
		switch b.kind {
		case wordKindInt:
			out = wordMemoryAddr(a.asMemoryAddr + b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asMemoryAddr) + b.asFloat)
		case wordKindInstAddr:
			out = wordInt(a.asMemoryAddr + b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordMemoryAddr(a.asMemoryAddr + b.asMemoryAddr)
		}
	}
	return out
}

// Returns the difference of two words.
func subWord(a word, b word) (out word) {
	switch a.kind {
	case wordKindInt:
		switch b.kind {
		case wordKindInt:
			out = wordInt(a.asInt - b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asInt) - b.asFloat)
		case wordKindInstAddr:
			out = wordInstAddr(a.asInt - b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordMemoryAddr(a.asInt - b.asMemoryAddr)
		}
	case wordKindFloat:
		switch b.kind {
		case wordKindInt:
			out = wordFloat(a.asFloat - float64(b.asInt))
		case wordKindFloat:
			out = wordFloat(a.asFloat - b.asFloat)
		case wordKindInstAddr:
			out = wordFloat(a.asFloat - float64(b.asInstAddr))
		case wordKindMemoryAddr:
			out = wordFloat(a.asFloat - float64(b.asMemoryAddr))
		}
	case wordKindInstAddr:
		switch b.kind {
		case wordKindInt:
			out = wordInstAddr(a.asInstAddr - b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asInstAddr) - b.asFloat)
		case wordKindInstAddr:
			out = wordInstAddr(a.asInstAddr - b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordInt(a.asInstAddr - b.asMemoryAddr)
		}
	case wordKindMemoryAddr:
		switch b.kind {
		case wordKindInt:
			out = wordMemoryAddr(a.asMemoryAddr - b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asMemoryAddr) - b.asFloat)
		case wordKindInstAddr:
			out = wordInt(a.asMemoryAddr - b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordMemoryAddr(a.asMemoryAddr - b.asMemoryAddr)
		}
	}
	return out
}

// Returns the product of two words.
func mulWord(a word, b word) (out word) {
	switch a.kind {
	case wordKindInt:
		switch b.kind {
		case wordKindInt:
			out = wordInt(a.asInt * b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asInt) * b.asFloat)
		case wordKindInstAddr:
			out = wordInt(a.asInt * b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordInt(a.asInt * b.asMemoryAddr)
		}
	case wordKindFloat:
		switch b.kind {
		case wordKindInt:
			out = wordFloat(a.asFloat * float64(b.asInt))
		case wordKindFloat:
			out = wordFloat(a.asFloat * b.asFloat)
		case wordKindInstAddr:
			out = wordFloat(a.asFloat * float64(b.asInstAddr))
		case wordKindMemoryAddr:
			out = wordFloat(a.asFloat * float64(b.asMemoryAddr))
		}
	case wordKindInstAddr:
		switch b.kind {
		case wordKindInt:
			out = wordInt(a.asInstAddr * b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asInstAddr) * b.asFloat)
		case wordKindInstAddr:
			out = wordInt(a.asInstAddr * b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordInt(a.asInstAddr * b.asMemoryAddr)
		}
	case wordKindMemoryAddr:
		switch b.kind {
		case wordKindInt:
			out = wordInt(a.asMemoryAddr * b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asMemoryAddr) * b.asFloat)
		case wordKindInstAddr:
			out = wordInt(a.asMemoryAddr * b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordInt(a.asMemoryAddr * b.asMemoryAddr)
		}
	}
	return out
}

// Returns the division of two words.
func divWord(a word, b word) (out word) {
	if (b.kind == wordKindInt && b.asInt == 0) ||
		(b.kind == wordKindFloat && b.asFloat == 0) ||
		(b.kind == wordKindInstAddr && b.asInstAddr == 0) ||
		(b.kind == wordKindMemoryAddr && b.asMemoryAddr == 0) {
		panic(fmt.Sprintf("divide by zero %s", b))
	}

	switch a.kind {
	case wordKindInt:
		switch b.kind {
		case wordKindInt:
			out = wordInt(a.asInt / b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asInt) / b.asFloat)
		case wordKindInstAddr:
			out = wordInt(a.asInt / b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordInt(a.asInt / b.asMemoryAddr)
		}
	case wordKindFloat:
		switch b.kind {
		case wordKindInt:
			out = wordFloat(a.asFloat / float64(b.asInt))
		case wordKindFloat:
			out = wordFloat(a.asFloat / b.asFloat)
		case wordKindInstAddr:
			out = wordFloat(a.asFloat / float64(b.asInstAddr))
		case wordKindMemoryAddr:
			out = wordFloat(a.asFloat / float64(b.asMemoryAddr))
		}
	case wordKindInstAddr:
		switch b.kind {
		case wordKindInt:
			out = wordInt(a.asInstAddr / b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asInstAddr) / b.asFloat)
		case wordKindInstAddr:
			out = wordInt(a.asInstAddr / b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordInt(a.asInstAddr / b.asMemoryAddr)
		}
	case wordKindMemoryAddr:
		switch b.kind {
		case wordKindInt:
			out = wordInt(a.asMemoryAddr / b.asInt)
		case wordKindFloat:
			out = wordFloat(float64(a.asMemoryAddr) / b.asFloat)
		case wordKindInstAddr:
			out = wordInt(a.asMemoryAddr / b.asInstAddr)
		case wordKindMemoryAddr:
			out = wordInt(a.asMemoryAddr / b.asMemoryAddr)
		}
	}
	return out
}
