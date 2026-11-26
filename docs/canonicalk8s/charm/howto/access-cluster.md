---
myst:
  html_meta:
    description: "How to access a Canonical Kubernetes charm cluster with kubectl, either from a control-plane unit or from your own workstation using the get-kubeconfig action."
---

# How to access the cluster with kubectl

This guide shows how to run `kubectl` commands against a {{product}} cluster
deployed with the `k8s` charm. You can use the `kubectl` snap on your own
machine with a kubeconfig retrieved from the cluster, or run `kubectl` on a
control-plane unit.

## Prerequisites

- A running Canonical Kubernetes cluster deployed with the `k8s` charm
- The Juju client, with access to the model the cluster is deployed in.

## Use the `kubectl` snap

To run `kubectl` from your own machine you need to install the `kubectl` and `yq` snaps and extract the kubeconfig.

### Install `kubectl` and `yq`

`kubectl` is available as a snap. Other installation methods are available in the [Kubernetes documentation](https://kubernetes.io/docs/tasks/tools/#kubectl).
Install the snap with:

```
sudo snap install kubectl --classic
```

Install [`yq`](https://github.com/mikefarah/yq) to process the kubeconfig data easier.

```
sudo snap install yq
```

### Get the kubeconfig

Retrieve the kubeconfig with the `get-kubeconfig` action. The action returns the file under the `kubeconfig`
key, which you can extract with `yq`:

```
juju run k8s/leader get-kubeconfig | yq .kubeconfig > cluster-kubeconfig.yaml
chmod 600 cluster-kubeconfig.yaml
```

```{warning}
The kubeconfig file contains credentials that grant access to the cluster.
Store it securely and do not share or commit it.
```

By default, the kubeconfig points to the API server through the address of the
unit that runs the action, or through the external load balancer if one is
integrated with the cluster. To use a different endpoint, pass the `server`
parameter (see the {ref}`get-kubeconfig action reference <k8s-action-get-kubeconfig>`).

If the API server on one unit is unreachable, run the action on a different
control-plane unit, for example `juju run k8s/<UNIT_NUMBER> get-kubeconfig`.

### Use kubectl

Pass the kubeconfig file to `kubectl` with the `--kubeconfig` flag:

```
kubectl --kubeconfig cluster-kubeconfig.yaml get nodes
```

All nodes should report a `Ready` status.

Or set the `KUBECONFIG` environment variable for the current shell:

```
export KUBECONFIG=$PWD/cluster-kubeconfig.yaml
kubectl get nodes
```

## Use kubectl on a control-plane unit

Alternatively, every {{product}} node includes `kubectl`, so you do not need to install
anything. Run it through `juju ssh` on a control-plane unit:

```
juju ssh k8s/leader -- sudo k8s kubectl get nodes
```

All nodes should report a `Ready` status.
