#!/bin/bash -xeu

DIR="$(realpath `dirname "${0}"`)"

. "${DIR}/static-dqlite.sh"

DEBUG_BUILD=${DEBUG_BUILD:-"n"}
if [[ "$DEBUG_BUILD" == "y" ]]; then
  go build \
    -gcflags=all="-N -l" \
    -tags dqlite,libsqlite3 \
    -ldflags '--linkmode "external" -extldflags "-static"' \
    "${@}"
else
  go build \
    -tags dqlite,libsqlite3 \
    -ldflags '-s -w --linkmode "external" -extldflags "-static"' \
    "${@}"
fi
