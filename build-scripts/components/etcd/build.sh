#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

export GOTOOLCHAIN=local

if [ "${FLAVOR}" = "fips" ]; then
  export CGO_ENABLED=1
  export GOEXPERIMENT=opensslcrypto
  export GOFLAGS="-tags=linux,cgo"
fi

make build
cp bin/* "${INSTALL}/"
