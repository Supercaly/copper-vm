package casm

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"coppervm.com/coppervm/pkg/coppervm"
)

type Casm struct {
	Bindings         []Binding
	DeferredOperands []DeferredOperand

	Program []coppervm.InstDef

	HasEntry          bool
	Entry             int
	EntryLocation     FileLocation
	DeferredEntryName string
}

func (casm *Casm) SaveProgramToFile(filePath string) {
	fmt.Printf("Entry point: %d\n", casm.Entry)
	println("Dump Program")
	for _, i := range casm.Program {
		fmt.Printf("%s %d\n", i.Name, i.Operand)
	}
	println("End Dump Program")

	// TODO(#3): Save compiled program to vm's binary
	println("Save to " + filePath)
}

func (casm *Casm) TranslateSource(filePath string) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file '%s': %s", filePath, err)
	}
	source := string(bytes)

	// Linize the source
	lines := Linize(source, filePath)

	// First Pass
	for _, line := range lines {
		switch line.Kind {
		case LineKindLabel:
			if line.AsLabel.Name == "" {
				log.Fatalf("%s: [ERROR]: Empty labels are not supported!", line.Location)
			}

			casm.bindLabel(line.AsLabel.Name, len(casm.Program), line.Location)
		case LineKindInstruction:
			exist, instDef := coppervm.GetInstDefByName(line.AsInstruction.Name)
			if !exist {
				log.Fatalf("%s: [ERROR]: Unknown instruction '%s'",
					line.Location,
					line.AsInstruction.Name)
			}

			if instDef.HasOperand {
				operand := ParseExprFromString(line.AsInstruction.Operand, line.Location)
				if operand.Kind == ExpressionKindBinding {
					println("Push deferred operand " + operand.AsBinding)
					casm.DeferredOperands = append(casm.DeferredOperands,
						DeferredOperand{
							Name:     operand.AsBinding,
							Address:  len(casm.Program),
							Location: line.Location,
						})
				} else {
					instDef.Operand = operand.AsNumLit
				}
			}
			casm.Program = append(casm.Program, instDef)
		case LineKindDirective:
			if line.AsDirective.Name == "entry" {
				casm.bindEntry(line.AsDirective.Block, line.Location)
			} else if line.AsDirective.Name == "const" {
				casm.bindConst(line.AsDirective, line.Location)
			} else {
				log.Fatalf("%s: [ERROR]: Unknown directive '%s'", line.Location, line.AsDirective.Name)
			}
		}
	}

	// Second Pass
	for _, b := range casm.Bindings {
		// TODO(#2): Compute deferred operands in second pass
		fmt.Printf("%s %d %s\n", b.Name, b.Value, b.Location)
	}

	// Resolve entry point
	if casm.HasEntry && casm.DeferredEntryName != "" {
		exist, binding := casm.getBindingByName(casm.DeferredEntryName)
		if !exist {
			log.Fatalf("%s: [ERROR]: Unknown binding '%s'",
				casm.EntryLocation,
				casm.DeferredEntryName)
		}

		casm.Entry = binding.Value
	}

	fmt.Printf("Entry point: %d\n", casm.Entry)
	println("Dump Program")
	for _, i := range casm.Program {
		fmt.Printf("%s %d\n", i.Name, i.Operand)
	}
	println("End Dump Program")
}

// Returns a binding by its name.
// If the binding exist the first return parameter will be true,
// otherwise it'll be null.
func (casm *Casm) getBindingByName(name string) (bool, Binding) {
	for _, b := range casm.Bindings {
		if b.Name == name {
			return true, b
		}
	}
	return false, Binding{}
}

// Binds a label.
func (casm *Casm) bindLabel(name string, address int, location FileLocation) {
	exist, binding := casm.getBindingByName(name)
	if exist {
		log.Fatalf("%s: [ERROR]: Name is already bound at location '%s'", location, binding.Location)
	}

	casm.Bindings = append(casm.Bindings, Binding{
		Name: name,
		Value: Expression{
			Kind:     ExpressionKindNumLit,
			AsNumLit: address,
		},
		Location: location,
	})
}

// Binds a constant.
func (casm *Casm) bindConst(directive DirectiveLine, location FileLocation) {
	name, block := SplitByDelim(directive.Block, ' ')
	name = strings.TrimSpace(name)
	block = strings.TrimSpace(block)

	exist, binding := casm.getBindingByName(name)
	if exist {
		log.Fatalf("%s: [ERROR]: Name '%s' is already bound at location '%s'",
			location,
			name,
			binding.Location)
	}

	casm.Bindings = append(casm.Bindings, Binding{
		Name:     name,
		Value:    ParseExprFromString(block, location),
		Location: location,
	})
}

// Binds an entry point.
func (casm *Casm) bindEntry(name string, location FileLocation) {
	if casm.HasEntry {
		log.Fatalf("%s: [ERROR]: Entry point is already set to '%s'",
			location,
			casm.EntryLocation)
	}

	casm.DeferredEntryName = name
	casm.HasEntry = true
	casm.EntryLocation = location
}
