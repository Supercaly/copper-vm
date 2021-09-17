package casm

import "fmt"

type FileLocation struct {
	FileName string
	Location int
	Col      int
	Row      int
}

func (fl FileLocation) String() string {
	return fmt.Sprintf("%s:%d:%d", fl.FileName, fl.Row+1, fl.Col+1)
}
