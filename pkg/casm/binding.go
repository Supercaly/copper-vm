package casm

type Binding struct {
	Name     string
	Value    Expression
	Location FileLocation
	IsLabel  bool
}

type DeferredOperand struct {
	Name     string
	Address  int
	Location FileLocation
}
