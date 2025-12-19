#!/bin/bash

# Store the current Go snap revision
INITIAL_GO_REVISION=$(snap list go | grep -E '^go\s' | awk '{print $3}')
echo "Current Go snap revision: ${INITIAL_GO_REVISION}"

# Extract Go version from go.mod (normalize to major.minor for snap channel)
GO_VERSION=$(grep -E '^go ' go.mod | awk '{print $2}' | cut -d. -f1-2)
if [ -z "${GO_VERSION}" ]; then
  echo "Error: Could not extract Go version from go.mod"
  exit 1
fi
echo "Go version from go.mod: ${GO_VERSION}"

# Refresh to go ${GO_VERSION}-fips/stable channel
echo "Refreshing to go ${GO_VERSION}-fips/stable channel..."
snap refresh go --channel=${GO_VERSION}-fips/stable

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
for bin in ctr containerd-shim-runc-v2; do
  export STATIC=1
  export CGO_ENABLED=0
  export GO_BUILDTAGS=
  export SHIM_CGO_ENABLED=0
  export SHIM_GO_BUILDTAGS=
  export GOEXPERIMENT=

  make "bin/${bin}"
  cp "bin/${bin}" "${INSTALL}/${bin}"
done

# Restore the initial Go snap revision
echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"
snap revert go --revision="${INITIAL_GO_REVISION}"
