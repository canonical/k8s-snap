#!/bin/bash

VERSION="${2}"

export INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

# Ensure `runc --version` prints the right commit hash from upstream
export COMMIT="$(git describe --long --always "${VERSION}")"

# runc is statically linked, and the Go 1.25+ OpenSSL backend dlopens libcrypto, which a static binary cannot do
export GOTOOLCHAIN=local
export MS_GO_NOSYSTEMCRYPTO=1
# Only needed for Go < 1.27; Go 1.27 removed this experiment and rejects it
if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
	export GOEXPERIMENT=nosystemcrypto
fi

make BUILDTAGS="seccomp apparmor" EXTRA_LDFLAGS="-s -w" static
cp runc "${INSTALL}/runc"
