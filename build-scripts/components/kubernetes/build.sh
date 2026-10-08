#!/bin/bash -x

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

export KUBE_GIT_VERSION_FILE="${PWD}/.version.sh"

for app in kubernetes; do
  export GOTOOLCHAIN=local
  # Go 1.27+ enables systemcrypto (opensslcrypto) by default and treats an
  # explicit GOEXPERIMENT=opensslcrypto as a hard error; only set it on
  # older toolchains (see runc/build.sh for the inverse case).
  go_minor=$(go env GOVERSION | sed -E 's/^go[0-9]+\.([0-9]+).*/\1/')
  if [ "${go_minor}" -lt 27 ]; then
    export GOEXPERIMENT=opensslcrypto
  fi
  export CGO_ENABLED=1
  make WHAT="cmd/${app}" KUBE_CGO_OVERRIDES="${app}" GOFLAGS="-tags=providerless,linux,cgo,ms_tls13kdf"
  cp _output/bin/"${app}" "${INSTALL}/${app}"
done

for app in kubectl kubelet kube-proxy kube-controller-manager kube-scheduler kube-apiserver; do
  ln -sf ./kubernetes "${INSTALL}/${app}"
done
