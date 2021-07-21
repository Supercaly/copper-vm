# Copper VM

A simple **Virtual Machine** with his own byte-code.

## Executables

The Copper VM ecosystem is composed by this executables:

* **casm** Assembler for the Copper VM custom byte-code 
* **deasm** Disassembler for te VM byte-code
* **emulator** VM emulator that runs any binary program

## Quick Start

To execute one of the main programs simply use go and run:

```console
$ go run cmd/casm/casm.go -o output.vm input.copper
```

```console
$ go run cmd/deasm/deasm.go input.vm 
```

```console
$ go run cmd/emulator/emulator.go input.vm
```
