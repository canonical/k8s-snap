# Annotations

Like annotations for other Kubernetes objects, CAPI annotations are key-value
pairs that can be used to reflect additional metadata for CAPI resources.

## Machine

The following annotations can be set on CAPI `Machine` resources.

### In-place upgrade

| Name                                                      | Description                                          | Values                                                                                                    | Set by user |
|-----------------------------------------------------------|------------------------------------------------------|-----------------------------------------------------------------------------------------------------------|-------------|
| `v1beta2.k8sd.io/in-place-upgrade-to`                     | Trigger a Kubernetes version upgrade on that machine | snap version e.g.:<br>- `localPath=/full/path/to/k8s.snap`<br>- `revision=123`<br>- `channel=latest/edge` | yes         |
| `v1beta2.k8sd.io/in-place-upgrade-status`                 | The status of the version upgrade                    | in-progress\|done\|failed                                                                                 | no          |
| `v1beta2.k8sd.io/in-place-upgrade-release`                | The current version on the machine                   | snap version e.g.:<br>- `localPath=/full/path/to/k8s.snap`<br>- `revision=123`<br>- `channel=latest/edge` | no          |
| `v1beta2.k8sd.io/in-place-upgrade-change-id`              | The ID of the currently running upgrade              | ID string                                                                                                 | no          |
| `v1beta2.k8sd.io/in-place-upgrade-last-failed-attempt-at` | The time of the last failed upgrade attempt          | RFC1123Z timestamp                                                                                        | no          |

### Refresh certificates

| Name                                          | Description                                                                    | Values                                                                                                                                                                                      | Set by user |
|-----------------------------------------------|--------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|
| `v1beta2.k8sd.io/refresh-certificates`        | The requested duration (TTL) that the refreshed certificates should expire in. | Duration (TTL) string. A number followed by a unit e.g.: `1mo`, `1y`, `90d`<br>Allowed units: Any unit supported by `time.ParseDuration` as well as `y` (year), `mo` (month) and `d` (day). | yes         |
| `v1beta2.k8sd.io/refresh-certificates-status` | The status of the certificate refresh request.                                 | in-progress\|done\|failed                                                                                                                                                                   | no          |

### Certificates expiry

| Name                                           | Description                                    | Values            | Set by user |
|------------------------------------------------|------------------------------------------------|-------------------|-------------|
| `machine.cluster.x-k8s.io/certificates-expiry` | Indicates the expiry date of the certificates. | RFC3339 timestamp | no          |

### Remediation

| Name                                                      | Description                                                   | Values      | Set by user |
|-----------------------------------------------------------|---------------------------------------------------------------|-------------|-------------|
| `controlplane.cluster.x-k8s.io/ck8s-server-configuration` | Stores the json-marshalled string of KCP ClusterConfiguration | JSON string | no          |
| `controlplane.cluster.x-k8s.io/remediation-in-progress`   | Keeps track that a KCP remediation is in progress             | JSON string | no          |
| `controlplane.cluster.x-k8s.io/remediation-for`           | Links a new machine to the unhealthy machine it is replacing  | JSON string | no          |
