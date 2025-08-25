#!/bin/bash -xe

DIR="$(realpath `dirname "${0}"`)"

. "${DIR}/dynamic-dqlite.sh"

go install \
  -tags dqlite,libsqlite3 \
  -ldflags "-s -w ${EXTRA_LDFLAGS}" \
  "${@}"
