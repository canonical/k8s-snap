#!/bin/bash

VERSION="${2}"

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

# these would very tedious to apply with a patch
sed -i 's/^package main/package plugin_main/' plugins/*/*/*.go
sed -i 's/^func main()/func Main()/' plugins/*/*/*.go

export CGO_ENABLED=0

go build -o cni -ldflags "-s -w -extldflags -static -X github.com/containernetworking/plugins/pkg/utils/buildversion.BuildVersion=${VERSION}" ./cni.go

cp cni "${INSTALL}/cni"
