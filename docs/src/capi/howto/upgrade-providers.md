# Upgrading the providers of a management cluster

This guide will walk you through the process of upgrading the 
providers of a management cluster.

## Prerequisites

- A {{product}} CAPI management cluster with providers installed and 
configured.

## Check for updates

With `clusterctl`, check if there are any new versions of the running 
providers:

```
clusterctl upgrade plan
```

The output shows the existing version of each provider as well 
as the next available version:

```text
NAME                    NAMESPACE       TYPE                     CURRENT VERSION   NEXT VERSION
canonical-kubernetes    cabpck-system   BootstrapProvider        v0.1.2            v0.2.0
canonical-kubernetes    cacpck-system   ControlPlaneProvider     v0.1.2            v0.2.0
cluster-api             capi-system     CoreProvider             v1.8.1            Already up to date
infrastructure-aws      capa-system     InfrastructureProvider   v2.6.1            Already up to date
```

## Trigger providers upgrade

To apply the upgrade plan recommended by `clusterctl upgrade plan`, simply:

```
clusterctl upgrade apply --contract v1beta1
```

To upgrade each provider one by one, issue:

```
clusterctl upgrade apply --bootstrap cabpck-system/canonical-kubernetes:v0.2.0
clusterctl upgrade apply --control-plane cacpck-system/canonical-kubernetes:v0.2.0
```

<!-- LINKS -->
[Cluster provisioning with CAPI and {{product}} tutorial]: ../tutorial/getting-started.md
