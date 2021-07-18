package casm

import "fmt"

type FileLocation struct {
	FileName string
	Location int
}

func (fl FileLocation) String() string {
	return fmt.Sprintf("%s:%d", fl.FileName, fl.Location)
}
