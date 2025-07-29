# How to recover a cluster after quorum loss

Highly available {{product}} clusters are designed to tolerate losing 
one or more nodes. Both [etcd] and [Dqlite] use [Raft] protocol
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

```{note}
{{product}} relies on two separate distributed datastores. The first 
is a Dqlite-based cluster datastore used to manage the {{product}} itself. 
The second is the Kubernetes backend datastore, which stores 
Kubernetes objects' state and can be either etcd (default) or Dqlite, depending 
on user configuration. For more information, please see [the architecture].
```

If you have set Dqlite as the datastore, please consult the 
[Dqlite configuration reference] before moving forward.

## etcd

### Take an etcd snapshot

Choose one of the remaining cluster nodes that has the most recent
version of the Raft log.

Install `etcdctl` and `etcdutl` binaries following the 
[etcd upstream installation instructions]. With sudo privilege, 
run `etcdctl` to verify you have access to etcd cluster node's data:

```
sudo etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt \
        --cert /etc/kubernetes/pki/apiserver-etcd-client.crt \
        --key /etc/kubernetes/pki/apiserver-etcd-client.key \
        snapshot save snapshot.db
```

Follow the [upstream instructions] to take a snapshot of the Keyspace. 

### Stop {{product}} services on all nodes

Before recovering the cluster, all remaining {{product}} services
must be stopped on every node:

```
sudo snap stop k8s
```

### Recover the k8sd database

Choose one of the remaining healthy cluster nodes that has the most recent
version of the Raft log. Use the `cluster-recover` command to reconfigure 
the Raft members and generate recovery tarballs that are used to restore the 
cluster datastore on lost nodes. The command is an interactive tool that 
allows you to modify the relevant files and provides useful hints at each step.
On the node with the most recent Raft logs, run:

```
sudo /snap/k8s/current/bin/k8sd cluster-recover \
    --state-dir=/var/snap/k8s/common/var/lib/k8sd/state \
    --k8s-dqlite-state-dir=/var/snap/k8s/common/var/lib/k8s-dqlite \
    --log-level 0
    --skip-k8s-dqlite
```

Use the command to update the ``cluster.yaml`` file, changing the role of the 
lost nodes to "spare". Additionally, verify the addresses and IDs specified 
in ``cluster.yaml``, ``info.yaml`` and ``daemon.yaml`` are correct, 
especially if database files were moved across nodes.

```{note}
By default, `cluster-recover` will recover both Dqlite cluster and Kubernetes 
datastores. When etcd is used, the ``--skip-k8s-dqlite`` flag is needed to 
instruct `cluster-recover` to ignore the Dqlite Kubernetes datastore.
```

Adjust the log level for additional debug messages by increasing its
value. Database backups are created by the command before making any changes.

Copy the generated ``recovery_db.tar.gz`` to all remaining nodes at
``/var/snap/k8s/common/var/lib/k8sd/state/recovery_db.tar.gz``. When the k8sd
service starts, it will load the archive and perform the necessary recovery
steps.

### Restore the etcd snapshot

Run the following command on all remainining live nodes to reconfigure the 
etcd membership:

```
etcdutl snapshot restore snapshot.db \
      --name=<NAME> \
      --initial-advertise-peer-urls=<ADVERTISE_PEER_URLS> \
      --initial-cluster=<INITIAL_CLUSTER> \
      --data-dir /var/snap/k8s/common/var/lib/etcd/data
```

Get the `<NAME>` and `<ADVERTISE_PEER_URLS>` for each node by 
executing the following command:

```
echo $(grep -E '^--(name|initial-advertise-peer-urls)=' /var/snap/k8s/common/args/etcd | xargs)
```

for each node the output could be something like:

```
--initial-advertise-peer-urls=https://10.246.154.125:2380 --name=node-1
```

The `<INITIAL_CLUSTER>` will be the comma-separated list of all remaining 
cluster members fetched from each node by running:

```
echo $(grep -E '^--(initial-cluster)=' /var/snap/k8s/common/args/etcd | xargs)
```

Aggregate the results from all remaining nodes to form the `<INITIAL_CLUSTER>`
that should have a format like below:

```
node-1=https://10.246.154.125:2380,node-2=https://10.246.154.126:2380,node-3=https://10.246.154.127:2380
```

### Start the services 

For each node, start the {{product}} services by running:

```
sudo snap start k8s
```

Ensure that the services started successfully by using
``sudo snap services k8s``. Use ``sudo k8s status --wait-ready`` to wait for the
cluster to become ready.

## Dqlite

### Stop {{product}} services on all nodes

Before recovering the cluster, all remaining {{product}} services
must be stopped on every node:

```
sudo snap stop k8s
```

### Recover the database

Choose one of the remaining healthy cluster nodes that has the most recent
version of the Raft log. Use the `cluster-recover` command to reconfigure 
the Raft members and generate recovery tarballs that are used to restore the 
cluster datastore on lost nodes. The command is an interactive tool that 
allows you modify the relevant files and provides useful hints at each step.
On the node with the most recent Raft logs, run:

```
sudo /snap/k8s/current/bin/k8sd cluster-recover \
    --state-dir=/var/snap/k8s/common/var/lib/k8sd/state \
    --k8s-dqlite-state-dir=/var/snap/k8s/common/var/lib/k8s-dqlite \
    --log-level 0
```

Use the command to update the ``cluster.yaml`` file, changing the role of the 
lost nodes to "spare". Additionally, verify the addresses and IDs specified 
in ``cluster.yaml``, ``info.yaml`` and ``daemon.yaml`` are correct, 
especially if database files were moved across nodes.

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

### Recover the remaining nodes

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
[etcd]: https://etcd.io/
[etcd upstream installation instructions]: https://etcd.io/docs/latest/install/
[upstream instructions]: https://etcd.io/docs/latest/op-guide/recovery/
[Dqlite configuration reference]: ../reference/dqlite.md
[Raft]: https://raft.github.io/
[the architecture]: ../explanation/architecture/
