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
.agents/skills/                  Agent playbooks: shared project skills + the triage pipeline
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

`.agents/skills/` holds markdown playbooks. The autonomous triage bot composes
them into its prompts, but they are written for **any agent** working in this
repository, so read the relevant one before starting:

- **`local-cluster/`**: bring up a real cluster from this checkout with
  `hack/cluster-up.sh`, drive the nodes, and the snap-build constraints
  (a rebuild is the only way a Go change reaches a cluster).
- **`inspection-report/`**: the layout of an inspection report tarball and the
  order to read it in to find a fault, including the traps that make an absent
  file look like a healthy one.
- **`triage/`**: the bot's own pipeline (`reproduce -> verify -> reproducer ->
  diagnose -> fix`), one file per step. These encode the contract with the
  orchestrator, so they are only directly useful when running that pipeline;
  the project knowledge they used to carry now lives in the skills above.

A step declares the shared skills it needs with a `> Uses:` line, and the runner
in `k8s_ai_agent_toolkit.triage.skills` appends each named skill to the prompt.

## Snap Channels

Channel format: `{major}.{minor}-{flavor}/{risk}`, e.g. `1.37-classic/stable`.
Flavors: `classic` (default), `strict`. Risk levels (ascending): `stable`, `candidate`, `beta`, `edge`.
