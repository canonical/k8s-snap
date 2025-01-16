# Migrate the management cluster

Management cluster migration allows admins to move the management cluster 
to a different substrate or perform maintenance tasks without disruptions.
This guide walks you through the migration of a management cluster.

## Prerequisites

- A {{product}} CAPI management cluster with Cluster API and providers 
installed and configured.

## Configure the target cluster

Before migrating a cluster, ensure that both the target and source management 
clusters run the same version of providers (infrastructure, bootstrap, 
control plane). Use `clusterctl init` to target the cluster::

```
clusterctl get kubeconfig <provisioned-cluster> > targetconfig
clusterctl init --kubeconfig=$PWD/targetconfig --bootstrap canonical-kubernetes --control-plane canonical-kubernetes --infrastructure <infra-provider-of-choice>
```

## Move the cluster

Simply call:

```
clusterctl move --to-kubeconfig=$PWD/targetconfig
```

<!-- LINKS -->
[Cluster provisioning with CAPI and {{product}} tutorial]: ../tutorial/getting-started.md
