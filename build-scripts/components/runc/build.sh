#!/bin/bash

# Check if Go is installed via snap
if snap list go >/dev/null 2>&1; then
    # Store the current Go snap revision
    INITIAL_GO_REVISION=$(snap list go | grep -E '^go\s' | awk '{print $3}')
    echo "Current Go snap revision: ${INITIAL_GO_REVISION}"
    
    # Refresh to go 1.23-fips/stable channel
    echo "Refreshing to go 1.23-fips/stable channel..."
    snap refresh go --channel=1.23-fips/stable
    GO_INSTALLED_VIA_SNAP=true
else
    # Install Go from source
    # NOTE(Hue): This is a workaround for CAPI, which runs this script in a Dockerfile
    # This version comes from https://github.com/microsoft/go/releases and later on needs to be
    # reverted back to whatever we're using in the https://github.com/canonical/cluster-api-k8s/blob/main/templates/docker/install-go.sh
    echo "Go snap not found, installing Go from source..."
    GO_VERSION="1.23.12-1"
    wget https://aka.ms/golang/release/latest/go$GO_VERSION.linux-amd64.tar.gz
    rm -rf /usr/local/go || true
    tar -C /usr/local -xzvf go$GO_VERSION.linux-amd64.tar.gz
    rm go$GO_VERSION.linux-amd64.tar.gz
    GO_INSTALLED_VIA_SNAP=false
fi

VERSION="${2}"

export INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

# Ensure `runc --version` prints the right commit hash from upstream
export COMMIT="$(git describe --long --always "${VERSION}")"

make BUILDTAGS="seccomp apparmor" EXTRA_LDFLAGS="-s -w" static
cp runc "${INSTALL}/runc"

# Restore Go state
if [ "$GO_INSTALLED_VIA_SNAP" = true ]; then
    # Restore the initial Go snap revision
    echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"
    snap revert go --revision="${INITIAL_GO_REVISION}"
else
    # Clean up downloaded Go tarball
    GO_VERSION="1.24.4-1"
    echo "Restoring Go to $GO_VERSION"
    wget https://aka.ms/golang/release/latest/go$GO_VERSION.linux-amd64.tar.gz
    rm -rf /usr/local/go || true
    tar -C /usr/local -xzvf go$GO_VERSION.linux-amd64.tar.gz
    rm go$GO_VERSION.linux-amd64.tar.gz
fi
