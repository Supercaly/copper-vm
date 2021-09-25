#! /bin/sh

set -e

CMD_DIR="cmd"

go install "$CMD_DIR/casm/casm.go"
go install "$CMD_DIR/deasm/deasm.go"
go install "$CMD_DIR/emulator/emulator.go"
go install "$CMD_DIR/copperdb/copperdb.go"

GOPATH=$(go env GOPATH)

cp -Rf "./stdlib" "$GOPATH"