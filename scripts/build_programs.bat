@echo off

set BUILD_DIR=build
set CMD_DIR=cmd

if not exist "%BUILD_DIR%" mkdir %BUILD_DIR%

go build -o "%BUILD_DIR%/casm.exe" "%CMD_DIR%/casm/casm.go"
go build -o "%BUILD_DIR%/deasm.exe" "%CMD_DIR%/deasm/deasm.go"
go build -o "%BUILD_DIR%/emulator.exe" "%CMD_DIR%/emulator/emulator.go"