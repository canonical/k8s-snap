# Refreshing Kubernetes Certificates

To keep your {{product}} cluster secure and functional, it is essential
to regularly refresh its certificates. Certificates in Kubernetes ensure
secure communication between the various components of the cluster. Expired
certificates lead to communication failures, disrupted services, and potential 
security risks. This how-to will walk you through
the steps to refresh the certificates for both control plane and worker
nodes in your {{product}} cluster.

## Prerequisites

- A running {{product}} cluster

```{note} To refresh the certificates in your cluster, make sure it was
initially set up with self-signed certificates during the bootstrap process.
```

### Refreshing Control Plane Node Certificates

To refresh the certificates on control plane nodes, perform the following steps
on each control plane node in your cluster:

1. Access each control plane node in your cluster.
2. Run the `refresh-certs` command:

```
sudo k8s refresh-certs --expires-in 1y --extra-sans mynode.local
```

This command refreshes the certificates for the control plane node, adding an
extra Subject Alternative Name (SAN) to the certificate. The `--expires-in`
flag sets the certificate's validity duration, which can be specified in years,
months, days, or any unit accepted by the [ParseDuration][] function in Go.

```{note} Ensure that you provide the same SANs that were used when the cluster
was bootstrapped. If you don't, the control plane may fail to communicate with
other nodes in the cluster.
```

3. The cluster will automatically update the certificates in the control plane
node and restart the necessary services. The new expiration date will be
displayed in the command output:

```
Certificates have been successfully refreshed, and will expire at 2025-08-27 21:00:00 +0000 UTC.
```

### Refreshing Worker Node Certificates

To refresh the certificates on worker nodes, perform the following steps on
each worker node in your cluster:

1. Access each worker node in your cluster.

2. Run the `refresh-certs` command:

```
sudo k8s refresh-certs --expires-in 10y --timeout 10m
```

This command refreshes the certificates for the worker node. The `--expires-in`
flag specifies the certificate's validity period, which can be set using any units 
accepted by the [ParseDuration][] function in Go, such as years, months, or days.

3. During the certificate refresh, multiple Certificate Signing Requests (CSRs) are
created. Follow the instructions in the command output to approve the CSRs on any
control plane node in the cluster.

```
sudo k8s refresh-certs --expires-in 10y --timeout 10s
The following CertificateSigningRequests should be approved. Run the following commands on any of the control plane nodes of the cluster:
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-serving
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-client
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kube-proxy-client
Waiting for certificates to be created...
```

4. Approve the CSRs by running the commands from the previous output on any control plane node:

```
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-serving
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-client
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kube-proxy-client
certificatesigningrequest.certificates.k8s.io/k8sd-3974895791729870959-worker-kubelet-serving approved
certificatesigningrequest.certificates.k8s.io/k8sd-3974895791729870959-worker-kubelet-client approved
certificatesigningrequest.certificates.k8s.io/k8sd-3974895791729870959-worker-kube-proxy-client approved
```

This command approves the CSRs, allowing the certificates to be created.

5. After approving all the requested CSRs, the worker node will automatically
refresh its certificates and restart the necessary services:

```
Certificates have been successfully refreshed, and will expire at 2034-08-27 21:00:00 +0000 UTC.
```

<!-- Links -->

[ParseDuration]: https://pkg.go.dev/time#ParseDuration
