# Cluster API - {{product}}

ClusterAPI (CAPI) is an open-source Kubernetes project that provides a
declarative API for cluster creation, configuration, and management. It is
designed to automate the creation and management of Kubernetes clusters in
various environments, including on-premises data centres, public clouds, and
edge devices.

CAPI abstracts away the details of infrastructure provisioning, networking, and
other low-level tasks, allowing users to define their desired cluster
configuration using simple YAML manifests. This makes it easier to create and
manage clusters in a repeatable and consistent manner, regardless of the
underlying infrastructure. In this way a wide range of infrastructure providers
has been made available, including but not limited to Amazon Web Services
(AWS), Microsoft Azure, Google Cloud Platform (GCP), and OpenStack.

CAPI also abstracts the provisioning and management of Kubernetes clusters
allowing for a variety of Kubernetes distributions to be delivered in all of
the supported infrastructure providers. {{product}} is one such Kubernetes
distribution that seamlessly integrates with Cluster API.

With {{product}} CAPI you can:

- provision a cluster with:
    - Kubernetes version 1.31 onwards
    - risk level of the track you want to follow (stable, candidate, beta, edge)
    - deploy behind proxies
- upgrade clusters with no downtime:
    - rolling upgrades for HA clusters and worker nodes
    - in-place upgrades for non-HA control planes and worker nodes

Please refer to the “Tutorial” section for concrete examples on CAPI deployments:


## CAPI architecture

Being a cloud-native framework, CAPI implements all its components as
controllers that run within a Kubernetes cluster. There is a separate
controller, called a ‘provider’, for each supported infrastructure substrate.
The infrastructure providers are responsible for provisioning physical or
virtual nodes and setting up networking elements such as load balancers and
virtual networks. In a similar way, each Kubernetes distribution that
integrates with ClusterAPI is managed by two providers: the control plane
provider and the bootstrap provider. The bootstrap provider is responsible for
delivering and managing Kubernetes on the nodes, while the control plane
provider handles the control plane’s specific lifecycle.

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
clusters.

Typically, the management cluster runs in a separate environment from the
clusters it manages, such as a public cloud or an on-premises data centre. It
serves as a centralised location for managing the configuration, policies, and
security of multiple managed clusters. By leveraging the management cluster,
users can easily create and manage a fleet of Kubernetes clusters in a
consistent and repeatable manner.

The {{product}} team maintains the two providers required for integrating with CAPI:

- The Cluster API Bootstrap Provider {{product}} (**CABPCK**) responsible for
  provisioning the nodes in the cluster and preparing them to be joined to the
  Kubernetes control plane. When you use the CABPCK you define a Kubernetes
  Cluster object that describes the desired state of the new cluster and
  includes the number and type of nodes in the cluster, as well as any
  additional configuration settings. The Bootstrap Provider then creates the
  necessary resources in the Kubernetes API server to bring the cluster up to
  the desired state. Under the hood, the Bootstrap Provider uses cloud-init to
  configure the nodes in the cluster. This includes setting up SSH keys,
  configuring the network, and installing necessary software packages.

- The Cluster API Control Plane Provider {{product}} (**CACPCK**) enables the
  creation and management of Kubernetes control planes using {{product}} as the
  underlying Kubernetes distribution. Its main tasks are to update the machine
  state and to generate the kubeconfig file used for accessing the cluster. The
  kubeconfig file is stored as a secret which the user can then retrieve using
  the `clusterctl` command.

```{figure} ../../assets/capi-ck8s.svg
   :width: 100%
   :alt: Deployment of components

   Deployment of components
```
