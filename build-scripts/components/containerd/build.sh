#!/bin/bash

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

VERSION="${2}"
REVISION=$(git rev-parse HEAD)

sed -i "s,^VERSION.*$,VERSION=${VERSION}," Makefile
sed -i "s,^REVISION.*$,REVISION=${REVISION}," Makefile

export GOTOOLCHAIN=local
# Only needed for Go < 1.27
if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
	export GOEXPERIMENT=opensslcrypto
fi
export CGO_ENABLED=1
export GO_BUILDTAGS="linux cgo ms_tls13kdf"
export SHIM_CGO_ENABLED=1
export SHIM_GO_BUILDTAGS="linux cgo ms_tls13kdf"
for bin in containerd ctr containerd-shim-runc-v2; do
  if [ "${bin}" = "ctr" ]; then
    # ctr only uses standard library cgo, so Go links it internally and
    # places the section headers at the start of the file. The LIEF-based
    # ELF patcher (build-scripts/hack/patchelf.py) leaves those headers
    # stale when it rewrites the program headers, corrupting the binary.
    # Link externally so the headers end up at the end of the file.
    EXTRA_LDFLAGS="-linkmode=external" make "bin/${bin}"
  else
    make "bin/${bin}"
  fi
  cp "bin/${bin}" "${INSTALL}/${bin}"
done
