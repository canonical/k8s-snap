## k8s get-join-token

Create a join token for a node to join the cluster

```
k8s get-join-token <node-name> [flags]
```

### Options

```
  -h, --help                   help for get-join-token
  -o, --output-format string   Specify in which format the output should be printed. One of plain, json or yaml (default "plain")
      --worker                 generate a join token for a worker node
```

### Options inherited from parent commands

```
  -d, --debug     Show all debug messages
  -v, --verbose   Show all information messages (default true)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

