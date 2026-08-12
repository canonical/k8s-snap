---
myst:
  html_meta:
    description: "How to configure a Canonical Kubernetes cluster using Juju, including DNS, networking, labels, and feature flags via charm configuration options."
---

# How to configure a {{ product }} cluster using Juju

This guide provides instructions for configuring a {{ product }} cluster using
Juju. The DNS feature is used as an example to demonstrate the various
configuration patterns and methods.

## Prerequisites

This guide assumes the following:

- [Juju][juju install] CLI installed on your machine
- A working Kubernetes cluster deployed with the `k8s` charm

## Understand the charm configuration

The `k8s` charm offers a wide range of configurable options and features
including networking, DNS, labels, taints, and more. Review the charm
documentation for [k8s][k8s configuration] and
[k8s-worker][k8s-worker configuration] to explore all available options
for customizing your cluster.

```{important}
When setting up your cluster for the first time, you will have access to
certain configuration settings (prefixed with `bootstrap-`) that cannot be
changed later. Be sure to check the charm's documentation beforehand to
understand the available configuration options.
```

The charm's configuration options include:

- **Cluster features**, prefixed by the feature name (e.g., DNS, gateway,
  ingress):
  - An enable/disable flag (e.g., `dns-enabled`)
  - Feature specific configuration options (e.g., `dns-cluster-domain`)
- **Cluster wide configurations** (e.g., labels, taints, and annotations).

### Cluster annotations

```{note}
v1alpha annotations are experimental and subject to change or removal in
future {{product}} releases.
```

The `cluster-annotations` config option sets the same
[cluster annotations][snap-annotations] used by the {{product}} snap. It
accepts a space-separated list of `key=value` pairs, and is only applied when
non-empty — clearing it (`juju config k8s --reset cluster-annotations`)
leaves any existing cluster annotations untouched rather than deleting them.

Some annotations, such as the [multi-peer BGP annotation][multi-peer-bgp],
require nested YAML that `cluster-annotations` cannot express. Set these by
running the `k8s` CLI on a unit instead:

```
juju exec --unit <k8s/unit#> -- sudo k8s set annotations="$(cat bgp-peers.yaml)"
```

## Apply the configuration

You can configure your cluster either during the initial deployment or by
updating an existing deployment.

### Option 1: During initial deployment

Use a YAML file or the `--config` flag to specify your desired configuration
during deployment.

**Using a YAML file:**

Create a configuration file with the desired configuration options. For example,
to enable DNS and set the cluster domain to `cluster.local`, create a file
`basic-config.yaml` with the following content:

```yaml
k8s:
  dns-enabled: true
  dns-cluster-domain: "cluster.local"
```

Deploy the `k8s` charm with the configuration file:

```
juju deploy k8s --base="ubuntu@24.04" --config ./basic-config.yaml
```

**Using the `--config` flag:**

Alternatively, deploy the application by specifying the configuration directly:

```
juju deploy k8s --base="ubuntu@24.04" --config dns-enabled=true --config dns-cluster-domain=cluster.local
```

### Option 2: Update an existing deployment

Modify the configuration of an existing deployment using a YAML file or the
`--config` flag.

**Using a YAML file:**

Apply the configuration from a YAML file:

```
juju config k8s --file ./basic-config.yaml
```

**Using the `--config` flag:**

Specify the configuration options directly:

```
juju config k8s dns-enabled=true dns-cluster-domain=cluster.local
```

### Monitor and verify the configuration

After applying the configuration, the charm will automatically apply the changes
and update the cluster. Monitor the progress by running:

```
juju status --watch 1s
```

Check the current configuration values with:

```
juju config k8s
juju config k8s-worker
```

<!-- LINKS -->
[juju install]: https://juju.is/docs/juju/install-and-manage-the-client
[k8s configuration]: https://charmhub.io/k8s/configurations
[k8s-worker configuration]: https://charmhub.io/k8s-worker/configurations
[snap-annotations]: /snap/reference/annotations.md
[multi-peer-bgp]: /snap/howto/networking/multi-peer-bgp.md
