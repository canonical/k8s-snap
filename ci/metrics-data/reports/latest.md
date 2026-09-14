# CI failure trend

## Headline metrics

| Period | Runs | Jobs | Failed | Fail % | Scheduled green | PR first-pass | Unclassified | Inspected | Top-5 conc. |
|---|---|---|---|---|---|---|---|---|---|
| 2026-09 | 111 | 29610 | 1519 | 5.1% | 0.0% | 8.8% | 8.7% | 100.0% | 80.3% |

## Failure attribution -- 2026-09

| Class | Count | Share of inspected | Owner |
|---|---|---|---|
| `product.bug` | 1276 | 84.0% | product team |
| `unknown` | 132 | 8.7% | unowned -- needs triage |
| `external.dependency` | 37 | 2.4% | external -- usually wait/retry |
| `ci.config` | 30 | 2.0% | CI owners |
| `infra.runner` | 24 | 1.6% | CI/infra |
| `infra.provisioning` | 11 | 0.7% | CI/infra |
| `test.bug` | 9 | 0.6% | test owners |

## Top signatures

| Signature | Count | Class | Rule | Age (d) | Tests | Example |
|---|---|---|---|---|---|---|
| `c55cefd2ca3f28e7` | 720 | product.bug/service_not_active | product-node-never-ready | 7 | 49 | [job](https://github.com/canonical/k8s-snap/actions/runs/34222910032/job/102055422547) |
| `8c531e6ba49b3382` | 181 | product.bug/upgrade_failure | product-retry-condition-never-met | 7 | 8 | [job](https://github.com/canonical/k8s-snap/actions/runs/34203776664/job/101993827926) |
| `38c23bd3a1bab957` | 148 | product.bug/service_not_active | product-kube-proxy-inactive | 7 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/34481480637/job/102893904891) |
| `3f2108396bd179b8` | 74 | product.bug/wait_for_timeout | product-wait-for-timeout | 5 | 7 | [job](https://github.com/canonical/k8s-snap/actions/runs/34338417019/job/102428642024) |
| `35b12bf41b664182` | 40 | product.bug/service_not_active | product-service-restarts | 7 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/34572278574/job/103869708381) |
| `ebd66004560a0708` | 25 | unknown/None | _unowned_ | 4 | 13 | [job](https://github.com/canonical/k8s-snap/actions/runs/34454500572/job/102802261908) |
| `0a7d38d398561dc3` | 24 | product.bug/join_failure | product-join-failure | 5 | 7 | [job](https://github.com/canonical/k8s-snap/actions/runs/34338421987/job/102428571751) |
| `7812166adb4d179b` | 24 | product.bug/cluster_never_ready | product-cluster-never-ready | 5 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/34338417019/job/102428642025) |
| `e78d17474b492b14` | 24 | external.dependency/snap_store | external-snap-store | 5 | 3 | [job](https://github.com/canonical/k8s-snap/actions/runs/34349363309/job/102464759410) |
| `2aa7b2c90cc3da81` | 21 | product.bug/join_failure | product-join-failure | 6 | 4 | [job](https://github.com/canonical/k8s-snap/actions/runs/34203776664/job/101993828002) |
| `3a51e6072640676d` | 16 | product.bug/wait_for_timeout | product-wait-for-timeout | 5 | 2 | [job](https://github.com/canonical/k8s-snap/actions/runs/34338421987/job/102428571847) |
| `0c4e7cc7cce4959c` | 12 | unknown/None | _unowned_ | 7 | 0 | [job](https://github.com/canonical/k8s-snap/actions/runs/34071033195/job/101588289507) |
| `5be67c2063b467fe` | 9 | test.bug/timeout_too_short | test-timeout-too-short | 6 | 0 | [job](https://github.com/canonical/k8s-snap/actions/runs/34203883006/job/101993843416) |
| `3a64943759434bc6` | 6 | unknown/None | _unowned_ | 5 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/34338421987/job/102428572044) |
| `867525576ccca976` | 6 | product.bug/upgrade_failure | product-service-restarts-rollback | 6 | 1 | [job](https://github.com/canonical/k8s-snap/actions/runs/34454500572/job/102802188983) |

## Integrity

These guard against apparent improvement caused by running less, rather than by breaking less.

- Unclassified rate: 8.7%
- Jobs never run due to an upstream failure: 0
- Failed `Prepare Environment` jobs: 0
- Quarantined signatures: 0

_Generated from rollup v1 · ruleset v1. Rollups produced under different ruleset versions are not directly comparable._