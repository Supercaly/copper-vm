package coppervm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadProgramFromMeta(t *testing.T) {
	func() {
		vm := Coppervm{}
		m := FileMeta(2, []InstDef{
			InstDefs[InstAddInt],
			InstDefs[InstNoop],
			InstDefs[InstPush],
			InstDefs[InstHalt],
		}, []byte{1, 2, 3}, DebugSymbols{})
		vm.loadProgramFromMeta(m)

		assert.Equal(t, InstAddr(2), vm.Ip)
		assert.Equal(t, 4, len(vm.Program))
		assert.Equal(t, byte(1), vm.Memory[0])
		assert.Equal(t, byte(2), vm.Memory[1])
		assert.Equal(t, byte(3), vm.Memory[2])
	}()

	func() {
		defer func() { recover() }()
		vm := Coppervm{}
		m := FileMeta(0,
			[]InstDef{},
			make([]byte, CoppervmMemoryCapacity+1),
			DebugSymbols{})
		vm.loadProgramFromMeta(m)

		t.Error("expecting an error")
	}()
}

func TestExecuteInstruction(t *testing.T) {
	assert := assert.New(t)
	vm := Coppervm{}
	tests := []struct {
		prog       []InstDef
		stack      []Word
		memory     []byte
		additional func(vm Coppervm)
		err        CoppervmErrorKind
	}{
		// noop
		{
			[]InstDef{{Kind: InstNoop}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(1), vm.Ip)
			},
			ErrorKindOk,
		},
		// push
		{
			[]InstDef{{Kind: InstPush, Operand: WordU64(1)}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordU64(1), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstPush, Operand: WordU64(1)}},
			make([]Word, CoppervmStackCapacity),
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackOverflow,
		},
		// swap
		{
			[]InstDef{{Kind: InstSwap, Operand: WordU64(1)}},
			[]Word{WordU64(1), WordU64(2)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(WordU64(2), vm.Stack[0])
				assert.Equal(WordU64(1), vm.Stack[1])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstSwap, Operand: WordU64(1)}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// dup
		{
			[]InstDef{{Kind: InstDup}},
			[]Word{WordU64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(2), vm.StackSize)
				assert.Equal(vm.Stack[0], vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDup}},
			make([]Word, CoppervmStackCapacity),
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackOverflow,
		},
		{
			[]InstDef{{Kind: InstDup}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// drop
		{
			[]InstDef{{Kind: InstDrop}},
			[]Word{WordU64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(0), vm.StackSize)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDrop}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// add int
		{
			[]InstDef{{Kind: InstAddInt}},
			[]Word{WordU64(1), WordU64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordU64(2), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstAddInt}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// sub int
		{
			[]InstDef{{Kind: InstSubInt}},
			[]Word{WordU64(1), WordU64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordU64(0), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstSubInt}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// mul int unsigned
		{
			[]InstDef{{Kind: InstMulInt}},
			[]Word{WordU64(3), WordU64(2)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordU64(6), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMulInt}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// mul int signed
		{
			[]InstDef{{Kind: InstMulIntSigned}},
			[]Word{WordU64(1), WordI64(-2)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordI64(-2), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMulIntSigned}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// div int unsigned
		{
			[]InstDef{{Kind: InstDivInt}},
			[]Word{WordU64(4), WordU64(2)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordU64(2), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDivInt}},
			[]Word{WordU64(1), WordU64(0)},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstDivInt}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// div int signed
		{
			[]InstDef{{Kind: InstDivIntSigned}},
			[]Word{WordI64(-2), WordU64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordI64(-2), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDivIntSigned}},
			[]Word{WordU64(1), WordU64(0)},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstDivIntSigned}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// mod int unsigned
		{
			[]InstDef{{Kind: InstModInt}},
			[]Word{WordU64(5), WordU64(3)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordU64(2).AsU64, vm.Stack[0].AsU64)
				assert.Equal(WordU64(2).AsI64, vm.Stack[0].AsI64)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstModInt}},
			[]Word{WordU64(1), WordU64(0)},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstModInt}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// mod int signed
		{
			[]InstDef{{Kind: InstModIntSigned}},
			[]Word{WordI64(-5), WordI64(3)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordI64(-2).AsI64, vm.Stack[0].AsI64)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstModIntSigned}},
			[]Word{WordU64(1), WordU64(0)},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstModIntSigned}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// add float
		{
			[]InstDef{{Kind: InstAddFloat}},
			[]Word{WordF64(1.0), WordF64(1.0)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordF64(2.0), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstAddFloat}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// sub float
		{
			[]InstDef{{Kind: InstSubFloat}},
			[]Word{WordF64(1.0), WordF64(1.0)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordF64(0.0), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstSubFloat}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// mul float
		{
			[]InstDef{{Kind: InstMulFloat}},
			[]Word{WordF64(2.0), WordF64(3.0)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordF64(6.0), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMulFloat}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// div float
		{
			[]InstDef{{Kind: InstDivFloat}},
			[]Word{WordF64(1.0), WordF64(2.0)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordF64(0.5), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDivFloat}},
			[]Word{WordF64(1.0), WordF64(0.0)},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstDivFloat}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// cmp
		{
			[]InstDef{{Kind: InstCmp}},
			[]Word{WordI64(1), WordI64(2)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(WordI64(-1), vm.Stack[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstCmp}},
			[]Word{WordI64(2)},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// jmp
		{
			[]InstDef{{Kind: InstJmp, Operand: WordU64(5)}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(5), vm.Ip)
			},
			ErrorKindOk,
		},
		// jmp zero
		{
			[]InstDef{{Kind: InstJmpZero, Operand: WordU64(5)}},
			[]Word{WordU64(0)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(5), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpZero, Operand: WordU64(5)}},
			[]Word{WordU64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(1), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpZero}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// jmp not zero
		{
			[]InstDef{{Kind: InstJmpNotZero, Operand: WordU64(5)}},
			[]Word{WordU64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(5), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpNotZero, Operand: WordU64(5)}},
			[]Word{WordU64(0)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(1), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpNotZero}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// jmp greather
		{
			[]InstDef{{Kind: InstJmpGreater, Operand: WordU64(5)}},
			[]Word{WordU64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(5), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpGreater, Operand: WordU64(5)}},
			[]Word{WordI64(-1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(1), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpGreater}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// jmp less
		{
			[]InstDef{{Kind: InstJmpLess, Operand: WordU64(5)}},
			[]Word{WordI64(-1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(5), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpLess, Operand: WordU64(5)}},
			[]Word{WordI64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(1), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpLess}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// jmp greather equal
		{
			[]InstDef{{Kind: InstJmpGreaterEqual, Operand: WordU64(5)}},
			[]Word{WordU64(0)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(5), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpGreaterEqual, Operand: WordU64(5)}},
			[]Word{WordI64(-1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(1), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpGreaterEqual}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// jmp less equal
		{
			[]InstDef{{Kind: InstJmpLessEqual, Operand: WordU64(5)}},
			[]Word{WordU64(0)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(5), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpLessEqual, Operand: WordU64(5)}},
			[]Word{WordI64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(1), vm.Ip)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpLessEqual}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// call
		{
			[]InstDef{{Kind: InstFunCall, Operand: WordU64(5)}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(5), vm.Ip)
				assert.Equal(int64(1), vm.StackSize)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstFunCall}},
			make([]Word, CoppervmStackCapacity),
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackOverflow,
		},
		// ret
		{
			[]InstDef{{Kind: InstFunReturn}},
			[]Word{WordU64(5)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(InstAddr(5), vm.Ip)
				assert.Equal(int64(0), vm.StackSize)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstFunReturn}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// mem read
		{
			[]InstDef{{Kind: InstMemRead}},
			[]Word{WordU64(1)},
			[]byte{0x01, 0x22},
			func(vm Coppervm) {
				assert.Equal(int64(1), vm.StackSize)
				assert.Equal(uint64(0x22), vm.Stack[0].AsU64)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMemRead}},
			[]Word{WordU64(uint64(CoppervmMemoryCapacity))},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindIllegalMemoryAccess,
		},
		{
			[]InstDef{{Kind: InstMemRead}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// mem write
		{
			[]InstDef{{Kind: InstMemWrite}},
			[]Word{WordU64(0x22), WordU64(0)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(0), vm.StackSize)
				assert.Equal(byte(0x22), vm.Memory[0])
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMemWrite}},
			[]Word{WordU64(0), WordU64(uint64(CoppervmMemoryCapacity))},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindIllegalMemoryAccess,
		},
		{
			[]InstDef{{Kind: InstMemWrite}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// TODO: Test syscalls
		// print
		{
			[]InstDef{{Kind: InstPrint}},
			[]Word{WordU64(1)},
			[]byte{},
			func(vm Coppervm) {
				assert.Equal(int64(0), vm.StackSize)
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstPrint}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindStackUnderflow,
		},
		// invalid
		{
			[]InstDef{{Kind: InstCount}},
			[]Word{},
			[]byte{},
			func(vm Coppervm) {},
			ErrorKindInvalidInstruction,
		},
	}

	for _, test := range tests {
		vm.Ip = 0
		vm.Program = test.prog
		copy(vm.Stack[:], test.stack)
		copy(vm.Memory[:], test.memory)
		vm.StackSize = int64(len(test.stack))

		err := vm.ExecuteInstruction()
		if err.Kind != test.err {
			t.Errorf("expecting '%s' but got '%s'", test.err, err.Kind)
		}
		test.additional(vm)
	}
}
