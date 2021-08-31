#! /bin/sh

set -e

BUILD_DIR=build
CMD_DIR=cmd

case "$1" in
"") BUILD_TAGS="release" ;;
"debug"|"release") BUILD_TAGS="$1" ;;
*)
    echo "$0 [debug|release]"
    exit 1
    ;;
esac

mkdir -p $BUILD_DIR

go build -tags "$BUILD_TAGS" -o $BUILD_DIR/casm $CMD_DIR/casm/casm.go
go build -tags "$BUILD_TAGS" -o $BUILD_DIR/deasm $CMD_DIR/deasm/deasm.go
go build -tags "$BUILD_TAGS" -o $BUILD_DIR/emulator $CMD_DIR/emulator/emulator.go
go build -tags "$BUILD_TAGS" -o $BUILD_DIR/copperdb $CMD_DIR/copperdb/copperdb.go

echo "programs built in $BUILD_TAGS mode"
