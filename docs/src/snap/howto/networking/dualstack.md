# How to enable Dual-Stack networking

Dual-stack networking allows Kubernetes to support both IPv4 and IPv6 addresses
simultaneously. This means that your pods, services can be assigned
both IPv4 and IPv6 addresses, allowing them to communicate over either protocol.
This document will guide you through enabling dual-stack, including necessary
configurations, known limitations, and common issues.

### Prerequisites

Before enabling dual-stack, ensure that your environment supports IPv6, and
that your network configuration (including any underlying infrastructure) is
compatible with dual-stack operation.

### Enabling Dual-Stack

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
   manifest that creates a service with `ipFamilyPolicy: RequireDualStack`:
   ```
    apiVersion: apps/v1
    kind: Deployment
    metadata:
    name: nginxdualstack
    spec:
    selector:
        matchLabels:
        run: nginxdualstack
    replicas: 1
    template:
        metadata:
        labels:
            run: nginxdualstack
        spec:
        containers:
        - name: nginxdualstack
            image: rocks.canonical.com/cdk/diverdane/nginxdualstack:1.0.0
            ports:
            - containerPort: 80
    ---
    apiVersion: v1
    kind: Service
    metadata:
    name: nginx-dualstack
    labels:
        run: nginxdualstack
    spec:
    type: NodePort
    ipFamilies:
    - IPv4
    - IPv6
    ipFamilyPolicy: RequireDualStack
    ports:
    - port: 80
        protocol: TCP
    selector:
        run: nginxdualstack

   ```

1. **Check IPv6 Connectivity**

   Retrieve the service details and ensure that an IPv6 address is assigned:

   ```sh
   sudo k8s kubectl get service -A
   ```

   The output should be similar to:
   ```
   root@k8s-dualstack:/k8s-snap# sudo k8s kubectl get svc -A
   NAMESPACE     NAME                                TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)         AGE
   default       kubernetes                          ClusterIP   10.152.183.1     <none>        443/TCP         4m12s
   default       nginx6                              NodePort    fd98::7534       <none>        80:32748/TCP    8s
   kube-system   ck-storage-rawfile-csi-controller   ClusterIP   None             <none>        <none>          4m11s
   kube-system   ck-storage-rawfile-csi-node         ClusterIP   10.152.183.172   <none>        9100/TCP        4m11s
   kube-system   coredns                             ClusterIP   10.152.183.69    <none>        53/UDP,53/TCP   4m12s
   kube-system   hubble-peer                         ClusterIP   10.152.183.217   <none>        443/TCP         4m11s
   kube-system   metrics-server                      ClusterIP   10.152.183.108   <none>        443/TCP         4m11s
   ```

   Test the connectivity to the deployed application using the IPv6 address
   from the retrieved output:

   ```sh
   curl http://[fd98::7534]/
   ```

   You should see a response from the Nginx server, confirming that IPv6 is
   working.


### CIDR Size Limitations

When setting up dual-stack networking, it is important to consider the
limitations regarding CIDR size:

- **/64 is too large for the Service CIDR**: Using a `/64` CIDR for services
may cause issues like failure to initialize the IPv6 allocator. This is due
to the CIDR size being too large for Kubernetes to handle efficiently.
