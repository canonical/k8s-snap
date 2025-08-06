#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

export GOTOOLCHAIN=local
export CGO_ENABLED=1
#export GOEXPERIMENT=opensslcrypto
export GOFLAGS="-tags=linux,cgo"

make build
cp bin/* "${INSTALL}/"
