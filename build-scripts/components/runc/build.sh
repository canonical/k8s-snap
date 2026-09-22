#!/bin/bash

# Store the current Go snap revision
INITIAL_GO_REVISION=$(snap list go | grep -E '^go\s' | awk '{print $3}')
echo "Current Go snap revision: ${INITIAL_GO_REVISION}"

# Refresh to go fips stable channel
maj_min=$(awk '/^go /{print $2}' go.mod | cut -d. -f1,2)
echo "Refreshing to go ${maj_min}-fips/stable channel..."
snap refresh go --channel=${maj_min}-fips/stable

VERSION="${2}"

export INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

# Ensure `runc --version` prints the right commit hash from upstream
export COMMIT="$(git describe --long --always "${VERSION}")"

# runc is statically linked, and the Go 1.25+ OpenSSL backend dlopens libcrypto, which a static binary cannot do
export GOTOOLCHAIN=local
export GOEXPERIMENT=nosystemcrypto

make BUILDTAGS="seccomp apparmor" EXTRA_LDFLAGS="-s -w" static
cp runc "${INSTALL}/runc"

# Restore the initial Go snap revision
echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"

# Snap revert fails if the revision is already the current one, so we ignore errors
snap revert go --revision="${INITIAL_GO_REVISION}" || true
