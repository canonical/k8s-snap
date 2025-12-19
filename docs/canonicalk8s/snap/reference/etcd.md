# Etcd database

{{product}} uses **etcd** as the default Kubernetes datastore.
etcd is a distributed key-value store that holds all Kubernetes
cluster state data.

## Database files

The etcd database has its state directory under the snap path:

* `/var/snap/k8s/common/var/lib/etcd/data`

This directory normally contains:

* `member/` – etcd member data, including the write-ahead log (WAL)
and snapshots

  * `snap/` – snapshot files of the key-value store
  * `wal/` – write-ahead logs for durability and recovery
* `server.crt`, `server.key` – TLS certificates for secure communication
* `ca.crt` – certificate authority for validating peer and client connections

## Inspecting the database

The `etcdctl` command-line client can be used to directly interact with etcd.
It is **not included** in the {{product}} snap.
You will need to obtain `etcdctl` separately from the [official etcd releases]
or your OS package manager.

Once `etcdctl` is available, you can connect to the etcd database as follows:

```
etcdctl \
  --endpoints=https://127.0.0.1:2379 \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  get "" --prefix --keys-only
```

To retrieve values for a given prefix:

```
etcdctl \
  --endpoints=https://127.0.0.1:2379 \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  get /registry/pods/default --prefix
```

To check the current cluster leader and member status:

```
etcdctl \
  --endpoints=https://127.0.0.1:2379 \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  endpoint status --write-out=table
```

## Notes

* etcd in {{product}} runs with TLS enabled by default to secure communication
between members and clients.
* All Kubernetes API objects are stored as serialized JSON in etcd, indexed by
their API path keys.
* Care should be taken when directly modifying etcd data, as it may corrupt
cluster state.

<!-- LINKS -->

[official etcd releases]: https://etcd.io/docs/
