#!/bin/bash -x

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

export KUBE_GIT_VERSION_FILE="${PWD}/.version.sh"

for app in kubernetes; do
  export GOTOOLCHAIN=local
	# Only needed for Go < 1.27
	if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
		export GOEXPERIMENT=opensslcrypto
	fi
  export CGO_ENABLED=1
  make WHAT="cmd/${app}" KUBE_CGO_OVERRIDES="${app}" GOFLAGS="-tags=providerless,linux,cgo,ms_tls13kdf"
  cp _output/bin/"${app}" "${INSTALL}/${app}"
done

for app in kubectl kubelet kube-proxy kube-controller-manager kube-scheduler kube-apiserver; do
  ln -sf ./kubernetes "${INSTALL}/${app}"
done
