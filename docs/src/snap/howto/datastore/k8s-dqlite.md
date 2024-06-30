# How to use the k8s-dqlite datastore

This guide walks you through bootstrapping a Canonical Kubernetes cluster
using the k8s-dqlite datastore.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have installed the Canonical Kubernetes snap
  (see How-to [Install Canonical Kubernetes from a snap][snap-install-howto]).
- You have not bootstrapped the Canonical Kubernetes cluster yet

## Adjust the bootstrap configuration

k8s-dqlite is the default datastore for Canonical Kubernetes. In case you need
to adjust any of its defaults, create a configuration file and insert the
contents below:

```yaml
# must be set to "k8s-dqlite"
datastore-type: k8s-dqlite

# port number to use for k8s-dqlite peer traffic (default is 9000)
k8s-dqlite-port: 9000
```

## Bootstrap the cluster

The next step is to bootstrap the cluster with our configuration file:

```
sudo k8s bootstrap --file /path/to/config.yaml
```

```{note}
The datastore can only be configured through the `--file` file option,
and is not available in interactive mode.
```

## Confirm the cluster is ready

It is recommended to ensure that the cluster initialises properly and is
running without issues. Run the command:

```
sudo k8s status --wait-ready
```

This command will wait until the cluster is ready and then display
the current status. The command will time-out if the cluster does not reach a
ready state.

## Operations

In the following section, common operations for interacting with the managed
k8s-dqlite datastore are documented.

### How to use the dqlite CLI

You can interact with the dqlite cluster using the `dqlite` CLI like so:

```bash
sudo /snap/k8s/current/bin/dqlite k8s \
  -s "file:///var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml" \
  -c "/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt" \
  -k "/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key"
```

Type `.help` to see a list of available commands.

<!-- LINKS -->

[snap-install-howto]: ./install/snap
