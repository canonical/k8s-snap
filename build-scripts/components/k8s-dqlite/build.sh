#!/bin/bash

INSTALL="${1}"
mkdir -p "${INSTALL}/bin" "${INSTALL}/usr/lib"

make static

cp bin/static/dqlite "${INSTALL}/bin/dqlite"
cp bin/static/k8s-dqlite "${INSTALL}/bin/k8s-dqlite"
