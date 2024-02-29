# How to use default storage

Canonical Kubernetes offers a local storage option to quickly set up and run a
cluster, especially for single-node support. This guide walks you through
enabling and configuring this feature.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstraped Canonical Kubernetes cluster (see the
  [getting-started-guide])


## Enable Storage

When bootstrapping the snap, the storage functionality is not enabled by
default. To enable it, execute the following command:

```sh
sudo k8s enable storage
```

## Configure Storage

While the storage option comes with sensible defaults, you can customise it to
meet your requirements. Obtain the current configuration by running:

```sh
sudo k8s get storage
```

You can modify the configuration using the `set` command. For example, to
change the local storage path:

```sh
sudo k8s set storage.local-path=/path/to/new/folder
```

The storage functionality provides the following configuration options:

- **local-path**: path where the local files will be created.
- **reclaim-policy**: set the reclaim policy of the persistent volumes
  provisioned. It should be one of "Retain", "Recycle", or "Delete".
- **set-default**: set the local-storage storage class to be the default. If
  this flag is not set and the cluster has already a default storage class it
  is not changed. If this flag is not set and the cluster does not have a
  default class set then the class from the local-storage becomes the default
  one.

## Disable Storage

The local storage option is suitable for single-node clusters and development
environments, but it has inherent limitations. For a production environment you
typically want a more sophisticated storage solution. To disable the storage
functionality, run:

```
sudo k8s disable storage
```

Note that this will only remove the CSI driver. The persististent volume claim
will still be there and your data remain on disk.


<!-- LINKS -->
[getting-started-guide]: ../tutorial/getting-started.md