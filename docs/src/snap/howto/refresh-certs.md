# Refreshing Kubernetes Certificates

Keeping your {{product}} cluster secure and functional requires
regularly refreshing its certificates. Certificates in Kubernetes ensure
secure communication between the various components of the cluster. If
these certificates expire, it can lead to communication failures, disrupted
services, and potential security risks. This how-to will walk you through
the steps to refresh the certificates for both control plane and worker
nodes in your {{product}} cluster.

## What you will need
- A running {{product}} cluster

```{note} To refresh the certificates in your cluster, make sure it was
initially set up with self-signed certificates or with your own CA certificates
and keys during the bootstrap process.
```

### Refreshing Control Plane Node Certificates
To refresh the certificates on control plane nodes, follow these steps:

1. Access each control plane node in your cluster.
2. Run the `refresh-certs` command:

```bash
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
node and restart the necessary services. The command output will show the new
expiration date:
```bash
Certificates have been successfully refreshed, and will expire at 2025-08-27 21:00:00 +0000 UTC.
```

### Refreshing Worker Node Certificates

To refresh the certificates on worker nodes, follow these steps:

1. Access each worker node in your cluster.

2. Run the `refresh-certs` command:
```bash
sudo k8s refresh-certs --expires-in 10y --timeout 10m
```
This command refreshes the certificates for the worker node. The `--expires-in`
flag specifies the certificate's validity period. As mentioned in the control
plane section, this duration can be set using any unit accepted by the
[ParseDuration][] function in Go, in addition to years, months, or days.

3. During this process, multiple Certificate Signing Requests (CSRs) are created.
The command output will guide you on how to approve the CSRs on any control plane
node in the cluster.
```bash
root@w-1:~# k8s refresh-certs --expires-in 10y --timeout 10s
The following CertificateSigningRequests should be approved. Run the following commands on any of the control plane nodes of the cluster:
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-serving
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-client
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kube-proxy-client
Waiting for certificates to be created...
```

4. On any control plane node, run the commands provided in the output:
```bash
root@t-1:~# k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-serving
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-client
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kube-proxy-client
certificatesigningrequest.certificates.k8s.io/k8sd-3974895791729870959-worker-kubelet-serving approved
certificatesigningrequest.certificates.k8s.io/k8sd-3974895791729870959-worker-kubelet-client approved
certificatesigningrequest.certificates.k8s.io/k8sd-3974895791729870959-worker-kube-proxy-client approved
```
This command approves the CSRs, allowing the certificates to be created.

5. After approving all the requested CSRs, the worker node will automatically
refresh its the certificates and restart the necessary services. You should see
a confirmation message similar to the following:
```bash
Certificates have been successfully refreshed, and will expire at 2034-08-27 21:00:00 +0000 UTC.
```

## Summing Up
By following this guide, you ensure that your {{product}} cluster
remains secure and operational, avoiding potential disruptions caused by expired
certificates. Regularly refreshing these certificates is a crucial part of
maintaining a healthy and secure Kubernetes environment.

<!-- Links -->

[ParseDuration]: https://pkg.go.dev/time#ParseDuration
