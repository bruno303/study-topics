#!/bin/bash

set -e

BIN_DIR="./bin"

echo "preparing..."
if [ -d "$BIN_DIR" ]; then
  rm -r $BIN_DIR
fi
mkdir $BIN_DIR

echo "compiling..."
go run toy-compiler.go
clang -g "$BIN_DIR/output.ll" -o "$BIN_DIR/output"

# old
# llc ./bin/output.ll -filetype=obj -O0 -g -o ./bin/output.o
# clang ./bin/output.o -o ./bin/output
