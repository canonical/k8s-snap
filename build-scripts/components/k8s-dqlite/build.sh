#!/bin/bash

INSTALL="${1}/bin"

## Use built dqlite dependencies (if any)
if [ -d "${SNAPCRAFT_STAGE}/dynamic-dqlite-deps" ]; then
  export DQLITE_BUILD_SCRIPTS_DIR="${SNAPCRAFT_STAGE}/dynamic-dqlite-deps"
fi

export GOTOOLCHAIN=local
# Only needed for Go < 1.27
if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
  export GOEXPERIMENT=opensslcrypto
fi
export CGO_ENABLED=1
export TAGS="libsqlite3,ms_tls13kdf"
make dynamic -j

mkdir -p "${INSTALL}"
for binary in k8s-dqlite dqlite; do
  cp -P "bin/dynamic/${binary}" "${INSTALL}/${binary}"
done
