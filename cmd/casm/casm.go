package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	au "github.com/Supercaly/coppervm/internal"
	c "github.com/Supercaly/coppervm/pkg/casm"
)

func usage(stream io.Writer, program string) {
	fmt.Fprintf(stream, "Usage: %s [OPTIONS] <input.copper>\n", program)
	fmt.Fprintf(stream, "[OPTIONS]: \n")
	fmt.Fprintf(stream, "    -I <include/path>    Add include path.\n")
	fmt.Fprintf(stream, "    -o <out.vm>          Specify the output path.\n")
	fmt.Fprintf(stream, "    -h                   Print this help message.\n")
}

func main() {
	casm := c.Casm{}
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
		} else if flag == "-I" {
			if len(args) == 0 {
				usage(os.Stderr, program)
				log.Fatalf("[ERROR]: No argument provided for flag `%s`\n", flag)
			}

			var includePath string
			includePath, args = au.Shift(args)
			casm.IncludePaths = append(casm.IncludePaths, includePath)
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

	if outputFilePath == "" {
		fileName := filepath.Base(inputFilePath)
		fileDir := filepath.Dir(inputFilePath)
		fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".vm"
		outputFilePath = filepath.Join(fileDir, fileName)
	}

	casm.TranslateSource(inputFilePath)
	casm.SaveProgramToFile(outputFilePath)
}
