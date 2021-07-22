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
		Kind:       InstDrop,
		HasOperand: false,
		Name:       "drop",
	},
	{
		Kind:       InstAddInt,
		HasOperand: false,
		Name:       "add",
	},
	{
		Kind:       InstSubInt,
		HasOperand: false,
		Name:       "sub",
	},
	{
		Kind:       InstMulInt,
		HasOperand: false,
		Name:       "mul",
	},
	{
		Kind:       InstMulIntSigned,
		HasOperand: false,
		Name:       "imul",
	},
	{
		Kind:       InstDivInt,
		HasOperand: false,
		Name:       "div",
	},
	{
		Kind:       InstDivIntSigned,
		HasOperand: false,
		Name:       "idiv",
	},
	{
		Kind:       InstModInt,
		HasOperand: false,
		Name:       "mod",
	},
	{
		Kind:       InstModIntSigned,
		HasOperand: false,
		Name:       "imod",
	},
	{
		Kind:       InstAddFloat,
		HasOperand: false,
		Name:       "fadd",
	},
	{
		Kind:       InstSubFloat,
		HasOperand: false,
		Name:       "fsub",
	},
	{
		Kind:       InstMulFloat,
		HasOperand: false,
		Name:       "fmul",
	},
	{
		Kind:       InstDivFloat,
		HasOperand: false,
		Name:       "fdiv",
	},
	{
		Kind:       InstJmp,
		HasOperand: true,
		Name:       "jmp",
	},
	{
		Kind:       InstJmpNotZero,
		HasOperand: true,
		Name:       "jnz",
	},
	{
		Kind:       InstFunCall,
		HasOperand: true,
		Name:       "call",
	},
	{
		Kind:       InstFunReturn,
		HasOperand: false,
		Name:       "ret",
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

	// Basic instructions
	InstPush
	InstSwap
	InstDup
	InstDrop
	InstHalt

	// Integer arithmetics
	InstAddInt
	InstSubInt
	InstMulInt
	InstMulIntSigned
	InstDivInt
	InstDivIntSigned
	InstModInt
	InstModIntSigned

	// Floating point arithmetics
	InstAddFloat
	InstSubFloat
	InstMulFloat
	InstDivFloat

	// Flow control
	InstJmp
	InstJmpNotZero

	// Functions
	InstFunCall
	InstFunReturn

	InstPrint

	InstCount
)

type InstDef struct {
	Kind       InstKind
	HasOperand bool
	Name       string
	Operand    Word
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
