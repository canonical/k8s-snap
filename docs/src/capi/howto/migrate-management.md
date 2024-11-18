# Migrate the management cluster

Management cluster migration is a really powerful operation in the cluster’s lifecycle as it allows admins
to move the management cluster in a more reliable substrate or perform maintenance tasks without disruptions.
In this guide we will walk through the migration of a management cluster.

## Prerequisites

In the [Cluster provisioning with CAPI and {{product}} tutorial] we showed how to provision a workloads cluster. Here, we start from the point where the workloads cluster is available and we will migrate the management cluster to the one cluster we just provisioned.

## Install the same set of providers to the provisioned cluster

Before migrating a cluster, we must make sure that both the target and source management clusters run the same version of providers (infrastructure, bootstrap, control plane). To do so, `clusterctl init` should be called against the target cluster:

```
clusterctl get kubeconfig <provisioned-cluster> > targetconfig
clusterctl init --kubeconfig=$PWD/targetconfig --bootstrap ck8s --control-plane ck8s --infrastructure <infra-provider-of-choice>
```

## Move the cluster

Simply call:

```
clusterctl move --to-kubeconfig=$PWD/targetconfig
```

<!-- LINKS -->
[Cluster provisioning with CAPI and {{product}} tutorial]: ../tutorial/getting-started.md
