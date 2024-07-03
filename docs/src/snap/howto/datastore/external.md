# How to use an external datastore

This guide walks you through bootstrapping a Canonical Kubernetes cluster
using an external etcd datastore.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have an external etcd cluster
- You have installed the Canonical Kubernetes snap
  (see How-to [Install Canonical Kubernetes from a snap][snap-install-howto]).
- You have not bootstrapped the Canonical Kubernetes cluster yet

## Adjust the bootstrap configuration

To use an external datastore, a configuration file that contains the required
datastore parameters needs to be provided to the bootstrap command.
Create a configuration file and insert the contents below while replacing
the placeholder values based on the configuration of your etcd cluster.

```yaml
# must be set to "external"
datastore-type: external

# comma-seperated list of etcd server URLs
# datastore-url: "https://10.0.0.11:2379,htps://10.0.0.12:2379"
datastore-url: "<etcd-member-addresses>"

# CA certificate for the etcd cluster, in PEM format.
datastore-ca-crt: |
  -----BEGIN CERTIFICATE-----
  .....
  -----END CERTIFICATE-----

# Client certificate and private key to authenticate with the etcd cluster, in
# PEM format. Must be signed by the CA certificate.
datastore-client-crt: |
  -----BEGIN CERTIFICATE-----
  .....
  -----END CERTIFICATE-----

datastore-client-key: |
  -----BEGIN RSA PRIVATE KEY-----
  .....
  -----END RSA PRIVATE KEY-----
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

In the following section, common operations for managing the external datastore
are documented.

### Edit the etcd servers or client certificates

When using an external datastore, it is possible that the etcd server URLs
change, or that the client certificates need to be rotated. In that case, to
update the etcd credentials used by the cluster, the following steps are
required:

1.  Create a `config.json` file with the new certificates and list of etcd servers:

    ```json
    {
      "datastore": {
        "type": "external",
        "servers": "https://10.0.0.11:2379,https://10.0.0.12:2379,https://10.0.0.13:2379",
        "ca-crt": "------BEGIN CERTIFICATE------\n.....\n-----END CERTIFICATE-----",
        "client-crt": "------BEGIN CERTIFICATE------\n.....\n-----END CERTIFICATE-----",
        "client-key": "------BEGIN RSA PRIVATE KEY------\n.....\n-----END RSA PRIVATE KEY-----",
      }
    }
    ```

2.  Apply the new configuration using the k8sd API directly. You must run this
    command on _one_ of the control plane nodes, as Canonical Kubernetes will
    sync the changes to other cluster nodes as needed:

    ```bash
    curl \
      -X PUT \
      --header "Content-type: application/json" \
      --data @config.json \
      --unix-socket /var/snap/k8s/common/var/lib/k8sd/state/control.socket \
      http://localhost/1.0/k8sd/cluster/config
    ```

You can verify the changes have been applied by looking at the following files:

- `/var/snap/k8s/common/args/kube-apiserver` should have the new etcd servers.
- `/etc/kubernetes/pki/etcd/ca.crt` should have the new CA certificate.
- `/etc/kubernetes/pki/apiserver-etcd-client.{crt,key}` should have the new
  client certificate and key.

<!-- LINKS -->

[snap-install-howto]: ./install/snap
