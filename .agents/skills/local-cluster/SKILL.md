<!--
Copyright 2026 Canonical, Ltd.
-->
# Local clusters from this checkout

`hack/cluster-up.sh` builds a real Canonical Kubernetes cluster in LXD
containers, straight from this checkout, using the same LXD profile, image
and install path as `tests/integration`. Use it whenever you need a live
cluster to reproduce a behaviour, drive it with `k8s`/`kubectl`, or check that
a change actually does something, without standing up the full end-to-end
suite.

## Bringing up a cluster

```bash
hack/cluster-up.sh --control-plane 1 --workers 1
hack/cluster-up.sh --status
hack/cluster-up.sh --destroy
```

The real options, from `hack/cluster-up.sh --help`:

```
Options:
  --control-plane N  control-plane nodes (default 1)
  --workers N        worker nodes (default 0)
  --snap PATH        k8s snap to install (default ./k8s.snap, built if absent)
  --image IMAGE      LXD image (default ubuntu:22.04)
  --prefix NAME      node name prefix (default k8s-triage)
  --status           show the current cluster and exit
  --destroy          delete every node of this prefix and exit
```

`-h`/`--help` prints that text and exits 0. An unrecognized flag is a hard
error ("unknown option: ... (try --help)"), never a silent skip.
`--control-plane` must be at least 1; both node counts must be non-negative
integers.

Three of those defaults are actually environment variables with a fallback,
so a flag you don't pass may not be the literal default printed above:

- `--prefix` defaults to `$CLUSTER_PREFIX`, falling back to `k8s-triage`.
- `--image` defaults to `$TEST_LXD_IMAGE`, falling back to `ubuntu:22.04` --
  the same variable and default `tests/integration` uses.
- `--snap`'s default resolution is more involved; see "The snap build
  reality" below.

Two more variables tune behaviour with no matching flag: `READY_TIMEOUT`
(default `10m`) is how long the script waits for `k8s status --wait-ready`
and for every node to reach `Ready`; `PER_NODE_GB` (default `8`) is the disk
budget per node used only for a pre-flight capacity estimate.

Map a described cluster shape onto the two counts directly: "two nodes, one
control plane and one worker" is `--control-plane 1 --workers 1`; three
control planes and two workers is `--control-plane 3 --workers 2`.

It is safe to re-run, including with different flags to grow a cluster you
already have: every step checks first and is skipped once already
satisfied. In order:

- `ensure_tooling`: installs LXD via snap if `lxc` is missing, runs
  `lxd init --auto` (a no-op once already initialized), and confirms
  `lxc list` works.
- `ensure_capacity`: refuses to continue when free disk is under
  `(control-plane + workers) * PER_NODE_GB`, rather than let a node silently
  fail with `DiskPressure` much later.
- `ensure_snap`: resolves or builds `k8s.snap` (see "The snap build
  reality" below).
- `ensure_profile`: creates (or overwrites) an LXD profile named after the
  prefix, from `tests/integration/lxd-profile.yaml` -- the same profile
  content `tests/integration` applies under its own profile name.
- each node: launched only if `lxc info <node>` doesn't already find it
  (`lxc launch <image> <node> -p default -p <prefix>`); the snap is then
  installed on it (`snap install --classic --dangerous`, then
  `/snap/k8s/current/k8s/hack/init.sh` to connect interfaces, exactly as
  `tests/integration` does after installing by path) only if
  `/snap/bin/k8s` isn't already present.
- the first control-plane node is bootstrapped (`k8s bootstrap`) only if
  `k8s status` isn't already working on it; every other node joins
  (`k8s get-join-token` / `k8s join-cluster`) only if it isn't already
  visible to `k8s kubectl get node`.

## Node names and driving the cluster

Node names are always derived from the **effective** prefix, so never assume
the literal default: `$CLUSTER_PREFIX` overrides it, and an automated caller
routinely sets that to scope a cluster to one job. Check it before you touch
a node:

```bash
echo "${CLUSTER_PREFIX:-k8s-triage}"
hack/cluster-up.sh --status          # lists the nodes that actually exist
```

For `--control-plane N --workers M`, the script derives:

- control-plane nodes: `<prefix>-cp1` .. `<prefix>-cp<N>`. `<prefix>-cp1` is
  always the first node: it is the one bootstrapped, and every other node's
  join token comes from it.
- worker nodes: `<prefix>-w1` .. `<prefix>-w<M>`.

So `--control-plane 1 --workers 1` gives `<prefix>-cp1` and `<prefix>-w1`:
`k8s-triage-cp1` and `k8s-triage-w1` under the default prefix, but
`k8s-triage-42-cp1` and `k8s-triage-42-w1` when `CLUSTER_PREFIX` is
`k8s-triage-42`. A hardcoded name from the wrong prefix just fails with
"instance not found".

Drive any node with `lxc exec <node> -- k8s ...` -- `k8s` is the CLI the
snap installs:

```bash
CP="${CLUSTER_PREFIX:-k8s-triage}-cp1"
lxc exec "$CP" -- k8s status
lxc exec "$CP" -- k8s kubectl get pods -A -o wide
```

or drop to a plain shell on it: `lxc exec "$CP" -- bash`.

## The snap build reality

A cold build is `snapcraft --use-lxd` from the repo root (needs LXD and
Snapcraft installed, per `CONTRIBUTING.md`). It launches an LXD build
instance and assembles every part of the snap inside it -- `etcd`,
`containerd`, `runc`, the Kubernetes binaries, `k8sd`, `cni`, `helm`, and
more, per `snap/snapcraft.yaml` -- which takes real time: `hack/cluster-up.sh`
itself warns it is building "tens of minutes" before doing so. The finished
package lands in the repo root as `k8s_<version>_<arch>.snap`
(`k8s_v1.35.3_multi.snap` in `CONTRIBUTING.md`'s example). Snapcraft's LXD
build instance (`snapcraft-k8s`) is stopped afterward, not deleted, so later
builds reuse it; remove it yourself with `lxc delete snapcraft-k8s` once you
no longer want it.

`hack/cluster-up.sh` avoids paying that cost on every run. Its `ensure_snap`
resolution, in order:

1. `--snap PATH` if you gave one (must exist).
2. Otherwise `$REPO_ROOT/k8s.snap` -- next to `hack/`, in whichever checkout
   you invoked the script from.
3. If that file is missing and the checkout is a git worktree, the
   `k8s.snap` sitting in the **primary** checkout instead -- the first entry
   of `git worktree list --porcelain` -- reused in place.
4. Only if none of those exist does it fall back to building, and that
   build runs `snapcraft --use-lxd` in step 2's `$REPO_ROOT`: wherever you
   invoked the script from, **not** the primary checkout.

Step 4 is exactly what breaks when your working tree is a git worktree
backed by an sshfs or multipass mount (a common shape for an automated
checkout). `snapcraft --use-lxd` copies every project file into its LXD
build instance, and on such a mount that copy has been observed to fail
with a permission error, taking the build with it. Treat it as a constraint
of this project rather than a theory to re-test.
`hack/cluster-up.sh` does not route around this for you -- step 3 only
reuses a snap that is already sitting in the primary checkout, it never
builds one there on your behalf.

So: if no `k8s.snap` exists anywhere yet and your working tree is such a
worktree, build it yourself in the primary checkout first (or pass `--snap`
pointing at the result), rather than letting the script's own fallback
build run in the worktree and fail. Find the primary checkout with:

```bash
PRIMARY="$(git worktree list --porcelain | sed -n '1s/^worktree //p')"
```

and build there:

```bash
(cd "$PRIMARY" && snapcraft --use-lxd && mv k8s_*.snap k8s.snap)
```

Once that snap exists, `hack/cluster-up.sh`'s own resolution (step 3) finds
and reuses it automatically on later runs, from any worktree, with no
`--snap` flag needed.

A change to a Go component such as `k8sd`
(built from `build-scripts/components/k8sd`, pinned to
`github.com/canonical/k8sd`) only reaches a cluster through a rebuild of the
snap. Testing against a stale `k8s.snap` after editing that source proves
nothing about the change -- rebuild first.

## This cluster vs. the integration test suite

`hack/cluster-up.sh` and `tests/integration` both drive LXD from the same
`tests/integration/lxd-profile.yaml` content, but they are two different
clusters, and mixing them up wastes time:

- `hack/cluster-up.sh` builds a **persistent** cluster, named by its prefix
  (`k8s-triage-*` by default), that only you drive by hand and only
  `--destroy` removes.
- `tests/integration` builds and tears down its **own** instances per test
  run, from the `harness.Instance` fixtures, named
  `k8s-integration-<n>-<hex>` under a separate LXD profile,
  `k8s-integration`. A test gets its cluster from the `instances` fixture,
  never from a cluster you built by hand.

Distinct names and profiles let both exist on the same LXD daemon at once,
but they still compete for the same machine's CPU and disk. If a run is
starved, destroy the `hack/cluster-up.sh` cluster first -- the test suite
never needed it.

For everything about writing or running the integration tests themselves --
fixtures, markers, the `TEST_*` variables, tagging -- see
`tests/integration/AGENTS.md`; it owns those conventions. The one thing
worth repeating here is pointing a run at a snap you already built:

```bash
export TEST_SNAP=$PWD/k8s.snap TEST_SUBSTRATE=lxd
cd tests/integration && tox -e integration -- tests/ -k <test_name>
```

## Tearing down

`hack/cluster-up.sh --prefix <name> --destroy` deletes every LXD instance
whose name starts with `<name>-` (`lxc delete --force`), control-plane and
worker alike. It is idempotent: with nothing matching, it logs "nothing to
delete" and exits 0, so it is always safe to run again, or unconditionally,
as cleanup.

Check what exists first with `hack/cluster-up.sh --prefix <name> --status`:
with a matching cluster it lists the LXD instances and, best-effort, runs
`k8s kubectl get nodes -o wide` from the first one.

The prefix is also the isolation between concurrent clusters: the script
escapes it before matching node names against it, specifically so that one
prefix's `--destroy` can never match, and delete, another prefix's nodes.
Give concurrent clusters different `--prefix` values (or `CLUSTER_PREFIX`)
and neither can interfere with the other, including on cleanup.
