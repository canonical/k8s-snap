## k8s bootstrap

Bootstrap a k8s cluster on this node.

### Synopsis

Initialize the necessary folders, permissions, service arguments, certificates and start up the Kubernetes services.

```
k8s bootstrap [flags]
```

### Options

```
  -h, --help               help for bootstrap
      --interactive        Interactively configure the most important cluster options.
      --timeout duration   The max time to wait for k8s to bootstrap. (default 1m30s)
```

### Options inherited from parent commands

```
  -d, --debug     Show all debug messages
  -v, --verbose   Show all information messages (default true)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

