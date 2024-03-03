## k8s join-cluster

Join a cluster

```
k8s join-cluster <join-token> [flags]
```

### Options

```
      --address string     The address (IP:Port) on which the nodes REST API should be available
  -h, --help               help for join-cluster
      --name string        The name of the joining node. defaults to hostname
      --timeout duration   The max time to wait for the node to be ready. (default 1m30s)
```

### Options inherited from parent commands

```
  -d, --debug     Show all debug messages
  -v, --verbose   Show all information messages (default true)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

