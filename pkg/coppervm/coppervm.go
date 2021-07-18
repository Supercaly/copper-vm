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
	Operand    int
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
