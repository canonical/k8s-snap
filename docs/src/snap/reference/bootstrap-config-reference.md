# Bootstrap configuration file reference

A YAML file can be supplied to the `k8s join-cluster` command to configure and
customise the cluster. This reference section provides the format of this file
by listing all available options and their details. See below for an example.

## Configuration options

```{include} ../../_parts/bootstrap_config.md
```


## Example

The following example configures and enables certain features, sets an external
cloud provider, marks the control plane nodes as unschedulable, changes the pod
and service CIDRs from the defaults and adds an extra SAN to the generated
certificates. It is also available to download [here][example-config].

```{literalinclude} /src/assets/example-bootstrap-config.yaml
:language: yaml
```

<!-- LINKS -->
[example-config]: /src/assets/example-bootstrap-config.yaml
