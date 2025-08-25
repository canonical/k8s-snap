#!/bin/bash

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

VERSION="${2}"
REVISION=$(git rev-parse HEAD)

sed -i "s,^VERSION.*$,VERSION=${VERSION}," Makefile
sed -i "s,^REVISION.*$,REVISION=${REVISION}," Makefile

export GOTOOLCHAIN=local

export STATIC=1
for bin in ctr containerd-shim containerd-shim-runc-v1 containerd-shim-runc-v2; do
  make "bin/${bin}"
  cp "bin/${bin}" "${INSTALL}/${bin}"
done

if [ "${FLAVOR}" = "fips" ]; then
  # Shims can be built statically as they do not contain any crypto functions
  export STATIC=0
  export GOEXPERIMENT=opensslcrypto
  export CGO_ENABLED=1
  export GO_BUILDTAGS="linux cgo"
fi

for bin in containerd; do
  make "bin/${bin}"
  cp "bin/${bin}" "${INSTALL}/${bin}"
done
