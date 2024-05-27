#!/bin/bash -eu

# Run regsync
TOOLS_DIR="$(realpath `dirname "${0}"`)"
(
  cd "${TOOLS_DIR}"
  go run github.com/regclient/regclient/cmd/regsync "${@}"
)
