#!/bin/bash

INSTALL="${1}/bin"

## Use built dqlite dependencies (if any)
if [ -d "${SNAPCRAFT_STAGE}/dynamic-dqlite-deps" ]; then
  export DQLITE_BUILD_SCRIPTS_DIR="${SNAPCRAFT_STAGE}/dynamic-dqlite-deps"
fi

export GOEXPERIMENT=opensslcrypto
make dynamic -j

mkdir -p "${INSTALL}"
for binary in k8s-dqlite dqlite; do
  cp -P "bin/dynamic/${binary}" "${INSTALL}/${binary}"
done
