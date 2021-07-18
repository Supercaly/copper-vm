package casm

import (
	"fmt"
	"io/ioutil"
	"log"
)

type Casm struct {
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

	lines := Linize(source)
	for _, l := range lines {
		fmt.Printf("%s\n", l.Kind)
	}
}
