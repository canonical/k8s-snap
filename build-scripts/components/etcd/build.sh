#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

export GOTOOLCHAIN=local
# (berkayoz): We're hitting https://github.com/NixOS/patchelf/issues/146 with patchelf
export CGO_ENABLED=0
# # export GOEXPERIMENT=opensslcrypto
make build
cp bin/* "${INSTALL}/"
