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

# Go 1.25+ enables the OpenSSL backend by default, and it dlopens libcrypto,
# which a statically linked binary cannot do
export GOTOOLCHAIN=local
export GOEXPERIMENT=nosystemcrypto

# libpathrs is not in the core22 archive as a C library, so build it from
# source the same way upstream's own release builds do
RUST_BIN="$(ls -d /usr/lib/rust-*/bin 2>/dev/null | sort -V | tail -1 || true)"
if [ -n "${RUST_BIN}" ]; then export PATH="${RUST_BIN}:${PATH}"; fi
LIBPATHRS_VERSION="0.2.5"
LIBPATHRS_DIR="$(mktemp -d)"
./script/build-libpathrs.sh "${LIBPATHRS_VERSION}" "${LIBPATHRS_DIR}"
export LD_LIBRARY_PATH="${LIBPATHRS_DIR}/lib"
export PKG_CONFIG_PATH="${LIBPATHRS_DIR}/lib/pkgconfig"

make BUILDTAGS="seccomp apparmor libpathrs" EXTRA_LDFLAGS="-s -w" static
cp runc "${INSTALL}/runc"

# Restore the initial Go snap revision
echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"

# Snap revert fails if the revision is already the current one, so we ignore errors
snap revert go --revision="${INITIAL_GO_REVISION}" || true
