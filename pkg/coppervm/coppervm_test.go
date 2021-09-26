package coppervm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadProgramFromFile(t *testing.T) {
	tests := []struct {
		path     string
		hasError bool
	}{
		{"testdata/test.notcopper", true},
		{"testdata/test1.copper", true},
		{"testdata/test.copper", false},
	}
	vm := Coppervm{}

	for _, test := range tests {
		_, err := vm.LoadProgramFromFile(test.path)
		if test.hasError {
			assert.Error(t, err, test)
		} else {
			assert.NoError(t, err, test)
		}
	}
}

func TestLoadProgramFromMeta(t *testing.T) {
	func() {
		vm := Coppervm{}
		meta := FileMeta(2, []InstDef{
			InstDef{
				Kind:       InstAddInt,
				HasOperand: false,
				Name:       "add",
			},
			InstDef{
				Kind:       InstNoop,
				HasOperand: false,
				Name:       "noop",
			},
			InstDef{
				Kind:       InstPush,
				HasOperand: true,
				Name:       "push",
			},
			InstDef{
				Kind:       InstHalt,
				HasOperand: false,
				Name:       "halt",
			},
		}, []byte{1, 2, 3}, DebugSymbols{})
		vm.loadProgramFromMeta(meta)

		assert.Equal(t, InstAddr(2), vm.Ip)
		assert.Len(t, vm.Program, 4)
		assert.Equal(t, byte(1), vm.Memory[0])
		assert.Equal(t, byte(2), vm.Memory[1])
		assert.Equal(t, byte(3), vm.Memory[2])
	}()

	func() {
		defer func() { recover() }()
		vm := Coppervm{}
		meta := FileMeta(0,
			[]InstDef{},
			make([]byte, CoppervmMemoryCapacity+1),
			DebugSymbols{})
		vm.loadProgramFromMeta(meta)

		assert.Fail(t, "expecting an error")
	}()
}

var instructionsTests = []struct {
	prog       []InstDef
	stack      []Word
	memory     []byte
	additional func(t assert.TestingT, vm Coppervm)
	err        CoppervmErrorKind
}{
	// noop
	{
		[]InstDef{{Kind: InstNoop}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(1), vm.Ip)
		},
		ErrorKindOk,
	},
	// push
	{
		[]InstDef{{Kind: InstPush, Operand: WordU64(1)}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordU64(1), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstPush, Operand: WordU64(1)}},
		make([]Word, CoppervmStackCapacity),
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackOverflow,
	},
	// swap
	{
		[]InstDef{{Kind: InstSwap, Operand: WordU64(1)}},
		[]Word{WordU64(1), WordU64(2)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, WordU64(2), vm.Stack[0])
			assert.Equal(t, WordU64(1), vm.Stack[1])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstSwap, Operand: WordU64(1)}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// dup
	{
		[]InstDef{{Kind: InstDup}},
		[]Word{WordU64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(2), vm.StackSize)
			assert.Equal(t, vm.Stack[0], vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstDup}},
		make([]Word, CoppervmStackCapacity),
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackOverflow,
	},
	{
		[]InstDef{{Kind: InstDup}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// drop
	{
		[]InstDef{{Kind: InstDrop}},
		[]Word{WordU64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(0), vm.StackSize)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstDrop}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// add int
	{
		[]InstDef{{Kind: InstAddInt}},
		[]Word{WordU64(1), WordU64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordU64(2), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstAddInt}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// sub int
	{
		[]InstDef{{Kind: InstSubInt}},
		[]Word{WordU64(1), WordU64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordU64(0), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstSubInt}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mul int unsigned
	{
		[]InstDef{{Kind: InstMulInt}},
		[]Word{WordU64(3), WordU64(2)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordU64(6), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstMulInt}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mul int signed
	{
		[]InstDef{{Kind: InstMulIntSigned}},
		[]Word{WordU64(1), WordI64(-2)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordI64(-2), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstMulIntSigned}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// div int unsigned
	{
		[]InstDef{{Kind: InstDivInt}},
		[]Word{WordU64(4), WordU64(2)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordU64(2), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstDivInt}},
		[]Word{WordU64(1), WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindDivideByZero,
	},
	{
		[]InstDef{{Kind: InstDivInt}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// div int signed
	{
		[]InstDef{{Kind: InstDivIntSigned}},
		[]Word{WordI64(-2), WordU64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordI64(-2), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstDivIntSigned}},
		[]Word{WordU64(1), WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindDivideByZero,
	},
	{
		[]InstDef{{Kind: InstDivIntSigned}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mod int unsigned
	{
		[]InstDef{{Kind: InstModInt}},
		[]Word{WordU64(5), WordU64(3)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordU64(2).AsU64, vm.Stack[0].AsU64)
			assert.Equal(t, WordU64(2).AsI64, vm.Stack[0].AsI64)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstModInt}},
		[]Word{WordU64(1), WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindDivideByZero,
	},
	{
		[]InstDef{{Kind: InstModInt}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mod int signed
	{
		[]InstDef{{Kind: InstModIntSigned}},
		[]Word{WordI64(-5), WordI64(3)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordI64(-2).AsI64, vm.Stack[0].AsI64)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstModIntSigned}},
		[]Word{WordU64(1), WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindDivideByZero,
	},
	{
		[]InstDef{{Kind: InstModIntSigned}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// add float
	{
		[]InstDef{{Kind: InstAddFloat}},
		[]Word{WordF64(1.0), WordF64(1.0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordF64(2.0), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstAddFloat}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// sub float
	{
		[]InstDef{{Kind: InstSubFloat}},
		[]Word{WordF64(1.0), WordF64(1.0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordF64(0.0), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstSubFloat}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mul float
	{
		[]InstDef{{Kind: InstMulFloat}},
		[]Word{WordF64(2.0), WordF64(3.0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordF64(6.0), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstMulFloat}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// div float
	{
		[]InstDef{{Kind: InstDivFloat}},
		[]Word{WordF64(1.0), WordF64(2.0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordF64(0.5), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstDivFloat}},
		[]Word{WordF64(1.0), WordF64(0.0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindDivideByZero,
	},
	{
		[]InstDef{{Kind: InstDivFloat}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// cmp
	{
		[]InstDef{{Kind: InstCmp}},
		[]Word{WordI64(1), WordI64(2)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, WordI64(-1), vm.Stack[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstCmp}},
		[]Word{WordI64(2)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// jmp
	{
		[]InstDef{{Kind: InstJmp, Operand: WordU64(5)}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(5), vm.Ip)
		},
		ErrorKindOk,
	},
	// jmp zero
	{
		[]InstDef{{Kind: InstJmpZero, Operand: WordU64(5)}},
		[]Word{WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(5), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpZero, Operand: WordU64(5)}},
		[]Word{WordU64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(1), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpZero}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// jmp not zero
	{
		[]InstDef{{Kind: InstJmpNotZero, Operand: WordU64(5)}},
		[]Word{WordU64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(5), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpNotZero, Operand: WordU64(5)}},
		[]Word{WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(1), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpNotZero}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// jmp greather
	{
		[]InstDef{{Kind: InstJmpGreater, Operand: WordU64(5)}},
		[]Word{WordU64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(5), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpGreater, Operand: WordU64(5)}},
		[]Word{WordI64(-1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(1), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpGreater}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// jmp less
	{
		[]InstDef{{Kind: InstJmpLess, Operand: WordU64(5)}},
		[]Word{WordI64(-1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(5), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpLess, Operand: WordU64(5)}},
		[]Word{WordI64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(1), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpLess}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// jmp greather equal
	{
		[]InstDef{{Kind: InstJmpGreaterEqual, Operand: WordU64(5)}},
		[]Word{WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(5), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpGreaterEqual, Operand: WordU64(5)}},
		[]Word{WordI64(-1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(1), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpGreaterEqual}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// jmp less equal
	{
		[]InstDef{{Kind: InstJmpLessEqual, Operand: WordU64(5)}},
		[]Word{WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(5), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpLessEqual, Operand: WordU64(5)}},
		[]Word{WordI64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(1), vm.Ip)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstJmpLessEqual}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// call
	{
		[]InstDef{{Kind: InstFunCall, Operand: WordU64(5)}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(5), vm.Ip)
			assert.Equal(t, int64(1), vm.StackSize)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstFunCall}},
		make([]Word, CoppervmStackCapacity),
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackOverflow,
	},
	// ret
	{
		[]InstDef{{Kind: InstFunReturn}},
		[]Word{WordU64(5)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, InstAddr(5), vm.Ip)
			assert.Equal(t, int64(0), vm.StackSize)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstFunReturn}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mem read
	{
		[]InstDef{{Kind: InstMemRead}},
		[]Word{WordU64(1)},
		[]byte{0x01, 0x22},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, uint64(0x22), vm.Stack[0].AsU64)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstMemRead}},
		[]Word{WordU64(uint64(CoppervmMemoryCapacity))},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindIllegalMemoryAccess,
	},
	{
		[]InstDef{{Kind: InstMemRead}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mem iread
	{
		[]InstDef{{Kind: InstMemReadInt}},
		[]Word{WordU64(0)},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, int64(5), vm.Stack[0].AsI64)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstMemReadInt}},
		[]Word{WordU64(uint64(CoppervmMemoryCapacity))},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindIllegalMemoryAccess,
	},
	{
		[]InstDef{{Kind: InstMemReadInt}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mem fread
	{
		[]InstDef{{Kind: InstMemReadFloat}},
		[]Word{WordU64(0)},
		[]byte{0x40, 0x2, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(1), vm.StackSize)
			assert.Equal(t, float64(2.3), vm.Stack[0].AsF64)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstMemReadFloat}},
		[]Word{WordU64(uint64(CoppervmMemoryCapacity))},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindIllegalMemoryAccess,
	},
	{
		[]InstDef{{Kind: InstMemReadFloat}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mem write
	{
		[]InstDef{{Kind: InstMemWrite}},
		[]Word{WordU64(0x22), WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(0), vm.StackSize)
			assert.Equal(t, byte(0x22), vm.Memory[0])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstMemWrite}},
		[]Word{WordU64(0), WordU64(uint64(CoppervmMemoryCapacity))},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindIllegalMemoryAccess,
	},
	{
		[]InstDef{{Kind: InstMemWrite}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mem iwrite
	{
		[]InstDef{{Kind: InstMemWriteInt}},
		[]Word{WordU64(5), WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(0), vm.StackSize)
			assert.Equal(t, byte(0x0), vm.Memory[0])
			assert.Equal(t, byte(0x0), vm.Memory[1])
			assert.Equal(t, byte(0x0), vm.Memory[2])
			assert.Equal(t, byte(0x0), vm.Memory[3])
			assert.Equal(t, byte(0x0), vm.Memory[4])
			assert.Equal(t, byte(0x0), vm.Memory[5])
			assert.Equal(t, byte(0x0), vm.Memory[6])
			assert.Equal(t, byte(0x5), vm.Memory[7])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstMemWriteInt}},
		[]Word{WordU64(0), WordU64(uint64(CoppervmMemoryCapacity))},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindIllegalMemoryAccess,
	},
	{
		[]InstDef{{Kind: InstMemWriteInt}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// mem fwrite
	{
		[]InstDef{{Kind: InstMemWriteFloat}},
		[]Word{WordF64(2.3), WordU64(0)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(0), vm.StackSize)
			assert.Equal(t, byte(0x40), vm.Memory[0])
			assert.Equal(t, byte(0x02), vm.Memory[1])
			assert.Equal(t, byte(0x66), vm.Memory[2])
			assert.Equal(t, byte(0x66), vm.Memory[3])
			assert.Equal(t, byte(0x66), vm.Memory[4])
			assert.Equal(t, byte(0x66), vm.Memory[5])
			assert.Equal(t, byte(0x66), vm.Memory[6])
			assert.Equal(t, byte(0x66), vm.Memory[7])
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstMemWriteFloat}},
		[]Word{WordF64(0.0), WordU64(uint64(CoppervmMemoryCapacity))},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindIllegalMemoryAccess,
	},
	{
		[]InstDef{{Kind: InstMemWriteFloat}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// TODO: Test syscalls
	// print
	{
		[]InstDef{{Kind: InstPrint}},
		[]Word{WordU64(1)},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {
			assert.Equal(t, int64(0), vm.StackSize)
		},
		ErrorKindOk,
	},
	{
		[]InstDef{{Kind: InstPrint}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindStackUnderflow,
	},
	// invalid
	{
		[]InstDef{{Kind: InstCount}},
		[]Word{},
		[]byte{},
		func(t assert.TestingT, vm Coppervm) {},
		ErrorKindInvalidInstruction,
	},
}

func TestExecuteInstruction(t *testing.T) {

	for _, test := range instructionsTests {
		vm := Coppervm{}
		vm.Program = test.prog
		copy(vm.Stack[:], test.stack)
		copy(vm.Memory[:], test.memory)
		vm.StackSize = int64(len(test.stack))

		err := vm.ExecuteInstruction()

		assert.Equal(t, test.err, err.Kind)
		test.additional(t, vm)
	}
}

func TestPushStack(t *testing.T) {
	vm := Coppervm{Program: []InstDef{{}}}
	err := vm.pushStack(WordU64(1))
	assert.Equal(t, ErrorKindOk, err.Kind)

	vm.StackSize = CoppervmStackCapacity + 1
	err = vm.pushStack(WordU64(1))
	assert.Equal(t, ErrorKindStackOverflow, err.Kind)
}

func TestReset(t *testing.T) {
	vm := Coppervm{}
	vm.LoadProgramFromFile("testdata/test.copper")
	res := vm.ExecuteProgram(-1)

	assert.Equal(t, ErrorKindOk, res.Kind)
	assert.NotZero(t, vm.Ip)
	assert.NotZero(t, vm.StackSize)

	vm.Reset()
	assert.Zero(t, vm.Ip)
	assert.Zero(t, vm.StackSize)
}
