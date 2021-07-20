package casm

type Binding struct {
	Name     string
	Value    Expression
	Location FileLocation
}

type DeferredOperand struct {
	Name     string
	Address  int
	Location FileLocation
}
