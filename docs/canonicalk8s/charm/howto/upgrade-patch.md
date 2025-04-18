# How to upgrade {{product}} to the next patch revision

It is recommended that you keep your Kubernetes deployment
updated to the latest available stable version. You should
also update the other applications deployed in your Kubernetes
cluster. Keeping up-to-date ensures you have the latest bug-fixes
and security patches for smooth operation of your cluster.

Kubernetes releases patch versions approximately every month. These updates
focus on making minor improvements without introducing new features. You
can check the latest release version on the [Kubernetes release page].

```{note} Kubernetes will not automatically handle patch
releases. The cluster will not perform an unattended automatic
upgrade between patch versions, e.g. 1.31.0 to 1.31.1.
Attended upgrades are required when you wish to upgrade
whether to a patch or minor version.
```

You can see which version of each application is currently deployed by running:

<!-- markdownlint-disable -->
```sh
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


## Before you begin

As with all upgrades, there is a possibility that there may be
unforeseen difficulties. It is highly recommended that you make
a backup of any important data, including any running workloads.
For more details on creating backups, see the separate
[docs on backups][backup-restore].


You should also make sure:

* The machine from which you will perform the backup has sufficient
  internet access to retrieve updated software
* Your cluster is running normally
* Your Juju client and controller/models are running the same,
  stable version of Juju (see the [Juju docs][juju-docs])
* You read the [Upstream release notes][upstream-notes] for details
  of Kubernetes deprecation notices and API changes that may impact
  your workloads


It is also important to understand that Kubernetes will only
upgrade and if necessary migrate, components relating specifically
to elements of Kubernetes installed and configured as part of Kubernetes.
This may not include any customized configuration of Kubernetes,
or no-build-in generated objects (e.g. storage classes) or deployments which
rely on deprecated APIs.

## Specific upgrade instructions

### Decide if an upgrade is available

Juju will contact [charmhub] daily to find new revisions of charms
deployed in your models. To see if the `k8s` or `k8s-worker` charms
can be upgraded, set with the following:

```sh
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

This outputs list of applications in the model:

* the name of the application (ex. `k8s`)
* the charm used by the application (ex. `k8s`)
* the kubernetes channel this charm follows (ex. `1.31/stable`)
* the current charm revision  (ex. `1001`)
* the next potential charm revision (ex. `ch:amd64/k8s-1002`)

If the `can-upgrade-to` revision is `null`, the charm is on the most
stable release within this channel and there is no patch upgrade to
complete.

Otherwise continue with the pre-upgrade-check.


### The pre-upgrade-check

Before running an upgrade, check that the cluster is
steady and ready for upgrade. The charm will perform checks
necessary to confirm the cluster is in safe working order before
upgrading.

```sh
juju run k8s/leader pre-upgrade-check
```

If no error appears, the `pre-upgrade-check` completed successfully.


### Refresh charm applications

#### Control Plane units (k8s)

Following the `pre-upgrade-check` update the control-plane nodes.

```sh
juju refresh k8s
juju status k8s --watch 5s
```

The `refresh` command instructs the juju controller to use the new charm
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

#### Worker units (k8s-worker)

After updating the control-plane applications, worker nodes may be upgraded
by running the `pre-upgrade-check` action.

```sh
juju run k8s-worker/leader pre-upgrade-check
juju refresh k8s-worker
juju status k8s-worker --watch 5s
```

The `refresh` command instructs the juju controller to use the new charm
revision of the application's charm channel to upgrade each unit. The
charm code is simultaneously replaced on each unit, then the `k8s`
snap is updated unit-by-unit starting with the Juju leader unit for the
application.

After the `k8s-worker` charm is upgraded, the application `Version` from
`juju status`
will reflect the updated version of the worker nodes making up the cluster.

```{note} Repeat [this section](#worker-units-k8s-worker) for every
application using the k8s-worker charm, if multiple k8s-worker
applications appear in the same model.
```


## Verify an upgrade

Once an upgrade is complete, confirm the successful upgrade by running:

<!-- markdownlint-disable -->
```sh
juju status
```
<!-- markdownlint-restore -->

... should indicate that all units are active/idle and the correct
version of **Kubernetes** is listed in the application's **Version**

It is recommended that you run a [cluster validation][cluster-validation]
to ensure that the cluster is fully functional.

<!-- LINKS -->
[Kubernetes release page]: https://kubernetes.io/releases/
[backup-restore]:     ../../snap/howto/backup-restore
[charmhub]:            https://charmhub.io/k8s
[cluster-validation]: ./validate
[juju-docs]:          https://documentation.ubuntu.com/juju/3.6/howto/manage-models/#how-to-manage-models
[release-notes]:      ../reference/releases
[upstream-notes]:     https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG
