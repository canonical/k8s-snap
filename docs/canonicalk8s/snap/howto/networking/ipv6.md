# How to set up an IPv6-only cluster

An IPv6-only Kubernetes cluster operates exclusively using IPv6 addresses,
without support for IPv4. This configuration is ideal for environments that
are transitioning away from IPv4 or want to take full advantage of IPv6's
expanded address space. This document explains how to set up
an IPv6-only cluster, including key configurations and necessary checks
to ensure proper setup.

## Prerequisites

Before setting up an IPv6-only cluster, ensure that:

- Your environment supports IPv6.
- Network infrastructure, such as routers, firewalls, and DNS, are configured
to handle IPv6 traffic.
- Any underlying infrastructure (e.g. cloud providers, bare metal setups)
must be IPv6-compatible.

## Set up an IPv6-only cluster

The process of creating an IPv6-only cluster involves specifying only IPv6
CIDRs for pods and services during the bootstrap process. Unlike dual-stack,
only IPv6 CIDRs are used.

1. **Bootstrap Kubernetes with IPv6 CIDRs**

Start by bootstrapping the Kubernetes cluster and providing only IPv6
CIDRs for pods and services:

```{important}
To run an IPv6-only cluster on a host with both IPv4 and IPv6 interfaces,
you must include `--address ::/0` in the bootstrap and join commands.
Without this option, the snap will attempt to bind to the IPv4 interface,
which will fail if only IPv6 CIDRs are configured.
```

```bash
sudo k8s bootstrap --timeout 10m --interactive
```

When asked `Which features would you like to enable?`, press Enter to enable
the default components.

When prompted, set the pod and service CIDRs to IPv6 ranges. For example:

```
Please set the Pod CIDR: fd01::/108
Please set the Service CIDR: fd98::/108
```

Alternatively, these values can be configured in a bootstrap configuration file
named `bootstrap-config.yaml` in this example:

```yaml
pod-cidr: fd01::/108
service-cidr: fd98::/108
```

Specify the configuration file during the bootstrapping process:

```bash
sudo k8s bootstrap --file bootstrap-config.yaml
```

2. **Verify Pod and Service creation**

Once the cluster is up, verify that all pods are running:

```sh
sudo k8s kubectl get pods -A
```

Deploy a pod with an nginx web-server and expose it via a service to verify
connectivity of the IPv6-only cluster:

```sh
sudo k8s kubectl apply -f https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/assets/how-to-ipv6-only-manifest.yaml
```

3. **Verify IPv6 connectivity**

Retrieve the service details to confirm an IPv6 address is assigned:

```sh
sudo k8s kubectl get service nginx-ipv6 -n default
```

Obtain the service’s IPv6 address from the output:

```
NAME         TYPE       CLUSTER-IP    EXTERNAL-IP   PORT(S)        AGE
nginx-ipv6   NodePort   fd98::7534    <none>        80:32248/TCP   2m
```

Use the assigned IPv6 address to test connectivity:

```bash
curl http://[fd98::7534]/
```

A welcome message from the nginx web-server is displayed when IPv6
connectivity is set up correctly.

## IPv6-only cluster considerations

**Service and Pod CIDR sizing**

Use `/108` as the maximum size for Service CIDRs. Larger ranges (e.g., `/64`)
may lead to allocation errors or Kubernetes failing to initialise the IPv6
address allocator.
