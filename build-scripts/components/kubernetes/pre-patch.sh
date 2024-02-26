#!/bin/bash -x

KUBE_ROOT="${PWD}"

# Remove .version.sh from previous runs (if any)
[ -f "${KUBE_ROOT}/.version.sh" ] && rm "${KUBE_ROOT}/.version.sh"

# Ensure clean Kubernetes version
source "${KUBE_ROOT}/hack/lib/version.sh"
kube::version::get_version_vars
kube::version::save_version_vars "${KUBE_ROOT}/.version.sh"
