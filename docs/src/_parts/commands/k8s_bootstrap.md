## k8s bootstrap

Bootstrap a k8s cluster on this node

### Synopsis

Initialize the necessary folders, permissions, service arguments, certificates and start up the Kubernetes services.

```
k8s bootstrap [flags]
```

### Options

```
      --config string   path to the YAML file containing your custom cluster bootstrap configuration.
  -h, --help            help for bootstrap
      --interactive     interactively configure the most important cluster options
```

### Options inherited from parent commands

```
  -o, --output-format string   set the output format to one of plain, json or yaml (default "plain")
  -t, --timeout duration       the max time to wait for the command to execute (default 1m30s)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

