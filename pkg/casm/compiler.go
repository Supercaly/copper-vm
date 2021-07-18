package casm

import (
	"fmt"
	"io/ioutil"
	"log"

	"coppervm.com/coppervm/pkg/coppervm"
)

type Casm struct {
	Bindings []Binding

	Program []coppervm.InstDef

	HasEntry          bool
	Entry             int
	EntryLocation     FileLocation
	DeferredEntryName string
}

func (casm *Casm) SaveProgramToFile(filePath string) {
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
				log.Fatalf("%s: [ERROR]: Unknown instruction '%s'", line.Location, line.AsInstruction.Name)
			}

			if instDef.HasOperand {
				// TODO: Parse instruction operand
				instDef.Operand = -1
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
func (casm *Casm) bindLabel(name string, position int, location FileLocation) {
	exist, binding := casm.getBindingByName(name)
	if exist {
		log.Fatalf("%s: [ERROR]: Name is already bound at location '%s'", location, binding.Location)
	}

	casm.Bindings = append(casm.Bindings, Binding{
		Name:     name,
		Value:    position,
		Location: location,
	})
}

// Binds a constant.
func (casm *Casm) bindConst(directive DirectiveLine, location FileLocation) {
	exist, binding := casm.getBindingByName(directive.Block)
	if exist {
		log.Fatalf("%s: [ERROR]: Name is already bound at location '%s'", location, binding.Location)
	}

	casm.Bindings = append(casm.Bindings, Binding{
		Name:     directive.Block,
		Value:    -1,
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
