package casm

import "fmt"

type IRKind int

const (
	IRKindLabel IRKind = iota
	IRKindInstruction
	IRKindEntry
	IRKindConst
	IRKindMemory
)

type IR struct {
	Location      FileLocation
	Kind          IRKind
	AsLabel       LabelIR
	AsInstruction InstructionIR
	AsEntry       EntryIR
	AsConst       ConstIR
	AsMemory      MemoryIR
}

type LabelIR struct {
	Name string
}

type InstructionIR struct {
	Name       string
	Operand    Expression
	HasOperand bool
}

type EntryIR struct {
	Name string
}

type ConstIR struct {
	Name  string
	Value Expression
}

type MemoryIR struct {
	Name  string
	Value Expression
}

func (ir IR) String() (out string) {
	out += "{"
	out += fmt.Sprintf("Kind: %s, ", ir.Kind)
	switch ir.Kind {
	case IRKindLabel:
		out += fmt.Sprintf("AsLabel: %s", ir.AsLabel)
	case IRKindInstruction:
		out += fmt.Sprintf("AsInstruction: %s", ir.AsInstruction)
	case IRKindEntry:
		out += fmt.Sprintf("AsEntry: %s", ir.AsEntry)
	case IRKindConst:
		out += fmt.Sprintf("AsConst: %s", ir.AsConst)
	case IRKindMemory:
		out += fmt.Sprintf("AsMemory: %s", ir.AsMemory)
	}
	out += "}"
	return out
}

func (kind IRKind) String() string {
	return [...]string{
		"IRKindLabel",
		"IRKindInstruction",
		"IRKindEntry",
		"IRKindConst",
		"IRKindMemory",
	}[kind]
}

func (l LabelIR) String() string {
	return fmt.Sprintf("{%s}", l.Name)
}

func (inst InstructionIR) String() (out string) {
	out += "{"
	out += inst.Name
	if inst.HasOperand {
		out += fmt.Sprintf(", %s", inst.Operand)
	}
	out += "}"
	return out
}

func (e EntryIR) String() string {
	return fmt.Sprintf("{%s}", e.Name)
}

func (c ConstIR) String() string {
	return fmt.Sprintf("{%s, %s}", c.Name, c.Value)
}

func (m MemoryIR) String() string {
	return fmt.Sprintf("{%s, %s}", m.Name, m.Value)
}
