package coppervm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordU64(t *testing.T) {
	w := WordU64(5)
	assert.Equal(t, uint64(5), w.AsU64)
	assert.Equal(t, int64(5), w.AsI64)
	assert.Equal(t, float64(5.0), w.AsF64)
}

func TestWordI64(t *testing.T) {
	w := WordI64(5)
	assert.Equal(t, uint64(5), w.AsU64)
	assert.Equal(t, int64(5), w.AsI64)
	assert.Equal(t, float64(5.0), w.AsF64)
}

func TestWordF64(t *testing.T) {
	w := WordF64(5.0)
	assert.Equal(t, uint64(5), w.AsU64)
	assert.Equal(t, int64(5), w.AsI64)
	assert.Equal(t, float64(5.0), w.AsF64)
}

func TestAddWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		t   TypeRepresentation
		res Word
	}{
		{WordU64(5), WordU64(3), TypeU64, WordU64(8)},
		{WordI64(-5), WordI64(3), TypeI64, WordI64(-2)},
		{WordF64(5.3), WordF64(3.2), TypeF64, WordF64(8.5)},
	}

	for _, test := range tests {
		result := AddWord(test.a, test.b, test.t)
		assert.Equal(t, test.res.AsU64, result.AsU64)
		assert.Equal(t, test.res.AsI64, result.AsI64)
		assert.Condition(t, func() (success bool) {
			return test.res.AsF64-result.AsF64 < 0.01
		})
	}
}

func TestSubWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		t   TypeRepresentation
		res Word
	}{
		{WordU64(5), WordU64(3), TypeU64, WordU64(2)},
		{WordI64(-5), WordI64(3), TypeI64, WordI64(-8)},
		{WordF64(5.3), WordF64(3.2), TypeF64, WordF64(2.10)},
	}

	for _, test := range tests {
		result := SubWord(test.a, test.b, test.t)
		assert.Equal(t, test.res.AsU64, result.AsU64)
		assert.Equal(t, test.res.AsI64, result.AsI64)
		assert.Condition(t, func() (success bool) {
			return test.res.AsF64-result.AsF64 < 0.01
		})
	}
}

func TestMulWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		t   TypeRepresentation
		res Word
	}{
		{WordU64(5), WordU64(3), TypeU64, WordU64(15)},
		{WordI64(-5), WordI64(3), TypeI64, WordI64(-15)},
		{WordF64(5.3), WordF64(3.2), TypeF64, WordF64(16.96)},
	}

	for _, test := range tests {
		result := MulWord(test.a, test.b, test.t)
		assert.Equal(t, test.res.AsU64, result.AsU64)
		assert.Equal(t, test.res.AsI64, result.AsI64)
		assert.Condition(t, func() (success bool) {
			return test.res.AsF64-result.AsF64 < 0.01
		})
	}
}

func TestDivWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		t   TypeRepresentation
		res Word
	}{
		{WordU64(16), WordU64(2), TypeU64, WordU64(8)},
		{WordI64(-6), WordI64(3), TypeI64, WordI64(-2)},
		{WordF64(8.2), WordF64(3.1), TypeF64, WordF64(2.65)},
	}

	for _, test := range tests {
		result := DivWord(test.a, test.b, test.t)
		assert.Equal(t, test.res.AsU64, result.AsU64)
		assert.Equal(t, test.res.AsI64, result.AsI64)
		assert.Condition(t, func() (success bool) {
			return test.res.AsF64-result.AsF64 < 0.01
		})
	}
}
func TestModWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		t   TypeRepresentation
		res Word
	}{
		{WordU64(15), WordU64(2), TypeU64, WordU64(1)},
		{WordI64(-6), WordI64(3), TypeI64, WordI64(0)},
	}

	for _, test := range tests {
		result := ModWord(test.a, test.b, test.t)
		assert.Equal(t, test.res.AsU64, result.AsU64)
		assert.Equal(t, test.res.AsI64, result.AsI64)
		assert.Condition(t, func() (success bool) {
			return test.res.AsF64-result.AsF64 < 0.01
		})
	}

	func() {
		defer func() { recover() }()
		ModWord(WordF64(8.2), WordF64(3.1), TypeF64)
		t.Error("expecting an error")
	}()
}
