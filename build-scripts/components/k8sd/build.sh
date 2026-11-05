#!/bin/bash

# Store the current Go snap revision
INITIAL_GO_REVISION=$(snap list go | grep -E '^go\s' | awk '{print $3}')
echo "Current Go snap revision: ${INITIAL_GO_REVISION}"

# Restore the initial Go snap revision on exit, even if the build fails.
restore_go() {
  echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"
  # Snap revert fails if the revision is already the current one, so we ignore errors
  snap revert go --revision="${INITIAL_GO_REVISION}" || true
}
trap restore_go EXIT

# k8sd is built with the Go version required by its go.mod, which can be
# newer than the shared toolchain used by the other components.
maj_min=$(awk '/^go /{print $2}' go.mod | cut -d. -f1,2)
echo "Refreshing to go ${maj_min}-fips/stable channel..."
snap refresh go --channel="${maj_min}"-fips/stable

VERSION="${2}"
INSTALL="${1}"

mkdir -p "${INSTALL}"

export CGO_ENABLED=1
export GOTOOLCHAIN=local
# Only needed for Go < 1.27
if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
	export GOEXPERIMENT=opensslcrypto
fi

make dynamic -j

mkdir -p "${INSTALL}/bin"
mkdir -p "${INSTALL}/lib"
for binary in k8s k8sd k8s-apiserver-proxy; do
cp -P "bin/dynamic/${binary}" "${INSTALL}/bin/${binary}"
done

# k8sd builds the dqlite shared libraries that we need to include.
cp -P bin/dynamic/lib/*.so* "${INSTALL}/lib/"

LD_LIBRARY_PATH="${INSTALL}/lib" "${INSTALL}/bin/k8s" list-images > "${INSTALL}/images.txt"
