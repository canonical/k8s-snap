# external datastore

Canonical Kubernetes supports using an external etcd cluster as the underlying
datastore of the cluster.

This page explains the behaviour of Canonical Kubernetes when using an external
etcd cluster. See How-To
[Configure Canonical Kubernetes with an external datastore][how-to-external] for
steps to deploy Canonical Kubernetes with an external etcd datastore.

## Topology

When using an external etcd datastore, the control plane nodes of the cluster
will only run the Kubernetes services. The cluster administrator is responsible
for deploying, managing and operating the external etcd datastore.

The control plane nodes are expected to be able to reach the external etcd
cluster over the network.

## TLS

For production deployments, it is highly recommended that the etcd cluster uses
TLS for both client and peer traffic. It is the responsibility of the cluster
administrator to deploy the external etcd cluster accordingly.

## Clustering

When using an external etcd datastore, the cluster administrator provides the
known etcd server URLs, as well as any required client certificates when
bootstrapping the cluster.

When adding a new control plane node to the cluster, Canonical Kubernetes will
configure it to use the same list of etcd servers and client certificates.

Removing a cluster node using `k8s remove-node` will not have any side-effect
on the external datastore.

## configuration and data directories

- `/etc/kubernetes/pki/etcd/ca.crt`: This is the CA certificate of the etcd
  cluster certificate. This will be created by Canonical Kubernetes, and contain
  the CA certificate specified when bootstrapping the cluster.
- `/etc/kubernetes/pki/apiserver-etcd-client.{crt,key}`: This is the client
  certificate and key used by `kube-apiserver` to authenticate with the etcd
  cluster.

<!-- LINKS -->

[how-to-external]: /snap/howto/datastore/external
