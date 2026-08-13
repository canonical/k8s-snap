# AGENTS.md - k8s-snap

Canonical Kubernetes snap. Bundles upstream Kubernetes, etcd, containerd, runc, CNI plugins,
helm, and k8sd into a single classic/strict snap package.

Before working in a subfolder, read its `AGENTS.md`:

| Folder | AGENTS.md | What it covers |
|--------|-----------|----------------|
| `build-scripts/` | [build-scripts/AGENTS.md](build-scripts/AGENTS.md) | component structure, patching algorithm |
| `k8s/` | [k8s/AGENTS.md](k8s/AGENTS.md) | bash library conventions, runtime paths, services |
| `tests/integration/` | [tests/integration/AGENTS.md](tests/integration/AGENTS.md) | test harness, tagging, utilities, linting |
| `docs/` | [docs/AGENTS.md](docs/AGENTS.md) | design proposals, Spread tests, docs build |

## Repository Layout

```
snap/snapcraft.yaml              snap definition; parts reference build-scripts/components/*
build-scripts/                   component build scripts (see build-scripts/AGENTS.md)
k8s/                             snap runtime: lib.sh, wrappers, manifests (see k8s/AGENTS.md)
.agents/skills/triage/           Playbooks for all agents: reproduce, test, diagnose, fix
ci/                              CI automation Python (GitHub Actions, tox, Mattermost)
tests/integration/               pytest integration tests (see tests/integration/AGENTS.md)
docs/                            MkDocs user docs and proposals (see docs/AGENTS.md)
```

## Multi-Repo Dependencies

The snap coordinates three repos:

| Repo | Purpose |
|------|---------|
| `github.com/canonical/k8s-snap` | this repo — snap shell, build scripts, integration tests |
| `github.com/canonical/k8sd` | Kubernetes daemon (Go backend) |
| `github.com/canonical/k8s-snap-api` | shared Go API types between k8sd and snap clients |

API changes require PRs in all three. During development, add a `replace` directive in
`k8sd/go.mod` pointing to a local k8s-snap-api checkout; remove it before merging.

## Snap Build

Built with `snapcraft --use-lxd`. Architectures: `amd64`, `arm64`, `ppc64el`, `s390x` (see `snap/snapcraft.yaml` for snap base and other build metadata).
Uses `go/<version>-fips/stable` for all Go component builds. FIPS mode is a first-class concern.

## Agent Skills

The repository includes documented workflows in `.agents/skills/triage/` that
are used by the autonomous triage bot, but they serve as definitive playbooks
for **any agent** working in this repository:

- **`reproduce.md`**: How to use `hack/cluster-up.sh` to build a cluster and manually reproduce a reported issue.
- **`reproducer.md`**: How to write, isolate, and run end-to-end tests in `tests/integration/`.
- **`diagnose.md`**: How to root-cause failures, particularly those crossing the boundary between the snap shell and the `k8sd` backend.
- **`fix.md`**: How to cleanly implement fixes across the multi-repo structure, rebuild the snap from source, and verify the fix against the tests.

Whenever you are asked to reproduce a bug, write a test, or fix an issue, you should consult these skill files for the exact commands, constraints, and architecture rules you must follow.
## Snap Channels

Channel format: `{major}.{minor}-{flavor}/{risk}`, e.g. `1.35-classic/stable`.
Flavors: `classic` (default), `strict`. Risk levels (ascending): `stable`, `candidate`, `beta`, `edge`.
