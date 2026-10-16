# CI failure report

Data period: `2026-10` (through `2026-10-15T16:31:53Z`)

## Attribution and delta

- Observed failed jobs: 213/10557 = 2.0% (n=10557 jobs)
- Failure pressure including blocked jobs: 2.0% (n=10557 jobs plus blocked)
- Top attribution: `unknown` owned by unowned triage (52.1% (n=213 facts) of inspected failures).

## Headline metrics

| Period | Runs | Failed jobs | Failure rate | Scheduled green | PR first-pass | Unclassified | M12 fact coverage |
|---|---:|---:|---|---|---|---|---|
| 2026-10 | 5 | 213 | 2.0% (n=10557 jobs) | 0.0% (n=5 runs) | n/a | 52.1% (n=213 facts) | 100.0% (n=213 expected failed jobs) |

## Failure attribution

| Class | Count | Share | Owner |
|---|---:|---|---|
| `unknown` | 111 | 52.1% (n=213 facts) | unowned triage |
| `external.dependency` | 87 | 40.8% (n=213 facts) | external dependency |
| `product.bug` | 9 | 4.2% (n=213 facts) | product team |
| `infra.provisioning` | 4 | 1.9% (n=213 facts) | CI/infra |
| `infra.runner` | 2 | 0.9% (n=213 facts) | CI/infra |

## Integrity metrics

These are report metrics, not appendix material: they show whether apparent improvement came from running less.

- M9 unclassified share: 52.1% (n=213 failed job facts)
- M11 jobs blocked by failed `Prepare Environment`: 0 across 0 prepare job(s).
- M12 failed-job fact coverage: 100.0% (n=213 expected failed jobs).
- M12 log availability among failed job facts: 94.4% (n=213 failed job facts).

### M10 tests collected per config

| Config | Min | Median | Max |
|---|---:|---:|---:|
| `jammy/amd64/latest/edge/unknown-flavor/multipass` | 71 | 71 | 71 |
| `noble/amd64/latest/edge/unknown-flavor/multipass` | 71 | 71 | 71 |
| `ubuntu:22.04/amd64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 42 |
| `ubuntu:22.04/amd64/1.32-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:22.04/amd64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 48 |
| `ubuntu:22.04/amd64/1.33-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:22.04/amd64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 53 |
| `ubuntu:22.04/amd64/1.34-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:22.04/amd64/1.35-classic/edge/unknown-flavor/lxd` | 63 | 63 | 63 |
| `ubuntu:22.04/amd64/1.35-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:22.04/amd64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:22.04/amd64/1.36-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:22.04/amd64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:22.04/amd64/1.37-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:22.04/amd64/latest/edge/unknown-flavor/lxd` | 1 | 68 | 68 |
| `ubuntu:22.04/arm64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 42 |
| `ubuntu:22.04/arm64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 48 |
| `ubuntu:22.04/arm64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 53 |
| `ubuntu:22.04/arm64/1.35-classic/edge/unknown-flavor/lxd` | 63 | 63 | 63 |
| `ubuntu:22.04/arm64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:22.04/arm64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:22.04/arm64/latest/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:24.04/amd64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 42 |
| `ubuntu:24.04/amd64/1.32-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:24.04/amd64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 48 |
| `ubuntu:24.04/amd64/1.33-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:24.04/amd64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 53 |
| `ubuntu:24.04/amd64/1.34-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:24.04/amd64/1.35-classic/edge/unknown-flavor/lxd` | 63 | 63 | 63 |
| `ubuntu:24.04/amd64/1.35-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:24.04/amd64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:24.04/amd64/1.36-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:24.04/amd64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:24.04/amd64/1.37-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:24.04/amd64/latest/edge/unknown-flavor/lxd` | 1 | 68 | 68 |
| `ubuntu:24.04/arm64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 42 |
| `ubuntu:24.04/arm64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 48 |
| `ubuntu:24.04/arm64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 53 |
| `ubuntu:24.04/arm64/1.35-classic/edge/unknown-flavor/lxd` | 63 | 63 | 63 |
| `ubuntu:24.04/arm64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:24.04/arm64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:24.04/arm64/latest/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:26.04/amd64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 42 |
| `ubuntu:26.04/amd64/1.32-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:26.04/amd64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 48 |
| `ubuntu:26.04/amd64/1.33-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:26.04/amd64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 53 |
| `ubuntu:26.04/amd64/1.34-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:26.04/amd64/1.35-classic/edge/unknown-flavor/lxd` | 63 | 63 | 63 |
| `ubuntu:26.04/amd64/1.35-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:26.04/amd64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:26.04/amd64/1.36-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:26.04/amd64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:26.04/amd64/1.37-classic/stable/unknown-flavor/lxd` | 1 | 1 | 1 |
| `ubuntu:26.04/amd64/latest/edge/unknown-flavor/lxd` | 1 | 68 | 68 |
| `ubuntu:26.04/arm64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 42 |
| `ubuntu:26.04/arm64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 48 |
| `ubuntu:26.04/arm64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 53 |
| `ubuntu:26.04/arm64/1.35-classic/edge/unknown-flavor/lxd` | 63 | 63 | 63 |
| `ubuntu:26.04/arm64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:26.04/arm64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `ubuntu:26.04/arm64/latest/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `unknown-os/amd64/latest/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `unknown-os/arm64/latest/edge/unknown-flavor/lxd` | 68 | 68 | 68 |
| `unknown-os/unknown-arch/artifact/unknown-flavor/lxd` | 19 | 19 | 19 |

## Top signatures

| Signature | Occurrences | Runs | Class | Rule | Tests | Configs | Example |
|---|---:|---:|---|---|---:|---:|---|
| `d87b7eabd6dd7094` | 81 | 4 | external.dependency/apt | external-apt | 2 | 24 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386376312) |
| `da6ba7b982998d76` | 54 | 4 | unknown | _unowned_ | 1 | 30 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386405521) |
| `b2fdc843a67d0cf8` | 26 | 3 | unknown | _unowned_ | 1 | 12 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172684836) |
| `53e161f93a3b80b6` | 6 | 1 | unknown | _unowned_ | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724191713/job/116385943967) |
| `7296136bc38d37ff` | 5 | 4 | unknown | _unowned_ | 1 | 5 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386356369) |
| `cfdad54817091717` | 3 | 1 | unknown | _unowned_ | 3 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38488046943/job/115606595549) |
| `edd44007115ea059` | 3 | 1 | unknown | _unowned_ | 1 | 3 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724191713/job/116386101876) |
| `07c6112aab64f685` | 2 | 1 | unknown | _unowned_ | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172569408) |
| `1876eb4438ae5539` | 2 | 2 | product.bug/snap_install_failure | product-snap-install-failure | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386405428) |
| `e0ac6b603e274f49` | 2 | 2 | unknown | _unowned_ | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172512657) |
| `e5c836d564d13b3e` | 2 | 1 | product.bug/snap_install_failure | product-snap-install-failure | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38325492624/job/115063241749) |
| `0e90cd6fe44a85cb` | 1 | 1 | unknown | _unowned_ | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172651379) |
| `2968f86e61478acc` | 1 | 1 | unknown | _unowned_ | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38488046943/job/115606486593) |
| `2d7b7be43282962c` | 1 | 1 | product.bug/snap_install_failure | product-snap-install-failure | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38325492624/job/115063247952) |
| `61f4723c5bc89911` | 1 | 1 | product.bug/snap_install_failure | product-snap-install-failure | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38325492624/job/115063250347) |

## Open CI-filed issues

_No open CI-filed issues._

_Derived from durable facts at report time. Rollup v1._
