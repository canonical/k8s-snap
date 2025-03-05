# How to add worker nodes with custom configurations

This guide will walk you through how to deploy multiple `k8s-worker`
applications with different configurations, to create node groups with specific
capabilities or requirements.

## Prerequisites

This guide assumes the following:

- A working Kubernetes cluster deployed with the `k8s` charm

## Example worker configuration

In this example, we will create two `k8s-worker` applications with different
configuration options.

```{note}
The configurations shown below are examples to demonstrate the deployment
pattern. You should adjust the node configurations, labels, and other
parameters according to your specific infrastructure requirements, workload
needs, and organizational policies. Review the [charm configuration] options
documentation to understand all available parameters that can be customized for
your worker nodes.
```

1. Workers for memory-intensive workloads (`worker-memory-config.yaml`):

```yaml
memory-workers:
  bootstrap-node-taints: "workload=memory:NoSchedule"
  kubelet-extra-args: "system-reserved=memory=2Gi"
```

2. Workers for GPU workloads (`worker-gpu-config.yaml`):

```yaml
gpu-workers:
  bootstrap-node-taints: "accelerator=nvidia:NoSchedule"
  node-labels: "gpu=true"
```

Deploy the worker applications with the custom configurations and integrate them
with the `k8s` application:

```bash
juju deploy k8s-worker memory-workers --config ./worker-memory-config.yaml
juju integrate k8s memory-workers:cluster
juju integrate k8s memory-workers:containerd
juju integrate k8s memory-workers:cos-tokens

juju deploy k8s-worker gpu-workers --config ./worker-gpu-config.yaml
juju integrate k8s gpu-workers:cluster
juju integrate k8s gpu-workers:containerd
juju integrate k8s gpu-workers:cos-tokens
```

Monitor the installation progress by running the following command:

```bash
juju status --watch 1s
```

<!-- LINKS -->
[charm configuration]: https://charmhub.io/k8s/configurations
