#!/bin/bash

set -ex

DIR=$(realpath $(dirname "${0}"))

BUILD_DIRECTORY="${SNAPCRAFT_PART_BUILD:-${DIR}/.build}"
INSTALL_DIRECTORY="${SNAPCRAFT_PART_INSTALL:-${DIR}/.install}"

mkdir -p "${BUILD_DIRECTORY}" "${INSTALL_DIRECTORY}"

COMPONENT_NAME="${1}"
COMPONENT_DIRECTORY="${DIR}/components/${COMPONENT_NAME}"

COMPONENT_BUILD_DIRECTORY="${BUILD_DIRECTORY}/${COMPONENT_NAME}"

# Detect source type and fetch sources
if [ -f "${COMPONENT_DIRECTORY}/repository" ]; then
  # Git-based build
  GIT_REPOSITORY=$(cat "${COMPONENT_DIRECTORY}/repository")
  GIT_TAG=$(cat "${COMPONENT_DIRECTORY}/version")

  # cleanup git repository if we cannot git checkout to the build tag
  if [ -d "${COMPONENT_BUILD_DIRECTORY}" ]; then
    cd "${COMPONENT_BUILD_DIRECTORY}"
    if ! git reset --hard "${GIT_TAG}"; then
      cd "${BUILD_DIRECTORY}"
      rm -rf "${COMPONENT_BUILD_DIRECTORY}"
    fi
  fi

  if [ ! -d "${COMPONENT_BUILD_DIRECTORY}" ]; then
    git clone "${GIT_REPOSITORY}" --depth 1 -b "${GIT_TAG}" "${COMPONENT_BUILD_DIRECTORY}"
  fi

  cd "${COMPONENT_BUILD_DIRECTORY}"
  VERSION="${GIT_TAG}"

elif [ -f "${COMPONENT_DIRECTORY}/deb-src" ]; then
  # deb-src build
  PACKAGE_NAME=$(cat "${COMPONENT_DIRECTORY}/deb-src")
  SRC_VERSION=$(cat "${COMPONENT_DIRECTORY}/version")

  cd "${BUILD_DIRECTORY}"
  rm -rf "${COMPONENT_BUILD_DIRECTORY}"

  # Validate version exists in repository using madison output
  if ! apt-cache madison "${PACKAGE_NAME}" | grep -q " ${SRC_VERSION} "; then
    echo "Error: Version ${SRC_VERSION} not found for package ${PACKAGE_NAME}"
    echo "Available versions:"
    apt-cache madison "${PACKAGE_NAME}" | head -10
    exit 1
  fi

  # Fetch the pinned source package version
  apt-get source -y "${PACKAGE_NAME}=${SRC_VERSION}"

  # Find extracted source directory and rename to COMPONENT_BUILD_DIRECTORY
  SOURCE_DIR=$(find . -maxdepth 1 -type d -name "${PACKAGE_NAME}-*" | head -1)
  if [ -z "${SOURCE_DIR}" ]; then
    echo "Error: Could not find extracted source directory for ${PACKAGE_NAME}"
    exit 1
  fi
  mv "${SOURCE_DIR}" "${COMPONENT_BUILD_DIRECTORY}"
  cd "${COMPONENT_BUILD_DIRECTORY}"

  # Get upstream version from debian/changelog
  VERSION=$(dpkg-parsechangelog -S Version)

  # Initialize git for patch management (Ubuntu patches already applied by dpkg-source)
  git init
  git add -A
  git commit -m "Import from Ubuntu source package"

else
  echo "Error: No 'repository' or 'deb-src' file found for component ${COMPONENT_NAME}"
  exit 1
fi

# Common logic for both source types
git config user.name "K8s builder bot"
git config user.email "k8s-bot@canonical.com"

if [ -e "${COMPONENT_DIRECTORY}/pre-patch.sh" ]; then
  bash -xe "${COMPONENT_DIRECTORY}/pre-patch.sh"
fi

for patch in $(python3 "${DIR}/print-patches-for.py" "${COMPONENT_NAME}" "${VERSION}"); do
  git am "${patch}"
done

bash -xe "${COMPONENT_DIRECTORY}/build.sh" "${INSTALL_DIRECTORY}" "${VERSION}"
