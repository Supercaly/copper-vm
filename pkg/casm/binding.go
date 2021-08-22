package casm

import (
	"fmt"

	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type Binding struct {
	Status        BindingStatus
	Name          string
	Value         Expression
	EvaluatedWord coppervm.Word
	Location      FileLocation
	IsLabel       bool
}

func (b Binding) String() string {
	return fmt.Sprintf("%s %s (%s) %s %s %t",
		b.Name,
		b.Value,
		b.EvaluatedWord,
		b.Status,
		b.Location,
		b.IsLabel)
}

type BindingStatus int

const (
	BindingUnevaluated BindingStatus = iota
	BindingEvaluating
	BindingEvaluated
)

func (state BindingStatus) String() string {
	return [...]string{
		"BindingUnevaluated",
		"BindingEvaluating",
		"BindingEvaluated",
	}[state]
}

type DeferredOperand struct {
	Name     string
	Address  int
	Location FileLocation
}
