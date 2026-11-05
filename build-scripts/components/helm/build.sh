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
make VERSION="${VERSION}" TAGS="linux,cgo,ms_tls13kdf"
cp bin/helm "${INSTALL}/helm"
