# Annotations

## Machine

The following annotations can be set on CAPI `Machine` resources.

| Name                                          | Description                                          | Values                       | Set by user |
|-----------------------------------------------|------------------------------------------------------|------------------------------|-------------|
| `v1beta2.k8sd.io/in-place-upgrade-to`         | Trigger a Kubernetes version upgrade on that machine | version e.g `v1.31.3`        | yes         |
| `v1beta2.k8sd.io/in-place-upgrade-status`     | The status of the version upgrade                    | in-progress\|done\|failed    | no          |
| `v1beta2.k8sd.io/in-place-upgrade-release`    | The current version on the machine                   | version e.g. `v1.31.3`       | no          |
| `v1beta2.k8sd.io/in-place-upgrade-refresh-id` | The ID of the currently running upgrade              | ID string                    | no          |
