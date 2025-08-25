#!/bin/bash -xe

DIR="$(realpath `dirname "${0}"`)"

. "${DIR}/dynamic-dqlite.sh"

go test \
  -tags dqlite,libsqlite3 \
  -ldflags "${EXTRA_LDFLAGS}" \
  "${@}"
