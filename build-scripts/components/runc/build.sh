#!/bin/bash

# Check if Go is installed via snap
if snap list go >/dev/null 2>&1; then
    # Store the current Go snap revision
    INITIAL_GO_REVISION=$(snap list go | grep -E '^go\s' | awk '{print $3}')
    echo "Current Go snap revision: ${INITIAL_GO_REVISION}"
    
    # Refresh to go 1.23-fips/stable channel
    echo "Refreshing to go 1.23-fips/stable channel..."
    snap refresh go --channel=1.23-fips/stable
fi

VERSION="${2}"

export INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

# Ensure `runc --version` prints the right commit hash from upstream
export COMMIT="$(git describe --long --always "${VERSION}")"

make BUILDTAGS="seccomp apparmor" EXTRA_LDFLAGS="-s -w" static
cp runc "${INSTALL}/runc"

# Restore Go state
if snap list go >/dev/null 2>&1; then
    # Restore the initial Go snap revision
    echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"
    snap revert go --revision="${INITIAL_GO_REVISION}"
fi
