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

// Create a new CoppervmFileMeta with given entry point, program, memory and debug symbols.
func FileMeta(entryPoint int, program []InstDef, memory []byte, symbols DebugSymbols) CoppervmFileMeta {
	return CoppervmFileMeta{
		Version:      CoppervmFileVersion,
		Entry:        entryPoint,
		Program:      program,
		Memory:       memory,
		DebugSymbols: symbols,
	}
}
