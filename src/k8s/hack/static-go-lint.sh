#!/bin/bash -xeu

DIR="$(realpath `dirname "${0}"`)"

. "${DIR}/static-dqlite.sh"
#	golangci-lint run -c ./../../golangci.yml -v

#chmod +x golangci-lint
golangci-lint run \
  --build-tags dqlite,libsqlite3 \
  "${@}"
