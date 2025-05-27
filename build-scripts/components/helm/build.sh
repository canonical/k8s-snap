#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

export GOTOOLCHAIN=local
export GOEXPERIMENT=opensslcrypto #,allowcrytofallback
export CGO_ENABLED=1
make VERSION="${VERSION}" TAGS="goexperiment.opensslcrypto,linux,cgo"
cp bin/helm "${INSTALL}/helm"
