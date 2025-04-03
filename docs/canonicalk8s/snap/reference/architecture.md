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

- K8s admin - interacts directly with the Kubernetes API server. {{product}}
provides out of the box admin access to the cluster to configure the cluster to
their needs.
- K8s user

Non-human users of the K8s snap:

- [`k8s-operator charm`][K8s charm].

Although {{product}} provides its own implementation of the following services,
external systems that can be easily integrated:

- Load Balancer
- Storage
- Identity management
- External datastore

## The k8s snap

What is contained within the K8s snap itself:

![cluster1][]

The `k8s` snap distribution includes the following:

- **Kubectl**: through which users and other systems interact with Kubernetes
  and drive the cluster operations.
- **K8s services**: These are all the Kubernetes services as well as core
  workloads built from upstream and shipped in the snap.
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

The cluster manager is also responsible for implementing the formation of the
cluster. This includes operations such as joining/removing nodes into the
cluster and reporting status.

This functionality is exposed via the following interfaces:

- The **CLI**: The CLI is available to only the root user on the K8s snap and
  all CLI commands are mapped to respective REST calls.

- The **API**: The API over HTTP serves the CLI and is also used to
  programmatically drive the Kubernetes cluster.

## Canonical K8s charms

Canonical `k8s` Charms encompass two primary components: the [`k8s` charm][K8s
charm] and the [`k8s-worker` charm][K8s-worker charm].

![cluster4][]

<!-- Charms are instantiated on a machine as a Juju unit, and a collection of units
constitutes an application. -->

Roles:

**`k8s`**

- Installs and manages the `k8s` snap
- Manages control plane node

**`k8s-worker`**

- Installs and manages the `k8s` snap
- Manages worker node

**Administrator**

- Manages the cluster via the Juju client

Both `k8s` and `k8s-worker` units are responsible
for installing and managing its machine's `k8s` snap, however the charm type
determines the node's role in the Kubernetes cluster. The `k8s` charm manages
`control-plane` nodes, whereas the `k8s-worker` charm manages Kubernetes
`worker` nodes. The administrator manages the cluster via the `juju` client,
directing the `juju` controller to reach the model's eventually consistent
state. For more detail on Juju's concepts, see the [Juju docs][].

The administrator may choose any supported cloud-types (OpenStack, MAAS, AWS,
GCP, Azure...) on which to manage the machines making up the Kubernetes
cluster. Juju selects a single leader unit per application to act as a
centralised figure with the model. The `k8s` leader oversees Kubernetes
bootstrapping and enlistment of new nodes. Follower `k8s` units will join the
cluster using secrets shared through relation data from the leader. The entire
lifecycle of the deployment is orchestrated by the `k8s` charm, with tokens and
cluster-related information being exchanged through Juju relation data.

Furthermore, the `k8s-worker` unit functions exclusively as a worker within the
cluster, establishing a relation with the `k8s` leader unit and requesting
tokens and cluster-related information through relation data. The `k8s` leader
is responsible for issuing these tokens and revoking them if a unit
administratively departs the cluster.

The `k8s` charm also supports the integration of other compatible charms,
enabling integrations such as connectivity with an external `etcd` datastore
and the sharing of observability data with the [`Canonical Observability Stack
(COS)`][COS docs]. This modular and integrated approach facilitates a robust
and flexible {{product}} deployment managed through Juju.

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
