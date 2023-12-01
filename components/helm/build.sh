#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

make VERSION="${VERSION}"
cp bin/helm "${INSTALL}/helm"
