@echo off

set CMD_DIR="cmd"

go install "%CMD_DIR%/casm/casm.go"
go install "%CMD_DIR%/deasm/deasm.go"
go install "%CMD_DIR%/emulator/emulator.go"
go install "%CMD_DIR%/copperdb/copperdb.go"

for /f %%i in ('go env GOPATH') do set GOPATH=%%i

xcopy.exe .\stdlib %GOPATH%\stdlib\ /E /Y /F