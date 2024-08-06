# Upgrade the Kubernetes version of a cluster

This guide walks you through the steps to rollout an upgrade for a
Cluster API managed Kubernetes cluster. The upgrade process includes updating
the Kubernetes version of the control plane and worker nodes.

## Prerequisites

To follow this guide, you will need:

- A Kubernetes management cluster with Cluster API and providers installed
  and configured.
- A target workload cluster managed by CAPI.
- `kubectl` installed and configured to access your management cluster.
- The workload cluster kubeconfig. We will refer to it as `c1-kubeconfig.yaml`
  in the following steps.

Please refer to the [getting-started guide][getting-started] for instructions
on how to set this up.
In the following steps we will refer to the workload cluster as `c1` and its
kubeconfig as `c1-kubeconfig.yaml`.

## Check the current cluster status

Before starting the upgrade, ensure your management cluster is in a healthy
state.

```
kubectl get nodes -o wide
```

Also, verify that the workload cluster runs on the expected Kubernetes version

```
kubectl --kubeconfig c1-kubeconfig.yaml get nodes -o wide
```

## Update the CK8sControlPlane resource

The first step in upgrading the control plane is to update the CK8sControlPlane
resource with the new Kubernetes version. In this example, the control plane
is called `c1-control-plane`.

```
kubectl edit ck8scontrolplane c1-control-plane
```

Replace the `spec.version` field with the desired Kubernetes version.

```yaml
spec:
  version: v1.30.3
```

Save and exit the editor.

## Monitor the upgrade process

CAPI will handle the rolling upgrade of control plane nodes.
You can watch the progress with

```
kubectl get ck8scontrolplane c1-control-plane -w
```

Inspect the current machines with

```
kubectl --kubeconfig c1-kubeconfig.yaml get nodes -o wide
```

The machines will be replaces one after each other until all machines run on
the desired version.

## Update the MachineDeployment resource

After upgrading the control plane, proceed with upgrading the worker nodes
in a similar fashion by updating the `MachineDeployment` resource. For
instance, we will be updating the `c1-worker-md`.

```
kubectl edit machinedeployment c1-worker-md
```

Update the `spec.template.spec.version` field with the desired
Kubernetes version.

```yaml
spec:
  template:
    spec:
      version: v1.30.3
```

Save and exit the editor.

## Monitor the worker node upgrade

Similar to the control-planes, you can monitor the upgrade with

```
kubectl get machinedeployment c1-worker-md
```

## Verify the upgrade

Finally, verify that all nodes are running the new version and are healthy:

```
kubectl --kubeconfig c1-kubeconfig.yaml get nodes -o wide
```

and that no old machines are left behind:

```
kubectl get machines -A
```

<!-- LINKS -->
[getting-started]: ../tutorial/getting-started.md
