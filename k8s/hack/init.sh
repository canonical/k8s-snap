#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNAP_DIR="${SNAP:-$(cd "${SCRIPT_DIR}/../.." && pwd)}"
SNAP_INSTANCE="${SNAP_INSTANCE_NAME:-$(basename "$(dirname "$SNAP_DIR")")}"

if grep -q '^confinement:[[:space:]]*strict' "${SNAP_DIR}/meta/snap.yaml" 2>/dev/null; then
  connections="$(snap connections "${SNAP_INSTANCE}")"
  while IFS= read -r plug; do
    if [ -n "$plug" ]; then
      snap connect "$plug"
    fi
  done <<< "$(awk 'NR>1 && $3 == "-" {print $2}' <<< "$connections")"
fi
