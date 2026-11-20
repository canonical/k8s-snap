#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNAP_DIR="${SNAP:-$(cd "${SCRIPT_DIR}/../.." && pwd)}"
SNAP_INSTANCE="${SNAP_INSTANCE_NAME:-$(basename "$(dirname "$SNAP_DIR")")}"

if grep -q '^confinement:[[:space:]]*strict' "${SNAP_DIR}/meta/snap.yaml" 2>/dev/null; then
  connected_any=0
  connections="$(snap connections "${SNAP_INSTANCE}")"
  while IFS= read -r plug; do
    if [ -n "$plug" ]; then
      snap connect "$plug"
      connected_any=1
    fi
  done <<< "$(awk 'NR>1 && $3 == "-" {print $2}' <<< "$connections")"

  # Services started during `snap install --dangerous` launched before plugs
  # were connected; restart so they pick up connected AppArmor, seccomp, and
  # device cgroup policies.
  if [ "$connected_any" -eq 1 ]; then
    snap restart "${SNAP_INSTANCE}"
    sock="/var/snap/${SNAP_INSTANCE}/common/var/lib/k8sd/state/control.socket"
    for _ in $(seq 1 60); do
      [ -S "$sock" ] && break
      sleep 0.5
    done
  fi
fi
