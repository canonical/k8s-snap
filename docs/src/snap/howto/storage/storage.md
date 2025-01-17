# How to use default storage

{{product}} offers a local-storage option to quickly set up and run a
cluster, especially for single-node support. This guide walks you through
enabling and configuring this feature.

```{warning}
Local storage will not survive node removal, so it may not be suitable for
certain setups, such as production multi-node clusters.
```

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstrapped {{product}} cluster (see the
  [getting-started-guide])

## Enable Local Storage

The storage feature is enabled by default when bootstrapping the snap. However,
if you have used a custom bootstrap configuration or disabled the feature, you
can enable it using the following command:

```
sudo k8s enable local-storage
```

## Configure Local Storage

While the storage option comes with sensible defaults, you can customise it to
meet your requirements. Obtain the current configuration by running:

```
sudo k8s get local-storage
```

You can modify the configuration using the `set` command. For example, to
change the local storage path:

```
sudo k8s set local-storage.local-path=/path/to/new/folder
```

The local-storage feature provides the following configuration options:

- `local-path`: path where the local files will be created.
- `reclaim-policy`: set the reclaim policy of the persistent volumes
  provisioned. It should be one of "Retain", "Recycle", or "Delete".
- `default`: set the local-storage storage class to be the default. If
  this flag is not set and the cluster already has a default storage class it
  is not changed. If this flag is not set and the cluster does not have a
  default class set then the class from the local-storage becomes the default.

## Disable Local Storage

The local storage option is only suitable for single-node clusters and
development environments as it has no multi node data replication. For a
production environment you may want a more sophisticated storage solution. To
disable local-storage, run:

```
sudo k8s disable local-storage
```

Disabling storage only removes the CSI driver. The persistent volume claims
will still be available and your data will remain on disk.

<!-- LINKS -->
[getting-started-guide]: ../../tutorial/getting-started.md
