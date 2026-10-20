# CI failure report

Data period: `2026-10` (through `2026-10-20T09:32:24Z`)

## Attribution and delta

- Observed failed jobs: 318/18222 = 1.7% (n=18222 jobs)
- Failure pressure including blocked jobs: 1.7% (n=18222 jobs plus blocked)
- Top attribution: `product.bug` owned by product team (51.9% (n=318 facts) of inspected failures).

## Headline metrics

| Period | Runs | Failed jobs | Failure rate | Scheduled green | PR first-pass | Unclassified | M12 fact coverage |
|---|---:|---:|---|---|---|---|---|
| 2026-10 | 8 | 318 | 1.7% (n=18222 jobs) | 0.0% (n=8 runs) | n/a | 5.0% (n=318 facts) | 100.0% (n=318 expected failed jobs) |

## Failure attribution

| Class | Count | Share | Owner |
|---|---:|---|---|
| `product.bug` | 165 | 51.9% (n=318 facts) | product team |
| `ci.config` | 103 | 32.4% (n=318 facts) | CI owners |
| `external.dependency` | 23 | 7.2% (n=318 facts) | external dependency |
| `unknown` | 16 | 5.0% (n=318 facts) | unowned triage |
| `infra.runner` | 7 | 2.2% (n=318 facts) | CI/infra |
| `infra.provisioning` | 4 | 1.3% (n=318 facts) | CI/infra |

## Integrity metrics

These are report metrics, not appendix material: they show whether apparent improvement came from running less.

- M9 unclassified share: 5.0% (n=318 failed job facts)
- M11 jobs blocked by failed `Prepare Environment`: 0 across 0 prepare job(s).
- M12 failed-job fact coverage: 100.0% (n=318 expected failed jobs).
- M12 log availability among failed job facts: 94.7% (n=318 failed job facts).

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
| `d87b7eabd6dd7094` | 103 | 5 | ci.config/apt_index_missing | ci-apt-index-missing | 2 | 24 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386376312) |
| `da6ba7b982998d76` | 77 | 7 | product.bug/upgrade_failure | product-feature-upgrade-incomplete | 1 | 30 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386405521) |
| `b2fdc843a67d0cf8` | 51 | 6 | product.bug/service_panic | product-service-panic | 1 | 12 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172684836) |
| `7296136bc38d37ff` | 11 | 6 | product.bug/service_restart | product-service-restarts | 1 | 10 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386356369) |
| `e0ac6b603e274f49` | 10 | 5 | product.bug/service_restart | product-service-restarts | 1 | 9 | [job](https://github.com/canonical/k8s-snap/actions/runs/38657906287/job/116172512657) |
| `53e161f93a3b80b6` | 6 | 1 | external.dependency/docs_links | external-docs-linkcheck | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724191713/job/116385943967) |
| `31bad0ec48a537c6` | 3 | 2 | product.bug/apiserver_unreachable | product-apiserver-unreachable | 1 | 3 | [job](https://github.com/canonical/k8s-snap/actions/runs/39081428042/job/117539067980) |
| `73cf55cac38adedf` | 3 | 3 | unknown | _unowned_ | 2 | 3 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386267312) |
| `cfdad54817091717` | 3 | 1 | external.dependency/pypi | external-pypi | 3 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38488046943/job/115606595549) |
| `edd44007115ea059` | 3 | 1 | product.bug/conformance | product-cncf-conformance | 1 | 3 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724191713/job/116386101876) |
| `2968f86e61478acc` | 2 | 2 | product.bug/containerd | product-containerd-image-import | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38488046943/job/115606486593) |
| `35868fe40d79058d` | 2 | 2 | product.bug/service_not_active | product-service-not-active | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/39179160270/job/117828258388) |
| `40a71c1cc7cfb11b` | 2 | 1 | external.dependency/snap_store | external-snap-store | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38325492624/job/115063241749) |
| `753ae9e6b747fa57` | 2 | 1 | unknown | _unowned_ | 2 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/39081428042/job/117539051480) |
| `baa0ed1c73cbdc79` | 2 | 2 | external.dependency/snap_store | external-snap-store | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/38724235406/job/116386405428) |

## Open CI-filed issues

_No open CI-filed issues._

_Derived from durable facts at report time. Rollup v1._
