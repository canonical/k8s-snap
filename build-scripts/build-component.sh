#!/bin/bash

set -ex

# Retry a command a bounded number of times on transient failures (e.g.
# network blips against an upstream host or the snap store) instead of
# failing the whole build outright. Exported so build.sh scripts invoked
# below (as a separate bash process) can call it directly too.
retry() {
  local attempts="${1}" delay="${2}"
  shift 2
  local i
  for i in $(seq 1 "${attempts}"); do
    if "$@"; then
      return 0
    elif [ "${i}" -eq "${attempts}" ]; then
      echo "Failed after ${attempts} attempts: $*" >&2
      return 1
    else
      echo "Retrying (${i}/${attempts}) in ${delay}s: $*" >&2
      sleep "${delay}"
    fi
  done
}
export -f retry

DIR=`realpath $(dirname "${0}")`

BUILD_DIRECTORY="${SNAPCRAFT_PART_BUILD:-${DIR}/.build}"
INSTALL_DIRECTORY="${SNAPCRAFT_PART_INSTALL:-${DIR}/.install}"

mkdir -p "${BUILD_DIRECTORY}" "${INSTALL_DIRECTORY}"

COMPONENT_NAME="${1}"
COMPONENT_DIRECTORY="${DIR}/components/${COMPONENT_NAME}"

GIT_REPOSITORY="$(cat "${COMPONENT_DIRECTORY}/repository")"
GIT_TAG="$(cat "${COMPONENT_DIRECTORY}/version")"

COMPONENT_BUILD_DIRECTORY="${BUILD_DIRECTORY}/${COMPONENT_NAME}"

# cleanup git repository if we cannot git checkout to the build tag
if [ -d "${COMPONENT_BUILD_DIRECTORY}" ]; then
  cd "${COMPONENT_BUILD_DIRECTORY}"
  if ! git reset --hard "${GIT_TAG}"; then
    cd "${BUILD_DIRECTORY}"
    rm -rf "${COMPONENT_BUILD_DIRECTORY}"
  fi
fi

if [ ! -d "${COMPONENT_BUILD_DIRECTORY}" ]; then
  retry 5 5 git clone "${GIT_REPOSITORY}" --depth 1 -b "${GIT_TAG}" "${COMPONENT_BUILD_DIRECTORY}"
fi

cd "${COMPONENT_BUILD_DIRECTORY}"
echo "Building ${COMPONENT_NAME} at commit $(git rev-parse HEAD)"
git config user.name "K8s builder bot"
git config user.email "k8s-bot@canonical.com"

if [ -e "${COMPONENT_DIRECTORY}/pre-patch.sh" ]; then
  bash -xe "${COMPONENT_DIRECTORY}/pre-patch.sh"
fi

for patch in $(python3 "${DIR}/print-patches-for.py" "${COMPONENT_NAME}" "${GIT_TAG}"); do
  git am "${patch}"
done

bash -xe "${COMPONENT_DIRECTORY}/build.sh" "${INSTALL_DIRECTORY}" "${GIT_TAG}"
