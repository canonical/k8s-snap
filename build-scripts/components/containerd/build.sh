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
