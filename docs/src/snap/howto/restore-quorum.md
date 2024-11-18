# Recovering a Cluster After Quorum Loss

Highly available {{product}} clusters can survive losing one or more
nodes. [Dqlite], the default datastore, implements a [Raft] based protocol
where an elected leader holds the definitive copy of the database, which is
then replicated on two or more secondary nodes.

When the a majority of the nodes are lost, the cluster becomes unavailable.
If at least one database node survived, the cluster can be recovered using the
steps outlined in this document.

```{note}
This guide can be used to recover the default {{product}} datastore,
dqlite. Persistent volumes on the lost nodes are *not* recovered.
```

## Dqlite Configuration

Be aware that {{product}} uses not one, but two dqlite databases:

* k8s-dqlite - used by Kubernetes itself
* k8sd - Kubernetes cluster management data

Each database has its own state directory:

* ``/var/snap/k8s/common/var/lib/k8s-dqlite``
* ``/var/snap/k8s/common/var/lib/k8sd/state``

The state directory normally contains:

* ``info.yaml`` - the id, address and cluster role of this node
* ``cluster.yaml`` - the state of the cluster, as seen by this dqlite node.
  It includes the same information as info.yaml, but for all cluster nodes.
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

## Stop {{product}} Services on All Nodes

Before recovering the cluster, all remaining {{product}} services
must be stopped. Use the following command on every node:

```
sudo snap stop k8s
```

## Recover the Database

Choose one of the remaining alive cluster nodes that has the most recent
version of the Raft log.

Update the ``cluster.yaml`` files, changing the role of the lost nodes to
"spare" (2). Additionally, double check the addresses and IDs specified in
``cluster.yaml``, ``info.yaml`` and ``daemon.yaml``, especially if database
files were moved across nodes.

The following command guides us through the recovery process, prompting a text
editor with informative inline comments for each of the dqlite configuration
files.

```
sudo /snap/k8s/current/bin/k8sd cluster-recover \
    --state-dir=/var/snap/k8s/common/var/lib/k8sd/state \
    --k8s-dqlite-state-dir=/var/snap/k8s/common/var/lib/k8s-dqlite \
    --log-level 0
```

Please adjust the log level for additional debug messages by increasing its
value. The command creates database backups before making any changes.

The above command will reconfigure the Raft members and create recovery
tarballs that are used to restore the lost nodes, once the Dqlite
configuration is updated.

```{note}
By default, the command will recover both Dqlite databases. If one of the
databases needs to be skipped, use the ``--skip-k8sd`` or ``--skip-k8s-dqlite``
flags. This can be useful when using an external Etcd database.
```

```{note}
Non-interactive mode can be requested using the ``--non-interactive`` flag.
In this case, no interactive prompts or text editors will be displayed and
the command will assume that the configuration files have already been updated.

This allows automating the recovery procedure.
```

Once the "cluster-recover" command completes, restart the k8s services on the
node:

```
sudo snap start k8s
```

Ensure that the services started successfully by using
``sudo snap services k8s``. Use ``k8s status --wait-ready`` to wait for the
cluster to become ready.

You may notice that we have not returned to an HA cluster yet:
``high availability: no``. This is expected as we need to recover

## Recover the remaining nodes

The k8s-dqlite and k8sd recovery tarballs need to be copied over to all cluster
nodes.

For k8sd, copy ``recovery_db.tar.gz`` to
``/var/snap/k8s/common/var/lib/k8sd/state/recovery_db.tar.gz``. When the k8sd
service starts, it will load the archive and perform the necessary recovery
steps.

The k8s-dqlite archive needs to be extracted manually. First, create a backup
of the current k8s-dqlite state directory:

```
sudo mv /var/snap/k8s/common/var/lib/k8s-dqlite \
  /var/snap/k8s/common/var/lib/k8s-dqlite.bkp
```

Then, extract the backup archive:

```
sudo mkdir /var/snap/k8s/common/var/lib/k8s-dqlite
sudo tar xf  recovery-k8s-dqlite-$timestamp-post-recovery.tar.gz \
  -C /var/snap/k8s/common/var/lib/k8s-dqlite
```

Node specific files need to be copied back to the k8s-dqlite state directory:

```
sudo cp /var/snap/k8s/common/var/lib/k8s-dqlite.bkp/cluster.crt \
  /var/snap/k8s/common/var/lib/k8s-dqlite
sudo cp /var/snap/k8s/common/var/lib/k8s-dqlite.bkp/cluster.key \
  /var/snap/k8s/common/var/lib/k8s-dqlite
sudo cp /var/snap/k8s/common/var/lib/k8s-dqlite.bkp/info.yaml \
  /var/snap/k8s/common/var/lib/k8s-dqlite
```

Once these steps are completed, restart the k8s services:

```
sudo snap start k8s
```

Repeat these steps for all remaining nodes. Once a quorum is achieved,
the cluster will be reported as "highly available":

```
$ sudo k8s status
cluster status:           ready
control plane nodes:      10.80.130.168:6400 (voter),
                          10.80.130.167:6400 (voter),
                          10.80.130.164:6400 (voter)
high availability:        yes
datastore:                k8s-dqlite
network:                  enabled
dns:                      enabled at 10.152.183.193
ingress:                  disabled
load-balancer:            disabled
local-storage:            enabled at /var/snap/k8s/common/rawfile-storage
gateway                   enabled
```


<!-- LINKS -->
[Dqlite]: https://dqlite.io/
[Raft]: https://raft.github.io/
