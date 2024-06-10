#!/bin/bash

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

go build -ldflags '-s -w' ./cmd/pebble
cp pebble "${INSTALL}/pebble"
