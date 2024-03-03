## k8s status

Retrieve the current status of the cluster

```
k8s status [flags]
```

### Options

```
      --format string      Specify in which format the output should be printed. One of plain, json or yaml (default "plain")
  -h, --help               help for status
      --timeout duration   The max time to wait for the K8s API server to be ready. (default 1m30s)
      --wait-ready         If set, the command will block until at least one cluster node is ready.
```

### Options inherited from parent commands

```
  -d, --debug     Show all debug messages
  -v, --verbose   Show all information messages (default true)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

