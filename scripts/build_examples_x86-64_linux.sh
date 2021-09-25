#! /bin/sh

set -e

mkdir -p examples/bin

examples=$(find examples/ -maxdepth 1 -name "*.casm" -type f)
for e in $examples; do
    name=$(basename $e)
    name=${name%.casm}
    ./build/casm -t x86-64 -o "examples/bin/$name.asm" $e -I stdlib/
    nasm -felf64 "examples/bin/$name.asm"
    ld -o "examples/bin/$name" "examples/bin/$name.o"
done