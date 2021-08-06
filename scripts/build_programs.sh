#! /bin/sh

set -e

BUILD_DIR=build
CMD_DIR=cmd

mkdir -p $BUILD_DIR

go build -o $BUILD_DIR/casm $CMD_DIR/casm/casm.go
go build -o $BUILD_DIR/deasm $CMD_DIR/deasm/deasm.go
go build -o $BUILD_DIR/emulator $CMD_DIR/emulator/emulator.go