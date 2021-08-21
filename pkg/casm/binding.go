package casm

import "github.com/Supercaly/coppervm/pkg/coppervm"

type Binding struct {
	Status        BindingStatus
	Name          string
	Value         Expression
	EvaluatedWord coppervm.Word
	Location      FileLocation
	IsLabel       bool
}

type BindingStatus int

const (
	BindingUnevaluated BindingStatus = iota
	BindingEvaluating
	BindingEvaluated
)

type DeferredOperand struct {
	Name     string
	Address  int
	Location FileLocation
}
