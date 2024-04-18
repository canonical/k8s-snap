#!/bin/bash

VERSION="${2}"

export INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

# Pin runc to Go 1.21: https://github.com/opencontainers/runc/issues/4233
if ! which go_121; then
  snap download go --channel 1.21 --basename go
  snap install ./go.snap --classic --dangerous --name go_121
fi
export GO=go_121

# Ensure `runc --version` prints the right commit hash from upstream
export COMMIT="$(git describe --long --always "${VERSION}")"

make BUILDTAGS="seccomp apparmor" EXTRA_LDFLAGS="-s -w" static
cp runc "${INSTALL}/runc"
