# Perform an in-place upgrade for a machine

This guide walks you through the steps to perform an in-place upgrade for a
Cluster API managed Machine.

## Prerequisites

To follow this guide, you will need:

- A Kubernetes management cluster with Cluster API and providers installed
  and configured.
- A target workload cluster managed by CAPI.
- `kubectl` installed and configured to access your management cluster.
- The workload cluster kubeconfig.

Please refer to the [getting-started guide][getting-started] for further
details on the required setup.
This guide refers to the workload cluster as `c1` and its
kubeconfig as `c1-kubeconfig.yaml`.

## Check the current cluster status

Prior to the upgrade, ensure that the management cluster is in a healthy
state.

```
kubectl get nodes -o wide
```

Confirm the Kubernetes version of the workload cluster:

```
kubectl --kubeconfig c1-kubeconfig.yaml get nodes -o wide
```

## Annotate the machine

In this first step, annotate the Machine resource with 
the in-place upgrade annotation. In this example, the machine
is called `c1-control-plane-xyzbw`.

```
kubectl annotate machine c1-control-plane-xyzbw "v1beta2.k8sd.io/in-place-upgrade-to=<upgrade-option>"
```

`<upgrade-option>` can be one of:
* `channel=<snap-channel>` which refreshes k8s to the given snap channel. e.g. `channel=1.30-classic/stable`
* `revision=<revision>` which refreshes k8s to the given revision. e.g. `revision=123`
* `localPath=<path>` which refreshes k8s with the snap file from the given absolute path. `localPath=full/path/to/k8s.snap`

Please refer to the [ClusterAPI Annotations Reference][capi-annotations-reference] for further
details on these options.

## Monitor the in-place upgrade

Watch the status of the in-place upgrade for the machine, by running the
following command and checking the `v1beta2.k8sd.io/in-place-upgrade-status` annotation:

```
kubectl get machine c1-control-plane-xyzbw -o yaml
```

On a successful upgrade:
* Value of the `v1beta2.k8sd.io/in-place-upgrade-status` annotation will be changed to `done`
* Value of the `v1beta2.k8sd.io/in-place-upgrade-release` annotation will be changed to the `<upgrade-option>` used to perform the upgrade.

## Cancelling a failing upgrade
The upgrade is retried periodically if the operation was unsuccessful.

The upgrade can be cancelled by running the following commands that remove the annotations:

```
kubectl annotate machine c1-control-plane-xyzbw "v1beta2.k8sd.io/in-place-upgrade-to-"
kubectl annotate machine c1-control-plane-xyzbw "v1beta2.k8sd.io/in-place-upgrade-change-id-"
```

## Verify the Kubernetes upgrade

Confirm that the node is healthy and runs on the new Kubernetes version:

```
kubectl --kubeconfig c1-kubeconfig.yaml get nodes -o wide
```


<!-- LINKS -->
[getting-started]: ../tutorial/getting-started.md
[capi-annotations-reference]: ../reference/annotations.md
