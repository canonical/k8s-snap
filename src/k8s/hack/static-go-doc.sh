#!/bin/bash -xeu

DIR="$(realpath `dirname "${0}"`)"

. "${DIR}/static-dqlite.sh"

go run -C ${DIR} \
  -tags dqlite,libsqlite3 \
  -ldflags '-linkmode "external" -extldflags "-static"' \
  "${@}"
