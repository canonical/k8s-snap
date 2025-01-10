# Upgrading with {{product}}

Upgrading your cluster to the lastest Kubernetes version is a critical part of
maintaining a healthy cluster. It provides enhanced security in the form
of the latest security patches as well as increased stability and performance
by incorporating the lastest optimizations and features. The [release cadence]
of {{product}} follows the [release cycle] set by upstream Kubernetes.


## Patch vs minor upgrades

Patch upgrades address bug fixes and are typically safe, introducing no
breaking changes.
Patch upgrades do not change the channel selected (for example a patch upgrade
from stable/1.32.0 to stable/1.32.1). Minor versions upgrades add new
features or deprecate existing features without breaking changes. These upgrades
change the version of the channel selected
(for example a minor upgrade from stable/1.32.x to stable/1.33.x).
For more information on channels see the [explanation page].

## Sequential upgrades

Upgrades must always be done sequentially. It is not recommended to upgrade by
more than one minor version at a time (for example upgrading from 1.31 to 1.33
is not supported). Changing between minor releases (for example upgrading
from 1.31 to 1.32) is also only recommended when you are on the
latest patch of the current version (for example 1.31.x where x is the latest
available patch in 1.31 channel). This is in line with upstream Kubernetes
[version skew policy].

## Upgrade approach

It is important to upgrade your cluster with the same method you used to install
the cluster. If you have deployed {{product}} using Juju, you must refresh
using Juju controller commands. Similarly with snap and CAPI deployments,
the original install method must be used.


## Snap upgrades

Updates for the {{product}} snap are checked several times a day with the
default [snap refersh fucntionality]. Patch upgrades will be applied
automatically unless they are manually stopped. Changes applied during the
latest refresh to the snap can be viewed using the command `snap change`.
Minor version upgrades need manual intervention and must be carried out on all
snaps individually. The length of time taken to upgrade the snap will depend
largely on your local setup, the services running on the cluster, resources
available to the cluster,the intensity of the workloads etc.

If you would like to manage the upgrading of your cluster using the snap please
see the how-to guide on [managing cluster upgrades].

## Charm upgrades

Upgrading the charm must be done manually for both patch and minor upgrades on
a per model basis. To ensure that the revision of the charm always matches the
revision of the snap it orchestrates, the `refresh` function of the snap has
been frozen and all upgrades must be done through Juju. The `juju refresh`
command instructs the Juju controller
to use the new charm revision within the current charm channel or to upgrade to
the next channel based on the paramaters provided. The charm code
is simultaneously replaced on each unit. Then, the k8s snap is updated
unit-by-unit. This is in order to maintain a highly-available kube-api-server
endpoint, starting with the Juju leader unit for the application. To ensure a
smoothe upgrade process, the pre-upgrade-check should be run before
all upgrades. This ensures that you cluster is in a stable and ready state for
an upgrade.

If you would like to upgrade your cluster using Juju please see the how-to guide
on [upgrading your cluster by a minor version].

<!-- TODO CAPI Upgrades-->

## High availability with cluster upgrades

High availability is automatically enabled in {{ product }} for clusters with
three or more nodes independent of the deployment method. Clusters without high
availability must be extra viligent on backing up cluster data before starting
the upgrade process and also must be aware of potential service disruptions
during cluster upgades. It is also important to understand that Kubernetes will
only upgrade and if necessary migrate, components relating specifically to
elements of Kubernetes installed and configured as part of Kubernetes. This may
not include any customized configuration of Kubernetes, or no-build-in
generated objects (e.g. storage classes) or deployments which rely on
deprecated APIs.

## Updating a cluster best practices

When upgrading with {{product}} best practices should be followed:

- Check the release notes associated with the release and pay
particular attention to any deprecations or imcompatible changes being
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
[snap refersh fucntionality]:https://snapcraft.io/docs/refresh-awareness
[version skew policy]: https://kubernetes.io/releases/version-skew-policy/
[explanation page]: channels.md
