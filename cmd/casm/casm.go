package main

import (
	"fmt"
	"io"
	"log"
	"os"

	au "coppervm.com/coppervm/internal"
	c "coppervm.com/coppervm/pkg/casm"
)

func usage(stream io.Writer, program string) {
	fmt.Fprintf(stream, "Usage: %s [OPTIONS] <input.copper>\n", program)
	fmt.Fprintf(stream, "[OPTIONS]: \n")
	fmt.Fprintf(stream, "    -o <out.vm>    Specify the output path.\n")
	fmt.Fprintf(stream, "    -h             Print this help message.\n")
}

func main() {
	args := os.Args
	var program string
	program, args = au.Shift(args)
	var inputFilePath string
	var outputFilePath string

	for len(args) > 0 {
		var flag string
		flag, args = au.Shift(args)

		if flag == "-h" {
			usage(os.Stdout, program)
			os.Exit(0)
		} else if flag == "-o" {
			if len(args) == 0 {
				usage(os.Stderr, program)
				log.Fatalf("[ERROR]: No argument provided for flag `%s`\n", flag)
			}

			outputFilePath, args = au.Shift(args)
		} else {
			if inputFilePath != "" {
				usage(os.Stderr, program)
				log.Fatalf("[ERROR]: input file is already provided as `%s`.\n", inputFilePath)
			}

			inputFilePath = flag
		}
	}

	if inputFilePath == "" {
		usage(os.Stderr, program)
		log.Fatalf("[ERROR]: input was not provided\n")
	}

	// TODO(#17): Use input path to determine output if not specified by the user
	if outputFilePath == "" {
		usage(os.Stderr, program)
		log.Fatalf("[ERROR]: output was not provided\n")
	}

	casm := c.Casm{}
	casm.TranslateSource(inputFilePath)
	casm.SaveProgramToFile(outputFilePath)
}
