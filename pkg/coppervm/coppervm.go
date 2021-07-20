package coppervm

type InstKind int

const (
	InstNoop InstKind = iota
	InstPush
	InstAdd
	InstHalt
	InstCount
)

type InstDef struct {
	Kind       InstKind
	HasOperand bool
	Name       string
	// TODO(#5): Use Word as operand type
	Operand int
}

var InstDefs = [InstCount]InstDef{
	{
		Kind:       InstNoop,
		HasOperand: false,
		Name:       "noop",
	},
	{
		Kind:       InstPush,
		HasOperand: true,
		Name:       "push",
	},
	{
		Kind:       InstAdd,
		HasOperand: false,
		Name:       "add",
	},
	{
		Kind:       InstHalt,
		HasOperand: false,
		Name:       "halt",
	},
}

func GetInstDefByName(name string) (bool, InstDef) {
	for _, inst := range InstDefs {
		if inst.Name == name {
			return true, inst
		}
	}
	return false, InstDef{}
}

const (
	CoppervmFileVersion int = 1
)

type CoppervmFileMeta struct {
	Version int       `json:"version"`
	Entry   int       `json:"entry_point"`
	Program []InstDef `json:"program"`
}

func FileMeta(entryPoint int, program []InstDef) CoppervmFileMeta {
	return CoppervmFileMeta{
		Version: CoppervmFileVersion,
		Entry:   entryPoint,
		Program: program,
	}
}
