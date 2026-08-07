#!/usr/bin/env bash
#
# Copyright 2026 Canonical, Ltd.
#
# Bring up a Canonical Kubernetes cluster in LXD containers for local issue
# reproduction. Uses the same profile, image and install path as
# tests/integration, so a cluster built here behaves like one the e2e suite
# builds. Safe to re-run: every step is skipped when already satisfied.
#
#   hack/cluster-up.sh --control-plane 1 --workers 1
#   hack/cluster-up.sh --status
#   hack/cluster-up.sh --destroy
set -euo pipefail

CONTROL_PLANE=1
WORKERS=0
PREFIX="${CLUSTER_PREFIX:-k8s-triage}"
SNAP=""
IMAGE="${TEST_LXD_IMAGE:-ubuntu:22.04}"
# `k8s status --wait-ready` defaults to 90s, which a small machine routinely
# overruns while the CNI settles, so wait long enough to mean something.
READY_TIMEOUT="${READY_TIMEOUT:-10m}"
# Rough disk budget per node (snap, container images, etcd, logs). Only used
# for the pre-flight estimate.
PER_NODE_GB="${PER_NODE_GB:-8}"
ACTION="up"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROFILE_YAML="$REPO_ROOT/tests/integration/lxd-profile.yaml"

log() { printf '[cluster-up] %s\n' "$*"; }
die() {
  printf '[cluster-up] error: %s\n' "$*" >&2
  exit 1
}

# Escapes BRE metacharacters so PREFIX (user-supplied via --prefix/
# CLUSTER_PREFIX) can never widen the grep below into matching -- and, on
# --destroy, deleting -- containers outside its own prefix.
re_escape() {
  local s=$1
  s=${s//\\/\\\\}
  s=${s//./\\.}
  s=${s//\*/\\*}
  s=${s//\[/\\[}
  s=${s//^/\\^}
  s=${s//\$/\\$}
  printf '%s' "$s"
}

usage() {
  cat <<'EOF'
Bring up a Canonical Kubernetes cluster in LXD containers for local issue
reproduction, using the same profile, image and install path as
tests/integration. Safe to re-run: each step is skipped when already satisfied.

  hack/cluster-up.sh --control-plane 1 --workers 1
  hack/cluster-up.sh --status
  hack/cluster-up.sh --destroy

Options:
  --control-plane N  control-plane nodes (default 1)
  --workers N        worker nodes (default 0)
  --snap PATH        k8s snap to install (default ./k8s.snap, built if absent)
  --image IMAGE      LXD image (default ubuntu:22.04)
  --prefix NAME      node name prefix (default k8s-triage)
  --status           show the current cluster and exit
  --destroy          delete every node of this prefix and exit
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
  --control-plane) CONTROL_PLANE="$2" && shift 2 ;;
  --workers) WORKERS="$2" && shift 2 ;;
  --snap) SNAP="$2" && shift 2 ;;
  --image) IMAGE="$2" && shift 2 ;;
  --prefix) PREFIX="$2" && shift 2 ;;
  --status) ACTION="status" && shift ;;
  --destroy) ACTION="destroy" && shift ;;
  -h | --help)
    usage
    exit 0
    ;;
  *) die "unknown option: $1 (try --help)" ;;
  esac
done

[ "$CONTROL_PLANE" -ge 1 ] || die "--control-plane must be at least 1"

FIRST="${PREFIX}-cp1"
nodes() { lxc list --format=csv -c n | grep "^$(re_escape "$PREFIX")-" || true; }

# --- tooling -----------------------------------------------------------------

ensure_tooling() {
  if ! command -v lxc >/dev/null; then
    log "installing lxd"
    sudo snap install lxd
  fi
  # Exits non-zero once already initialised, which is the common case.
  sudo lxd init --auto >/dev/null 2>&1 || true
  lxc list >/dev/null 2>&1 ||
    die "cannot reach lxd as $(id -un): add yourself to the 'lxd' group and re-login"
  command -v jq >/dev/null || sudo apt-get install -y -qq jq
}

ensure_capacity() {
  # A node that runs out of disk does not fail loudly: the kubelet reports
  # DiskPressure, the CNI image pull backs off, and the node simply never
  # becomes Ready. Refuse up front rather than time out much later.
  local want=$(((CONTROL_PLANE + WORKERS) * PER_NODE_GB))
  local where avail
  where="$(lxc storage get default source 2>/dev/null || true)"
  [ -d "$where" ] || where=/
  avail="$(df -BG --output=avail "$where" | tail -1 | tr -dc '0-9')"
  [ -n "$avail" ] || return 0
  log "disk: ${avail}G free on ${where}, need ~${want}G"
  [ "$avail" -ge "$want" ] || die "$(
    printf 'not enough disk for %d node(s): ~%dG needed, %dG free on %s.\n' \
      "$((CONTROL_PLANE + WORKERS))" "$want" "$avail" "$where"
    printf 'Grow the disk, delete old instances (lxc list), or lower the node '
    printf 'count. Override the estimate with PER_NODE_GB.'
  )"
}

ensure_snap() {
  if [ -n "$SNAP" ]; then
    [ -f "$SNAP" ] || die "--snap $SNAP does not exist"
  else
    SNAP="$REPO_ROOT/k8s.snap"
    # Run from a git worktree (as the triage bot does), the snap normally sits
    # in the primary checkout. Reuse it rather than spend tens of minutes
    # rebuilding a byte-identical artefact.
    if [ ! -f "$SNAP" ]; then
      primary="$(git -C "$REPO_ROOT" worktree list --porcelain 2>/dev/null |
        sed -n '1s/^worktree //p')"
      if [ -n "$primary" ] && [ -f "$primary/k8s.snap" ]; then
        SNAP="$primary/k8s.snap"
      fi
    fi
  fi
  if [ ! -f "$SNAP" ]; then
    log "no $SNAP yet, building it (tens of minutes)"
    command -v snapcraft >/dev/null || sudo snap install snapcraft --classic
    (cd "$REPO_ROOT" && snapcraft --use-lxd && mv k8s_*.snap k8s.snap)
  fi
  SNAP="$(readlink -f "$SNAP")"
  log "snap under test: $SNAP"
}

ensure_profile() {
  [ -f "$PROFILE_YAML" ] || die "missing $PROFILE_YAML"
  lxc profile show "$PREFIX" >/dev/null 2>&1 || lxc profile create "$PREFIX" >/dev/null
  lxc profile edit "$PREFIX" <"$PROFILE_YAML"
}

# --- nodes -------------------------------------------------------------------

launch_node() {
  local name="$1"
  if lxc info "$name" >/dev/null 2>&1; then
    log "$name: already exists"
  else
    log "$name: launching $IMAGE"
    lxc launch "$IMAGE" "$name" -p default -p "$PREFIX" >/dev/null
  fi
  lxc exec "$name" -- cloud-init status --wait >/dev/null 2>&1 || true
}

install_snap_on() {
  local name="$1"
  if lxc exec "$name" -- test -x /snap/bin/k8s 2>/dev/null; then
    log "$name: k8s snap already installed"
    return
  fi
  log "$name: installing k8s snap"
  lxc file push "$SNAP" "$name/root/k8s.snap" >/dev/null
  lxc exec "$name" -- snap install --classic --dangerous /root/k8s.snap
  # Connects interfaces and applies the network prerequisites, exactly as
  # tests/integration does after installing by path.
  lxc exec "$name" -- /snap/k8s/current/k8s/hack/init.sh >/dev/null
}

joined() { lxc exec "$FIRST" -- k8s kubectl get node "$1" >/dev/null 2>&1; }

bootstrap_first() {
  if lxc exec "$FIRST" -- k8s status >/dev/null 2>&1; then
    log "$FIRST: already bootstrapped"
    return
  fi
  log "$FIRST: bootstrapping"
  lxc exec "$FIRST" -- k8s bootstrap --name "$FIRST"
  lxc exec "$FIRST" -- k8s status --wait-ready --timeout "$READY_TIMEOUT" >/dev/null
}

join_node() {
  local name="$1" role="$2" token_args=()
  if joined "$name"; then
    log "$name: already in the cluster"
    return
  fi
  [ "$role" = "worker" ] && token_args=(--worker)
  log "$name: joining as $role"
  local token
  token="$(lxc exec "$FIRST" -- k8s get-join-token "$name" "${token_args[@]}")"
  lxc exec "$name" -- k8s join-cluster "$token" --name "$name"
}

summary() {
  # `--wait-ready` is satisfied by a single ready node, so a freshly joined
  # worker can still be NotReady. Wait for all of them: a successful exit
  # should mean the whole cluster is usable.
  lxc exec "$FIRST" -- k8s status --wait-ready --timeout "$READY_TIMEOUT" >/dev/null
  log "waiting for all nodes to become Ready"
  lxc exec "$FIRST" -- k8s kubectl wait --for=condition=Ready nodes --all \
    --timeout="$READY_TIMEOUT" >/dev/null
  echo
  lxc exec "$FIRST" -- k8s kubectl get nodes -o wide
  cat <<EOF

cluster:    ${CONTROL_PLANE} control-plane, ${WORKERS} worker(s), prefix '${PREFIX}'
first node: ${FIRST}
run k8s:    lxc exec ${FIRST} -- k8s status
run kubectl:lxc exec ${FIRST} -- k8s kubectl get pods -A
node shell: lxc exec <node> -- bash
destroy:    hack/cluster-up.sh --prefix ${PREFIX} --destroy
EOF
}

# --- actions -----------------------------------------------------------------

case "$ACTION" in
status)
  ensure_tooling
  found="$(nodes)"
  [ -n "$found" ] || {
    log "no nodes with prefix '${PREFIX}'"
    exit 0
  }
  echo "$found"
  lxc exec "$FIRST" -- k8s kubectl get nodes -o wide 2>/dev/null || true
  exit 0
  ;;
destroy)
  ensure_tooling
  found="$(nodes)"
  [ -n "$found" ] || {
    log "nothing to delete"
    exit 0
  }
  # shellcheck disable=SC2086 # deliberate word splitting: one arg per node
  lxc delete --force $found
  log "deleted: $(echo "$found" | tr '\n' ' ')"
  exit 0
  ;;
esac

ensure_tooling
ensure_capacity
ensure_snap
ensure_profile

for i in $(seq 1 "$CONTROL_PLANE"); do
  launch_node "${PREFIX}-cp${i}"
  install_snap_on "${PREFIX}-cp${i}"
done
for i in $(seq 1 "$WORKERS"); do
  launch_node "${PREFIX}-w${i}"
  install_snap_on "${PREFIX}-w${i}"
done

bootstrap_first
for i in $(seq 2 "$CONTROL_PLANE"); do
  join_node "${PREFIX}-cp${i}" control-plane
done
for i in $(seq 1 "$WORKERS"); do
  join_node "${PREFIX}-w${i}" worker
done

summary
