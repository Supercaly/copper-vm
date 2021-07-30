#! /bin/sh

set -e

mkdir -p examples/bin

examples=$(find examples/ -name "*.copper" -type f)
for e in $examples; do
    name=$(basename $e)
    name=${name%.copper}
    ./scripts/linux/casm.sh -o "examples/bin/$name.vm" $e -I examples/
done