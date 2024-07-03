# How to use the embedded etcd datastore

This guide walks you through bootstrapping a Canonical Kubernetes cluster
using the embedded etcd datastore.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have installed the Canonical Kubernetes snap
  (see How-to [Install Canonical Kubernetes from a snap][snap-install-howto]).
- You have not bootstrapped the Canonical Kubernetes cluster yet

## Adjust the bootstrap configuration

To use the embedded etcd datastore, a configuration file that contains the
required datastore parameters needs to be provided to the bootstrap command.
Create a configuration file and insert the contents below while replacing
the placeholder values based on the configuration of your etcd cluster.

```yaml
# must be set to "etcd"
datastore-type: etcd

# port number that will be used for client traffic (default is 2379)
etcd-port: 2379

# port number that will be used for peer traffic (default is 2380)
etcd-peer-port: 2380

# (optional) custom CA certificate and private key to use to generate TLS
# certificates for the etcd cluster, in PEM format. If not specified, a
# self-signed CA will be used instead.
etcd-ca-crt: |
  -----BEGIN CERTIFICATE-----
  .....
  -----END CERTIFICATE-----

etcd-ca-key: |
  -----BEGIN RSA PRIVATE KEY-----
  .....
  -----END RSA PRIVATE KEY-----
```

```{note}
the embedded etcd cluster will always be configured with TLS.
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
etcd datastore are documented.

### How to use etcdctl

You can interact with the embedded etcd cluster using the standard `etcdctl` CLI
tool. `etcdctl` is not included in Canonical Kubernetes and needs to be
installed separately if needed. To point `etcdctl` to the embedded cluster, you
need to set the following arguments:

```bash
sudo ETCDCTL_API=3 etcdctl \
  --endpoints https://${nodeip}:2379 \
  --cacert /etc/kubernetes/pki/etcd/ca.crt \
  --cert /etc/kubernetes/pki/apiserver-etcd-client.crt \
  --key /etc/kubernetes/pki/apiserver-etcd-client.key \
  member list
```

### Using k8s-dqlite dbctl

There is a `k8s-dqlite dbctl` subcommand that can be used from control
plane nodes to directly interact with the datastore if required. This tool is
supposed to be a lightweight alternative to common `etcdctl` commands:

```bash
sudo /snap/k8s/current/bin/k8s-dqlite dbctl --help
```

Some examples are shown below:

#### List cluster members

```bash
sudo /snap/k8s/current/bin/k8s-dqlite dbctl member list
```

#### Create a database snapshot

```bash
sudo /snap/k8s/current/bin/k8s-dqlite dbctl snapshot save ./file.db
```

The created `file.db` contains a point-in-time backup snapshot of the etcd
cluster, and can be used to restore the cluster if needed.

<!-- LINKS -->

[snap-install-howto]: ./install/snap
