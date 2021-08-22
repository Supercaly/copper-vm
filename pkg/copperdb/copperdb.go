package copperdb

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Supercaly/coppervm/internal"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type Copperdb struct {
	InputFile string

	vm *coppervm.Coppervm

	breakpoints       Breakpoints
	breakpointCount   uint
	currentBreakpoint Breakpoint

	debugSymbols coppervm.DebugSymbols

	sessionHalt bool
}

func NewCopperdb(inputFile string) Copperdb {
	return Copperdb{
		InputFile: inputFile,
		vm:        &coppervm.Coppervm{},
	}
}

// Start the debugger session.
// In this method the debugger will promp the user for commands
// and execute them.
func (db *Copperdb) StartDebugSession() error {
	meta, err := db.vm.LoadProgramFromFile(db.InputFile)
	if err != nil {
		return err
	}
	db.vm.Halt = true
	db.debugSymbols = meta.DebugSymbols

	// Start db session
	reader := bufio.NewReader(os.Stdin)
	for !db.sessionHalt {
		fmt.Print("(coppervm) ")
		str, err := reader.ReadString('\n')
		str = strings.TrimSuffix(str, "\n")
		str = strings.TrimSpace(str)
		if err != nil {
			return fmt.Errorf("something went wrong with the debugger: %s", err)
		}
		db.executeCommand(str)
	}

	return nil
}

// Execute a command in input.
func (db *Copperdb) executeCommand(input string) {
	cmd, args := internal.SplitByDelim(input, ' ')
	cmd = strings.TrimSpace(cmd)
	args = strings.TrimSpace(args)

	// Skip if the command is empty
	if cmd == "" {
		return
	}

	switch cmd {
	case "r":
		db.runProgram()
	case "c":
		db.continueProgram()
	case "s":
		db.stepProgram()
	case "b":
		addr, err := db.stringToBrAddress(args)
		if err != nil {
			fmt.Println(err)
		} else {
			db.addBreakpoint(coppervm.InstAddr(addr))
		}
	case "d":
		num, err := db.stringToBrNum(args)
		if err != nil {
			fmt.Printf("Invalid breakpoint number '%s'\n", args)
		} else {
			db.removeBreakpoint(num)
		}
	case "l":
		db.listBreakpoints()
	case "p":
		db.vm.DumpStack()
		fmt.Println()
	case "m":
		db.vm.DumpMemory()
		fmt.Println()
	case "x":
		if !db.vm.Halt {
			fmt.Printf("[%d] -> %s\n", db.vm.Ip, db.vm.Program[db.vm.Ip])
		} else {
			fmt.Println("The program is not being run. Use 'r' to run it first.")
		}
	case "q":
		if !db.vm.Halt {
			fmt.Println("A debugging session is still active")
			if !internal.AskConfirmation("Quit anyway?") {
				return
			}
		}
		fmt.Println("Bye!")
		db.sessionHalt = true
	case "h":
		db.PrintHelp()
	default:
		fmt.Printf("Unknown instruction '%s'. Try 'h'.\n", cmd)
	}
}

// Reset the debugger status.
func (db *Copperdb) Reset() {
	db.currentBreakpoint = EmptyBreakpoint()
	db.vm.Reset()
}

// Executes count instructions of the debugged program.
func (db *Copperdb) executeInstructions(count int) {
	for count != 0 && !db.vm.Halt {
		if db.brakeAtAddr(db.vm.Ip) {
			// Reached a breakpoint
			db.currentBreakpoint = db.breakpoints[db.breakpoints.GetIndexByAddress(db.vm.Ip)]
			fmt.Printf("\nBreakpoint %d, %d\n",
				db.currentBreakpoint.Number,
				db.currentBreakpoint.Addr)
			return
		} else {
			db.currentBreakpoint = EmptyBreakpoint()
			// Execute current instruction
			err := db.vm.ExecuteInstruction()
			if err.Kind != coppervm.ErrorKindOk {
				fmt.Println("Error", err)
				db.vm.Halt = true
				return
			}
		}
		count--
	}
	if db.vm.Halt {
		fmt.Printf("Program executed with exit code '%d'\n", db.vm.ExitCode)
	}
}

// Returns true if the current address is a brakepoint not reached,
// false otherwise.
func (db *Copperdb) brakeAtAddr(addr coppervm.InstAddr) bool {
	brIdx := db.breakpoints.GetIndexByAddress(addr)
	if brIdx == -1 {
		return false
	}
	return db.breakpoints[brIdx] != db.currentBreakpoint
}

// Run the debugged program if it's not already running.
func (db *Copperdb) runProgram() {
	if !db.vm.Halt {
		fmt.Println("The program has been started already.")
		if internal.AskConfirmation("Start it from the beginning?") {
			db.vm.Halt = true
			db.runProgram()
		}
	} else {
		fmt.Printf("Starting program '%s'\n", db.InputFile)
		db.Reset()
		db.executeInstructions(-1)
	}
}

// Continue the execution of the program if it's running.
func (db *Copperdb) continueProgram() {
	if db.vm.Halt {
		fmt.Println("The program is not being run. Use 'r' to run it first.")
	} else {
		fmt.Println("Continuing.")
		db.executeInstructions(-1)
	}
}

// Execute a single instruction of the program.
func (db *Copperdb) stepProgram() {
	if db.vm.Halt {
		fmt.Println("The program is not being run. Use 'r' to run it first.")
	} else {
		fmt.Println("Stepping.")
		db.executeInstructions(1)
	}
}

// Print all set breakpoints.
func (db *Copperdb) listBreakpoints() {
	if len(db.breakpoints) > 0 {
		fmt.Println("Num\tAddress")
		for _, br := range db.breakpoints {
			fmt.Printf("%d\t%d\n", br.Number, br.Addr)
		}
	} else {
		fmt.Println("No breakpoints.")
	}
}

// Add a new breakpoint at given address.
func (db *Copperdb) addBreakpoint(addr coppervm.InstAddr) {
	db.breakpointCount++
	newBr := Breakpoint{
		Number: db.breakpointCount,
		Addr:   addr,
	}
	db.breakpoints = append(db.breakpoints, newBr)
	fmt.Println("Breakpoint", newBr.Number, "at", newBr.Addr)
}

// Remove breakpoint withgiven number.
func (db *Copperdb) removeBreakpoint(brNum uint) {
	brIdx := db.breakpoints.GetIndexByNumber(brNum)
	if brIdx == -1 {
		fmt.Println("No breakpoint number", brNum)
	} else {
		removedBr := db.breakpoints[brIdx]
		db.breakpoints[brIdx] = db.breakpoints[len(db.breakpoints)-1]
		db.breakpoints = db.breakpoints[:len(db.breakpoints)-1]
		sort.Sort(db.breakpoints)
		fmt.Println("Delete breakpoint", removedBr.Number, "at", removedBr.Addr)
	}
}

// Print a help message.
func (db *Copperdb) PrintHelp() {
	fmt.Println("r           -- Start debugged program.")
	fmt.Println("c           -- Continue program being debugged after breakpoint.")
	fmt.Println("s           -- Step program to next instruction.")
	fmt.Println("b <loc|sym> -- Set a new breakpoint at specified location or symbol.")
	fmt.Println("d <loc>     -- Delete breakpoint at specified location.")
	fmt.Println("l           -- List all breakpoints.")
	fmt.Println("p           -- Dump the stack.")
	fmt.Println("m           -- Dump the memory.")
	fmt.Println("x           -- Print the instruction at ip.")
	fmt.Println("q           -- Quit the debugger.")
	fmt.Println("h           -- Print this help message.")
}

// Parse a string to an breakpoint number.
func (db *Copperdb) stringToBrNum(str string) (uint, error) {
	out64, err := strconv.ParseUint(str, 10, 64)
	return uint(out64), err
}

// Parse a string to an address.
func (db *Copperdb) stringToBrAddress(str string) (out int, err error) {
	// Try parse as address
	out, err = strconv.Atoi(str)
	if err != nil {
		// Try parse as symbol name
		idx := db.debugSymbols.GetIndexByName(str)
		if idx == -1 {
			return out, fmt.Errorf("no debug symbol or address '%s'", str)
		}
		out = int(db.debugSymbols[idx].Address)
	}
	return out, nil
}
