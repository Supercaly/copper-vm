#! /bin/sh

set -e

bins=$(find examples/bin -maxdepth 1 -name "*.vm" -type f)
for binary in $bins; do
    file_name=$(basename $binary)
    file_name=${file_name%.vm}

    echo "Run '$binary'"
    output=$(./build/emulator $binary)

    value=$(cat "examples/test/$file_name.txt")
    if [ "$value" != "$output" ]; then
        echo "'$binary' produced a different output from what expected"
        exit 1
    fi
done