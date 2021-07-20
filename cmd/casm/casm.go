package main

import c "coppervm.com/coppervm/pkg/casm"

func main() {
	// TODO(#8): Get casm program form command line arguments
	casm := c.Casm{}
	casm.TranslateSource("examples/123.copper")
	casm.SaveProgramToFile("examples/bin/123.vm")
}
