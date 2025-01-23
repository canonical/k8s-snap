# Dqlite database

{{product}} uses not one, but two Dqlite databases:

* k8s-dqlite - used by Kubernetes itself (as an ETCD replacement)
* k8sd - Kubernetes cluster management data

## Database files

Each database has its own state directory:

* ``/var/snap/k8s/common/var/lib/k8s-dqlite``
* ``/var/snap/k8s/common/var/lib/k8sd/state``

The state directory normally contains:

* ``info.yaml`` - the id, address and cluster role of this node
* ``cluster.yaml`` - the state of the cluster, as seen by this Dqlite node.
  It includes the same information as info.yaml, but for all cluster nodes
* ``00000abcxx-00000abcxx``, ``open-abc`` - database segments
* ``cluster.crt``, ``cluster.key`` - node certificates
* ``snapshot-abc-abc-abc.meta``
* ``metadata{1,2}``
* ``*.sock`` - control unix sockets

K8sd contains additional files to manage cluster memberships and member's
secure communication:

* ``server.crt``, ``server.key`` certificates
* ``truststore`` folder, containing trusted certificates
* ``daemon.yaml`` - k8sd daemon configuration
* separate ``database`` folder

Dqlite cluster members have one of the following roles:

| Role enum | Role name | Replicates database | Voting in leader elections |
|-----------|-----------|---------------------|----------------------------|
| 0         | voter     | yes                 | yes                        |
| 1         | stand-by  | yes                 | no                         |
| 2         | spare     | no                  | no                         |

## Inspecting the databases

Use the following command to connect to the k8s-dqlite database:

```
sudo /snap/k8s/current/bin/dqlite \
  -s file:///var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml \
  -c /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt \
  -k /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key \
  k8s
```

The ``.leader`` command displays the current cluster leader.

The kine key-value pairs are stored in the ``kine`` database and can be retrieved
like so:

```
select id, name, value from kine limit 100;
```

Use ``/snap/k8s/current/bin/k8sd sql`` to issue k8sd sql commands.
Note that a very limited subset of SQL syntax is available, however the
following can be used to enumerate the tables:

```
/snap/k8s/current/bin/k8sd sql \
  --state-dir /var/snap/k8s/common/var/lib/k8sd/state
  "SELECT * FROM sqlite_master WHERE type='table'"
```
