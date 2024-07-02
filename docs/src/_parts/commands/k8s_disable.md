## k8s disable

Disable core cluster features

### Synopsis

Disable one of network, dns, gateway, ingress, local-storage, load-balancer.

```
k8s disable [network|dns|gateway|ingress|local-storage|load-balancer] ... [flags]
```

### Options

```
  -h, --help                   help for disable
      --output-format string   set the output format to one of plain, json or yaml (default "plain")
      --timeout duration       the max time to wait for the command to execute (default 1m30s)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI
