@echo off

set BUILD_DIR=build
set CMD_DIR=cmd

if "%1" == "" set BUILD_TAGS="release"
if "%1" == "debug" set BUILD_TAGS="%1"
if "%1" == "release" set BUILD_TAGS="%1"
if not "%1" == "" if not "%1" == "debug" if not "%1" == "release" echo "%~0 [debug|release]" & exit 0

if not exist "%BUILD_DIR%" mkdir %BUILD_DIR%

go build -tags "%BUILD_TAGS%" -o "%BUILD_DIR%/casm.exe" "%CMD_DIR%/casm/casm.go"
go build -tags "%BUILD_TAGS%" -o "%BUILD_DIR%/deasm.exe" "%CMD_DIR%/deasm/deasm.go"
go build -tags "%BUILD_TAGS%" -o "%BUILD_DIR%/emulator.exe" "%CMD_DIR%/emulator/emulator.go"
go build -tags "%BUILD_TAGS%" -o "%BUILD_DIR%/copperdb.exe" "%CMD_DIR%/copperdb/copperdb.go"

echo programs built in %BUILD_TAGS% mode