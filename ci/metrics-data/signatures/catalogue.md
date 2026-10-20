# CI failure signature catalogue

| Signature | Occurrences | Runs | Class | Rule | State | First seen | Last seen | Contexts |
|---|---:|---:|---|---|---|---|---|---|
| `d87b7eabd6dd7094` | 103 | 5 | ci.config/apt_index_missing | ci-apt-index-missing | active | 2026-10-14T00:04:38Z | 2026-10-19T00:05:04Z | lxd:100, multipass:3 |
| `da6ba7b982998d76` | 77 | 7 | product.bug/upgrade_failure | product-feature-upgrade-incomplete | active | 2026-10-13T00:05:23Z | 2026-10-20T00:04:20Z | lxd:77 |
| `b2fdc843a67d0cf8` | 51 | 6 | product.bug/service_panic | product-service-panic | active | 2026-10-13T00:05:23Z | 2026-10-20T00:04:20Z | lxd:51 |
| `7296136bc38d37ff` | 11 | 6 | product.bug/service_restart | product-service-restarts | active | 2026-10-13T00:05:23Z | 2026-10-20T00:04:20Z | lxd:11 |
| `e0ac6b603e274f49` | 10 | 5 | product.bug/service_restart | product-service-restarts | active | 2026-10-13T00:05:23Z | 2026-10-20T00:04:20Z | lxd:10 |
| `53e161f93a3b80b6` | 6 | 1 | external.dependency/docs_links | external-docs-linkcheck | active | 2026-10-15T10:59:57Z | 2026-10-15T10:59:57Z | :6 |
| `31bad0ec48a537c6` | 3 | 2 | product.bug/apiserver_unreachable | product-apiserver-unreachable | active | 2026-10-15T00:04:42Z | 2026-10-18T00:04:44Z | lxd:3 |
| `73cf55cac38adedf` | 3 | 3 | unknown | _unowned_ | active | 2026-10-15T11:00:17Z | 2026-10-19T00:05:04Z | lxd:3 |
| `cfdad54817091717` | 3 | 1 | external.dependency/pypi | external-pypi | active | 2026-10-14T00:04:38Z | 2026-10-14T00:04:38Z | lxd:3 |
| `edd44007115ea059` | 3 | 1 | product.bug/conformance | product-cncf-conformance | active | 2026-10-15T10:59:57Z | 2026-10-15T10:59:57Z | lxd:3 |
| `2968f86e61478acc` | 2 | 2 | product.bug/containerd | product-containerd-image-import | active | 2026-10-14T00:04:38Z | 2026-10-20T00:04:20Z | lxd:2 |
| `35868fe40d79058d` | 2 | 2 | product.bug/service_not_active | product-service-not-active | active | 2026-10-18T00:04:44Z | 2026-10-19T00:05:04Z | lxd:2 |
| `40a71c1cc7cfb11b` | 2 | 1 | external.dependency/snap_store | external-snap-store | active | 2026-10-13T00:05:23Z | 2026-10-13T00:05:23Z | lxd:2 |
| `753ae9e6b747fa57` | 2 | 1 | unknown | _unowned_ | active | 2026-10-18T00:04:44Z | 2026-10-18T00:04:44Z | lxd:2 |
| `baa0ed1c73cbdc79` | 2 | 2 | external.dependency/snap_store | external-snap-store | active | 2026-10-13T00:05:23Z | 2026-10-15T11:00:17Z | lxd:2 |
| `080cc59d6f79599f` | 1 | 1 | unknown | _unowned_ | active | 2026-10-19T00:05:04Z | 2026-10-19T00:05:04Z | lxd:1 |
| `0a3da64f34765623` | 1 | 1 | unknown | _unowned_ | active | 2026-10-19T00:05:04Z | 2026-10-19T00:05:04Z | lxd:1 |
| `0e90cd6fe44a85cb` | 1 | 1 | unknown | _unowned_ | active | 2026-10-15T00:04:42Z | 2026-10-15T00:04:42Z | lxd:1 |
| `238e8243f76ddc60` | 1 | 1 | external.dependency/snap_store | external-snap-store | active | 2026-10-13T00:05:23Z | 2026-10-13T00:05:23Z | lxd:1 |
| `2807cefe257bd4ca` | 1 | 1 | product.bug/join_failure | product-join-failure | active | 2026-10-18T00:04:44Z | 2026-10-18T00:04:44Z | lxd:1 |
| `338a064e4326070e` | 1 | 1 | product.bug/service_restart | product-service-restarts | active | 2026-10-19T00:05:04Z | 2026-10-19T00:05:04Z | lxd:1 |
| `462cdded92e7dfaf` | 1 | 1 | external.dependency/snap_store | external-snap-store | active | 2026-10-13T00:05:23Z | 2026-10-13T00:05:23Z | lxd:1 |
| `744894a2fd3c5e2c` | 1 | 1 | unknown | _unowned_ | active | 2026-10-15T10:59:57Z | 2026-10-15T10:59:57Z | multipass:1 |
| `7a7d07a706e36164` | 1 | 1 | unknown | _unowned_ | active | 2026-10-20T00:04:20Z | 2026-10-20T00:04:20Z | lxd:1 |
| `7c4a0bbf6eec3050` | 1 | 1 | unknown | _unowned_ | active | 2026-10-13T00:05:23Z | 2026-10-13T00:05:23Z | lxd:1 |
| `7d6a50cfd6936b87` | 1 | 1 | unknown | _unowned_ | active | 2026-10-14T00:04:38Z | 2026-10-14T00:04:38Z | lxd:1 |
| `8b9dd86abcd0a895` | 1 | 1 | external.dependency/snap_store | external-snap-store | active | 2026-10-15T11:00:17Z | 2026-10-15T11:00:17Z | lxd:1 |
| `a934445a5476c39d` | 1 | 1 | product.bug/service_not_active | product-service-should-be-inactive | active | 2026-10-18T00:04:44Z | 2026-10-18T00:04:44Z | lxd:1 |
| `a9849d4684ceb073` | 1 | 1 | external.dependency/snap_store | external-snap-store | active | 2026-10-13T00:05:23Z | 2026-10-13T00:05:23Z | lxd:1 |
| `bd6eaa1c1177a033` | 1 | 1 | unknown | _unowned_ | active | 2026-10-20T00:04:20Z | 2026-10-20T00:04:20Z | lxd:1 |
| `cbb87f101d781266` | 1 | 1 | product.bug/node_not_ready | product-node-never-ready | active | 2026-10-13T00:05:23Z | 2026-10-13T00:05:23Z | lxd:1 |
| `ce41aa9f5a1fbfa9` | 1 | 1 | unknown | _unowned_ | active | 2026-10-15T00:04:42Z | 2026-10-15T00:04:42Z | lxd:1 |
| `d722795db15ed246` | 1 | 1 | product.bug/networking | product-gateway-ip-missing | active | 2026-10-14T00:04:38Z | 2026-10-14T00:04:38Z | lxd:1 |
| `de9d743139d8133c` | 1 | 1 | unknown | _unowned_ | active | 2026-10-20T00:04:20Z | 2026-10-20T00:04:20Z | lxd:1 |
| `e444e8bbba287346` | 1 | 1 | product.bug/node_not_ready | product-node-never-ready | active | 2026-10-14T00:04:38Z | 2026-10-14T00:04:38Z | lxd:1 |
| `f1ae9e093d594491` | 1 | 1 | unknown | _unowned_ | active | 2026-10-15T10:59:57Z | 2026-10-15T10:59:57Z | multipass:1 |
