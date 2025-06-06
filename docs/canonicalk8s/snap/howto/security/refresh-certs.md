# How to refresh Kubernetes certificates

To keep your {{product}} cluster secure and functional, it is essential
to regularly refresh its certificates. Certificates in Kubernetes ensure
secure communication between the various components of the cluster. Expired
certificates lead to communication failures, disrupted services, and potential
security risks. This how-to will walk you through
the steps to refresh the certificates for both control plane and worker
nodes in your {{product}} cluster.

```{warning}
Only Kubernetes component certificates refreshes are supported with the
 `k8s refresh-certs` command. Microcluster and k8s-dqlite certificates' expiration
 is set to 20 years, so renewal is not typically necessary. They are not automatically
 renewed by the command and currently cannot be refreshed manually.
 Additionally, like upstream Kubernetes, rotating the Certificate Authority (CA) is not supported.
```

## Prerequisites

- A running {{product}} cluster that has self-signed certificates enabled.
 This would have been set during the bootstrap process.

```{note} To refresh the certificates in your cluster, make sure it was
initially set up with self-signed certificates during the bootstrap process.
```

### Refresh Control Plane node certificates

1. To refresh the certificates on control plane nodes, perform the following
steps on each control plane node in your cluster:

```
sudo k8s refresh-certs --expires-in 1y --extra-sans mynode.local
```

**`--extra-sans`**

This command refreshes the certificates for the control plane node, adding an
extra [Subject Alternative Name][] (SAN) to the certificate. Check the
current SANs on your node by running the following command:

```
openssl x509 -in /etc/kubernetes/pki/apiserver.crt -noout -text | grep -A 1 "Subject Alternative Name"
```

If your node setup includes additional SANs, be sure to provide the
specific SANs for each node as needed using the `--extra-sans` flag. While this
is not required, omitting them could impact your node's ability to communicate
with other components in the cluster.

**`--expires-in`**

The `--expires-in` flag sets the certificate's validity duration, which can
be specified in years, months, days, or any other unit accepted by the
[ParseDuration][] function in Go.

**`--certificates`**

By default, all internal certificates are refreshed on the control plane node
when you run `refresh-certs`.
You can however selectively refresh certificates using the `--certificates` flag
and specify the certificates to be refreshed. Run `k8s refresh-certs -h` to
see available options.

2. The cluster will automatically update the certificates in the control plane
node and restart the necessary services. The new expiration date will be
displayed in the command output:

```
Certificates have been successfully refreshed, and will expire at 2025-08-27 21:00:00 +0000 UTC.
```

### Refresh Worker node certificates

1. To refresh the certificates on worker nodes, perform the following steps on
each worker node in your cluster:

```
sudo k8s refresh-certs --expires-in 10y --timeout 10m
```

**`--expires-in`**

This command refreshes the certificates for the worker node. The `--expires-in`
flag specifies the certificate's validity period, which can be set using any
units accepted by the [ParseDuration][] function in Go, such as years, months,
or days.

**`--certificates`**

Worker nodes support selective certificate renewal too. Use the
`--certificates` flag to choose which certificates to refresh. For details, see
`k8s refresh-certs -h`.

2. During the certificate refresh, multiple Certificate Signing Requests (CSRs)
are created. Follow the instructions in the command output to approve the CSRs
on any control plane node in the cluster.

```
sudo k8s refresh-certs --expires-in 10y --timeout 10s
The following CertificateSigningRequests should be approved. Run the following commands on any of the control plane nodes of the cluster:
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-serving
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-client
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kube-proxy-client
Waiting for certificates to be created...
```

3. Approve the CSRs by running the following commands on any control plane
node, which allows the certificates to be created:

```
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-serving
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kubelet-client
k8s kubectl certificate approve k8sd-3974895791729870959-worker-kube-proxy-client
certificatesigningrequest.certificates.k8s.io/k8sd-3974895791729870959-worker-kubelet-serving approved
certificatesigningrequest.certificates.k8s.io/k8sd-3974895791729870959-worker-kubelet-client approved
certificatesigningrequest.certificates.k8s.io/k8sd-3974895791729870959-worker-kube-proxy-client approved
```

4. After approving all the requested CSRs, the worker node will automatically
refresh its certificates and restart the necessary services:

```
Certificates have been successfully refreshed, and will expire at 2034-08-27 21:00:00 +0000 UTC.
```

<!-- Links -->

[ParseDuration]: https://pkg.go.dev/time#ParseDuration
[Subject Alternative Name]: https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.6
