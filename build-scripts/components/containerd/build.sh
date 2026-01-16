#!/bin/bash

# Store the current Go snap revision
INITIAL_GO_REVISION=$(snap list go | grep -E '^go\s' | awk '{print $3}')
echo "Current Go snap revision: ${INITIAL_GO_REVISION}"

# Refresh to go fips stable channel
maj_min=$(awk '/^go /{print $2}' go.mod | cut -d. -f1,2)
echo "Refreshing to go ${maj_min}-fips/stable channel..."
snap refresh go --channel=${maj_min}-fips/stable

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

VERSION="${2}"
REVISION=$(git rev-parse HEAD)

sed -i "s,^VERSION.*$,VERSION=${VERSION}," Makefile
sed -i "s,^REVISION.*$,REVISION=${REVISION}," Makefile

# -static is hardcoded in the SHIM_GO_LDFLAGS Makefile variable, so we need to remove it to build dynamically linked binaries
# See https://github.com/containerd/containerd/blob/442cb34bda9a6a0fed82a2ca7cade05c5c749582/Makefile#L105
sed -i 's/-extldflags "-static"//' Makefile

export GOTOOLCHAIN=local
export GOEXPERIMENT=opensslcrypto
export CGO_ENABLED=1
export GO_BUILDTAGS="linux cgo ms_tls13kdf"
export SHIM_CGO_ENABLED=1
export SHIM_GO_BUILDTAGS="linux cgo ms_tls13kdf"

for bin in containerd ctr containerd-shim-runc-v1 containerd-shim-runc-v2; do
  make "bin/${bin}"
  cp "bin/${bin}" "${INSTALL}/${bin}"
done

# Restore the initial Go snap revision
echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"
snap revert go --revision="${INITIAL_GO_REVISION}"
