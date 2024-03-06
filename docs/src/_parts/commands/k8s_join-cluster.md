## k8s join-cluster

Join a cluster using the provided token

```
k8s join-cluster <join-token> [flags]
```

### Options

```
      --address string   the address (IP:Port) on which the nodes REST API should be available
  -h, --help             help for join-cluster
      --name string      the name of the joining node. defaults to hostname
```

### Options inherited from parent commands

```
  -o, --output-format string   set the output format to one of plain, json or yaml (default "plain")
  -t, --timeout duration       the max time to wait for the command to execute (default 1m30s)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

