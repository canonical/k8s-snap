# How to use default Gateway

{{product}} enables you to configure advanced networking of your cluster using
[gateway API]. When enabled, the necessary CRDs and GatewayClass are generated
to enable the CNI controllers configure traffic and provision infrastructure to
the cluster.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstrapped {{product}} cluster (see the
[Getting Started][getting-started-guide] guide).

## Check Gateway status

Gateway should be enabled by default. Find out whether Gateway is enabled or
disabled with the following command:

```
sudo k8s status
```

Please ensure that Gateway is enabled on your cluster.

## Enable Gateway

To enable Gateway, run:

```
sudo k8s enable gateway
```

## Disable gateway

You can `disable` the built-in Gateway:

``` {warning}
If you have an active cluster, disabling Gateway may impact external access to services within your cluster. Ensure that you have alternative configurations in place before disabling Gateway.
```

```
sudo k8s disable gateway
```
<!-- LINKS -->
[gateway API]:https://gateway-api.sigs.k8s.io/
[getting-started-guide]: ../../tutorial/getting-started
[kubectl-guide]: ../../tutorial/kubectl
