# Copper Instruction Set

## Basic Instructions

| Mnemonic | Operand | Description |
| --- | :---: | --- |
| noop | - | does nothing |
| push | value | push a value to the stack |
| swap | n | swap the stack top with the n-th element |
| dup | - | duplicate the stack top |
| drop | - | pop the stack top |
| halt | - | stops the virtual machine execution |

## Integer arithmetics

| Mnemonic | Operand | Description |
| --- | :---: | --- |
| add | - | integer addition of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 
| sub | - | integer subtract of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 
| mul | - | integer multiplication (unsigned) of first two elements on the stack, the result is pushed on stack top and the elements are consumed |
| imul | - | integer multiplication (signed) of first two elements on the stack, the result is pushed on stack top and the elements are consumed |
| div | - | integer division (unsigned) of first two elements on the stack, the result is pushed on stack top and the elements are consumed |
| idiv | - | integer division (signed) of first two elements on the stack, the result is pushed on stack top and the elements are consumed |
| mod | - | integer modulo (unsigned) of first two elements on the stack, the result is pushed on stack top and the elements are consumed |
| imod | - | integer modulo (signed) of first two elements on the stack, the result is pushed on stack top and the elements are consumed |

## Floating point arithmetics

| Mnemonic | Operand | Description |
| --- | :---: | --- |
| fadd | - | floating point addition of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 
| fsub | - | floating point subtract of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 
| fmul | - | floating point multiplication of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 
| fdiv | - | floating point division of first two elements on the stack, the result is pushed on stack top and the elements are consumed |

## Flow control
| Mnemonic | Operand | Description |
| --- | :---: | --- |
| jmp | location | jump unconditionally to location | 
| jnz | location | jump to location if stack top is not zero, the top is consumed | 

## Functions
| Mnemonic | Operand | Description |
| --- | :---| ---|
| call | location | moves the ip to given location; it's like jmp, but before moving push the current ip to the stack so ret can go back |
| ret | - | set the ip to the stack top |