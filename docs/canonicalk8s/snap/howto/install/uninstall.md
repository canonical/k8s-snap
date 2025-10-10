# How to uninstall the {{product}} snap

This guide provides step-by-step instructions for removing the {{ product }}
snap from your system.

## Remove the node from the cluster

```{important}
You must remove the node from the cluster **before** deleting the snap.
Deleting the snap while the node is part of the cluster can cause the node to enter an unreachable state.
In this case, the node needs to be removed with the `--force` flag as explained in the [troubleshooting page].
```

From any control plane node:

```
sudo k8s remove-node <NODE_NAME>
```

Ensure the node has been removed from the cluster:

```
sudo k8s kubectl get nodes
```

## Remove the k8s snap

Uninstall the `k8s` snap:

```
sudo snap remove k8s
```

This command uninstalls the snap but may leave some configurations and data
files on the system.
For a complete removal, including all cluster data, use the `--purge` option:

```
sudo snap remove k8s --purge
```

## Confirm snap removal

To confirm the snap is successfully removed, check the list of installed
snaps:

```
snap list k8s
```

This command should produce an output similar to:

```
error: no matching snaps installed
```

<!-- Links -->
[troubleshooting page]: ../../reference/troubleshooting.md#remove-a-permanently-lost-node-from-the-cluster
