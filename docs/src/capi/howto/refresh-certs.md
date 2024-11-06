# Refreshing Workload Cluster Certificates

This how-to will walk you through the steps to refresh the certificates for
both control plane and worker nodes in your {{product}} Cluster API cluster.

## Prerequisites

- A Kubernetes management cluster with Cluster API and Canonical K8s providers
  installed and configured.
- A target workload cluster managed by Cluster API.
- `kubectl` installed and configured to access your management cluster.

Please refer to the [getting-started guide][getting-started] for further
details on the required setup.
This guide refers to the workload cluster as `c1`.

```{note} To refresh the certificates in your cluster, make sure it was
initially set up with self-signed certificates. You can verify this by
checking the `CK8sConfigTemplate` resource for the cluster to see if a
`BootstrapConfig` value was provided with the necessary certificates.
```

### Refresh Control Plane Node Certificates

To refresh the certificates on control plane nodes, follow these steps for each
control plane node in your workload cluster:

1. First, check the names of the control plane machines in your cluster:

```
clusterctl describe cluster c1
```

2. For each control plane machine, annotate the machine resource with the
`v1beta2.k8sd.io/refresh-certificates` annotation. The value of the annotation
should specify the duration for which the certificates will be valid. For
example, to refresh the certificates for a control plane machine named
`c1-control-plane-nwlss` to expire in 10 years, run the following command:

```
kubectl annotate machine c1-control-plane-nwlss v1beta2.k8sd.io/refresh-certificates=10y
```

```{note} The value of the annotation can be specified in years (y), months
(mo), (d) days, or any unit accepted by the [ParseDuration][] function in
Go.
```

The Cluster API provider will automatically refresh the certificates on the
control plane node and restart the necessary services. To track the progress of
the certificate refresh, check the events for the machine resource:

```
kubectl get events --field-selector involvedObject.name=c1-control-plane-nwlss
```

The machine will be ready once the event `CertificatesRefreshDone` is
displayed.

3. After the certificate refresh is complete, the new expiration date will be
displayed in the `machine.cluster.x-k8s.io/certificates-expiry` annotation of
the machine resource:

```
"machine.cluster.x-k8s.io/certificates-expiry": "2034-10-25T14:25:23-05:00"
```

### Refresh Worker Node Certificates

To refresh the certificates on worker nodes, follow these steps for each worker
node in your workload cluster:

1. Check the names of the worker machines in your cluster:

```
clusterctl describe cluster c1
```

2. Add the `v1beta2.k8sd.io/refresh-certificates` annotation to each worker
machine, specifying the desired certificate validity duration. For example, to
set the certificates for `c1-worker-md-0-4lxb7-msq44` to expire in 10 years:

```
kubectl annotate machine c1-worker-md-0-4lxb7-msq44 v1beta2.k8sd.io/refresh-certificates=10y
```

The ClusterAPI provider will handle the certificate refresh and restart
necessary services. Track the progress by checking the machine's events:

```
kubectl get events --field-selector involvedObject.name=c1-worker-md-0-4lxb7-msq44
```

The machine will be ready once the event `CertificatesRefreshDone` is
displayed.

3. After the certificate refresh is complete, the new expiration date will be
displayed in the `machine.cluster.x-k8s.io/certificates-expiry` annotation of
the machine resource:

```
"machine.cluster.x-k8s.io/certificates-expiry": "2034-10-25T14:33:04-05:00"
```

<!-- Links -->
[getting-started]: ../tutorial/getting-started.md
[ParseDuration]: https://pkg.go.dev/time#ParseDuration
