#!/bin/bash -x

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

export KUBE_GIT_VERSION_FILE="${PWD}/.version.sh"

for app in kubectl kube-apiserver kubelet kube-scheduler kube-proxy kube-controller-manager; do
  make WHAT="cmd/${app}" KUBE_STATIC_OVERRIDES="${app}"
  cp _output/bin/"${app}" "${INSTALL}/${app}"
done
