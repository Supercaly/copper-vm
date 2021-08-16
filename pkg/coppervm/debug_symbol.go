package coppervm

type DebugSymbol struct {
	Name    string
	Address InstAddr
}

type DebugSymbols []DebugSymbol

// Returns the index of a debug symbol with given name
// or -1 if it's not present.
func (ds DebugSymbols) GetIndexByName(name string) int {
	for i, s := range ds {
		if s.Name == name {
			return i
		}
	}
	return -1
}
