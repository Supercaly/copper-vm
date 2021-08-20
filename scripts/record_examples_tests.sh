#! /bin/sh

set -e

mkdir -p examples/test

bins=$(find examples/bin -maxdepth 1 -name "*.vm" -type f)
for binary in $bins; do
    file_name=$(basename $binary)
    file_name=${file_name%.vm}

    echo "Record '$binary'"
    program_output=$(build/emulator $binary)
    echo "$program_output" > "examples/test/$file_name.txt"
done