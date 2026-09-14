#!/usr/bin/env bash
#
# Copyright 2026 Canonical, Ltd.
#
# Bring up a Canonical Kubernetes cluster in LXD for local issue reproduction.
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
# Wait for CNI to settle.
READY_TIMEOUT="${READY_TIMEOUT:-10m}"
# Disk budget per node for the pre-flight check. A control-plane node carries
# etcd on top of the snap, container images and logs, so it needs materially
# more than a worker; 8G for either has been observed to run out under etcd.
PER_CP_GB="${PER_CP_GB:-20}"
PER_WORKER_GB="${PER_WORKER_GB:-10}"
ACTION="up"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROFILE_YAML="$REPO_ROOT/tests/integration/lxd-profile.yaml"

log() { printf '[cluster-up] %s\n' "$*"; }
die() {
  printf '[cluster-up] error: %s\n' "$*" >&2
  exit 1
}

# Escape BRE metacharacters in prefix.
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

case "$CONTROL_PLANE" in
'' | *[!0-9]*) die "--control-plane must be a non-negative integer, got '$CONTROL_PLANE'" ;;
esac
case "$WORKERS" in
'' | *[!0-9]*) die "--workers must be a non-negative integer, got '$WORKERS'" ;;
esac
case "$PER_CP_GB" in
'' | *[!0-9]*) die "PER_CP_GB must be a non-negative integer, got '$PER_CP_GB'" ;;
esac
case "$PER_WORKER_GB" in
'' | *[!0-9]*) die "PER_WORKER_GB must be a non-negative integer, got '$PER_WORKER_GB'" ;;
esac
[ "$CONTROL_PLANE" -ge 1 ] || die "--control-plane must be at least 1"
# The prefix reaches a grep pattern and `lxc delete`, so constrain it to the
# shape a node name can actually have. Without this, an embedded newline
# splits the pattern in two ("^x" plus a bare "-"), and the second pattern
# matches unrelated containers that --destroy would then delete.
case "$PREFIX" in
'' | *[!a-zA-Z0-9_-]*) die "--prefix must be non-empty [a-zA-Z0-9_-] only, got '$PREFIX'" ;;
# A leading '-' makes every generated node name look like an option to lxc,
# which parses it as a flag rather than an instance name.
-*) die "--prefix must not start with '-', got '$PREFIX'" ;;
esac

FIRST="${PREFIX}-cp1"
# Anchored to the exact node shape this script creates, so a parent prefix
# cannot match (and --destroy cannot delete) another run's nodes: plain
# "^k8s-triage-" would also match every "k8s-triage-<issue>-cp1".
# Selects by name shape *and* this script's ownership marker, so a
# pre-existing container that merely happens to be called <prefix>-cp1 is
# never installed into, joined to the cluster, or deleted by --destroy.
node_is_ours() {
  [ "$(lxc config get "$1" user.managed-by 2>/dev/null)" = "$NODE_MARKER" ]
}

nodes() {
  local name
  while read -r name; do
    [ -n "$name" ] || continue
    if node_is_ours "$name"; then
      printf '%s\n' "$name"
    fi
  done < <(lxc list --format=csv -c n |
    grep -E -- "^$(re_escape "$PREFIX")-(cp|w)[0-9]+$" || true)
  # Always succeed: with `set -e`, a trailing unowned container would
  # otherwise make the whole function fail and abort the script.
  return 0
}

# --- tooling -----------------------------------------------------------------

ensure_tooling() {
  if ! command -v lxc >/dev/null; then
    log "installing lxd"
    sudo snap install lxd
  fi
  # Ignore if already initialized.
  sudo lxd init --auto >/dev/null 2>&1 || true
  lxc list >/dev/null 2>&1 ||
    die "cannot reach lxd as $(id -un): add yourself to the 'lxd' group and re-login"
}

ensure_capacity() {
  # Fail early if disk space is insufficient. Only nodes that do not exist yet
  # are counted: on a re-run the existing ones are already on disk, and
  # charging for them again would refuse a resume that is actually fine.
  local want=0 i
  for i in $(seq 1 "$CONTROL_PLANE"); do
    node_is_ours "${PREFIX}-cp${i}" || want=$((want + PER_CP_GB))
  done
  for i in $(seq 1 "$WORKERS"); do
    node_is_ours "${PREFIX}-w${i}" || want=$((want + PER_WORKER_GB))
  done
  [ "$want" -gt 0 ] || return 0
  local where avail
  where="$(lxc storage get default source 2>/dev/null || true)"
  [ -d "$where" ] || where=/
  avail="$(df -BG --output=avail "$where" | tail -1 | tr -dc '0-9')"
  [ -n "$avail" ] || return 0
  log "disk: ${avail}G free on ${where}, need ~${want}G for new nodes"
  [ "$avail" -ge "$want" ] || die "$(
    printf 'not enough disk: ~%dG needed for the nodes still to create, %dG free on %s.\n' \
      "$want" "$avail" "$where"
    printf 'Grow the disk, delete old instances (lxc list), or lower the node '
    printf 'count. Override the estimate with PER_CP_GB / PER_WORKER_GB.'
  )"
}


ensure_snap() {
  if [ -n "$SNAP" ]; then
    [ -f "$SNAP" ] || die "--snap $SNAP does not exist"
  else
    SNAP="$REPO_ROOT/k8s.snap"
    # Reuse snap from primary checkout when run from a worktree.
    if [ ! -f "$SNAP" ]; then
      primary="$(
        git -C "$REPO_ROOT" worktree list --porcelain 2>/dev/null |
          sed -n '1s/^worktree //p'
      )" || true
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

# Marks a profile as created by this script, so a prefix that happens to
# collide with a pre-existing profile is never silently overwritten -- and
# --destroy never deletes a profile it did not create.
PROFILE_MARKER="managed by hack/cluster-up.sh"
NODE_MARKER="managed by hack/cluster-up.sh"

profile_is_ours() {
  [ "$(lxc profile get "$1" user.managed-by 2>/dev/null)" = "$PROFILE_MARKER" ]
}

ensure_profile() {
  [ -f "$PROFILE_YAML" ] || die "missing $PROFILE_YAML"
  if lxc profile show "$PREFIX" >/dev/null 2>&1; then
    profile_is_ours "$PREFIX" ||
      die "LXD profile '$PREFIX' already exists and was not created by this script; choose another --prefix"
  else
    lxc profile create "$PREFIX" >/dev/null
  fi
  lxc profile edit "$PREFIX" <"$PROFILE_YAML"
  lxc profile set "$PREFIX" user.managed-by "$PROFILE_MARKER"
}

# --- nodes -------------------------------------------------------------------

launch_node() {
  local name="$1"
  if lxc info "$name" >/dev/null 2>&1; then
    node_is_ours "$name" ||
      die "container '$name' already exists and was not created by this script; choose another --prefix"
    log "$name: already exists"
  else
    log "$name: launching $IMAGE"
    lxc launch "$IMAGE" "$name" -p default -p "$PREFIX" >/dev/null
    lxc config set "$name" user.managed-by "$NODE_MARKER"
  fi
  lxc exec "$name" -- cloud-init status --wait >/dev/null 2>&1 || true
}

install_snap_on() {
  local name="$1" want have
  # Compare the artefact, not just "is something installed": a rerun after a
  # rebuild (or with a different --snap) must not leave the old snap in the
  # cluster, or the triage test validates stale code and a fix looks unproven.
  want="$(sha256sum "$SNAP" | cut -d' ' -f1)"
  have="$(lxc exec "$name" -- cat /root/k8s.snap.sha256 2>/dev/null || true)"
  if [ "$want" = "$have" ] && lxc exec "$name" -- test -x /snap/bin/k8s 2>/dev/null; then
    log "$name: k8s snap already installed (matching build)"
    return
  fi
  [ -z "$have" ] || log "$name: installed snap differs from $SNAP, reinstalling"
  log "$name: installing k8s snap"
  lxc file push "$SNAP" "$name/root/k8s.snap" >/dev/null
  # `snap install --dangerous` fails with "already installed" when the node
  # already carries a build, so replacing a stale one needs refresh.
  if lxc exec "$name" -- test -x /snap/bin/k8s 2>/dev/null; then
    lxc exec "$name" -- snap refresh --classic --dangerous --amend /root/k8s.snap
  else
    lxc exec "$name" -- snap install --classic --dangerous /root/k8s.snap
  fi
  # Initialize interfaces and network prerequisites.
  lxc exec "$name" -- /snap/k8s/current/k8s/hack/init.sh >/dev/null
  # Recorded only once init has succeeded too: marking the snap earlier would
  # let a rerun skip both reinstall and initialisation on a node whose
  # interfaces were never connected.
  printf '%s' "$want" | lxc exec "$name" -- tee /root/k8s.snap.sha256 >/dev/null
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
  # Wait for all nodes to become ready.
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
  if [ -n "$found" ]; then
    # shellcheck disable=SC2086 # deliberate word splitting: one arg per node
    lxc delete --force $found
    log "deleted: $(echo "$found" | tr '\n' ' ')"
  else
    log "nothing to delete"
  fi
  # The profile is created per prefix, so it leaks on a long-lived host once
  # the nodes are gone. Only ever remove one this script created: a prefix
  # can collide with an unrelated profile (including `default`), and deleting
  # that would be someone else's outage. Removing fails harmlessly while
  # anything still references it.
  if profile_is_ours "$PREFIX"; then
    lxc profile delete "$PREFIX" >/dev/null 2>&1 && log "deleted profile: $PREFIX" || true
  fi
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
