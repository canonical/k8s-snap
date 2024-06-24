# Canonical Kubernetes with a managed etcd datastore

Canonical Kubernetes supports using a managed etcd cluster as the underlying
datastore of the cluster.

This page explains the behaviour of the managed etcd cluster. See How-To
[Configure Canonical Kubernetes with etcd][how-to-etcd] for steps to deploy
Canonical Kubernetes with a managed etcd datastore.

## Topology

When using the managed etcd datastore, all the control plane nodes of the
cluster will be running an etcd instance. The etcd cluster is configured with
TLS for both client and peer traffic.

The etcd datastore uses ports 2379 (for client) and 2380 (for peer) traffic.
These ports can be configured when bootstrapping the cluster.

## TLS

Canonical Kubernetes will generate a separate self-signed CA certificate for
the etcd cluster. If needed, it is possible to specify a custom CA certificate
when bootstrapping the cluster. Any of the following scenarios are supported:

- No certificates are given, Canonical Kubernetes will generate self-signed CA
  and server certificates as needed.
- A custom CA certificate and private key are given during bootstrap. Canonical
  Kubernetes will then use this to generate server and peer certificates as
  needed.
- A custom CA certificate is passed. In this scenario, the server and peer
  certificates and private must also be specified. This is required for the
  bootstrap node, as well as any control plane nodes that join the cluster. In
  case any required certificate is not specified, the bootstrap or join process
  will fail.

## Clustering

When adding a new control plane node to the cluster, Canonical Kubernetes will
perform the following steps:

1. Use the etcd CA to generate peer and server certificates for the new node.
2. The new node will automatically register itself on the etcd cluster (by
   performing the equivalent of `etcdctl member add --peer-url ...`).
3. The new node will start and join the cluster quorum. If necessary, it will
   force a new leader election in the etcd cluster (e.g. while transitioning
   from 1 to 2 control plane nodes).

Similarly, when removing a cluster node from the cluster using `k8s remove-node`,
Canonical Kubernetes will make sure that the node is also removed from the etcd
cluster.

Canonical Kubernetes will also keep track of the active members of the etcd
cluster, and will periodically update the list of `--etcd-servers` in the
kube-apiserver arguments. This assures that if the etcd service on the local
node misbehaves, then `kube-apiserver` can still work by reaching the rest of
the etcd cluster members.

## Quorum

When using the managed etcd datastore, all nodes participate equally in the
raft quorum. That means an odd number of **2k + 1** nodes is needed to maintain
a fault tolerance of **k** nodes (such that the rest **k + 1** nodes maintain
an active quorum).

## Directories and paths

The etcd configuration and data directories to be aware of are:

- `/var/snap/k8s/common/var/lib/etcd/etcd.yaml`: YAML file with etcd
  cluster configuration. This contains information for the initial cluster
  members, TLS certificate paths and member peer and client URLs.
- `/var/snap/k8s/comonn/var/lib/etcd/data`: etcd data directory.
- `/etc/kubernetes/pki/etcd`: contains certificates for the etcd cluster
  (etcd CA certificate, server certificate and key, peer certificate and key).

<!-- LINKS -->

[how-to-etcd]: /snap/howto/datastore/etcd
