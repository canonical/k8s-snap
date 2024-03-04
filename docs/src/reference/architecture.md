# Architecture

A system architecture document is the starting point for many interested
participants in a project, whether you intend contributing or simply want to
understand how the software is structured. This documentation lays out the
current design of Canonical Kubernetes, following the [C4 model]. 

##  System context 

This overview of Canonical Kubernetes demonstrates the interactions of
Kubernetes with users and with other systems.

```{kroki} ../assets/overview.puml
```

Two actors interact with the Kubernetes snap:

- **K8s admin**: The administrator of the cluster interacts directly with the
  Kubernetes API server. Out of the box our K8s distribution offers admin
  access to the cluster. That initial user is able to configure the cluster to
  match their needs and of course create other users that may or may not have
  admin privileges. The K8s admin is also able to maintain workloads running
  in the cluster.

- **K8s user**: A user consuming the workloads hosted in the cluster. Users do
  not have access to the Kubernetes API server. They need to access the cluster
  through the options (nodeport, ingress, load-balancer) offered by the
  administrator who deployed the workload they are interested in.

There are non-human users of the K8s snap, for example the [`k8s-operator
charm`][K8s charm]. The K8s charm needs to drive the Kubernetes cluster and to
orchestrate the multi-node clustering operations.

A set of external systems need to be easily integrated with our K8s
distribution. We have identified the following:
 - **Load Balancer**: Although the K8s snap distribution comes with a
   load balancer we expect the end customer environment to have a load balancer
   and thus we need to integrate with it.
- **Storage**: Kubernetes typically expects storage to be external to the
  cluster. The K8s snap comes with a local storage option but we still need to
  offer proper integration with any storage solution.
- **Identity management**: Out of the box the K8s snap offers credentials for
  an admin user. The admin user can complete the integration with any identity
  management system available or do user management manually.
- **External datastore**: By default, Kubernetes uses etcd to keep track of
  state. Our K8s snap comes with `dqlite` as its datastore. We should however
  be able to use any end client owned datastore installation. That should
  include an external `postgresql` or `etcd`.

## The k8s snap

Looking more closely at what is contained within the K8s snap itself:

```{kroki} ../assets/k8s-container.puml
```

The `k8s` snap distribution includes the following:

- **Kubectl**: through which users and other systems interact with Kubernetes
  and drive the cluster operations.
- **K8s services**: These are all the Kubernetes services as well as core workloads
  built from upstream and shipped in the snap.
- State is backed up by **dqlite** by default, which keeps that state of the
  Kubernetes cluster as well as the state we maintain for the needs of the
  cluster operations. The cluster state may optionally be stored in a
  different, external datastore.
- **Runtime**: `containerd` and `runc` are the shipped container runtimes.
- **K8sd**: which implements the operations logic and exposes that
  functionality via CLIs and APIs.

## K8sd

K8sd is the component that implements and exposes the operations functionality
needed for managing the Kubernetes cluster.

```{kroki} ../assets/k8sd-component.puml
```

At the core of the `k8sd` functionality we have the cluster
manager that is responsible for configuring the services, workload and features we
deem important for a Kubernetes cluster. Namely:

- Kubernetes systemd services
- DNS
- CNI
- ingress
- gateway API
- load-balancer
- local-storage
- metrics-server

The cluster manager is also responsible for implementing the formation of the
cluster. This includes operations such as joining/removing nodes into the
cluster and reporting status.

This functionality is exposed via the following interfaces:

- The **CLI**: The CLI is available to only the root user on the K8s snap and
  all CLI commands are mapped to respective REST calls.

- The **API**: The API over HTTP serves the CLI and is also used to
  programmatically drive the Kubernetes cluster. 

<!-- LINKS -->
[C4 model]: https://c4model.com/
[K8s charm]: https://charmhub.io/k8s
