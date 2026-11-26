# CI failure report

Data period: `2026-11` (through `2026-11-26T04:54:37Z`)

## Attribution and delta

- Observed failed jobs: 425/14126 = 3.0% (n=14126 jobs) (+1.3 pp)
- Failure pressure including blocked jobs: 3.0% (n=14126 jobs plus blocked) (+1.3 pp)
- Top attribution: `product.bug` owned by product team (84.5% (n=425 facts) of inspected failures).

## Headline metrics

| Period | Runs | Failed jobs | Failure rate | Scheduled green | PR first-pass | Unclassified | M12 fact coverage |
|---|---:|---:|---|---|---|---|---|
| 2026-10 | 8 | 318 | 1.7% (n=18222 jobs) | 0.0% (n=8 runs) | n/a | 5.0% (n=318 facts) | 100.0% (n=318 expected failed jobs) |
| 2026-11 | 5 | 425 | 3.0% (n=14126 jobs) | 0.0% (n=5 runs) | n/a | 15.5% (n=425 facts) | 100.0% (n=425 expected failed jobs) |

## Failure attribution

| Class | Count | Share | Owner |
|---|---:|---|---|
| `product.bug` | 359 | 84.5% (n=425 facts) | product team |
| `unknown` | 66 | 15.5% (n=425 facts) | unowned triage |

## Integrity metrics

These are report metrics, not appendix material: they show whether apparent improvement came from running less.

- M9 unclassified share: 15.5% (n=425 failed job facts) (+10.5 pp)
- M11 jobs blocked by failed `Prepare Environment`: 0 across 0 prepare job(s).
- M12 failed-job fact coverage: 100.0% (n=425 expected failed jobs).
- M12 log availability among failed job facts: 100.0% (n=425 failed job facts).

### M10 tests collected per config

| Config | Min | Median | Max |
|---|---:|---:|---:|
| `ubuntu:22.04/amd64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 66 |
| `ubuntu:22.04/amd64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 71 |
| `ubuntu:22.04/amd64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 81 |
| `ubuntu:22.04/amd64/1.35-classic/edge/unknown-flavor/lxd` | 65 | 65 | 93 |
| `ubuntu:22.04/amd64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 91 |
| `ubuntu:22.04/amd64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 96 |
| `ubuntu:22.04/amd64/latest/edge/unknown-flavor/lxd` | 71 | 73 | 102 |
| `ubuntu:22.04/arm64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 66 |
| `ubuntu:22.04/arm64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 71 |
| `ubuntu:22.04/arm64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 81 |
| `ubuntu:22.04/arm64/1.35-classic/edge/unknown-flavor/lxd` | 65 | 65 | 93 |
| `ubuntu:22.04/arm64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 91 |
| `ubuntu:22.04/arm64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 96 |
| `ubuntu:22.04/arm64/latest/edge/unknown-flavor/lxd` | 71 | 73 | 102 |
| `ubuntu:24.04/amd64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 66 |
| `ubuntu:24.04/amd64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 71 |
| `ubuntu:24.04/amd64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 81 |
| `ubuntu:24.04/amd64/1.35-classic/edge/unknown-flavor/lxd` | 65 | 65 | 93 |
| `ubuntu:24.04/amd64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 91 |
| `ubuntu:24.04/amd64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 96 |
| `ubuntu:24.04/amd64/latest/edge/unknown-flavor/lxd` | 71 | 73 | 102 |
| `ubuntu:24.04/arm64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 66 |
| `ubuntu:24.04/arm64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 71 |
| `ubuntu:24.04/arm64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 81 |
| `ubuntu:24.04/arm64/1.35-classic/edge/unknown-flavor/lxd` | 65 | 65 | 93 |
| `ubuntu:24.04/arm64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 91 |
| `ubuntu:24.04/arm64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 96 |
| `ubuntu:24.04/arm64/latest/edge/unknown-flavor/lxd` | 71 | 73 | 102 |
| `ubuntu:26.04/amd64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 66 |
| `ubuntu:26.04/amd64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 71 |
| `ubuntu:26.04/amd64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 81 |
| `ubuntu:26.04/amd64/1.35-classic/edge/unknown-flavor/lxd` | 65 | 65 | 93 |
| `ubuntu:26.04/amd64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 91 |
| `ubuntu:26.04/amd64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 96 |
| `ubuntu:26.04/amd64/latest/edge/unknown-flavor/lxd` | 71 | 73 | 102 |
| `ubuntu:26.04/arm64/1.32-classic/edge/unknown-flavor/lxd` | 42 | 42 | 66 |
| `ubuntu:26.04/arm64/1.33-classic/edge/unknown-flavor/lxd` | 48 | 48 | 71 |
| `ubuntu:26.04/arm64/1.34-classic/edge/unknown-flavor/lxd` | 53 | 53 | 81 |
| `ubuntu:26.04/arm64/1.35-classic/edge/unknown-flavor/lxd` | 65 | 65 | 93 |
| `ubuntu:26.04/arm64/1.36-classic/edge/unknown-flavor/lxd` | 68 | 68 | 91 |
| `ubuntu:26.04/arm64/1.37-classic/edge/unknown-flavor/lxd` | 68 | 68 | 96 |
| `ubuntu:26.04/arm64/latest/edge/unknown-flavor/lxd` | 71 | 73 | 102 |
| `unknown-os/unknown-arch/artifact/unknown-flavor/lxd` | 19 | 19 | 50 |

## Top signatures

| Signature | Occurrences | Runs | Class | Rule | Tests | Configs | Example |
|---|---:|---:|---|---|---:|---:|---|
| `b23345bfac804d7b` | 24 | 5 | unknown | _unowned_ | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138881993795) |
| `bb40ddc3376ea7c1` | 12 | 1 | product.bug/node_not_ready | product-node-never-ready | 2 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873388968) |
| `eb8fa9ef8c5c6a46` | 12 | 3 | unknown | _unowned_ | 1 | 8 | [job](https://github.com/canonical/k8s-snap/actions/runs/45725848029/job/139595584529) |
| `7a4b63232208221d` | 8 | 3 | product.bug/cluster_never_ready | product-cluster-never-ready | 1 | 7 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873384206) |
| `00b444820009a354` | 6 | 1 | product.bug/node_not_ready | product-node-never-ready | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873390577) |
| `025f8f4cf9ef8622` | 6 | 1 | product.bug/wait_for_timeout | product-wait-for-timeout | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873384044) |
| `08c0bd527e13771e` | 6 | 1 | product.bug/node_not_ready | product-node-never-ready | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873388501) |
| `092cc8cd74024629` | 6 | 1 | product.bug/node_not_ready | product-node-never-ready | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873379151) |
| `0a66efd9a2002f65` | 6 | 1 | product.bug/node_not_ready | product-node-never-ready | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873388047) |
| `0fd8ad952ddf64bc` | 6 | 1 | product.bug/node_not_ready | product-node-never-ready | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873379183) |
| `170c68613040961c` | 6 | 1 | product.bug/join_failure | product-join-failure | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873471686) |
| `1b353994352e5e1c` | 6 | 1 | product.bug/node_not_ready | product-node-never-ready | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873381530) |
| `2ddc73e81db98d31` | 6 | 1 | product.bug/join_failure | product-join-failure | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873383971) |
| `35f73fbfa626fe68` | 6 | 1 | product.bug/wait_for_timeout | product-wait-for-timeout | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873386402) |
| `3a64943759434bc6` | 6 | 1 | unknown | _unowned_ | 1 | 6 | [job](https://github.com/canonical/k8s-snap/actions/runs/45514926115/job/138873382369) |

## Open CI-filed issues

_No open CI-filed issues._

_Derived from durable facts at report time. Rollup v1._
