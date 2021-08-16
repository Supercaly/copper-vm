package coppervm

import (
	"testing"
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

		if vm.Ip != 2 {
			t.Errorf("expecting %d but got %d", 2, vm.Ip)
		}
		if len(vm.Program) != 4 {
			t.Errorf("expecting %d elements but got %d", 4, len(vm.Program))
		}
		if vm.Memory[0] != 1 && vm.Memory[1] != 2 && vm.Memory[2] != 3 {
			t.Error("memory values not correct")
		}
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
	vm := Coppervm{}
	tests := []struct {
		prog  []InstDef
		stack []Word
		err   CoppervmErrorKind
	}{
		// noop
		{
			[]InstDef{InstDef{Kind: InstNoop}},
			[]Word{},
			ErrorKindOk,
		},
		// push
		{
			[]InstDef{InstDef{Kind: InstPush, Operand: WordU64(1)}},
			[]Word{},
			ErrorKindOk,
		},
		{
			[]InstDef{InstDef{Kind: InstPush, Operand: WordU64(1)}},
			make([]Word, CoppervmStackCapacity),
			ErrorKindStackOverflow,
		},
		// swap
		{
			[]InstDef{{Kind: InstSwap, Operand: WordU64(1)}},
			[]Word{
				WordU64(1),
				WordU64(2),
			},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstSwap, Operand: WordU64(1)}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// dup
		{
			[]InstDef{{Kind: InstDup}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDup}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// drop
		{
			[]InstDef{{Kind: InstDrop}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDrop}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// add int
		{
			[]InstDef{{Kind: InstAddInt}},
			[]Word{WordU64(1), WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstAddInt}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// sub int
		{
			[]InstDef{{Kind: InstSubInt}},
			[]Word{WordU64(1), WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstSubInt}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// mul int unsigned
		{
			[]InstDef{{Kind: InstMulInt}},
			[]Word{WordU64(1), WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMulInt}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// mul int signed
		{
			[]InstDef{{Kind: InstMulIntSigned}},
			[]Word{WordU64(1), WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMulIntSigned}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// div int unsigned
		{
			[]InstDef{{Kind: InstDivInt}},
			[]Word{WordU64(1), WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDivInt}},
			[]Word{WordU64(1), WordU64(0)},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstDivInt}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// div int signed
		{
			[]InstDef{{Kind: InstDivIntSigned}},
			[]Word{WordU64(1), WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDivIntSigned}},
			[]Word{WordU64(1), WordU64(0)},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstDivIntSigned}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// mod int unsigned
		{
			[]InstDef{{Kind: InstModInt}},
			[]Word{WordU64(1), WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstModInt}},
			[]Word{WordU64(1), WordU64(0)},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstModInt}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// mod int signed
		{
			[]InstDef{{Kind: InstModIntSigned}},
			[]Word{WordU64(1), WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstModIntSigned}},
			[]Word{WordU64(1), WordU64(0)},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstModIntSigned}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// add float
		{
			[]InstDef{{Kind: InstAddFloat}},
			[]Word{WordF64(1.0), WordF64(1.0)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstAddFloat}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// sub float
		{
			[]InstDef{{Kind: InstSubFloat}},
			[]Word{WordF64(1.0), WordF64(1.0)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstSubFloat}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// mul float
		{
			[]InstDef{{Kind: InstMulFloat}},
			[]Word{WordF64(1.0), WordF64(1.0)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMulFloat}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// div float
		{
			[]InstDef{{Kind: InstDivFloat}},
			[]Word{WordF64(1.0), WordF64(1.0)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstDivFloat}},
			[]Word{WordF64(1.0), WordF64(0.0)},
			ErrorKindDivideByZero,
		},
		{
			[]InstDef{{Kind: InstDivFloat}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// cmp
		{
			[]InstDef{{Kind: InstCmp}},
			[]Word{WordI64(1), WordI64(2)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstCmp}},
			[]Word{WordI64(2)},
			ErrorKindStackUnderflow,
		},
		// jmp
		{
			[]InstDef{{Kind: InstJmp}},
			[]Word{},
			ErrorKindOk,
		},
		// jmp zero
		{
			[]InstDef{{Kind: InstJmpZero}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpZero}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// jmp not zero
		{
			[]InstDef{{Kind: InstJmpNotZero}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpNotZero}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// jmp greather
		{
			[]InstDef{{Kind: InstJmpGreater}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpGreater}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// jmp less
		{
			[]InstDef{{Kind: InstJmpLess}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpLess}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// jmp greather equal
		{
			[]InstDef{{Kind: InstJmpGreaterEqual}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpGreaterEqual}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// jmp less equal
		{
			[]InstDef{{Kind: InstJmpLessEqual}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstJmpLessEqual}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// call
		{
			[]InstDef{{Kind: InstFunCall}},
			[]Word{},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstFunCall}},
			make([]Word, CoppervmStackCapacity),
			ErrorKindStackOverflow,
		},
		// ret
		{
			[]InstDef{{Kind: InstFunReturn}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstFunReturn}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// mem read
		{
			[]InstDef{{Kind: InstMemRead}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMemRead}},
			[]Word{WordU64(uint64(CoppervmMemoryCapacity))},
			ErrorKindIllegalMemoryAccess,
		},
		{
			[]InstDef{{Kind: InstMemRead}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// mem write
		{
			[]InstDef{{Kind: InstMemRead}},
			[]Word{WordU64(1), WordU64(0)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstMemRead}},
			[]Word{WordU64(0), WordU64(uint64(CoppervmMemoryCapacity))},
			ErrorKindIllegalMemoryAccess,
		},
		{
			[]InstDef{{Kind: InstMemRead}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// print
		{
			[]InstDef{{Kind: InstPrint}},
			[]Word{WordU64(1)},
			ErrorKindOk,
		},
		{
			[]InstDef{{Kind: InstPrint}},
			[]Word{},
			ErrorKindStackUnderflow,
		},
		// invalid
		{
			[]InstDef{{Kind: InstCount}},
			[]Word{},
			ErrorKindInvalidInstruction,
		},
	}

	for _, test := range tests {
		vm.Ip = 0
		vm.Program = test.prog
		copy(vm.Stack[:], test.stack)
		vm.StackSize = int64(len(test.stack))

		if err := vm.ExecuteInstruction(); err.Kind != test.err {
			t.Errorf("expecting '%s' but got '%s'", test.err, err.Kind)
		}
	}
}
