#!/bin/bash -x

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

export KUBE_GIT_VERSION_FILE="${PWD}/.version.sh"

for app in kubernetes; do
  # We are setting allowcryptofallback here. Normally this is not recommended because the user could
  # be led to think their system is compliant when the binaries have actually fallen back to non-FIPS compliant crypto.
  # We do this because we want to have a single branch/track for FIPS and non-FIPS deployments. In this scenario,
  # we don't have a choice to allow fallback.
  make GOEXPERIMENT=opensslcrypto WHAT="cmd/${app}" KUBE_CGO_OVERRIDES="${app}" GOFLAGS="-tags=providerless,goexperiment.systemcrypto,linux,cgo,allowcryptofallback"
  cp _output/bin/"${app}" "${INSTALL}/${app}"
done

for app in kubectl kubelet kube-proxy kube-controller-manager kube-scheduler kube-apiserver; do
  ln -sf ./kubernetes "${INSTALL}/${app}"
done
