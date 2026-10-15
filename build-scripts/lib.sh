#!/bin/bash
#
# Shared bash function library for build-scripts.

# Retry a command a bounded number of times, waiting between attempts.
# Example: 'retry 5 5 git clone https://example.com/repo.git'
retry() {
  local attempts="${1}" delay="${2}"
  shift 2
  local i
  for i in $(seq 1 "${attempts}"); do
    if "$@"; then
      return 0
    elif [ "${i}" -eq "${attempts}" ]; then
      echo "Failed after ${attempts} attempts: $*" >&2
      return 1
    else
      echo "Retrying (${i}/${attempts}) in ${delay}s: $*" >&2
      sleep "${delay}"
    fi
  done
}
