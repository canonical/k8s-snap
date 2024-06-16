# How to use an external datastore

Canonical Kubernetes supports using an external datastore such as etcd
instead of the bundled dqlite datastore.
This guide walks you through configuring an external etcd datastore.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have an external etcd cluster
- You have installed the Canonical Kubernetes snap
  (see How-to [Install Canonical Kubernetes from a snap][snap-install-howto]).
- You have not bootstrapped the Canonical Kubernetes cluster yet

```{warning}
The selection of the backing datastore can only be changed during the bootstrap process.
There is no migration path between the bundled dqlite and the external datastores.
```

## Adjust the bootstrap configuration

To use an external datastore, a configuration file that contains the required
datastore parameters needs to be provided to the bootstrap command.
Create a configuration file and insert the contents below while replacing
the placeholder values based on the configuration of your etcd cluster.

```yaml
datastore-type: external
datastore-url: "<etcd-member-addresses>"
datastore-ca-crt: |
  <etcd-root-ca-certificate>
datastore-client-crt: |
  <etcd-client-certificate>
datastore-client-key: |
  <etcd-client-key>
```

* `datastore-url` expects a comma seperated list of addresses
  (e.g. `https://10.42.254.192:2379,https://10.42.254.193:2379,https://10.42.254.194:2379`)
* `datastore-ca-crt` expects a certificate for the CA in PEM format
* `datastore-client-crt` expects a certificate that's signed by the root CA
  for the client in PEM format
* `datastore-client-key` expects a key for the client in PEM format

```{note}
`datastore-ca-crt`, `datastore-client-crt` and `datastore-client-key` options
can be omitted if the etcd cluster is not configured to use secure connections.
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

<!-- LINKS -->

[snap-install-howto]: ./install/snap
