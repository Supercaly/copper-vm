# Copper Instruction Set

## Basic Instructions

| Mnemonic | Operand | Description |
| --- | :---: | --- |
| noop | - | does nothing |
| push | value | push a value to the stack |
| swap | n | swap the stack top with the n-th element |
| dup | - | duplicate the stack top |
| halt | - | stops the virtual machine execution |

## Integer arithmetics

| Mnemonic | Operand | Description |
| --- | :---: | --- |
| add | - | integer addition of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 
| sub | - | integer subtract of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 
| mul | - | integer multiply (unsigned) of first two elements on the stack, the result is pushed on stack top and the elements are consumed |
| imul | - | integer multiply (signed) of first two elements on the stack, the result is pushed on stack top and the elements are consumed |

## Floating point arithmetics

| Mnemonic | Operand | Description |
| --- | :---: | --- |
| fadd | - | floating point addition of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 
| fsub | - | floating point subtract of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 
| fmul | - | floating point multiply of first two elements on the stack, the result is pushed on stack top and the elements are consumed | 

## Flow control
| Mnemonic | Operand | Description |
| --- | :---: | --- |
| jmp | location | jump unconditionally to location | 
| jnz | location | jump to location if stack top is not zero, the top is consumed | 