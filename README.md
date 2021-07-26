# Copper VM

A simple **Virtual Machine** with his own byte-code.

## Executables

The Copper VM ecosystem is composed by this executables:

* **casm** Assembler for the VM custom byte-code 
* **deasm** Disassembler for the VM byte-code
* **emulator** VM emulator that runs any binary program

## Quick Start

To execute one of the programs you can use the scripts on linux or windows:

```console
$ ./scripts/linux/casm.sh -o output.vm input.copper
$ ./scripts/linux/deasm.sh input.vm 
$ ./scripts/linux/emulator.sh input.vm
```

```console
$> .\scripts\windows\casm.bat -o output.vm input.copper
$> .\scripts\windows\deasm.bat input.vm 
$> .\scripts\windows\emulator.bat input.vm
```
