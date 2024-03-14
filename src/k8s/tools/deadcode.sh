#!/bin/bash -eu

# Build deadcode
TOOLS_DIR="$(realpath `dirname "${0}"`)"
(
  cd "${TOOLS_DIR}"
  go install golang.org/x/tools/cmd/deadcode
)

# Run deadcode
DIR="${TOOLS_DIR}/../hack"
. "${DIR}/static-dqlite.sh"
(
  cd "${TOOLS_DIR}/.."
  x="$($(go env GOPATH)/bin/deadcode -tags dqlite,libsqlite3 -test ./...)"
  if [ ! -z "$x" ]; then
    echo "$x"
    exit 1
  fi
)
