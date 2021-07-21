package coppervm

// List of all existing instructions
// of a coppervm program.
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
		Kind:       InstSwap,
		HasOperand: true,
		Name:       "swap",
	},
	{
		Kind:       InstDup,
		HasOperand: false,
		Name:       "dup",
	},
	{
		Kind:       InstAdd,
		HasOperand: false,
		Name:       "add",
	},
	{
		Kind:       InstSub,
		HasOperand: false,
		Name:       "sub",
	},
	{
		Kind:       InstJnz,
		HasOperand: true,
		Name:       "jnz",
	},
	{
		Kind:       InstPrint,
		HasOperand: false,
		Name:       "write",
	},
	{
		Kind:       InstHalt,
		HasOperand: false,
		Name:       "halt",
	},
}

type InstKind int

const (
	// TODO(#9): Add more instructions
	InstNoop InstKind = iota

	InstPush
	InstSwap
	InstDup

	InstAdd
	InstSub

	InstJnz

	InstPrint

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

// Return an instruction definition by it's string
// representation.
// This function return true if the instruction exist,
// false otherwise.
func GetInstDefByName(name string) (bool, InstDef) {
	for _, inst := range InstDefs {
		if inst.Name == name {
			return true, inst
		}
	}
	return false, InstDef{}
}
