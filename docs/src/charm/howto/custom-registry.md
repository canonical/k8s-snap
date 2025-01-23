# How to configure a custom registry

The `k8s` charm can be configured to use a custom container registry for its
container images. This is particularly useful if you have a private registry or
operate in an air-gapped environment where you need to pull images from a
different registry. This guide will walk you through the steps to set up `k8s`
charm to pull images from a custom registry.

## Prerequisites

- A running `k8s` charm cluster.
- Access to a custom container registry from the cluster (e.g., docker registry
  or Harbor).

## Configure the charm

To configure the charm to use a custom registry, you need to set the
`containerd_custom_registries` configuration option. This options allows
the charm to configure `containerd` to pull images from registries that require
authentication. This configuration option should be a JSON-formatted array of
credential objects. For more details on the `containerd_custom_registries`
option, refer to the [charm configurations] documentation.

For example, to configure the charm to use a custom registry at
`myregistry.example.com:5000` with the username `myuser` and password
`mypassword`, set the `containerd_custom_registries` configuration option as
follows:

```
juju config k8s containerd_custom_registries='[{
    "url": "http://myregistry.example.com:5000",
    "host": "myregistry.example.com:5000",
    "username": "myuser",
    "password": "mypassword"
}]'
```

Allow the charm to apply the configuration changes and wait for Juju to
indicate that the changes have been successfully applied. You can monitor the
progress by running:

```
juju status --watch 2s
```

## Verify the configuration

Once the charm is configured and active, verify that the custom registry is
configured correctly by creating a new workload and ensuring that the images
are being pulled from the custom registry.

For example, to create a new workload using the `nginx:latest` image that you
have previously pushed to the `myregistry.example.com:5000` registry, run the
following command:

```
kubectl run nginx --image=myregistry.example.com:5000/nginx:latest
```

To confirm that the image has been pulled from the custom registry and that the
workload is running, use the following command:

```
kubectl get pod nginx -o jsonpath='{.spec.containers[*].image}{"->"}{.status.containerStatuses[*].ready}'
```

The output should indicate that the image was pulled from the custom registry
and that the workload is running.

```
myregistry.example.com:5000/nginx:latest->true
```

<!-- LINKS -->

[charm configurations]: https://charmhub.io/k8s/configurations
