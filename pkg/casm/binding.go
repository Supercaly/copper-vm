package casm

import (
	"fmt"

	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type binding struct {
	Status        bindingStatus
	Name          string
	Value         Expression
	EvaluatedWord coppervm.Word
	EvaluatedKind ExpressionKind
	Location      FileLocation
	IsLabel       bool
}

func (b binding) String() string {
	return fmt.Sprintf("%s %s (%s) %s %s %t",
		b.Name,
		b.Value,
		b.EvaluatedWord,
		b.Status,
		b.Location,
		b.IsLabel)
}

type bindingStatus int

const (
	bindingUnevaluated bindingStatus = iota
	bindingEvaluating
	bindingEvaluated
)

func (state bindingStatus) String() string {
	return [...]string{
		"BindingUnevaluated",
		"BindingEvaluating",
		"BindingEvaluated",
	}[state]
}

type deferredOperand struct {
	Name     string
	Address  int
	Location FileLocation
}
