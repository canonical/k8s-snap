#!/bin/bash -xeu

DIR="$(realpath `dirname "${0}"`)"

. "${DIR}/static-dqlite.sh"

$(go env GOPATH)/bin/deadcode \
  -tags dqlite,libsqlite3 \
  "${@}" 
