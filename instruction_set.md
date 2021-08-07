# Copper Instruction Set

## Basic Instructions

| Mnemonic | Operand | Description |
| --- | :---: | --- |
| noop | - | does nothing |
| push | value | push a value to the stack.<br/>The value could be: unsigned integer (15), signed integer (-3), floating point (2.6, -15.3)|
| swap | n | swap the stack top with the n-th element |
| dup | - | duplicate the stack top |
| drop | - | pop the stack top |
| halt | - | stops the virtual machine execution (it's the same as calling syscall exit with code 0) |

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
| cmp | - | compares first to elements on the stack, consumes them and push on stack top: 0 if a = b, 1 if a > b, -1 if b > a | 
| jmp | location | jump unconditionally to location | 
| jz | location | jump to location if stack top is zero, the top is consumed | 
| jnz | location | jump to location if stack top is not zero, the top is consumed | 
| jg | location | jump to location if stack top is greater then zero, the top is consumed | 
| jl | location | jump to location if stack top is less then zero, the top is consumed | 
| jge | location | jump to location if stack top is greater or equal then zero, the top is consumed | 
| jle | location | jump to location if stack top is less or equal then zero, the top is consumed | 

## Functions
| Mnemonic | Operand | Description |
| --- | :---: | --- |
| call | location | moves the ip to given location; it's like jmp, but before moving push the current ip to the stack so ret can go back |
| ret | - | set the ip to the stack top and pop it |

## Memory access
| Mnemonic | Operand | Description |
| --- | :---: | --- |
| read | - | reads a byte from the memory.<br/> The memory address to read is the stack top, that is replaced with the byte read after the instruction is executed |
| write | - | writes a byte to the memory.<br/> The value to write and his destination are the first two elements on the stack; the values are consumed after the instruction is executed. |

## System Calls
To interact with the underlying system you can use the `syscall` instruction which has one of the following as operands:

| Op | Name | Arg0 | Arg1 | Arg2 | Description |
| --- | :---: | :---: | :---: | :---: | --- |
| 0 | read | fd | buffer | count | reads count bytes form fd and put them to buffer.<br/>At the end pushes on stack top the number of bytes read of -1 in case of error |
| 1 | write | fd | buffer | count | writes count bytes from buffer to fd.<br/>At the end pushes on stack top the number of bytes written of -1 in case of error |
| 2 | open | file_name | - | - | opens a file with given name and returns his file descriptor or -1 on case of error | 
| 3 | close | fd | - | - | close file descriptor fd. At the end pushes on stack top 0 on success or -1 in case of error | 
| 4 | seek | fd | offset | whence | set the offset of the next read/write operation to offset, interpreted according to whence: 0 relative to file origin, 1 relative to current offset, 2 relative to file end. At the end pushes on stack top the new offset or -1 in case of error | 
| 5 | exit | status_code | - | - | stops the virtual machine execution with status_code |

## Debug
| Mnemonic | Operand | Description |
| --- | :---: | ---|
| print | - | debug print the stack top consuming it |