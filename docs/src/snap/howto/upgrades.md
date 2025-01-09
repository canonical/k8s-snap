# Upgrading the snap

Upgrading the Kubernetes version of a node is a critical operation that 
requires careful planning and execution. {{product}} is shipped as a snap,
which simplifies the upgrade process.
This how-to guide will cover the steps to upgrade the {{product}} snap to a 
new version as well as how to freeze upgrades.

## Important Considerations Before Upgrading

- According to the [upstream Kubernetes][1], skipping **minor** versions while 
upgrading is not supported. For more details, please visit the
[Version Skew Policy][2].
- Before performing an upgrade, it's important to back up the cluster data. 
This can be done by following the steps outlined in the [backup guide][3].
- For more information on managing snap updates, please refer to the 
[snap documentation][4].

## Patch Upgrade

Patch upgrades address bug fixes and are typically safe, introducing no 
breaking changes.
{{product}} snap is installed using a specific track (e.g. `1.32-classic`). 
Snaps automatically check for updates several times a day and apply 
them when available.
These updates ensure the latest changes in the installed track are applied.
Patch upgrades can also be triggered manually by following the steps below.

1. **List available revisions:**
```
snap info k8s
```

2. **Refresh the snap:**
```
snap refresh k8s
```

3. **Verify the upgrade:**
Ensure that the upgrade was successful by checking the version of the snap and 
confirming that the cluster is ready:
```
snap info k8s
sudo k8s status --wait-ready
```

## Minor Version Upgrade

Minor versions add new features or deprecate existing features without 
breaking changes.
To upgrade to a new minor version, the snap channel needs to be changed.


1. **List available channels:**
```
snap info k8s
```

2. **Change the snap channel:**
The {{product}} snap channel can be changed by using the `snap refresh` 
command.
```
snap refresh --channel=1.33/stable k8s
```

3. **Verify the upgrade:**
Ensure that the upgrade was successful by checking the version of the snap 
and confirming that the cluster is ready:
```
snap info k8s
sudo k8s status --wait-ready
```

```{note}
In a multi-node cluster, the upgrade should be performed on all nodes.
```

## Freezing Upgrades

To prevent automatic updates, the snap can be frozen to a specific revision. 
`snap refresh --hold[=<duration>]` holds refreshes for a specified duration 
(or forever, if no value is specified).
```
snap refresh k8s --hold
```
Or specify a time window:
```
snap refresh k8s --hold=24h
```

<!-- LINKS -->
[1]: https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/
[2]: https://kubernetes.io/docs/setup/release/version-skew-policy/
[3]: ./backup-restore.md
[4]: https://snapcraft.io/docs/managing-updates
[5]: ../../charm/index.md
[6]: ../../capi/index.md
