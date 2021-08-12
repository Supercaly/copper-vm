package internal

import "strings"

// Splits a line by delimiter.
// Given a line and a rune delimiter it returns two strings:
// the first is the part before the delimiter, the second is
// the part after the delimiter.
//
// Note the delimiter is kept in as the first character of the
// second string.
func SplitByDelim(input string, delim rune) (f string, s string) {
	idx := strings.IndexRune(input, delim)

	if idx == -1 {
		return input, ""
	}

	f = input[:idx]
	s = input[idx:]
	return f, s
}

// Splits a line by all characters matching predicate.
// Given a line and a predicate function it returns two strings:
// the first is the the part mathing the predicate, the second if
// the rest.
func SplitWhile(input string, predicate func(rune) bool) (f string, s string) {
	idx := 0
	for len(input) > idx && predicate(rune(input[idx])) {
		idx++
	}
	f = input[:idx]
	s = input[idx:]
	return f, s
}
