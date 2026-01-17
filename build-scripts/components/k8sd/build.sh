#!/bin/bash

VERSION="${2}"
INSTALL="${1}"

mkdir -p "${INSTALL}"

## Use built dqlite dependencies (if any)
if [ -d "${CRAFT_STAGE}/dynamic-dqlite-deps" ]; then
  export DQLITE_BUILD_SCRIPTS_DIR="${CRAFT_STAGE}/dynamic-dqlite-deps"
fi

export CGO_ENABLED=1
export GOTOOLCHAIN=local
export GOEXPERIMENT=opensslcrypto

make dynamic -j

mkdir -p "${INSTALL}/bin"
mkdir -p "${INSTALL}/lib"
for binary in k8s k8sd k8s-apiserver-proxy; do
cp -P "bin/dynamic/${binary}" "${INSTALL}/bin/${binary}"
done
cp -P bin/dynamic/lib/*.so* "${INSTALL}/lib/"

LD_LIBRARY_PATH="${DQLITE_BUILD_SCRIPTS_DIR}/.deps/dynamic/lib" "${INSTALL}/bin/k8s" list-images > "${INSTALL}/images.txt"
