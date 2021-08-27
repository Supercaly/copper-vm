#! /bin/sh

set -e

mkdir -p examples/bin

examples=$(find examples/ -maxdepth 1 -name "*.casm" -type f)
for e in $examples; do
    name=$(basename $e)
    name=${name%.casm}
    ./build/casm -o "examples/bin/$name.copper" $e -I stdlib/
done