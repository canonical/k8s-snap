# Two-Node Active-Active High-Availability using Dqlite

## Rationale

High availability (HA) is a mandatory requirement for most production-grade
Kubernetes deployments, usually implying three or more nodes.

Two-node HA clusters are sometimes preferred for cost savings and operational
efficiency considerations. Follow this guide to learn how Canonical Kubernetes
can achieve high availability with just two nodes while using the default
datastore, Dqlite.

Dqlite cannot achieve Raft quorum with less than three nodes. This means that
Dqlite will not be able to replicate data and the secondaries will simply
forward the queries to the primary node.

In the event of a node failure, database recovery will require following the
steps in the [Dqlite recovery guide].

## Proposed solution

Since Dqlite data replication is not available in this situation, we propose
using synchronous block level replication through
[Distributed Replicated Block Device] (DRBD).

The cluster monitoring and failover process will be handled by Pacemaker and
Corosync. After a node failure, the DRBD volume will be mounted on the standby
node, allowing access to the latest Dqlite database version.

Additional recovery steps are automated and invoked through Pacemaker.

## Alternatives

Another possible approach is to use PostgreSQL with Kine and logical
replication. However, it is outside the scope of this document.

See the [external datastore guide] for more information on how Canonical
Kubernetes can be configured to use other datastores.

## Guide

### Prerequisites

* Ensure both nodes are part of the Kubernetes cluster.
  See the [getting started] and [add/remove nodes] guides.
* The user associated with the HA service has SSH access to the peer node and
  passwordless sudo configured. For simplicity, the default "ubuntu" user can
  be used.
* We recommend using static IP configuration.

The [two-node-ha.sh script] automates most operations related to the two-node
HA scenario and is included in the snap.

The first step is to install the required packages:

```
/snap/k8s/current/k8s/hack/two-node-ha.sh install_packages
```

### DRBD

For the purpose of this guide, we are going to use a loopback device as DRBD
backing storage:

```
sudo dd if=/dev/zero of=/opt/drbd0-backstore bs=1M count=2000
```

Ensure that the loopback device is attached at boot time, before Pacemaker
starts.

```
cat <<EOF | sudo tee /etc/rc.local
#!/bin/sh
mknod /dev/lodrbd b 7 200
losetup /dev/lodrbd /opt/drbd0-backstore
EOF

sudo chmod +x /etc/rc.local

cat <<EOF | sudo tee /etc/systemd/system/rc-local.service
# This unit gets pulled automatically into multi-user.target by
# systemd-rc-local-generator if /etc/rc.local is executable.
[Unit]
Description=/etc/rc.local Compatibility
Documentation=man:systemd-rc-local-generator(8)
ConditionFileIsExecutable=/etc/rc.local
After=network.target

[Service]
Type=forking
ExecStart=/etc/rc.local start
TimeoutSec=0
RemainAfterExit=yes
GuessMainPID=no

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable rc-local.service
sudo systemctl start rc-local.service
```

Let's configure the DRBD block device that will hold the Dqlite data.
Ensure the correct node addresses are used.

```
# Disable the DRBD service, it will be managed through Pacemaker.
sudo systemctl disable drbd

HAONE_ADDR=<firstNodeAddress>
HATWO_ADDR=<secondNodeAddress>

cat <<EOF | sudo tee /etc/drbd.d/r0.res
resource r0 {
  on haone {
    device /dev/drbd0;
    disk /dev/lodrbd;
    address ${HAONE_ADDR}:7788;
    meta-disk internal;
  }
  on hatwo {
    device /dev/drbd0;
    disk /dev/lodrbd;
    address ${HATWO_ADDR}:7788;
    meta-disk internal;
  }
}
EOF

sudo drbdadm create-md r0
sudo drbdadm status
```

Let's create a mount point for the DRBD block device. Non-default mount points
need to be passed to the ``two-node-ha.sh`` script mentioned above, see the
script for the full list of configurable parameters.

```
DRBD_MOUNT_DIR=/mnt/drbd0
sudo mkdir -p $DRBD_MOUNT_DIR
```

Run the following *once* to initialize the filesystem.

```
sudo drbdadm up r0

sudo drbdadm -- --overwrite-data-of-peer primary r0/0
sudo mkfs.ext4 /dev/drbd0

sudo drbdadm down r0
```

Add the drbd device to the ``multipathd`` blacklist, ensuring that the multipath
service will not attempt to manage this device:

```
sudo cat <<EOF | sudo tee -a /etc/multipath.conf
blacklist {
    devnode "^drbd*"
}
EOF

sudo systemctl restart multipathd
```

### Corosync

Let's prepare the Corosync configuration. Again, make sure to use the correct
addresses.

```
HAONE_ADDR=<firstNodeAddress>
HATWO_ADDR=<secondNodeAddress>

sudo mv /etc/corosync/corosync.conf /etc/corosync/corosync.conf.orig

cat <<EOF | sudo tee /etc/corosync/corosync.conf
totem {
  version: 2
  cluster_name: ha
  secauth: off
  transport:udpu
  interface {
    ringnumber: 0
    bindnetaddr: ${netaddr}
    broadcast: yes
    mcastport: 5405
  }
}

nodelist {
  node {
    ring0_addr: ${HAONE_ADDR}
    name: haone
    nodeid: 1
  }
  node {
    ring0_addr: ${HATWO_ADDR}
    name: hatwo
    nodeid: 2
  }
}

quorum {
  provider: corosync_votequorum
  two_node: 1
  wait_for_all: 1
  last_man_standing: 1
  auto_tie_breaker: 0
}
EOF
```

Follow the above steps on both nodes before moving forward.

### Pacemaker

Let's define a Pacemaker resource for the DRBD block device, which
ensures that the block device will be mounted on the replica in case of a
primary node failure.

[Pacemaker fencing] (stonith) configuration is environment specific and thus
outside the scope of this guide. However, we highly recommend using fencing
if possible to reduce the risk of cluster split-brain situations.

```
HAONE_ADDR=<firstNodeAddress>
HATWO_ADDR=<secondNodeAddress>
DRBD_MOUNT_DIR=${DRBD_MOUNT_DIR:-"/mnt/drbd0"}

sudo crm configure <<EOF
property stonith-enabled=false
property no-quorum-policy=ignore
primitive drbd_res ocf:linbit:drbd params drbd_resource=r0 op monitor interval=29s role=Master op monitor interval=31s role=Slave
ms drbd_master_slave drbd_res meta master-max=1 master-node-max=1 clone-max=2 clone-node-max=1 notify=true
primitive fs_res ocf:heartbeat:Filesystem params device=/dev/drbd0 directory=${DRBD_MOUNT_DIR} fstype=ext4
colocation fs_drbd_colo INFINITY: fs_res drbd_master_slave:Master
order fs_after_drbd mandatory: drbd_master_slave:promote fs_res:start
commit
show
quit
EOF
```

Before moving forward, let's ensure that the DRBD Pacemaker resource runs on
the primary (voter) Dqlite node.

In this setup, only the primary node holds the latest Dqlite data, which will
be transferred to the DRBD device once the clustered service starts.
This is automatically handled by the ``two-node-ha.sh start_service`` command.

```
sudo k8s status
sudo drbadadm status
sudo crm status
```

If the DRBD device is assigned to the secondary Dqlite node (spare), move it
to the primary like so:

```
sudo crm resource move fs_res <primary_node_name>

# remove the node constraint.
sudo crm resource clear fs_res
```

### Kubernetes services

We can now turn our attention to the Kubernetes services. Ensure that the k8s
snap services no longer start automatically. Instead, they will be managed by a
wrapper service.

```
for f in `sudo snap services k8s  | awk 'NR>1 {print $1}'`; do
    echo "disabling snap.$f"
    sudo systemctl disable "snap.$f";
done
```

The next step is to define the wrapper service. Add the following to
``/etc/systemd/system/two-node-ha-k8s.service``. Note that the sample uses the
``ubuntu`` user, feel free to use a different one as long as the prerequisites
are met.

```
[Unit]
Description=K8s service wrapper handling Dqlite recovery for two-node HA setups.
After=network.target pacemaker.service

[Service]
User=ubuntu
Group=ubuntu
Type=oneshot
ExecStart=/bin/bash /snap/k8s/current/k8s/hack/two-node-ha.sh start_service
ExecStop=/bin/bash sudo snap stop k8s
RemainAfterExit=true

[Install]
WantedBy=multi-user.target
```

```{note}
The ``two-node-ha.sh start_service`` command used by the service wrapper automatically
detects the expected Dqlite role based on the DRBD state and takes the
necessary steps to bootstrap the Dqlite state directories, synchronize with the
peer node (if available) and recover the database.
```

We need the ``two-node-ha-k8s`` service to be restarted once a DRBD failover
occurs. For that, we are going to define a separate service that will be
invoked by Pacemaker. Create a file called
``/etc/systemd/system/two-node-ha-k8s-failover.service`` containing the following:

```
[Unit]
Description=Managed by Pacemaker, restarts two-node-ha-k8s on failover.
After=network.target home-ubuntu-workspace.mount

[Service]
Type=oneshot
ExecStart=systemctl restart two-node-ha-k8s
RemainAfterExit=true
```

Reload the systemd configuration and set ``two-node-ha-k8s`` to start
automatically. Notice that ``two-node-ha-k8s-failover`` must not be configured
to start automatically, but instead is going to be managed through Pacemaker.

```
sudo systemctl enable two-node-ha-k8s
sudo systemctl daemon-reload
```

Make sure that both nodes have been configured using the above steps before
moving forward.

We can now define a new Pacemaker resource that will invoke the
``two-node-ha-k8s-failover`` service when a DRBD failover occurs.

```
sudo crm configure <<EOF
primitive ha_k8s_failover_service systemd:two-node-ha-k8s-failover op start interval=0 timeout=120 op stop interval=0 timeout=30
order failover_after_fs mandatory: fs_res:start ha_k8s_failover_service:start
colocation fs_failover_colo INFINITY: fs_res ha_k8s_failover_service
commit
show
quit
EOF
```

The setup is ready, start the HA k8s service on both nodes:

```
sudo systemctl start two-node-ha-k8s
```

## Troubleshooting

### Dqlite recovery failing because of unexpected data segments

Dqlite recovery may fail if there are data segments past the latest snapshot.

```
Error: failed to recover k8s-dqlite, error: k8s-dqlite recovery failed, error:
recover failed with error code 1, error details: raft_recover(): io:
closed segment 0000000000002369-0000000000002655 is past last snapshot
snapshot-2-2048-642428, pre-recovery backup:
/var/snap/k8s/common/recovery-k8s-dqlite-2024-09-05T082644Z-pre-recovery.tar.gz
```

Remove the offending segments and restart the ``two-node-ha-k8s`` service.

### DRBD split brain

The DRBD cluster may enter a split brain state and stop synchronizing. The
chances increase if fencing (stonith) is not enabled.

```
ubuntu@hatwo:~$ sudo drbdadm status
r0 role:Primary
  disk:UpToDate

ubuntu@hatwo:~$ cat /proc/drbd
version: 8.4.11 (api:1/proto:86-101)
srcversion: C7B8F7076B8D6DB066D84D9
 0: cs:StandAlone ro:Secondary/Unknown ds:UpToDate/DUnknown   r-----
    ns:0 nr:0 dw:0 dr:0 al:0 bm:0 lo:0 pe:0 ua:0 ap:0 ep:1 wo:f oos:1802140

ubuntu@hatwo:~$ dmesg | grep "Split"
[  +0.000082] block drbd0: Split-Brain detected but unresolved, dropping connection!

```

To recover DRBD, use following procedure:

```
# On the stale node:
sudo drbdadm secondary r0 
sudo drbdadm disconnect r0
sudo drbdadm -- --discard-my-data connect r0

# On the node that contains the latest data
sudo drbdadm connect r0
```

<!--LINKS -->
[Distributed Replicated Block Device]: https://ubuntu.com/server/docs/distributed-replicated-block-device-drbd
[Dqlite recovery guide]: restore-quorum
[external datastore guide]: external-datastore
[two-node-ha.sh script]: https://github.com/canonical/k8s-snap/blob/main/k8s/hack/two-node-ha.sh
[getting started]: ../tutorial/getting-started
[add/remove nodes]: ../tutorial/add-remove-nodes
[Pacemaker fencing]: https://clusterlabs.org/pacemaker/doc/2.1/Pacemaker_Explained/html/fencing.html
