# How to Setup an IPv6-Only Cluster

An IPv6-only Kubernetes cluster operates exclusively using IPv6 addresses,
without support for IPv4. This configuration is ideal for environments that
are transitioning away from IPv4 or want to take full advantage of IPv6's
expanded address space. In this document, we’ll guide you through setting up
an IPv6-only cluster, including key configurations and necessary checks
to ensure proper setup.

## Prerequisites

Before setting up an IPv6-only cluster, ensure that:

- Your environment supports IPv6.
- Network infrastructure, such as routers, firewalls, and DNS, are configured
to handle IPv6 traffic.
- Any underlying infrastructure (e.g., cloud providers, bare metal setups)
must be IPv6-compatible.

## Setting Up an IPv6-Only Cluster

The process of creating an IPv6-only cluster involves specifying only IPv6
CIDRs for Pods and Services during the bootstrap process. Unlike dual-stack,
only IPv6 CIDRs are used.

1. **Bootstrap Kubernetes with IPv6 CIDRs**

You can start by bootstrapping the Kubernetes cluster and providing only IPv6
CIDRs for pods and services:

```bash
sudo k8s bootstrap --timeout 10m --interactive
```

When prompted, set the Pod and Service CIDRs to IPv6 ranges. For example:

```
Please set the Pod CIDR: [fd01::/108]
Please set the Service CIDR: [fd98::/108]
```

Alternatively, these values can be configured in a bootstrap configuration file
(`bootstrap-config.yaml`):

```yaml
pod-cidr: fd01::/108
service-cidr: fd98::/108
```

Then, use the configuration file during the bootstrapping process:

```bash
sudo k8s bootstrap --file bootstrap-config.yaml
```

2. **Verify Pod and Service Creation**

Once the cluster is up, verify that all pods are running:

```sh
sudo k8s kubectl get pods -A
```

Test the IPv6-only cluster by deploying a pod and exposing it via a service:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-ipv6
spec:
  selector:
    matchLabels:
      run: nginx-ipv6
  replicas: 1
  template:
    metadata:
      labels:
        run: nginx-ipv6
    spec:
      containers:
      - name: nginx-ipv6
        image: rocks.canonical.com/cdk/diverdane/nginxipv6:1.0.0
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-ipv6
  labels:
    run: nginx-ipv6
spec:
  type: NodePort
  ipFamilies:
  - IPv6
  ports:
  - port: 80
    protocol: TCP
  selector:
    run: nginx-ipv6
```

3. **Check IPv6 Connectivity**

Retrieve the service details to confirm an IPv6 address is assigned:

```sh
sudo k8s kubectl get service nginx-ipv6 -n default
```

The output should show the service’s IPv6 address:

```
NAME         TYPE       CLUSTER-IP    EXTERNAL-IP   PORT(S)        AGE
nginx-ipv6   NodePort   fd98::7534    <none>        80:32248/TCP   2m
```

Use the assigned IPv6 address to test connectivity:

```bash
curl http://[fd98::7534]/
```

If the Nginx server responds, IPv6 connectivity is working properly.

## IPv6-Only Cluster Considerations

1. **Service and Pod CIDR Sizing**

Use `/108` as the maximum size for Service CIDRs. Larger ranges (e.g., `/64`)
may lead to allocation errors or Kubernetes failing to initialize the IPv6
address allocator.

2. **Ingress and DNS**

When setting up an IPv6-only cluster, ensure that your ingress controllers and
DNS configurations are properly configured for IPv6, as many setups default to
IPv4.

3. **External IPv6 Access**

Verify that your external networking supports IPv6, especially if
your applications need to communicate beyond the cluster.
Ensure firewalls and load balancers are IPv6-compatible.
