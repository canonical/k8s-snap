#!/bin/bash

VERSION="${2}"
INSTALL="${1}"

mkdir -p "${INSTALL}"

export CGO_ENABLED=1
export GOTOOLCHAIN=local
# Go 1.27+ enables systemcrypto (opensslcrypto) by default and treats an
# explicit GOEXPERIMENT=opensslcrypto as a hard error; only set it on
# older toolchains (see runc/build.sh for the inverse case).
go_minor=$(go env GOVERSION | sed -E 's/^go[0-9]+\.([0-9]+).*/\1/')
if [ "${go_minor}" -lt 27 ]; then
  export GOEXPERIMENT=opensslcrypto
fi

make dynamic -j

mkdir -p "${INSTALL}/bin"
mkdir -p "${INSTALL}/lib"
for binary in k8s k8sd k8s-apiserver-proxy; do
cp -P "bin/dynamic/${binary}" "${INSTALL}/bin/${binary}"
done

# k8sd builds the dqlite shared libraries that we need to include.
cp -P bin/dynamic/lib/*.so* "${INSTALL}/lib/"

LD_LIBRARY_PATH="${INSTALL}/lib" "${INSTALL}/bin/k8s" list-images > "${INSTALL}/images.txt"
