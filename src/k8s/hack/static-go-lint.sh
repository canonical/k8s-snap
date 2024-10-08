#!/bin/bash -xeu

DIR="$(realpath `dirname "${0}"`)"

. "${DIR}/static-dqlite.sh"
#	golangci-lint run -c ./../../golangci.yml -v

go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

golangci-lint run \
  --build-tags dqlite,libsqlite3 \
  "${@}"
