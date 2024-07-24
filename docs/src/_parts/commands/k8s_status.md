## k8s status

Retrieve the current status of the cluster

### Synopsis

Retrieve the current status of the cluster. Also, The status for each
core feature is described with the following fields:
- **Enabled**: Specifies whether the feature is successfully deployed (does not guarantee that the feature is working as expected).
- **Message**: Describes the status in a human readable form. This field is only
meant to be informative and should not be programmatically parsed in any way.
- **Version**: Version of the feature.
- **UpdatedAt**: Timestamp of the lastest update to the feature status.

Can be used with `--output-format=plain` 
to show a compact, human readable output, or `--output-format=json` or `yaml` flag to show 
more information.

Example with `--output-format=plain` would be:
```text
cluster status:           ready
control plane nodes:      10.97.72.156:6400 (voter)
high availability:        no
datastore:                k8s-dqlite
network:                  enabled
dns:                      enabled at 10.152.183.26
ingress:                  enabled
load-balancer:            enabled, L2 mode
local-storage:            enabled at /var/snap/k8s/common/rawfile-storage
gateway                   enabled
```

```
k8s status [flags]
```

### Options

```
  -h, --help                   help for status
      --output-format string   set the output format to one of plain, json or yaml (default "plain")
      --timeout duration       the max time to wait for the command to execute (default 1m30s)
      --wait-ready             wait until at least one cluster node is ready
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

