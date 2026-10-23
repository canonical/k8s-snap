#!/bin/bash

VERSION="${2}"

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

# these would very tedious to apply with a patch
sed -i 's/^package main/package plugin_main/' plugins/*/*/*.go
sed -i 's/^func main()/func Main()/' plugins/*/*/*.go

export CGO_ENABLED=1
export GOTOOLCHAIN=local

# Only needed for Go < 1.27
if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
	export GOEXPERIMENT=opensslcrypto
fi

go build -tags "linux,cgo,ms_tls13kdf" -o cni -ldflags "-linkmode external -s -w -X github.com/containernetworking/plugins/pkg/utils/buildversion.BuildVersion=${VERSION}" ./cni.go

cp cni "${INSTALL}/cni"
