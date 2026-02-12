# Dqlite database

{{product}} uses Dqlite for k8sd, which manages Kubernetes cluster management
data.

## Database files

The k8sd database state directory is located at:

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
* ``server.crt``, ``server.key`` - certificates
* ``truststore`` - folder containing trusted certificates
* ``daemon.yaml`` - k8sd daemon configuration
* ``database`` - separate database folder

Dqlite cluster members have one of the following roles:

| Role enum | Role name | Replicates database | Voting in leader elections |
|-----------|-----------|---------------------|----------------------------|
| 0         | voter     | yes                 | yes                        |
| 1         | stand-by  | yes                 | no                         |
| 2         | spare     | no                  | no                         |

## Inspecting the database

Use ``/snap/k8s/current/bin/k8sd sql`` to issue SQL queries to the k8sd
Dqlite database. Note that a very limited subset of SQL syntax is available.

The following command can be used to enumerate the tables:

```
/snap/k8s/current/bin/k8sd sql \
  --state-dir /var/snap/k8s/common/var/lib/k8sd/state \
  "SELECT * FROM sqlite_master WHERE type='table'"
```
