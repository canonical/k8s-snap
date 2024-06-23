# Datastore

```{toctree}
:hidden:
Datastore <self>
```

One of the core components of a Kubernetes cluster is the datastore. The
datastore is where all of the cluster state is persisted. The `kube-apiserver`
communicates with the datastore using an [etcd API].

Canonical Kubernetes supports three different datastore types:

1. `k8s-dqlite` (**default**) (managed): Control plane nodes form a dqlite
   cluster and expose an etcd endpoint over a local unix socket. The dqlite
   cluster is automatically updated when adding or removing cluster members.

   For more details, see [k8s-dqlite].

2. `etcd` (managed): Control plane nodes form an etcd cluster. The etcd cluster
   is automatically updated when adding or removing cluster members.

   For more details, see [etcd].

3. `external`: Do not deploy or manage the datastore. The user is expected to
   provision and manage an external etcd datastore, and provide the connection
   credentials (URLs and client certificates) when bootstrapping the cluster.

   For more details, see [external].

```{warning}
The selection of the backing datastore can only be done during the bootstrap
process. It is not possible to change the datastore type of a running cluster.

Instead, a new cluster should be deployed and workloads should be migrated to it
using a blue-green deployment method.
```

```{toctree}
:titlesonly:

k8s-dqlite
etcd
external
```

<!-- LINKS -->

[etcd API]: https://etcd.io/docs/v3.5/learning/api/
[k8s-dqlite]: k8s-dqlite
[etcd]: etcd
[external]: external
