# Cluster API and {{product}}

## Cluster API

ClusterAPI (CAPI) is an open-source Kubernetes project that provides a
declarative API for cluster creation, configuration, and management. It is
designed to automate the creation and management of Kubernetes clusters in
various environments, including on-premises data centers, public clouds, and
edge devices.

CAPI abstracts away the details of infrastructure provisioning, networking, and
other low-level tasks, allowing users to define their desired cluster
configuration using simple YAML manifests. This makes it easier to create and
manage clusters in a repeatable and consistent manner, regardless of the
underlying infrastructure. In this way a wide range of infrastructure providers
has been made available, including but not limited to MAAS, Amazon Web Services
(AWS), Microsoft Azure, Google Cloud Platform (GCP), and OpenStack.

CAPI also abstracts the provisioning and management of Kubernetes clusters
allowing for a variety of Kubernetes distributions to be delivered in all of
the supported infrastructure providers. {{product}} is one such Kubernetes
distribution that seamlessly integrates with Cluster API.

## {{product}} and CAPI

With {{product}} CAPI you can:

- Provision a cluster that meets your needs
  - Choose a {{product}} version
  - Choose the risk level of the track (Kubernetes version) you want to follow -
  stable, candidate, beta or edge
  - Deploy behind proxies
- Upgrade clusters with no downtime
    - Carry out rolling upgrades for High Availability (HA) clusters and worker
    nodes
    - Carry out in-place upgrades for non-HA control planes and worker nodes

Please refer to the [getting started tutorial] for a concrete example of a CAPI
deployment.

## CAPI architecture

Being a cloud-native framework, CAPI implements all its components as
controllers that run within a Kubernetes cluster.

### Infrastructure provider

There is a separate controller, called a ‘provider’, for each supported
infrastructure substrate. The infrastructure providers are responsible for
provisioning physical or virtual nodes and setting up networking elements such
as load balancers and
virtual networks.

### Kubernetes distributions

In a similar way, each Kubernetes distribution that
integrates with ClusterAPI is managed by either the control plane provider,
the bootstrap provider, or both.

- **Control plane provider**: handles the control plane’s specific lifecycle.
- **Bootstrap provider**: responsible for delivering and managing Kubernetes on
the nodes.

### Management cluster

The CAPI providers operate within a Kubernetes cluster known as the management
cluster. The administrator is responsible for selecting the desired combination
of infrastructure and Kubernetes distribution by instantiating the respective
infrastructure, bootstrap, and control plane providers on the management
cluster.

The management cluster functions as the control plane for the ClusterAPI
operator, which is responsible for provisioning and managing the infrastructure
resources necessary for creating and managing additional Kubernetes clusters.
It is important to note that the management cluster is not intended to support
any other workload, as the workloads are expected to run on the provisioned
clusters. As a result, the provisioned clusters are referred to as workload
clusters. While CAPI providers mostly live on the management cluster, it's
also possible to maintain the them in the workload cluster.
Read more about this in the [upstream docs around pivoting].

## {{product}} providers

The {{product}} team maintains the two providers required for integrating
with CAPI:

### Cluster API Bootstrap Provider {{product}}

The Cluster API Bootstrap Provider {{product}} (**CABPCK**) is responsible for
provisioning the nodes in the cluster and preparing them to be joined to the
Kubernetes control plane. When you use the CABPCK, you define a Kubernetes
`Cluster` object that describes the desired state of the new cluster. This
includes the number and type of nodes in the cluster, as well as any
additional configuration settings. The Bootstrap Provider then creates the
necessary resources in the Kubernetes API server to bring the cluster up to
the desired state. Under the hood, the Bootstrap Provider uses cloud-init to
configure the nodes in the cluster. This includes setting up SSH keys,
configuring the network, and installing necessary software packages.

### Cluster API Control Plane Provider {{product}}

The Cluster API Control Plane Provider {{product}} (**CACPCK**) enables the
creation and management of Kubernetes control planes using {{product}} as the
underlying Kubernetes distribution. Its main tasks are to update the machine
state and to generate the kubeconfig file used for accessing the cluster. The
kubeconfig file is stored as a secret which the user can then retrieve using
the `clusterctl` command. This component also handles the upgrade process for
the control plane nodes.

## {{product}} CAPI architecture diagram

```{figure} ../../assets/capi-ck8s.svg
   :width: 100%
   :alt: Deployment of components

   Deployment of components
```

<!-- LINKS -->
[getting started tutorial]: /capi/tutorial/getting-started.md
[upstream docs around pivoting]: https://cluster-api.sigs.k8s.io/clusterctl/commands/move#pivot
