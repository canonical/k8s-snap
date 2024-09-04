#!/bin/bash -xeu

DIR="$(realpath `dirname "${0}"`)"

. "${DIR}/dynamic-dqlite.sh"

DEBUG_BUILD=${DEBUG_BUILD:-"n"}
if [[ "$DEBUG_BUILD" == "y" ]]; then
  go build \
  -gcflags=all="-N -l" \
  -tags dqlite,libsqlite3 \
  "${@}"
else
  go build \
  -tags dqlite,libsqlite3 \
  -ldflags '-s -w' \
  "${@}"
fi
