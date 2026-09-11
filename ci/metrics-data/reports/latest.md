# CI failure trend

## Headline metrics

| Period | Runs | Jobs | Failed | Fail % | Scheduled green | PR first-pass | Unclassified | Inspected | Top-5 conc. |
|---|---|---|---|---|---|---|---|---|---|
| 2026-09 | 53 | 8939 | 554 | 6.2% | 0.0% | 7.8% | 24.4% | 100.0% | 81.8% |

## Failure attribution -- 2026-09

| Class | Count | Share of inspected | Owner |
|---|---|---|---|
| `product.bug` | 391 | 70.6% | product team |
| `unknown` | 135 | 24.4% | unowned -- needs triage |
| `external.dependency` | 12 | 2.2% | external -- usually wait/retry |
| `infra.provisioning` | 8 | 1.4% | CI/infra |
| `infra.runner` | 8 | 1.4% | CI/infra |

## Top signatures

| Signature | Count | Class | Rule | Age (d) | Tests | Example |
|---|---|---|---|---|---|---|
| `c55cefd2ca3f28e7` | 288 | product.bug/service_not_active | product-node-never-ready | 1 | 49 | [job](https://github.com/canonical/k8s-snap/actions/runs/34452508229/job/102796001770) |
| `8c531e6ba49b3382` | 52 | product.bug/upgrade_failure | product-retry-condition-never-met | 1 | 3 | [job](https://github.com/canonical/k8s-snap/actions/runs/34452508229/job/102796004176) |
| `38c23bd3a1bab957` | 41 | product.bug/service_not_active | product-kube-proxy-inactive | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/34481480637/job/102893904891) |
| `3f2108396bd179b8` | 34 | unknown/None | _unowned_ | 1 | 7 | [job](https://github.com/canonical/k8s-snap/actions/runs/34464281672/job/102834393656) |
| `ebd66004560a0708` | 25 | unknown/None | _unowned_ | 1 | 13 | [job](https://github.com/canonical/k8s-snap/actions/runs/34454500572/job/102802261908) |
| `7812166adb4d179b` | 11 | unknown/None | _unowned_ | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/34464281672/job/102834393671) |
| `e78d17474b492b14` | 9 | external.dependency/snap_store | external-snap-store | 1 | 3 | [job](https://github.com/canonical/k8s-snap/actions/runs/34454500572/job/102802265906) |
| `35b12bf41b664182` | 8 | product.bug/service_not_active | product-service-restarts | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/34419715087/job/102692380079) |
| `3a51e6072640676d` | 8 | unknown/None | _unowned_ | 1 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/34464281672/job/102834393426) |
| `8a9224850d7c13ce` | 7 | unknown/None | _unowned_ | 1 | 0 | [job](https://github.com/canonical/k8s-snap/actions/runs/34452508229/job/102795886230) |
| `0c4e7cc7cce4959c` | 5 | unknown/None | _unowned_ | 1 | 0 | [job](https://github.com/canonical/k8s-snap/actions/runs/34422507815/job/102700789471) |
| `35868fe40d79058d` | 2 | unknown/None | _unowned_ | 0 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/34544940175/job/103095542454) |
| `3a64943759434bc6` | 2 | unknown/None | _unowned_ | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/34464281672/job/102834393686) |
| `5be67c2063b467fe` | 2 | unknown/None | _unowned_ | 1 | 0 | [job](https://github.com/canonical/k8s-snap/actions/runs/34464281672/job/102834254131) |
| `867525576ccca976` | 2 | product.bug/upgrade_failure | product-service-restarts-rollback | 1 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/34454500572/job/102802188983) |

## Integrity

These guard against apparent improvement caused by running less, rather than by breaking less.

- Unclassified rate: 24.4%
- Jobs never run due to an upstream failure: 0
- Failed `Prepare Environment` jobs: 0
- Quarantined signatures: 0

_Generated from rollup v1 · ruleset v1. Rollups produced under different ruleset versions are not directly comparable._