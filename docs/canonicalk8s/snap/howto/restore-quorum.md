# How to recover a cluster after quorum loss

Highly available {{product}} clusters can survive losing one or more
nodes. Both [etcd] and [Dqlite] implement a [Raft] based protocol
where an elected leader holds the definitive copy of the database, which is
then replicated on two or more secondary nodes.

When the a majority of the nodes are lost, the cluster becomes unavailable.
If at least one database node survived, the cluster can be recovered using the
steps outlined in this document.

```{note}
This guide can be used to recover the {{product}} managed datastore,
which can be either etcd or Dqlite. Persistent volumes on the lost nodes are 
*not* recovered.
```

If you have set Dqlite as the datastore, please consult the 
[Dqlite configuration reference] before moving forward.

# etcd

## Take an etcd snapshot

Choose one of the remaining alive cluster nodes that has the most recent
version of the Raft log.

Install `etcdctl` and `etcdutl` binaries following the 
[etcd upstream installation instructions]. Run this command with sudo 
privilege to verify you have access to etcd cluster node's data:

```
sudo etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt \
        --cert /etc/kubernetes/pki/apiserver-etcd-client.crt \
        --key /etc/kubernetes/pki/apiserver-etcd-client.key \
        snapshot save snapshot.db
```

Follow the [upstream instructions] to take an snapshot of the keyspace. 

## Stop {{product}} services on all nodes

Before recovering the cluster, all remaining {{product}} services
must be stopped. Use the following command on every node:

```
sudo snap stop k8s
```

## Recover the k8sd database

Update the ``cluster.yaml`` file, changing the role of the lost nodes to
"spare". Additionally, double check the addresses and IDs specified in
``cluster.yaml``, ``info.yaml`` and ``daemon.yaml``, especially if database
files were moved across nodes.

The following command guides us through the recovery process, prompting a text
editor with informative inline comments for each of the Dqlite configuration
files.

```
sudo /snap/k8s/current/bin/k8sd cluster-recover \
    --state-dir=/var/snap/k8s/common/var/lib/k8sd/state \
    --k8s-dqlite-state-dir=/var/snap/k8s/common/var/lib/k8s-dqlite \
    --log-level 0
    --skip-k8s-dqlite
```

Please adjust the log level for additional debug messages by increasing its
value. The command creates database backups before making any changes.

Copy the generated ``recovery_db.tar.gz`` to all remaining nodes at
``/var/snap/k8s/common/var/lib/k8sd/state/recovery_db.tar.gz``. When the k8sd
service starts, it will load the archive and perform the necessary recovery
steps.

## Restore the etcd snapshot

Run the following command on all remanining alive nodes to reconfigure the 
etcd membership:

```
etcdutl snapshot restore snapshot.db \
      --name=<NAME> \
      --initial-advertise-peer-urls=<ADVERTISE_PEER_URLS> \
      --initial-cluster=<INITIAL_CLUSTER> \
      --data-dir /var/snap/k8s/common/var/lib/etcd/data
```

You can get the `<NAME>` and `<ADVERTISE_PEER_URLS>` for each node by 
executing the following command:

```
args=$(grep -E '^--(name|initial-advertise-peer-urls)=' /var/snap/k8s/common/args/etcd | xargs)
```

The `<INITIAL_CLUSTER>` will be the comma-separated list of values fetched 
from each node by running:

```
args=$(grep -E '^--(name|initial-cluster)=' /var/snap/k8s/common/args/etcd | xargs)
```

## Start the services 

For each node, start the {{product}} services by running:

```
sudo snap start k8s
```

Ensure that the services started successfully by using
``sudo snap services k8s``. Use ``sudo k8s status --wait-ready`` to wait for the
cluster to become ready.

# Dqlite

## Stop {{product}} services on all nodes

Before recovering the cluster, all remaining {{product}} services
must be stopped. Use the following command on every node:

```
sudo snap stop k8s
```

## Recover the database

Choose one of the remaining alive cluster nodes that has the most recent
version of the Raft log.

Update the ``cluster.yaml`` files, changing the role of the lost nodes to
"spare" (2). Additionally, double check the addresses and IDs specified in
``cluster.yaml``, ``info.yaml`` and ``daemon.yaml``, especially if database
files were moved across nodes.

The following command guides us through the recovery process, prompting a text
editor with informative inline comments for each of the Dqlite configuration
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
``sudo snap services k8s``. Use ``sudo k8s status --wait-ready`` to wait for the
cluster to become ready.

You may notice that we have not returned to an HA cluster yet:
``high availability: no``. This is expected as we need to recover
the remaining nodes.

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

# Difference between etcd and Dqlite pertaining 2-node quorum

K8s-dqlite and etcd handle quorum formation differently in a two-node 
cluster configuration. With k8s-dqlite, quorum is not established until 
a third node is added. The second node that joins the cluster initially 
acts as a follower and is only promoted to a voter once the cluster reaches 
three nodes. In contrast, etcd adds the second node as a full voter right 
away, effectively forming a quorum with just two nodes.

This design difference impacts cluster behavior during node failure. If one 
node fails in a k8s-dqlite-backed cluster, the cluster can still operate as 
long as the remaining node is the leader. However, in an etcd-backed cluster, 
the system becomes unavailable regardless of which node goes down, since both 
nodes are needed to maintain quorum.

Etcd’s documentation explicitly warns against reconfiguring a two-member 
cluster by removing a member. Because quorum requires a majority of nodes, 
and the majority in a two-node setup is also two, any failure during the 
removal process can render the cluster inoperable.

<!-- LINKS -->
[Dqlite]: https://dqlite.io/
[etcd]: https://etcd.io/
[etcd upstream installation instructions]: https://etcd.io/docs/latest/install/
[upstream instructions]: https://etcd.io/docs/latest/op-guide/recovery/
[Dqlite configuration reference]: ../reference/dqlite.md
[Raft]: https://raft.github.io/
