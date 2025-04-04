# Architecture diagrams

A system architecture document is the starting point for many interested
participants in a project, whether you intend contributing or simply want to
understand how the software is structured. This documentation lays out the
current design of {{product}}, following the [C4 model].

## System context

This overview of {{product}} demonstrates the interactions of
Kubernetes with users and with other systems.

![cluster5][]

Actors that interact with the K8s snap:

- **K8s admin**:  interacts directly with the Kubernetes API server. {{product}}
provides out of the box admin access to configure the cluster.
- **K8s user** : consumes the workloads hosted in the cluster.

Non-human users of the K8s snap:

- [`k8s-operator charm`][K8s charm]: drives the Kubernetes cluster and
orchestrates multi-node clustering operations.

Although {{product}} provides its own implementation of the following services,
external systems that can be easily integrated:

- **Load Balancer**
- **Storage**
- **Identity management**
- **External datastore**

## The k8s snap

Contained within the K8s snap itself:

![cluster1][]

The `k8s` snap distribution includes the following:

- **Kubectl**: through which users and other systems interact with Kubernetes
and drive the cluster operations.
- **K8s control plane**: the Kubernetes services, as well as core
workloads built from upstream and shipped in the snap.
- **Kubernetes datastore**: uses dqlite to store data on the state of the
cluster. It can be replaced by an external datastore.
- **Cluster datastore**: uses dqlite as a replicated database to store cluster
configuration.
- **Runtime**: `containerd` is the shipped container runtime.
- **K8sd**: implements the operations logic and exposes that
functionality via CLIs and APIs.

## K8sd

K8sd is the component that implements and exposes the operations functionality
needed for managing the Kubernetes cluster.

![cluster2][]

Functionality provided by `k8sd` cluster manager is:

- Kubernetes systemd services
- DNS
- CNI
- ingress
- gateway API
- load-balancer
- local-storage
- metrics-server
- implementing the formation of the cluster
- reporting cluster status

This functionality is exposed via the following interfaces:

- The **CLI**: The CLI is available to only the root user on the K8s snap and
  all CLI commands are mapped to respective REST calls.

- The **API**: The API over HTTP serves the CLI and is also used to
  programmatically drive the Kubernetes cluster.

## Canonical K8s charms

Canonical `k8s` charms are the [`k8s` charm][K8s
charm] and the [`k8s-worker` charm][K8s-worker charm].

![cluster4][]

The {{product}} charms include the following:

- **`k8s`** : installs and manages the `k8s` snap on control plane node. Also
supports integration with other compatible charms.
- **`k8s-worker`**: installs and manages the `k8s` snap on worker node.
- **Administrator**: manages the cluster via the Juju client.
- **K8sd API Manager**: Makes API calls to the `k8s` snap
- **Relation Databags for `k8s` and `k8s-worker`**: Juju databags for sharing
information between the `k8s` and `k8s-worker` charms

<!-- IMAGES -->

[cluster1]: https://assets.ubuntu.com/v1/58712341-snap.svg
[cluster2]: https://assets.ubuntu.com/v1/d74833fe-k8sd.svg
[cluster4]: https://assets.ubuntu.com/v1/53a083a9-charms.svg
[cluster5]: https://assets.ubuntu.com/v1/bcfe150f-overview.svg

<!-- LINKS -->
[C4 model]:           https://c4model.com/
[K8s charm]:          https://charmhub.io/k8s
[K8s-Worker charm]:   https://charmhub.io/k8s-worker
[Juju docs]:          https://juju.is/docs/juju
[COS docs]:           https://ubuntu.com/observability
