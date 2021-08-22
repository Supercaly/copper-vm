package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Supercaly/coppervm/pkg/copperdb"
)

func usage(stream io.Writer, program string) {
	fmt.Fprintf(stream, "Usage: %s <input.copper>\n", program)
}

func main() {
	if len(os.Args) != 2 {
		usage(os.Stderr, os.Args[0])
		log.Fatalf("[ERROR]: input was not provided\n")
	}

	db := copperdb.NewCopperdb(os.Args[1])
	if err := db.StartDebugSession(); err != nil {
		log.Fatalf("[ERROR]: %s", err)
	}
}
