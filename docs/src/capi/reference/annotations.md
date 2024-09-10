# Annotations

Like annotations for other Kubernetes objects, CAPI annotations are key-value
pairs that can be used to reflect additional metadata for CAPI resources.

## Machine

The following annotations can be set on CAPI `Machine` resources.

| Name                                          | Description                                          | Values                       | Set by user |
|-----------------------------------------------|------------------------------------------------------|------------------------------|-------------|
| `v1beta2.k8sd.io/in-place-upgrade-to`         | Trigger a Kubernetes version upgrade on that machine | snap version e.g.:<br>- `localPath=/full/path/to/k8s.snap`<br>- `revision=123`<br>- `channel=latest/edge` | yes |
| `v1beta2.k8sd.io/in-place-upgrade-status`     | The status of the version upgrade                    | in-progress\|done\|failed    | no          |
| `v1beta2.k8sd.io/in-place-upgrade-release`    | The current version on the machine                   | snap version e.g.:<br>- `localPath=/full/path/to/k8s.snap`<br>- `revision=123`<br>- `channel=latest/edge` | no |
| `v1beta2.k8sd.io/in-place-upgrade-change-id`  | The ID of the currently running upgrade              | ID string                    | no          |
