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
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

func usage(stream io.Writer, program string) {
	fmt.Fprintf(stream, "Usage: %s [OPTIONS] <input.copper>\n", program)
	fmt.Fprintf(stream, "[OPTIONS]: \n")
	fmt.Fprintf(stream, "    -I <include/path>    Add include path.\n")
	fmt.Fprintf(stream, "    -o <out.vm>          Specify the output path.\n")
	fmt.Fprintf(stream, "    -d                   Add debug symbols to use with copperdb.\n")
	fmt.Fprintf(stream, "    -h                   Print this help message.\n")
}

func main() {
	casm := c.Casm{}
	args := os.Args
	var program string
	program, args = au.Shift(args)

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

			casm.OutputFile, args = au.Shift(args)
		} else if flag == "-d" {
			casm.AddDebugSymbols = true
		} else if flag == "-I" {
			if len(args) == 0 {
				usage(os.Stderr, program)
				log.Fatalf("[ERROR]: No argument provided for flag `%s`\n", flag)
			}

			var includePath string
			includePath, args = au.Shift(args)
			casm.IncludePaths = append(casm.IncludePaths, includePath)
		} else {
			if casm.InputFile != "" {
				usage(os.Stderr, program)
				log.Fatalf("[ERROR]: input file is already provided as `%s`.\n", casm.InputFile)
			}

			casm.InputFile = flag
		}
	}

	if casm.InputFile == "" {
		usage(os.Stderr, program)
		log.Fatalf("[ERROR]: input was not provided\n")
	}

	if casm.OutputFile == "" {
		fileName := filepath.Base(casm.InputFile)
		fileDir := filepath.Dir(casm.InputFile)
		fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName)) + coppervm.CoppervmFileExtention
		casm.OutputFile = filepath.Join(fileDir, fileName)
	}

	if err := casm.TranslateSourceFile(casm.InputFile); err != nil {
		log.Fatalf("[ERROR]: %s", err)
	}
	if err := casm.SaveProgramToFile(); err != nil {
		log.Fatalf("[ERROR]: %s", err)
	}
}
