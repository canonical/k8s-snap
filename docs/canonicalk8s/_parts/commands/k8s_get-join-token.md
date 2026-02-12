## k8s get-join-token

Create a token for a node to join the cluster

### Synopsis

Create a join token that allows a new node to join the cluster. The node name is required for control plane nodes but optional for worker nodes.

```
k8s get-join-token <node-name> [flags]
```

### Options

```
      --expires-in duration   the time until the token expires (default 24h0m0s)
  -h, --help                  help for get-join-token
      --timeout duration      the max time to wait for the command to execute (default 1m30s)
      --worker                generate a join token from a cluster control plane for a worker node. Specifying the node name of the worker node is optional
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

