package coppervm

const (
	CoppervmFileVersion int = 1
)

type CoppervmFileMeta struct {
	Version      int          `json:"version"`
	Entry        int          `json:"entry_point"`
	Program      []InstDef    `json:"program"`
	Memory       []byte       `json:"memory"`
	DebugSymbols DebugSymbols `json:"db_symbols"`
}

func FileMeta(entryPoint int, program []InstDef, memory []byte, symbols DebugSymbols) CoppervmFileMeta {
	return CoppervmFileMeta{
		Version:      CoppervmFileVersion,
		Entry:        entryPoint,
		Program:      program,
		Memory:       memory,
		DebugSymbols: symbols,
	}
}

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
