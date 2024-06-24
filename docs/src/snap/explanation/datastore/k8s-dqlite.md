# k8s-dqlite datastore

Canonical Kubernetes supports using a managed dqlite cluster as the underlying
datastore of the cluster. This is the default option when no configuration is
specified.

This page explains the behaviour of the managed dqlite cluster. See How-To
[Configure Canonical Kubernetes with dqlite][how-to-dqlite] for steps to
deploy Canonical Kubernetes with a managed etcd datastore.

## Topology

When using the managed dqlite datastore, all the control plane nodes of the
cluster will be running `k8s-dqlite`. Internal cluster communication happens
over TLS between the members. Each cluster member exposes a local unix socket
for `kube-apiserver` to access the datastore.

The dqlite cluster uses port 9000 on each node for cluster communication. This
port can be configured when bootstrapping the cluster.

## Clustering

When adding a new control plane node to the cluster, Canonical Kubernetes will
add the node to the dqlite cluster.

Similarly, when removing a node from the cluster using `k8s remove-node`,
Canonical Kubernetes will make sure that the node is also removed from the
k8s-dqlite cluster.

Since `kube-apiserver` instances access the datastore over a local unix socket,
no reconfiguration is needed on that front.

## configuration and data directory

The k8s-dqlite configuration and data paths to be aware of are:

- `/var/snap/k8s/common/args/k8s-dqlite`: Command line arguments for the
  `k8s-dqlite` service.
- `/var/snap/k8s/common/var/lib/k8s-dqlite`: Data directory.

<!-- LINKS -->

[how-to-dqlite]: /snap/howto/datastore/k8s-dqlite
