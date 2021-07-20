package casm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

// Save a copper vm program to binary file.
func (casm *Casm) SaveProgramToFile(filePath string) {
	fmt.Printf("Entry point: %d\n", casm.Entry)
	println("Dump Program")
	for _, i := range casm.Program {
		fmt.Printf("%s %d\n", i.Name, i.Operand)
	}
	println("End Dump Program")

	test, err := json.Marshal(coppervm.FileMeta(casm.Entry, casm.Program))
	if err != nil {
		log.Fatalf("[ERROR]: Error writign program to file %s", err)
	}

	fileErr := ioutil.WriteFile(filePath, []byte(test), os.ModePerm)
	if fileErr != nil {
		log.Fatalf("[ERROR]: Error saving file '%s': %s", filePath, fileErr)
	}
	println("[INFO]: Program saved to '" + filePath + "'")
}

// Translate a copper assembly file to copper vm's binary.
// Given a file path this function will read it and parse it
// as assembly code for the vm generating the correct program
// in-memory.
// Use SaveProgramToFile to save the program to binary file.
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

	println("Binding before second pass")
	for _, b := range casm.Bindings {
		fmt.Printf("Binding: %s (%d %d %s) %s\n",
			b.Name,
			b.Value.Kind,
			b.Value.AsNumLit,
			b.Value.AsBinding,
			b.Location)
	}

	// Second Pass
	for _, deferredOp := range casm.DeferredOperands {
		fmt.Printf("Resolve def_op: %s at %d\n", deferredOp.Name, deferredOp.Address)

		exist, binding := casm.getBindingByName(deferredOp.Name)
		if !exist {
			log.Fatalf("%s: [ERROR]: Unknown binding '%s'",
				deferredOp.Location,
				deferredOp.Name)
		}
		casm.Program[deferredOp.Address].Operand = casm.evaluateBinding(&binding, deferredOp.Location)
	}

	println("Binding after second pass")
	for _, b := range casm.Bindings {
		fmt.Printf("Binding: %s (%d %d %s) %s\n",
			b.Name,
			b.Value.Kind,
			b.Value.AsNumLit,
			b.Value.AsBinding,
			b.Location)
	}

	// Resolve entry point
	if casm.HasEntry && casm.DeferredEntryName != "" {
		exist, binding := casm.getBindingByName(casm.DeferredEntryName)
		if !exist {
			log.Fatalf("%s: [ERROR]: Unknown binding '%s'",
				casm.EntryLocation,
				casm.DeferredEntryName)
		}

		if binding.Value.Kind != ExpressionKindNumLit {
			log.Fatalf("%s: [ERROR]: Only label names can be set as entry point",
				casm.EntryLocation)
		}
		casm.Entry = casm.evaluateBinding(&binding, casm.EntryLocation)
	}
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

// Evaluate a binding to extract a word.
func (casm *Casm) evaluateBinding(binding *Binding, location FileLocation) (value int) {
	// TODO(#6): Cyclic bindings cause a stack overflow
	value = casm.evaluateExpression(binding.Value, location)
	return value
}

// Evaluate an expression to extract a word.
func (casm *Casm) evaluateExpression(expr Expression, location FileLocation) (ret int) {
	switch expr.Kind {
	case ExpressionKindBinding:
		exist, binding := casm.getBindingByName(expr.AsBinding)
		if !exist {
			log.Fatalf("%s: [ERROR]: Cannot find binding '%s'",
				location,
				expr.AsBinding)
		}
		ret = casm.evaluateBinding(&binding, location)
	case ExpressionKindNumLit:
		ret = expr.AsNumLit
	}
	return ret
}
