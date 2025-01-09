# How to Uninstall the {{ product }} Snap

This guide provides step-by-step instructions for removing the {{ product }}
snap from your system.

---

## Steps to Uninstall

### Remove the snap

To uninstall the `k8s` snap, run the following command:

```
sudo snap remove k8s
```

This command uninstalls the snap but may leave some configuration and data
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

which should produce an output similar to:

```
error: no matching snaps installed
```
