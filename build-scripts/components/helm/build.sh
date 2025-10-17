#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

# NOTE: Using the latest (2025-10-16) toolchain as the module
# defaults to 1.24.0
GOTOOLCHAIN=go1.24.9 make VERSION="${VERSION}"
cp bin/helm "${INSTALL}/helm"
