## k8s set

Set cluster configuration

### Synopsis

Configure one of network, dns, gateway, ingress, local-storage, load-balancer.
Use `k8s get` to explore configuration options.

```
k8s set <feature.key=value> ... [flags]
```

### Options

```
  -h, --help                   help for set
      --output-format string   set the output format to one of plain, json or yaml (default "plain")
      --timeout duration       the max time to wait for the command to execute (default 1m30s)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

