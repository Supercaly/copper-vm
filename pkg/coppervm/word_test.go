package coppervm

import (
	"testing"
)

func TestWordU64(t *testing.T) {
	w := WordU64(5)
	if w.AsU64 != 5 &&
		w.AsI64 != 5 &&
		w.AsF64 != 5.0 {
		t.Error("WordU64 not created correctly")
	}
}

func TestWordI64(t *testing.T) {
	w := WordI64(5)
	if w.AsU64 != 5 &&
		w.AsI64 != 5 &&
		w.AsF64 != 5.0 {
		t.Error("WordI64 not created correctly")
	}
}

func TestWordF64(t *testing.T) {
	w := WordF64(5.0)
	if w.AsU64 != 5 &&
		w.AsI64 != 5 &&
		w.AsF64 != 5.0 {
		t.Error("WordF64 not created correctly")
	}
}

func TestAddWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		res Word
	}{
		{WordU64(5), WordU64(3), WordU64(8)},
		{WordI64(-5), WordI64(3), WordI64(-2)},
		{WordF64(5.3), WordF64(3.2), WordF64(8.5)},
	}

	for i, test := range tests {
		result := AddWord(test.a, test.b)
		if !wordsEquals(result, test.res) {
			t.Errorf("expecting %d %#v but got %#v", i, test.res, result)
		}
	}
}

func TestSubWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		res Word
	}{
		{WordU64(5), WordU64(3), WordU64(2)},
		{WordI64(-5), WordI64(3), WordI64(-8)},
		{WordF64(5.3), WordF64(3.2), WordF64(2.10)},
	}

	for i, test := range tests {
		result := SubWord(test.a, test.b)
		if !wordsEquals(result, test.res) {
			t.Errorf("expecting %d %#v but got %#v", i, test.res, result)
		}
	}
}

func TestMulWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		res Word
	}{
		{WordU64(5), WordU64(3), WordU64(15)},
		{WordI64(-5), WordI64(3), WordI64(-15)},
		{WordF64(5.3), WordF64(3.2), Word{15, 15, 16.96}},
	}

	for i, test := range tests {
		result := MulWord(test.a, test.b)
		if !wordsEquals(result, test.res) {
			t.Errorf("expecting %d %#v but got %#v", i, test.res, result)
		}
	}
}

func TestDivWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		res Word
	}{
		{WordU64(16), WordU64(2), WordU64(8)},
		{WordI64(-6), WordI64(3), Word{0x5555555555555553, -2, -2}},
		{WordF64(8.2), WordF64(3.1), Word{2, 2, 2.65}},
	}

	for i, test := range tests {
		result := DivWord(test.a, test.b)
		if !wordsEquals(result, test.res) {
			t.Errorf("expecting %d %#v but got %#v", i, test.res, result)
		}
	}
}
func TestModWord(t *testing.T) {
	tests := []struct {
		a   Word
		b   Word
		res Word
	}{
		{WordU64(15), WordU64(2), WordU64(1)},
		{WordI64(-6), WordI64(3), Word{1, 0, 0}},
		{WordF64(8.2), WordF64(3.1), Word{2, 2, 0}},
	}

	for i, test := range tests {
		result := ModWord(test.a, test.b)
		if !wordsEquals(result, test.res) {
			t.Errorf("expecting %d %#v but got %#v", i, test.res, result)
		}
	}
}

func wordsEquals(a Word, b Word) bool {
	return a.AsU64 == b.AsU64 &&
		a.AsI64 == b.AsI64 &&
		a.AsF64-b.AsF64 < 0.01
}
