# Use default storage in Canonical Kubernetes

Canonical Kubernetes offers a local storage option to quickly set up and run a cluster, especially for single-node support. This guide walks you through enabling and configuring this feature.

Before you start, ensure you have a bootstrapped Canonical Kubernetes cluster ready.
Check out the [Getting Started](https://github.com/canonical/k8s-snap/blob/main/docs/src/tutorial/getting-started.md) guide to learn how to do this.

## Enable Storage
When bootstrapping the snap, the storage functionality is not enabled by default. To enable it, execute the following command:

```sh
sudo k8s enable storage
```

## Configure Storage
While the storage option comes with sensible defaults, you can customise it to meet your requirements. Obtain the current configuration by running:

```sh
sudo k8s get storage
```

> **Note**: For an explanation of each configuration option, refer to the [reference](#TODO) section. 

You can modify the configuration using the `set` command. For example, to change the local storage path:

```
sudo k8s set storage.local-path=/path/to/new/folder
```

## Disable Storage
The local storage option is suitable for single-node clusters and development environments, but it has inherent limitations. 
To disable the storage functionality, run:

```
sudo k8s disable storage
```
