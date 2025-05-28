#!/bin/bash

VERSION="${2}"

export INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

# Ensure `runc --version` prints the right commit hash from upstream
export COMMIT="$(git describe --long --always "${VERSION}")"

export GOTOOLCHAIN=local
export CGO_ENABLED=1
export GOEXPERIMENT=opensslcrypto
make EXTRA_BUILDTAGS="linux cgo apparmor" EXTRA_LDFLAGS="-s -w"
cp runc "${INSTALL}/runc"
