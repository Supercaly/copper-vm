package coppervm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIndexByName(t *testing.T) {
	ds := DebugSymbols{
		DebugSymbol{Name: "symbol1", Address: InstAddr(1)},
		DebugSymbol{Name: "symbol2", Address: InstAddr(2)},
		DebugSymbol{Name: "symbol3", Address: InstAddr(3)},
	}
	tests := []struct {
		name   string
		expect int
	}{
		{"symbol1", 0},
		{"symbol2", 1},
		{"symbol3", 2},
		{"symbol4", -1},
	}

	for _, test := range tests {
		assert.Equal(t, test.expect, ds.GetIndexByName(test.name))
	}
}
