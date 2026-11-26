#!/usr/bin/env bash

set -e

SNAP_DIR="${SNAP:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
SNAP_INSTANCE="${SNAP_INSTANCE_NAME:-$(basename "$(dirname "$SNAP_DIR")")}"

# Sideloaded strict snaps start with their plugs disconnected.
if grep -q 'confinement: strict' "${SNAP_DIR}/meta/snap.yaml"; then
  plugs="$(snap connections "${SNAP_INSTANCE}" | awk 'NR>1 && $3 == "-" {print $2}')"
  for plug in $plugs; do
    snap connect "$plug"
  done

  # Services were started before their plugs were connected, so restart them
  # to pick up the new AppArmor, seccomp and device cgroup policies.
  if [ -n "$plugs" ]; then
    snap restart "${SNAP_INSTANCE}"
    sock="/var/snap/${SNAP_INSTANCE}/common/var/lib/k8sd/state/control.socket"
    for _ in $(seq 1 60); do
      [ -S "$sock" ] && break
      sleep 0.5
    done
  fi
fi
