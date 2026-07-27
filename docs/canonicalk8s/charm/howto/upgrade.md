---
myst:
  html_meta:
    description: How to upgrade Canonical Kubernetes to new minor and patch versions.
---

# How to upgrade {{product}} 

It is recommended that you keep your Kubernetes deployment
updated to the latest available stable version. You should
also update the other applications deployed in your Kubernetes
cluster. Keeping up-to-date ensures you have the latest bug-fixes
and security patches for smooth operation of your cluster.


```{note} Kubernetes will not automatically handle release upgrades. The 
cluster will not perform an unattended
automatic upgrade between minor versions, e.g. 1.30.1 to 1.31.0 or between 
patch versions e.g. 1.31.0 to 1.31.1.
```

Check the latest release version on the [Kubernetes release page].

## Before you begin

As with all upgrades, there is a possibility that there may be
unforeseen difficulties. It is highly recommended to make
a backup of any important data, including any running workloads.
For more details on creating backups, see the separate
[docs on backups][backup-restore].

Verify that:

* The machine from which you will perform the backup has sufficient
  internet access to retrieve updated software
* Your cluster is running normally
* Your Juju client and controller/models are running the same,
  stable version of Juju (see the [Juju docs][juju-docs])
* You read the [upstream release notes][upstream-notes] for details
  of Kubernetes deprecation notices and API changes that may impact
  your workloads

It is also important to understand that Kubernetes will only
upgrade and if necessary migrate, components relating specifically
to elements of Kubernetes installed and configured as part of Kubernetes.
This may not include any customized configuration of Kubernetes,
or non-built-in generated objects (e.g. storage classes) or deployments which
rely on deprecated APIs.

## Get the current version 

Determine which version of each application is currently deployed by running:

<!-- markdownlint-disable -->

```
juju status
```

<!-- markdownlint-restore -->

The ‘App’ section of the output lists each application and its
version number. Note that this is the version of the upstream
application deployed. The version of the Juju charm is indicated
under the column titled ‘Rev’. The charms may be updated in
between new versions of the application.

<!-- markdownlint-disable -->

```
Model       Controller  Cloud/Region   Version  SLA          Timestamp
my-cluster  canonicaws  aws/us-east-1  3.6.0    unsupported  16:02:18-05:00

App      Version  Status  Scale  Charm    Channel        Rev  Exposed  Message
k8s      1.31.3   active      3  k8s      1.31/stable    123  yes      Ready

Unit        Workload  Agent  Machine  Public address  Ports     Message
k8s/0       active    idle   0        54.89.153.117   6443/tcp  Ready
k8s/1*      active    idle   1        3.238.230.3     6443/tcp  Ready
k8s/2       active    idle   2        34.229.202.243  6443/tcp  Ready

Machine  State    Address         Inst id              Base          AZ          Message
0        started  54.89.153.117   i-0b6fc845c28864913  ubuntu@22.04  us-east-1f  running
1        started  3.238.230.3     i-05439714c88bea35f  ubuntu@22.04  us-east-1f  running
2        started  34.229.202.243  i-07ecf97ed29860334  ubuntu@22.04  us-east-1c  running
```

<!-- markdownlint-restore -->

(patch-upgrades)=

## Patch upgrades 

Kubernetes releases patch versions approximately every month. These updates
focus on making minor improvements without introducing new features.

### Check if an upgrade is available

Juju will contact [Charmhub] daily to find new revisions of charms
deployed in your models. To see if the `k8s` or `k8s-worker` charms
can be upgraded, set with the following:

```
juju status --format=json | \
   jq '.applications |
        to_entries[] | {
           application: .key,
           "charm-name": .value["charm-name"],
           "charm-channel": .value["charm-channel"],
           "charm-rev": .value["charm-rev"],
           "can-upgrade-to": .value["can-upgrade-to"]
        }'
```

This outputs a list of applications in the model:

* the name of the application (ex. `k8s`)
* the charm used by the application (ex. `k8s`)
* the kubernetes channel this charm follows (ex. `1.31/stable`)
* the current charm revision  (ex. `1001`)
* the next potential charm revision (ex. `ch:amd64/k8s-1002`)

If the `can-upgrade-to` revision is `null`, the charm is on the most
stable release within this channel. If your aim is to get the latest patch 
updates in this charm channel, then there is nothing more to do. Otherwise 
continue with the `pre-upgrade-check`.

### Run the pre-upgrade-check

Before running any upgrade, check that the cluster is
steady and ready for upgrade. The charm will perform checks
necessary to confirm the cluster is in safe working order before
upgrading.

```
juju run k8s/leader pre-upgrade-check
```

If no error appears, the `pre-upgrade-check` completed successfully.

### Refresh control plane units (k8s)

Update the control-plane nodes:

```
juju refresh k8s
juju status k8s --watch 5s
```

The `refresh` command instructs the Juju controller to use the new charm
revision within the current charm `channel`. The charm code is simultaneously
replaced on each unit, then the `k8s` snap is updated unit-by-unit in order
to maintain a highly-available kube-api-server endpoint, starting with the
Juju leader unit for the application.

During the upgrade process, the application status message and the
`k8s` leader unit message will display the current progress,
listing the `k8s` and `k8s-worker` units still pending upgrades.

After the `k8s` charm is upgraded, the application `Version` from `juju status`
will reflect the updated version of the control-plane nodes making up the
cluster.

### Refresh worker units (k8s-worker)

After updating the control-plane applications, worker nodes may be upgraded
after running the `pre-upgrade-check` action.

```
juju run k8s-worker/leader pre-upgrade-check
juju refresh k8s-worker
juju status k8s-worker --watch 5s
```

After the `k8s-worker` charm is upgraded, the application `Version` from
`juju status`
will reflect the updated version of the worker nodes making up the cluster.

Repeat these steps on refreshing worker units for every
application using the k8s-worker charm, if multiple k8s-worker
applications appear in the same model.

### Verify the patch upgrade 

Once the patch upgrade is complete, confirm the successful upgrade by running:

```
juju status
```

... should indicate that all units are active/idle and the correct
version of **Kubernetes** is listed in the application's **Version**.

It is recommended that you run a [cluster validation][cluster-validation]
to ensure that the cluster is fully functional.

(minor-upgrades)=

## Minor upgrades

Once you are sure you have the latest patch updates in your current channel, now
you can upgrade the minor version. New minor 
versions of Kubernetes are set to release three
times per year. They typically contain significant changes and new functionality
while being backward compatible. 

```{caution} Only update the charm to the next minor version.
If the current `charm-channel` is `1.31/stable`, it's critical
to refresh to the `1.32/stable`. Skipping channels (e.g. 1.31 -> 1.33)
will result in the units blocking and indicating they cannot upgrade.

See Kubernetes' [version skew policy]
```

### Run the pre-upgrade-check

Before running any upgrade, check that the cluster is
steady and ready for upgrade. You must run this command between each upgrade. 

```
juju run k8s/leader pre-upgrade-check
```

If no error appears, the `pre-upgrade-check` completed successfully.

### Refresh the control plane units (k8s)

```sh
juju refresh k8s --channel ${NEXT_CHANNEL}
juju status k8s --watch 5s
```

The `refresh` command instructs the Juju controller to follow a new
charm `channel`. The Kubernetes charm will be upgraded to the latest
revision within that channel. The charm code is simultaneously replaced
on each unit, then the `k8s` snap is updated unit-by-unit in order to
maintain a highly-available kube-api-server endpoint, starting with the
Juju leader unit for each application.

During the upgrade process, the application status message and the
`k8s` leader unit message will display the current progress,
listing the `k8s` and `k8s-worker` units still pending upgrades.

After the `k8s` charm is upgraded, the application `Version` from `juju status`
will reflect the updated version of the control-plane nodes making up the
cluster.

### Refresh worker units (k8s-worker)

After updating the control-plane applications, worker nodes may be upgraded
after running the `pre-upgrade-check` action.

```sh
juju run k8s-worker/leader pre-upgrade-check
juju refresh k8s-worker --channel ${NEXT_CHANNEL}
juju status k8s-worker --watch 5s
```

After the `k8s-worker` charm is upgraded, the application `Version`
from `juju status`
will reflect the updated version of the worker nodes making up the cluster.

Repeat these steps on refreshing worker units for every
application using the k8s-worker charm, if multiple k8s-worker
applications appear in the same model.

### Verify the minor upgrade 

Once the minor upgrade is complete, confirm the successful upgrade by running:

```
juju status
```

... should indicate that all units are active/idle and the correct
version of **Kubernetes** is listed in the application's **Version**.

It is recommended that you run a [cluster validation][cluster-validation]
to ensure that the cluster is fully functional.

## Kubernetes workers' phased upgrade with node draining

By default, the charm upgrades the `k8s` snap on each node without draining
workloads first. The snap refresh stops and restarts all Kubernetes services
immediately, making the node temporarily unavailable to the scheduler with no
warning to running pods.

For clusters running sensitive workloads, a phased approach allows you to
cordon and drain each node before the snap is refreshed, giving pods time to
reschedule gracefully.

```{caution}
You must run `pre-upgrade-check` **before** issuing `juju refresh`. Without it,
the charm will set every unit to failed with `"Unit was upgraded without a
pre-upgrade-check"`.
```

```{note}
The phased process below can also be applied to Kubernetes control plane nodes,
except that they don't need to be drained. If your control plane nodes
also run as worker nodes, they will need to be drained.
```

### 1. Set the target channel

The `k8s` snap channel versioning *is not* the same as the charm channel
versions. Instead, it follows the following versioning scheme:

| Kubernetes Version | Charm Channel | Snap Channel |
|---|---|---|
| v1.34.4 | 1.34/stable | 1.34-classic/stable |

### 2. Run pre-upgrade-check

```sh
export CHARM_NEXT_CHANNEL=1.35/stable   # adjust to your target
export SNAP_NEXT_CHANNEL=1.35-classic/stable   # adjust to your target

juju run k8s-worker/leader pre-upgrade-check
```

### 3. Drain and upgrade worker nodes one at a time

Repeat the following block for **each worker node**, substituting
`NODE` and `UNIT`:

```sh
NODE="<kubectl-node-name>"       # from: kubectl get nodes
UNIT="k8s-worker/<N>"            # from: juju status

# Cordon – stop new pods scheduling on this node
kubectl cordon "$NODE"

# Drain – evict all workloads gracefully
kubectl drain "$NODE" \
  --ignore-daemonsets \
  --delete-emptydir-data \
  --grace-period=60 \
  --timeout=300s

# Confirm only DaemonSet pods remain
kubectl get pods -A --field-selector spec.nodeName="$NODE"

# Refresh the snap on this node only
juju exec --unit "$UNIT" \
  "sudo snap refresh k8s --channel=$SNAP_NEXT_CHANNEL \
  && sudo snap set k8s refresh.hold=forever"

# Wait for the node to re-join as Ready
kubectl wait node "$NODE" --for=condition=Ready --timeout=300s

# Uncordon – allow workloads back onto the node
kubectl uncordon "$NODE"

# Confirm pods are rescheduling before moving to the next node
kubectl get pods -A | grep -vE 'Running|Completed|Terminating'
```

### 4. Run juju refresh to update the charm code

All snaps are already at the target revision. The `juju refresh` below updates
only the charm code; the snap step is skipped on every unit.

```sh
juju refresh k8s-worker --channel ${CHARM_NEXT_CHANNEL}
juju status k8s-worker --watch 5s
```

### 5. Verify

```
juju status
```

All units should be `active/idle` and the correct version of **Kubernetes**
should be listed in the application's **Version** column.

It is recommended that you run a [cluster validation][cluster-validation]
to ensure the cluster is fully functional.

---

## Recover from a failed upgrade 

If anything goes wrong during the upgrade, the Juju application will surface 
the error in the application message, indicating that the unit failed to 
upgrade. See the Juju debug logs for more details.

If you are upgrading a control plane node:

```
juju debug-log --include=k8s
```

If you are upgrading a worker node:

```
juju debug-log --include=k8s-worker
```

<!-- LINKS -->

[Kubernetes release page]: https://kubernetes.io/releases/
[backup-restore]:      ../../snap/howto/backup-restore
[Charmhub]:            https://charmhub.io/k8s
[cluster-validation]:  ./validate
[juju-docs]:           https://documentation.ubuntu.com/juju/3.6/howto/manage-models/
[upstream-notes]:      https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.35.md#deprecation
[version skew policy]: https://kubernetes.io/releases/version-skew-policy/
