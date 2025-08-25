#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

export GOTOOLCHAIN=local

export TAGS=""

if [ "${FLAVOR}" = "fips" ]; then
  export CGO_ENABLED=1
  export GOEXPERIMENT=opensslcrypto
  export TAGS="linux,cgo"
fi

make VERSION="${VERSION}" TAGS="${TAGS}"
cp bin/helm "${INSTALL}/helm"
