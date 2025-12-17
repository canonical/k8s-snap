#!/bin/bash -xeu

DIR="$(realpath `dirname "${0}"`)"

. "${DIR}/static-dqlite.sh"

# Workaround for issue in Go 1.25 https://github.com/golang/go/issues/75031
go env -w GOTOOLCHAIN="go$(grep ^go go.mod | awk '{print $2;}')+auto"

go test \
  -tags dqlite,libsqlite3 \
  -ldflags '-linkmode "external" -extldflags "-static"' \
  "${@}"
