#!/bin/bash -x

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

export KUBE_GIT_VERSION_FILE="${PWD}/.version.sh"

export GOTOOLCHAIN=local

export TAGS="providerless"

if [ "${FLAVOR}" = "fips" ]; then
  export GOEXPERIMENT=opensslcrypto
  export CGO_ENABLED=1
  export TAGS="${TAGS},linux,cgo"
fi

for app in kubernetes; do
  make WHAT="cmd/${app}" KUBE_STATIC_OVERRIDES="${app}" KUBE_CGO_OVERRIDES="${app}" GOFLAGS="-tags=${TAGS}"
  cp _output/bin/"${app}" "${INSTALL}/${app}"
done

for app in kubectl kubelet kube-proxy kube-controller-manager kube-scheduler kube-apiserver; do
  ln -sf ./kubernetes "${INSTALL}/${app}"
done
