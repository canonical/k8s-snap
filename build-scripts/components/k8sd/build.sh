#!/bin/bash

VERSION="${2}"
INSTALL="${1}"

mkdir -p "${INSTALL}"

# Store the current Go snap revision
INITIAL_GO_REVISION=$(snap list go | grep -E '^go\s' | awk '{print $3}')
echo "Current Go snap revision: ${INITIAL_GO_REVISION}"

# k8sd (tracked at main, see components/k8sd/version) requires go >= 1.27.1,
# ahead of the go/1.26-fips/stable build-snap pinned for the rest of the
# snap's parts. go/1.27-fips/stable is not published yet, so use "candidate"
# until it is; switch this to "stable" once available.
echo "Refreshing to go 1.27-fips/candidate channel..."
snap refresh go --channel=1.27-fips/candidate

export CGO_ENABLED=1
export GOTOOLCHAIN=local

make dynamic -j

mkdir -p "${INSTALL}/bin"
mkdir -p "${INSTALL}/lib"
for binary in k8s k8sd k8s-apiserver-proxy; do
cp -P "bin/dynamic/${binary}" "${INSTALL}/bin/${binary}"
done

# k8sd builds the dqlite shared libraries that we need to include.
cp -P bin/dynamic/lib/*.so* "${INSTALL}/lib/"

LD_LIBRARY_PATH="${INSTALL}/lib" "${INSTALL}/bin/k8s" list-images > "${INSTALL}/images.txt"

# Restore the initial Go snap revision
echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"

# Snap revert fails if the revision is already the current one, so we ignore errors
snap revert go --revision="${INITIAL_GO_REVISION}" || true
