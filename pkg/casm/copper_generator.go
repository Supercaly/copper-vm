package casm

import (
	"encoding/json"
	"fmt"

	"github.com/Supercaly/coppervm/pkg/coppervm"
)

type copperGenerator struct {
	rep       *internalRep
	dbSymbols coppervm.DebugSymbols
	program   []coppervm.InstDef
}

func (gen *copperGenerator) saveProgram(addDebugSymbols bool) string {
	if addDebugSymbols {
		gen.addDebugSymbols()
	}

	meta := coppervm.FileMeta(gen.rep.entry, gen.program, gen.rep.memory, gen.dbSymbols)
	metaJson, err := json.Marshal(meta)
	if err != nil {
		panic(fmt.Errorf("error writing program to file %s", err))
	}

	return string(metaJson)
}

func (gen *copperGenerator) generateProgram() {
	gen.internalProgramToVMProgram()
}

func (gen *copperGenerator) addDebugSymbols() {
	for _, b := range gen.rep.bindings {
		if b.isLabel {
			gen.dbSymbols = append(gen.dbSymbols, coppervm.DebugSymbol{
				Name:    b.name,
				Address: coppervm.InstAddr(b.evaluatedWord.asInstAddr),
			})
		}
	}
}

func (gen *copperGenerator) internalProgramToVMProgram() {
	for _, i := range gen.rep.program {
		gen.program = append(gen.program, coppervm.InstDef{
			Kind:       i.kind,
			HasOperand: i.hasOperand,
			Name:       i.name,
			Operand:    i.operand.toCoppervmWord(),
		})
	}
}
