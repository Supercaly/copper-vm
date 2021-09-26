package casm

import (
	"fmt"
)

type binding struct {
	status        bindingStatus
	name          string
	value         Expression
	evaluatedWord word
	evaluatedKind ExpressionKind
	location      FileLocation
	isLabel       bool
}

func (b binding) String() string {
	return fmt.Sprintf("%s %s (%s) %s %s %t",
		b.name,
		b.value,
		b.evaluatedWord,
		b.status,
		b.location,
		b.isLabel)
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
