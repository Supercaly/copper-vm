package main

import c "coppervm.com/coppervm/pkg/casm"

func main() {
	casm := c.Casm{}
	casm.TranslateSource("examples/123.copper")
	casm.SaveProgramToFile("examples/bin/123.vm")
}
