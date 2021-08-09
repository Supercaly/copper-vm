package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	au "github.com/Supercaly/coppervm/internal"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

func usage(stream io.Writer, program string) {
	fmt.Fprintf(stream, "Usage: %s [OPTIONS] <input.vm>\n", program)
	fmt.Fprintf(stream, "OPTIONS:\n")
	fmt.Fprintf(stream, "    -l <limit>      Limit the steps of the emulation.\n")
	fmt.Fprintf(stream, "                    If negative no limit will be set.\n")
	fmt.Fprintf(stream, "    -h              Print this help message.\n")
}

func main() {
	args := os.Args
	var program string
	program, args = au.Shift(args)
	var inputFilePath string
	var limit int = -1

	for len(args) > 0 {
		var flag string
		flag, args = au.Shift(args)

		if flag == "-h" {
			usage(os.Stdout, program)
			os.Exit(0)
		} else if flag == "-l" {
			if len(args) == 0 {
				usage(os.Stderr, program)
				log.Fatalf("[ERROR]: No argument provided for flag `%s`\n", flag)
			}

			var limitStr string
			var err error
			limitStr, args = au.Shift(args)
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				log.Fatalf("[ERROR]: limit argument must be a number!")
			}
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

	vm := coppervm.Coppervm{}
	vm.LoadProgramFromFile(inputFilePath)
	if err := vm.ExecuteProgram(limit); err != coppervm.ErrorOk {
		log.Fatalf("%s: [ERROR]: '%s' at ip '%d'",
			inputFilePath,
			err,
			vm.Ip)
	}

	// Exit the program with vm's exit code
	os.Exit(vm.ExitCode)
}
