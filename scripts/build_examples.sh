#! /bin/sh

set -e

mkdir -p examples/bin

examples=$(find examples/ -maxdepth 1 -name "*.copper" -type f)
for e in $examples; do
    name=$(basename $e)
    name=${name%.copper}
    ./build/casm -o "examples/bin/$name.vm" $e -I stdlib/
done