package internal

import (
	"testing"
	"unicode"

	"github.com/Supercaly/coppervm/internal"
)

func TestSplitByDelim(t *testing.T) {
	tests := []struct {
		in        string
		out, rest string
	}{
		{"test.123", "test", ".123"},
		{"123.test.123", "123", ".test.123"},
		{"test", "test", ""},
		{"", "", ""},
	}

	for _, test := range tests {
		f, s := internal.SplitByDelim(test.in, '.')
		if f != test.out {
			t.Errorf("Expected '%s' but got '%s'", test.out, f)
		}
		if s != test.rest {
			t.Errorf("Expected '%s' but got '%s'", test.rest, s)
		}
	}
}

func TestSplitWhile(t *testing.T) {
	tests := []struct {
		in        string
		out, rest string
	}{
		{"test123", "test", "123"},
		{"123test123", "", "123test123"},
		{"test", "test", ""},
		{"123", "", "123"},
		{"", "", ""},
	}

	for _, test := range tests {
		f, s := internal.SplitWhile(test.in, func(r rune) bool {
			return unicode.IsLetter(r)
		})
		if f != test.out {
			t.Errorf("Expected '%s' but got '%s'", test.out, f)
		}
		if s != test.rest {
			t.Errorf("Expected '%s' but got '%s'", test.rest, s)
		}
	}
}
