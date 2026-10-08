#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

export GOTOOLCHAIN=local
export CGO_ENABLED=1
# Go 1.27+ enables systemcrypto (opensslcrypto) by default and treats an
# explicit GOEXPERIMENT=opensslcrypto as a hard error; only set it on
# older toolchains (see runc/build.sh for the inverse case).
go_minor=$(go env GOVERSION | sed -E 's/^go[0-9]+\.([0-9]+).*/\1/')
if [ "${go_minor}" -lt 27 ]; then
  export GOEXPERIMENT=opensslcrypto
fi
export GOFLAGS="-tags=linux,cgo,ms_tls13kdf"

make build
cp bin/* "${INSTALL}/"
