# How to enable Dual-Stack networking

Dual-stack networking allows Kubernetes to support both IPv4 and IPv6 addresses
simultaneously. This means that your pods, services can be assigned
both IPv4 and IPv6 addresses, allowing them to communicate over either protocol.
This document will guide you through enabling dual-stack, including necessary
configurations, known limitations, and common issues.

## Prerequisites

Before enabling dual-stack, ensure that your environment supports IPv6, and
that your network configuration (including any underlying infrastructure) is
compatible with dual-stack operation.

## Enabling Dual-Stack

Dual-stack can be enabled by specifying both IPv4 and IPv6 CIDRs during the
cluster bootstrap process. The key configuration parameters are:

- **Pod CIDR**: Defines the IP range for pods.
- **Service CIDR**: Defines the IP range for services.

1. **Bootstrap Kubernetes with Dual-Stack CIDRs**

   Bootstrap the cluster in interactive mode and set both IPv4 and
   IPv6 CIDRs:

   ```
   sudo k8s bootstrap --timeout 10m --interactive
   ```

   When asked `Which features would you like to enable?`, press Enter to enable
   the default components.

   When prompted, set the Pod CIDR and Service CIDR:

   ```
   Please set the Pod CIDR: [10.1.0.0/16]: 10.1.0.0/16,fd01::/108
   Please set the Service CIDR: [10.152.183.0/24]: 10.152.183.0/24,fd98::/108
   ```

   Alternatively, the CIDRs can be configured in a bootstrap configuration file:

   ```yaml
   pod-cidr: 10.1.0.0/16,fd01::/108
   service-cidr: 10.152.183.0/24,fd98::/108
   ```

   This configuration file, here called `bootstrap-config.yaml`, can then be
   applied during the cluster bootstrapping process:

   ```
   sudo k8s bootstrap --file bootstrap-config.yaml
   ```

1. **Verify Pod and Service Creation**

   Once the cluster is up and running, verify that all pods are running:

   ```sh
   sudo k8s kubectl get pods -A
   ```

   To test that the cluster is configured with dual-stack, apply the following
   manifest that creates a service with `ipFamilyPolicy: RequireDualStack`.
   It also creates an nginx deployment sample workload.

   ```
   sudo k8s kubectl apply -f https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/assets/how-to-dualstack-manifest.yaml
   ```

1. **Check IPv6 Connectivity**

   Retrieve the service details and ensure that an IPv6 address is assigned:

   ```sh
   sudo k8s kubectl describe service nginx-dualstack
   ```

   The output should contain a line like:

   ```
   IPs: 10.152.183.170,fd98::6f88
   ```

   Test the connectivity to the deployed application using the IPv6 address
   from the retrieved output:

   ```sh
   curl http://[fd98::6f88]/
   ```

   You should see a response from the Nginx server, confirming that IPv6 is
   working.


## CIDR Size Limitations

When setting up dual-stack networking, it is important to consider the
limitations regarding CIDR size:

- **/108 is the maximum size for the Service CIDR**
Using a smaller value than `/108` for service CIDRs
may cause issues like failure to initialise the IPv6 allocator. This is due
to the CIDR size being too large for Kubernetes to handle efficiently.

See upstream reference: [kube-apiserver validation][kube-apiserver-test]

<!-- LINKS -->

[kube-apiserver-test]: https://github.com/kubernetes/kubernetes/blob/master/cmd/kube-apiserver/app/options/validation_test.go#L435
