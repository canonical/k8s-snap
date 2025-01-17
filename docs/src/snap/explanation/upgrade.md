# Upgrading {{product}}

Upgrading your cluster to the latest Kubernetes version is a critical part of
maintaining a healthy cluster. It provides enhanced security in the form
of the latest security patches as well as increased stability and performance
by incorporating the latest optimizations and features. The [release cadence]
of {{product}} follows the [release cycle] set by upstream Kubernetes.


## Patch vs minor upgrades

Patch upgrades address bug fixes and are typically safe, introducing no
breaking changes.
Patch upgrades do not change the channel selected (for example a patch upgrade
from stable/1.32.0 to stable/1.32.1). Minor version upgrades add new
features or deprecate existing features without breaking changes. These upgrades
change the version of the channel selected
(for example a minor upgrade from stable/1.32.x to stable/1.33.x).
For more information on channels see the [explanation page].

## Sequential upgrades

Upgrades of control plane nodes
must always be done sequentially. It is not recommended to upgrade by
more than one minor version at a time (for example upgrading from 1.31 to 1.33
is **not** supported). When upgrading between minor releases, for instance from
1.31 to 1.32, first upgrade to the
latest patch of the current version. In this example you would check the 1.31
channel for the latest patch: 1.31.x. This is in line with upstream Kubernetes
[version skew policy]. For worker nodes, upgrading
with a skew of up to three versions is acceptable.

## Upgrade approach

It is important to upgrade your cluster with the same method that you used to install
the cluster. If you have deployed {{product}} using Juju, you should upgrade your cluster
using Juju controller commands. Similarly, with snap and CAPI deployments,
the original install method should be used.


## Snap upgrades

Updates for the {{product}} snap are checked several times a day with the
default [snap refresh functionality]. Patch upgrades will be applied
automatically unless they are manually stopped or a version of the snap is
pinned. It is not recommended to pin the version of your snap unless you have a
justified reason for doing so. Pinning the version leads to a less secure and
reliable deployment as the snap will not receive the latest updates. Changes
applied during the latest refresh to the snap can be viewed using the command
`snap change`. Minor version upgrades need manual intervention and must be
carried out on all snaps individually. The length of time taken to upgrade the
snap will depend largely on your local setup, the services running on the
cluster, the resources available to the cluster, the intensity of the workloads
etc.

If you would like to manage the upgrading of your cluster using the snap, please
see the how-to guide on [managing cluster upgrades].

## Charm upgrades

Upgrading the charm must be done manually for both patch and minor upgrades on
a per model basis. To ensure that the revision of the charm matches the
revision of the snap it orchestrates, the automatic `refresh` function of the
snap has been placed on hold in this deployment method. The snap can still be
manually updated by running a [targeted snap refresh], but all upgrades should
be done through Juju. The `juju refresh`command instructs the Juju controller
to use the new charm revision within the current charm channel or to upgrade to
the next channel based on the parameters provided. The charm code
is simultaneously replaced on each unit. Then, the k8s snap is updated
unit-by-unit. This is in order to maintain a highly-available kube-api-server
endpoint, starting with the Juju leader unit for the application. To ensure a
smooth upgrade process, the mandatory pre-upgrade-check is run before
all upgrades. In addition to verifying the cluster's health,
the pre-upgrade-check also creates the upgrade stack which
will be used after the charm is upgraded. The upgrade stack is
used to orchestrate the upgrade process across the k8s or k8s-worker units.

If you would like to upgrade your cluster using Juju please see the how-to guide
on [upgrading your cluster by a minor version].

<!-- TODO CAPI Upgrades-->

## High availability with cluster upgrades

High availability is automatically enabled in {{ product }} for clusters with
three or more nodes independent of the deployment method. Clusters without high
availability must be extra vigilant on [backing up cluster data] before starting
the upgrade process and also must be aware of potential service disruptions
during cluster upgades. It is also important to understand that Kubernetes will
only upgrade and if necessary migrate, components of Kubernetes installed and
configured as part of Kubernetes. This may
not include any customized configuration of Kubernetes, or no-build-in
generated objects (e.g. storage classes) or deployments which rely on
deprecated APIs. To find out more about high availability in a {{product}}
cluster see the [high availability explanation page].

## Updating a cluster best practices

When upgrading with {{product}} best practices should be followed:

- Check the [release notes] associated with the release and pay
particular attention to any deprecations or incompatible changes being
introduced that may affect your cluster.
- Create a backup of all the information in the cluster.
- Upgrade the control plane nodes before upgrading the worker nodes.
- Cluster administrators should identify critical workloads on the node to be
updated and decide if the node needs to be cordoned and drained before cluster
upgrade.
- Deploy a rolling upgrade strategy meaning you upgrade one node successfully
to completion before upgrading the next node.

<!-- LINKS -->
[release cadence]: https://ubuntu.com/about/release-cycle#canonical-kubernetes-release-cycle
[release cycle]: https://kubernetes.io/releases/release/
[managing cluster upgrades]: ../howto/upgrades
[upgrading your cluster by a minor version]: ../../charm/howto/upgrade-minor/
[snap refresh functionality]:https://snapcraft.io/docs/refresh-awareness
[version skew policy]: https://kubernetes.io/releases/version-skew-policy/
[explanation page]: channels.md
[high availability explanation page]: high-availability.md
[targeted snap refresh]:https://snapcraft.io/docs/managing-updates#p-32248-if-snaps-are-specified
[release notes]: /src/releases
[backing up cluster data]: /src/snap/howto/backup-restore.md
