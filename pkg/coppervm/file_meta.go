package coppervm

const (
	CoppervmFileVersion int = 1
)

type CoppervmFileMeta struct {
	Version int       `json:"version"`
	Entry   int       `json:"entry_point"`
	Program []InstDef `json:"program"`
	Memory  []byte    `json:"memory"`
}

func FileMeta(entryPoint int, program []InstDef, memory []byte) CoppervmFileMeta {
	return CoppervmFileMeta{
		Version: CoppervmFileVersion,
		Entry:   entryPoint,
		Program: program,
		Memory:  memory,
	}
}
