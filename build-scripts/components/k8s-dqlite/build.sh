#!/bin/bash

INSTALL="${1}/bin"

## Use built dqlite dependencies (if any)
if [ -d "${SNAPCRAFT_STAGE}/static-dqlite-deps" ]; then
  export DQLITE_BUILD_SCRIPTS_DIR="${SNAPCRAFT_STAGE}/static-dqlite-deps"
fi

make static -j

mkdir -p "${INSTALL}"
for binary in k8s-dqlite dqlite; do
  cp -P "bin/static/${binary}" "${INSTALL}/${binary}"
done
