package copperdb

import "github.com/Supercaly/coppervm/pkg/coppervm"

type Breakpoint struct {
	Number uint
	Addr   coppervm.InstAddr
}

func EmptyBreakpoint() Breakpoint { return Breakpoint{} }

type Breakpoints []Breakpoint

// Implement the sort interface
func (a Breakpoints) Len() int           { return len(a) }
func (a Breakpoints) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Breakpoints) Less(i, j int) bool { return a[i].Number < a[j].Number }

// Returns the index of the breakpoint with given number
// or -1 if it doesn't exist.
func (brs Breakpoints) GetIndexByNumber(num uint) int {
	for idx, br := range brs {
		if br.Number == num {
			return idx
		}
	}
	return -1
}

// Returns the index of the breakpoint with given address
// or -1 if it doesn't exist.
func (brs Breakpoints) GetIndexByAddress(addr coppervm.InstAddr) int {
	for idx, br := range brs {
		if br.Addr == addr {
			return idx
		}
	}
	return -1
}
