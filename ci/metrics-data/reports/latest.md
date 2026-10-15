# CI failure report

Data period: `2026-10` (through `2026-10-15T16:31:53Z`)

## Attribution and delta

- Observed failed jobs: 213/10557 = 2.0% (n=10557 jobs)
- Failure pressure including blocked jobs: 2.0% (n=10557 jobs plus blocked)
- Top attribution: `unknown` owned by unowned triage (90.1% (n=213 facts) of inspected failures).

## Headline metrics

| Period | Runs | Failed jobs | Failure rate | Scheduled green | PR first-pass | Unclassified | M12 fact coverage |
|---|---:|---:|---|---|---|---|---|
| 2026-10 | 5 | 213 | 2.0% (n=10557 jobs) | 0.0% (n=5 runs) | n/a | 90.1% (n=213 facts) | 100.0% (n=213 expected failed jobs) |

## Failure attribution

| Class | Count | Share | Owner |
|---|---:|---|---|
| `unknown` | 192 | 90.1% (n=213 facts) | unowned triage |
| `product.bug` | 9 | 4.2% (n=213 facts) | product team |
| `external.dependency` | 6 | 2.8% (n=213 facts) | external dependency |
| `infra.provisioning` | 4 | 1.9% (n=213 facts) | CI/infra |
| `infra.runner` | 2 | 0.9% (n=213 facts) | CI/infra |

## Integrity metrics

These are report metrics, not appendix material: they show whether apparent improvement came from running less.

- M9 unclassified share: 90.1% (n=213 failed job facts)
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
| `48188aba0f2b3176` | 81 | 4 | unknown | _unowned_ | 2 | 24 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386208355) |
| `30f9c0739d5c01a3` | 56 | 4 | unknown | _unowned_ | 3 | 30 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386234043) |
| `36be75d243c93b39` | 26 | 3 | unknown | _unowned_ | 1 | 12 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172507965) |
| `12b87410a67e88a1` | 6 | 1 | unknown | _unowned_ | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724191713/job/116385933789) |
| `52b477a901821adc` | 5 | 4 | unknown | _unowned_ | 1 | 5 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386237871) |
| `2be89bfd2b409c9c` | 3 | 1 | unknown | _unowned_ | 3 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38488046943/job/115606587247) |
| `611cf5b67506e5cf` | 3 | 1 | unknown | _unowned_ | 1 | 3 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724191713/job/116386063634) |
| `625ed3afafaceeaa` | 2 | 2 | unknown | _unowned_ | 2 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38488046943/job/115606540841) |
| `7182be53fd6c8293` | 2 | 1 | product.bug/snap_install_failure | product-snap-install-failure | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38325492624/job/115063228415) |
| `86320dba2e89666c` | 2 | 2 | unknown | _unowned_ | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172512657) |
| `c6b1e9b8ae19cfed` | 2 | 2 | product.bug/snap_install_failure | product-snap-install-failure | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386405428) |
| `e32ac84cc9437b1f` | 2 | 1 | unknown | _unowned_ | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172550701) |
| `002ffa955d3a1a22` | 1 | 1 | unknown | _unowned_ | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172651379) |
| `257f79da332b464b` | 1 | 1 | unknown | _unowned_ | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172538016) |
| `2860635ab44b8e63` | 1 | 1 | unknown | _unowned_ | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724191713/job/116386139293) |

## Open CI-filed issues

_No open CI-filed issues._

_Derived from durable facts at report time. Rollup v1._
