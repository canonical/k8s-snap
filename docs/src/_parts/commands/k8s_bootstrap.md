## k8s bootstrap

Bootstrap a new Kubernetes cluster

### Synopsis

Generate certificates, configure service arguments and start the Kubernetes services.

```
k8s bootstrap [flags]
```

### Options

```
      --address string   microcluster address, defaults to the node IP address
      --file string      path to the YAML file containing your custom cluster bootstrap configuration.
  -h, --help             help for bootstrap
      --interactive      interactively configure the most important cluster options
      --name string      node name, defaults to hostname
```

### Options inherited from parent commands

```
      --output-format string   set the output format to one of plain, json or yaml (default "plain")
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

