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

Please refer to the [getting-started guide][getting-started] for further details on the required setup.
This guide refers to the workload cluster as `c1` and its
kubeconfig as `c1-kubeconfig.yaml`.

## Check the current cluster status

Prior to the upgrade, ensure that the management cluster is in a healthy
state.

```
kubectl get nodes -o wide
```

Confirm that the Kubernetes version of the workload cluster:

```
kubectl --kubeconfig c1-kubeconfig.yaml get nodes -o wide
```

## Update the Control Plane

In this first step, update the CK8sControlPlane
resource with the new Kubernetes version. In this example, the control plane
is called `c1-control-plane`.

```
kubectl edit ck8scontrolplane c1-control-plane
```

Replace the `spec.version` field with the new Kubernetes version.

```yaml
spec:
  version: v1.30.3
```

Please safe your changes.

## Monitor the control plane upgrade

Watch CAPI handle the rolling upgrade of control plane nodes, by running the following command:

```
kubectl get ck8scontrolplane c1-control-plane -w
```

To inspect the current machines, execute:

```
kubectl --kubeconfig c1-kubeconfig.yaml get nodes -o wide
```

The machines will be replaced in turn until all machines run on
the desired version.

## Update the Worker nodes

After upgrading the control plane, proceed with upgrading the worker nodes
by updating the `MachineDeployment` resource. For
instance, we will be updating the `c1-worker-md`.

```
kubectl edit machinedeployment c1-worker-md
```

Update the `spec.template.spec.version` field with the new
Kubernetes version.

```yaml
spec:
  template:
    spec:
      version: v1.30.3
```

Please safe your changes.

## Monitor the worker node upgrade

Just like with the control planes, monitor the upgrade using:

```
kubectl get machinedeployment c1-worker-md
```

## Verify the Kubernetes upgrade

Confirm that all nodes are healthy and run on the new Kubernetes version:

```
kubectl --kubeconfig c1-kubeconfig.yaml get nodes -o wide
```

As a last step, ensure that no old machines are left behind:

```
kubectl get machines -A
```

<!-- LINKS -->
[getting-started]: ../tutorial/getting-started.md
