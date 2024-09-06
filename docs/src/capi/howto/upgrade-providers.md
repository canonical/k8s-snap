# Uprading the providers of a management cluster

In this guide we will go through the process of upgrading providers of a management cluster.

## Prerequisites

We assume we already have a management cluster and the infrastructure provider configured as described in the [Cluster provisioning with CAPI and {{product}} tutorial]. The selected infrastructure provider is AWS. We have not yet called clusterctl init to initialise the cluster.

## Initialise the cluster

To demonstrate the steps of upgrading the management cluster, we will begin by initialising a desired version of the {{product}} CAPI providers.

To set the version of the providers to be installed we use the following notation:

```
clusterctl init --bootstrap ck8s:v0.1.2 --control-plane ck8s:v0.1.2 --infrastructure <infra-provider-of-choice>
```

## Check for updates

With clusterctl we can check if there are any new versions on the running providers:

```
clusterctl upgrade plan
```

The output shows the existing version of each provider as well as the version that we can upgrade into:

```text
NAME                 NAMESPACE       TYPE                     CURRENT VERSION   NEXT VERSION
bootstrap-ck8s       cabpck-system   BootstrapProvider        v0.1.2            v0.2.0
control-plane-ck8s   cacpck-system   ControlPlaneProvider     v0.1.2            v0.2.0
cluster-api          capi-system     CoreProvider             v1.8.1            Already up to date
infrastructure-aws   capa-system     InfrastructureProvider   v2.6.1            Already up to date
```

## Trigger providers upgrade

To apply the upgrade plan recommended by clusterctl upgrade plan simply:

```
clusterctl upgrade apply --contract v1beta1
```

To upgrade each provider one by one, issue:

```
clusterctl upgrade apply --bootstrap cabpck-system/ck8s:v0.2.0
clusterctl upgrade apply --control-plane cacpck-system/ck8s:v0.2.0
```

<!-- LINKS -->
[Cluster provisioning with CAPI and {{product}} tutorial]: ../tutorial/getting-started.md
