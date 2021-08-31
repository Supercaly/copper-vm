package internal

import "fmt"

var printEnabled bool = false

func EnableDebugPrint() {
	printEnabled = true
}

func DebugPrintEnabled() bool {
	return printEnabled
}

func DebugPrint(format string, a ...interface{}) {
	if printEnabled {
		fmt.Printf(format, a...)
	}
}
