#!/bin/bash

set -ex

DIR=`realpath $(dirname "${0}")`
PRIME_DIRECTORY="${CRAFT_PRIME:-${DIR}/.prime}"

BASE="${1}"
BIN="${2}"
ARCH_TRIPLET="${3}"

# Note(ben): patchelf messes up the segments when patching the binary
# due to https://github.com/NixOS/patchelf/issues/446.
# Instead, we use a custom script based on LIEF to manually patch the rpath and interpreter.
echo "Patching ELF file: $BIN with LIEF"
# Only install Python deps on supported arches
pip3 install -r $DIR/hack/requirements.txt

if [ "$ARCH_TRIPLET" = "x86_64-linux-gnu" ]; then
python3 $DIR/hack/patchelf.py "$BIN" \
    --set-rpath /snap/$BASE/current/lib/x86_64-linux-gnu/ \
    --set-interpreter /snap/$BASE/current/lib64/ld-linux-x86-64.so.2
else
python3 $DIR/hack/patchelf.py "$BIN" \
    --set-rpath /snap/$BASE/current/lib/aarch64-linux-gnu/ \
    --set-interpreter /snap/$BASE/current/lib/ld-linux-aarch64.so.1
fi
echo "==> LIEF patching $BIN complete"
mkdir -p "$PRIME_DIRECTORY/bin"
cp "$BIN" "$PRIME_DIRECTORY/bin/"
