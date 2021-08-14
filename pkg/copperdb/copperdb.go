package copperdb

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Supercaly/coppervm/internal"
	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type Copperdb struct {
	InputFile       string
	Vm              *coppervm.Coppervm
	Breakpoints     Breakpoints
	BreakpointCount uint
}

// Start the debugger session.
// In this method the debugger will promp the user for commands
// and execute them.
func (db *Copperdb) StartProgramDebug() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("(coppervm) ")
		str, err := reader.ReadString('\n')
		str = strings.TrimSuffix(str, "\n")
		str = strings.TrimSpace(str)

		if err != nil {
			log.Fatalf("Something went wrong with the debugger: %s", err)
		}
		db.ExecuteCommand(str)
	}
}

// Execute a command in input.
func (db *Copperdb) ExecuteCommand(input string) {
	cmd, args := internal.SplitByDelim(input, ' ')
	cmd = strings.TrimSpace(cmd)
	args = strings.TrimSpace(args)

	// Skip if the command is empty
	if cmd == "" {
		return
	}

	switch cmd {
	case "r":
		db.RunProgram()
	case "c":
		db.ContinueProgram()
	case "s":
		db.StepProgram()
	case "b":
		addr, err := db.StringToBrAddress(args)
		if err != nil {
			fmt.Println(err)
		} else {
			db.AddBreakpoint(coppervm.InstAddr(addr))
		}
	case "d":
		num, err := db.StringToBrNum(args)
		if err != nil {
			fmt.Printf("Invalid breakpoint number '%s'\n", args)
		} else {
			db.RemoveBreakpoint(num)
		}
	case "l":
		db.ListBreakpoints()
	case "p":
		db.Vm.DumpStack()
		fmt.Println()
	case "m":
		db.Vm.DumpMemory()
		fmt.Println()
	case "x":
		fmt.Printf("[%d] -> %s\n", db.Vm.Ip, db.Vm.Program[db.Vm.Ip])
	case "q":
		if !db.Vm.Halt {
			fmt.Println("A debugging session is still active")
			if !AskConfirmation("Quit anyway?") {
				return
			}
		}
		fmt.Println("Bye!")
		os.Exit(0)
	case "h":
		db.PrintHelp()
	default:
		fmt.Printf("Unknown instruction '%s'. Try 'h'.\n", cmd)
	}
}

// Reset the debugger status.
func (db *Copperdb) Reset() {
	db.Breakpoints.Reset()
	db.Vm.Reset()
}

// Executes count instructions of the debugged program.
func (db *Copperdb) ExecuteInstructions(count int) {
	for count != 0 && !db.Vm.Halt {
		if db.Breakpoints.IsAddressNonReachedBr(db.Vm.Ip) {
			// Reached a breakpoint... stop the execution
			brIdx := db.Breakpoints.GetIndexByAddress(db.Vm.Ip)
			db.Breakpoints[brIdx].IsReached = true
			br := db.Breakpoints[brIdx]
			fmt.Printf("\nBreakpoint %d, %d\n", br.Number, br.Addr)
			return
		} else {
			// Execute current instruction
			err := db.Vm.ExecuteInstruction()
			if err.Kind != coppervm.ErrorKindOk {
				fmt.Println("Error", err)
			}
		}
		count--
	}
	if db.Vm.Halt {
		fmt.Printf("Program executed with exit code '%d'\n", db.Vm.ExitCode)
	}
}

// Run the debugged program if it's not already running.
func (db *Copperdb) RunProgram() {
	if !db.Vm.Halt {
		fmt.Println("The program has been started already.")
		if AskConfirmation("Start it from the beginning?") {
			db.Vm.Halt = true
			db.RunProgram()
		}
	} else {
		fmt.Printf("Starting program '%s'\n", db.InputFile)
		db.Reset()
		db.ExecuteInstructions(-1)
	}
}

// Continue the execution of the program if it's running.
func (db *Copperdb) ContinueProgram() {
	if db.Vm.Halt {
		fmt.Println("The program is not being run. Use 'r' to run it first.")
	} else {
		fmt.Println("Continuing.")
		db.ExecuteInstructions(-1)
	}
}

// Execute a single instruction of the program.
func (db *Copperdb) StepProgram() {
	if db.Vm.Halt {
		fmt.Println("The program is not being run. Use 'r' to run it first.")
	} else {
		fmt.Println("Stepping.")
		db.ExecuteInstructions(1)
	}
}

// Print all set breakpoints.
func (db *Copperdb) ListBreakpoints() {
	if len(db.Breakpoints) > 0 {
		fmt.Println("Num\tAddress\tBroke")
		for _, br := range db.Breakpoints {
			fmt.Printf("%d\t%d\t%t\n", br.Number, br.Addr, br.IsReached)
		}
	} else {
		fmt.Println("No breakpoints.")
	}
}

// Add a new breakpoint at given address.
func (db *Copperdb) AddBreakpoint(addr coppervm.InstAddr) {
	// TODO(#48): Add a table of compiler symbols to casm so we can keep track of labels
	db.BreakpointCount++
	newBr := Breakpoint{
		Number:    db.BreakpointCount,
		Addr:      addr,
		IsReached: false,
	}
	db.Breakpoints = append(db.Breakpoints, newBr)
	fmt.Println("Breakpoint", newBr.Number, "at", newBr.Addr)
}

// Remove breakpoint withgiven number.
func (db *Copperdb) RemoveBreakpoint(brNum uint) {
	brIdx := db.Breakpoints.GetIndexByNumber(brNum)
	if brIdx == -1 {
		fmt.Println("No breakpoint number", brNum)
	} else {
		removedBr := db.Breakpoints[brIdx]
		db.Breakpoints[brIdx] = db.Breakpoints[len(db.Breakpoints)-1]
		db.Breakpoints = db.Breakpoints[:len(db.Breakpoints)-1]
		sort.Sort(db.Breakpoints)
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
func (db *Copperdb) StringToBrNum(str string) (uint, error) {
	out64, err := strconv.ParseUint(str, 10, 64)
	return uint(out64), err
}

// Parse a string to an address.
func (db *Copperdb) StringToBrAddress(str string) (out int, err error) {
	// Try parse as address
	out, err = strconv.Atoi(str)
	if err != nil {
		// Try parse as symbol name
		idx := db.Vm.DebugSymbols.GetIndexByName(str)
		if idx == -1 {
			return out, fmt.Errorf("no debug symbol or address '%s'", str)
		}
		out = int(db.Vm.DebugSymbols[idx].Address)
	}
	return out, nil
}
