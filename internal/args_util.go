package internal

// Pop the first argument form comman line args list.
// Returns the first argument and the rest.
func Shift(args []string) (out string, rest []string) {
	out = args[0]
	rest = args[1:]
	return out, rest
}
