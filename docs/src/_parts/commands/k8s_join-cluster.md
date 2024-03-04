## k8s join-cluster

Join a cluster using the provided token

```
k8s join-cluster <join-token> [flags]
```

### Options

```
      --address string     the address (IP:Port) on which the nodes REST API should be available
  -h, --help               help for join-cluster
      --name string        the name of the joining node. defaults to hostname
      --timeout duration   the max time to wait for the node to be ready (default 1m30s)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

