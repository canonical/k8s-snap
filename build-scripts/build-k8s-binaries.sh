#!/bin/bash

INSTALL="${SNAPCRAFT_PART_INSTALL}"

export DQLITE_BUILD_SCRIPTS_DIR="${CRAFT_STAGE}/static-dqlite-deps"

export GOTOOLCHAIN=local

if [ "${FLAVOR}" = "fips" ]; then
  export GOEXPERIMENT=opensslcrypto
  export CGO_ENABLED=1
  export EXTRA_LDFLAGS="-X 'github.com/canonical/k8s/pkg/config.buildFlavor=fips'"

  make dynamic -j

  mkdir -p "${INSTALL}/bin"
  mkdir -p "${INSTALL}/lib"
  for binary in k8s k8sd k8s-apiserver-proxy; do
  cp -P "bin/dynamic/${binary}" "${INSTALL}/bin/${binary}"
  done
  cp -P bin/dynamic/lib/*.so* "${INSTALL}/lib/"

  LD_LIBRARY_PATH="${DQLITE_BUILD_SCRIPTS_DIR}/.deps/dynamic/lib" "${INSTALL}/bin/k8s" list-images > "${INSTALL}/images.txt"
else
  make static -j

  mkdir -p "${INSTALL}/bin"
  for binary in k8s k8sd k8s-apiserver-proxy; do
  cp -P "bin/static/${binary}" "${INSTALL}/bin/${binary}"
  done
  ./bin/static/k8s list-images > "${INSTALL}/images.txt"
fi
