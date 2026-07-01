# k8s

Snap runtime layer: entry points, service wrappers, bash library, bundled manifests, and
configuration templates. This directory is copied verbatim into the snap as `$SNAP/k8s/`.

## Directory Layout

```
lib.sh               shared bash function library, sourced by all wrappers
wrappers/commands/   snap app entry points (k8s CLI, dqlite)
wrappers/services/   snap daemon entry points (containerd, kubelet, kube-apiserver, …)
args/                default service argument files (overridden at runtime in SNAP_DATA)
manifests/charts/    bundled Helm charts deployed by k8sd
resources/
  configurations/    audit policies, pod security configs, DISA-STIG bootstrap templates
scripts/             install/remove helper scripts (inspect.sh, etc.)
hack/                low-level init helpers (connect-interfaces.sh, init.sh, …)
pebble/              pebble layer config for service supervision
profiles/            containerd base profile
systemd/             containerd systemd drop-in defaults
```

## Bash Library (`lib.sh`)

All functions are namespaced `k8s::<category>::<name>`. Never define bare function names.

| Namespace | Purpose |
|-----------|---------|
| `k8s::common::*` | env setup, FIPS check, env validation, snap confinement detection |
| `k8s::cmd::*` | thin wrappers: `kubectl`, `ctr`, `k8s`, `hostname` |
| `k8s::util::*` | system helpers: default interface, kernel module loading, wait loops |
| `k8s::remove::*` | cleanup hooks invoked during `snap remove` |
| `k8s::init::*` | bootstrap helpers (k8sd single-node init) |
| `k8s::kubelet::*` | kubelet-specific setup (shared root dir) |
| `k8s::containerd::*` | containerd systemd defaults |
| `k8s::apiserver::*` | feature-gate sanitization |

When adding a function, include a comment block immediately above it with a one-line
description and a usage example:

```bash
# Short description of what this does
# Example: 'k8s::util::my_helper arg1 arg2'
k8s::util::my_helper() {
  ...
}
```

## Snap Runtime Paths

| Path | Purpose |
|------|---------|
| `/snap/k8s/current/k8s/` | read-only snap content (this directory at runtime) |
| `/var/snap/k8s/common/args/` | service argument files (writable, override defaults) |
| `/var/snap/k8s/common/etc/` | configuration files |
| `/var/snap/k8s/common/var/lib/` | persistent service data |

`$SNAP_DATA` resolves to `/var/snap/k8s/<revision>/` (revision-scoped writable).
`$SNAP_COMMON` resolves to `/var/snap/k8s/common/` (shared across revisions).

## Snap Services

All services run under `snap services k8s.*`:

`containerd`, `etcd`, `k8sd`, `kubelet`, `kube-apiserver`, `kube-controller-manager`,
`kube-proxy`, `kube-scheduler`

Start order is declared in `snap/snapcraft.yaml` via `before`/`after` constraints.
Most services are `install-mode: disable` and are activated by k8sd during bootstrap.
