#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

export GOTOOLCHAIN=local
export CGO_ENABLED=1
# Only needed for Go < 1.27
if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
	export GOEXPERIMENT=opensslcrypto
fi
export GOFLAGS="-tags=linux,cgo,ms_tls13kdf"

make build
cp bin/* "${INSTALL}/"
