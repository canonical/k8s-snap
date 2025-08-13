#!/bin/bash

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

VERSION="${2}"
REVISION=$(git rev-parse HEAD)

sed -i "s,^VERSION.*$,VERSION=${VERSION}," Makefile
sed -i "s,^REVISION.*$,REVISION=${REVISION}," Makefile

export GOTOOLCHAIN=local
export GOEXPERIMENT=opensslcrypto
export CGO_ENABLED=1
export GO_BUILDTAGS="linux cgo ms_tls13kdf"
for bin in containerd; do
  make "bin/${bin}"
  cp "bin/${bin}" "${INSTALL}/${bin}"
done

# Shims can be built statically as they do not contain any crypto functions
for bin in ctr containerd-shim containerd-shim-runc-v1 containerd-shim-runc-v2; do
  export STATIC=1
  export CGO_ENABLED=0
  export GO_BUILDTAGS=
  export SHIM_CGO_ENABLED=0
  export SHIM_GO_BUILDTAGS=
  export GOEXPERIMENT=

  make "bin/${bin}"
  cp "bin/${bin}" "${INSTALL}/${bin}"
done
