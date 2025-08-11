# How to recover a cluster after quorum loss

Highly available {{product}} clusters are designed to tolerate the loss of 
one or more nodes. Both [etcd] and [Dqlite] use [Raft] protocol
where an elected leader holds the definitive copy of the database, which is
then replicated on two or more secondary nodes.

When the majority of the nodes are lost, the cluster becomes unavailable.
If at least one database node survived, the cluster can be recovered using the
steps outlined in this document.

{{product}} relies on two separate distributed datastores. The first 
is a Dqlite-based cluster datastore used to manage the distribution's state. 
The second database stores the Kubernetes objects' state and is either etcd 
by default or Dqlite, depending on user configuration. For more information, 
please see [the architecture guide].

```{warning}
This guide can be used to recover the cluster datastore and the {{product}} 
managed datastore, which can be either etcd or Dqlite. Persistent volumes on 
the lost nodes are *not* recovered.
```

If you have set up Dqlite as the datastore, please consult the 
[Dqlite configuration reference] before moving forward.


## Stop {{product}} services on all nodes

Before recovering the cluster, all remaining {{product}} services
must be stopped on every node:

```
sudo snap stop k8s
```

## Recover the cluster datastore 

Choose one of the remaining healthy cluster nodes that has the most recent
version of the Raft log. 

To find the node with the most recent log entries, navigate to
`/var/snap/k8s/common/var/lib/k8sd/state/database` on each node and
look at all the segment files matching the format 
`0000000000006145-0000000000006823` (this is just an example).
For each node, identify the segment file with the highest 
end-segment index (e.g., 6823 in the example), then compare across 
nodes and select the one with the highest index overall.

Use the `cluster-recover` command to reconfigure 
the Raft members and generate recovery tarballs that are used to restore the 
cluster datastore on lost nodes. The command is an interactive tool that 
allows you to modify the relevant files and provides useful hints at each step.
On the node with the most recent Raft logs, run:

`````{tabs}

````{group-tab} etcd

```
sudo /snap/k8s/current/bin/k8sd cluster-recover \
    --state-dir=/var/snap/k8s/common/var/lib/k8sd/state \
    --k8s-dqlite-state-dir=/var/snap/k8s/common/var/lib/k8s-dqlite \
    --log-level 0
    --skip-k8s-dqlite
```

````

````{group-tab} Dqlite

```
sudo /snap/k8s/current/bin/k8sd cluster-recover \
    --state-dir=/var/snap/k8s/common/var/lib/k8sd/state \
    --k8s-dqlite-state-dir=/var/snap/k8s/common/var/lib/k8s-dqlite \
    --log-level 0
```

````

`````

```{note}
By default, the command will recover both the cluster and Kubernetes datastore 
Dqlite databases. If one of the databases needs to be skipped, use the 
`--skip-k8sd` or `--skip-k8s-dqlite` flags. This can also be useful when 
using an external etcd database.
```

Use the command to update the ``cluster.yaml`` file, changing the role of the 
lost nodes to "spare". Additionally, verify the addresses and IDs specified 
in ``cluster.yaml``, ``info.yaml`` and ``daemon.yaml`` are correct, 
especially if database files were moved across nodes.

Adjust the log level for additional debug messages by increasing its
value. Database backups are created by the command before making any changes.

Copy the generated ``recovery_db.tar.gz`` to all remaining nodes at
``/var/snap/k8s/common/var/lib/k8sd/state/recovery_db.tar.gz``. When we 
restart k8sd in a later step, it will load the archive and perform the 
necessary recovery steps.

```{note}
Non-interactive mode can be requested using the ``--non-interactive`` flag.
In this case, no interactive prompts or text editors will be displayed and
the command will assume that the configuration files have already been updated.

This allows automating the recovery procedure.
```

## Recover the Kubernetes datastore

`````{tabs}

````{group-tab} etcd

### Take an etcd snapshot

Choose one of the remaining cluster nodes. Install `etcdctl` 
and `etcdutl` binaries following the 
[etcd upstream installation instructions]. With sudo privilege, 
run `etcdctl` to verify you have access to etcd cluster node's data:

```
sudo etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt \
        --cert /etc/kubernetes/pki/apiserver-etcd-client.crt \
        --key /etc/kubernetes/pki/apiserver-etcd-client.key \
        snapshot save snapshot.db
```

Follow the [upstream instructions] to take a snapshot of the Keyspace. 

### Restore the etcd snapshot

On all remaining live nodes, reconfigure the etcd membership:

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
cluster members. The list should be created by gathering the results from
running the following command on each node:

```
echo $(grep -E '^--(initial-cluster)=' /var/snap/k8s/common/args/etcd | xargs)
```

Aggregate the results from all remaining nodes to form the `<INITIAL_CLUSTER>`
that should have a format like below:

```
node-1=https://10.246.154.125:2380,node-2=https://10.246.154.126:2380,node-3=https://10.246.154.127:2380
```

Ultimately you would need to run a command similar to the following:

```
etcdutl snapshot restore snapshot.db \
      --name='node-1' \
      --initial-advertise-peer-urls='https://10.246.154.125:2380' \
      --initial-cluster='node-1=https://10.246.154.125:2380,node-2=https://10.246.154.126:2380,node-3=https://10.246.154.127:2380' \
      --data-dir /var/snap/k8s/common/var/lib/etcd/data
```

````

````{group-tab} Dqlite

The k8s-dqlite recovery tarballs that were created with the `cluster-recover` 
command need to be copied over to all cluster
nodes.

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

````

`````

## Start the services 

For each node, start the {{product}} services by running:

```
sudo snap start k8s
```

Confirm that the services started successfully by running
``sudo snap services k8s``. Use ``sudo k8s status --wait-ready`` to wait for the
cluster to become ready.

Once the nodes are up the cluster will be reported as "highly available" again:

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
[the architecture guide]: ../explanation/architecture/
