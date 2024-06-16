# How to use the embedded datastore

Canonical Kubernetes supports using an embedded datastore such as etcd
instead of the bundled dqlite datastore.
This guide walks you through configuring the embedded etcd datastore.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have installed the Canonical Kubernetes snap
  (see How-to [Install Canonical Kubernetes from a snap][snap-install-howto]).
- You have not bootstrapped the Canonical Kubernetes cluster yet

```{warning}
The selection of the backing datastore can only be changed during the bootstrap process.
There is no migration path between the bundled dqlite and the embedded datastore.
```

## Adjust the bootstrap configuration

To use the embedded datastore, a configuration file that contains the required
datastore parameters needs to be provided to the bootstrap command.
Create a configuration file and insert the contents below while replacing
the placeholder values based on the desired configuration.

```yaml
# required
datastore-type: embedded

# optional
datastore-embedded-port: 2379
datastore-embedded-peer-port: 2380
datastore-embedded-ca-crt: |
  <etcd-root-ca-certificate>
datastore-embedded-ca-key: |
  <etcd-root-ca-private-key>
```

* `datastore-type` must be set to `embedded`.
* `datastore-embedded-port` expects a port number that will be used for the
  client URLs of the embedded cluster. The default is port `2379`.
* `datastore-embedded-peer-port` expects a port number that will be used for
  the peer URLs of the embedded cluster. The default is port `2380`.
* `datastore-embedded-ca-crt` and `datastore-embedded-ca-key`: an optional
  custom CA certificate and private key to use for the embedded etcd cluster.
  If not specified, k8sd will automatically generate a self-signed CA.

```{note}
the embedded datastore will always be configured with TLS.
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

## Clustering

Control plane nodes that join with `k8s join-cluster` will be added as members
to the embedded datastore. Nodes that join the cluster start by registering
themselves in the cluster.

During the transition from 1 to 2 control plane nodes, the embedded cluster
will temporarily freeze and reject write operations, since after moving to a
2-node quorum, both nodes must be available for the raft protocol to proceed.

When removing a node with `k8s remove-node`, the node will also be removed
from the embedded datastore.

## Interacting with the embedded datastore

Under normal operation, you should not have to interact with the embedded
datastore directly. However, in case it is needed for debugging problems, or
operations like creating a cluster backup, you can use either `etcdctl` or
the built-in `k8s-dqlite embeddedctl` commands.

### Using etcdctl

You can interact with the embedded datastore using the standard `etcdctl` CLI
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

### Using k8s-dqlite embeddedctl

There is a `k8s-dqlite embeddedctl` subcommand that can be used from control
plane nodes to directly interact with the datastore if required. This tool is
supposed to be a lightweight alternative to common `etcdctl` commands, and
comes with the following subcommands:

- `k8s-dqlite embeddedctl member {add,list,remove}`: list the current cluster
  members, and add or remove a cluster member.
- `k8s-dqlite embeddedctl snapshot save "file.db"`: create a backup snapshot of
  the current database state and save it in `file.db`. this backup can then be
  used to do a point-in-time restoration of the database using `etcdutl`.

You can access the `k8s-dqlite embeddedctl` commands as shown below. Specify
the `--help` argument to see all available commands and supported arguments:

```bash
sudo /snap/k8s/current/bin/k8s-dqlite embeddedctl --help
```

<!-- LINKS -->

[snap-install-howto]: ./install/snap
