# Copper VM

A simple **Virtual Machine** with his own byte-code.

## Executables

The Copper VM ecosystem is composed by this executables:

* **casm** Assembler for the VM custom byte-code 
* **deasm** Disassembler for the VM byte-code
* **emulator** VM emulator that runs any binary program
* **copperdb** Debugger for the VM program

## Quick Start

Before executing the programs you need to build them; you can build them in a local `build` directory using:

```console
$ ./scripts/build_programs.sh
```

```console
$> .\scripts\build_programs.bat
```

or you can install them to the directory named by the `GOBIN` environment variable using:

```console
$ go install ./cmd/...
```