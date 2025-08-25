#!/bin/bash

INSTALL="${1}/bin"

export GOTOOLCHAIN=local

export BUILD_STRATEGY="static"

if [ "${FLAVOR}" = "fips" ]; then
  export GOEXPERIMENT=opensslcrypto
  export CGO_ENABLED=1
  export BUILD_STRATEGY="dynamic"
fi

## Use built dqlite dependencies (if any)
if [ -d "${SNAPCRAFT_STAGE}/${BUILD_STRATEGY}-dqlite-deps" ]; then
  export DQLITE_BUILD_SCRIPTS_DIR="${SNAPCRAFT_STAGE}/${BUILD_STRATEGY}-dqlite-deps"
fi

make $BUILD_STRATEGY -j

mkdir -p "${INSTALL}"
for binary in k8s-dqlite dqlite; do
  cp -P "bin/${BUILD_STRATEGY}/${binary}" "${INSTALL}/${binary}"
done
