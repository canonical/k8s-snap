# Architecture

A system architecture document is the starting point for many interested
participants in a project, whether you intend contributing or simply want to
understand how the software is structured. This documentation lays out the
current design of {{product}}, following the [C4 model].

## System context

This overview of {{product}} demonstrates the interactions of
Kubernetes with users and with other systems.

![cluster5][]

Two actors interact with the Kubernetes snap:

- **K8s admin**: The administrator of the cluster interacts directly with the
Kubernetes API server. Out of the box our K8s distribution offers admin
access to the cluster. That initial user is able to configure the cluster to
match their needs and of course create other users that may or may not have
admin privileges. The K8s admin is also able to maintain workloads running
in the cluster. If you deploy {{product}} from a snap, this is how the cluster
is manually orchestrated.

- **K8s user**: A user consuming the workloads hosted in the cluster. Users do
not have access to the Kubernetes API server. They need to access the cluster
through the options (NodePort, Ingress, load-balancer) offered by the
administrator who deployed the workload they are interested in.

There are non-human users of the K8s snap, for example the [`k8s-operator
charm`][K8s charm]. The K8s charm uses the snap to drive the Kubernetes cluster
and orchestrates multi-node clustering operations.

A set of external systems are easily integrated with our K8s
distribution:

- **Load Balancer**: Although the K8s snap distribution comes with a
load balancer, we allow the end customer environment to bring their own load
balancer solution.
- **Storage**: Kubernetes typically expects storage to be external to the
cluster. The K8s snap comes with a local storage option but also
allows integrations with alternative storage solutions.
- **Identity management**: Out of the box the K8s snap offers credentials for
an admin user. The admin user can complete the integration with any identity
management system available or do user management manually.
- **External datastore**: By default, Canonical Kubernetes uses etcd to
keep track of state. However, users can choose to switch to k8s-dqlite or
use an end client owned datastore installation such as the use of an
external `etcd`.

## The k8s snap

Looking more closely at what is contained within the K8s snap itself:

![cluster1][]

The `k8s` snap distribution includes the following:

- **Kubectl**: through which users and other systems interact with Kubernetes
and drive the cluster operations.
- **K8s core components**: These are all the Kubernetes services as well as core
workloads built from upstream and shipped in the snap.
- **Kubernetes datastore**: uses etcd to store data on the state of the
cluster. It can be replaced by [k8s-dqlite] or an external datastore.
- **Cluster datastore**: uses Dqlite managed by [Microcluster] as a replicated
database to store cluster configuration. It is used
by `k8sd` in order to carry out the orchestration of the additional Kubernetes
components included in {{product}} such as cluster membership management.
- **Container runtime**: `containerd` is the shipped container runtime.
- **K8sd**: implements the operations logic and exposes that
functionality via CLIs and APIs.

## K8sd

K8sd is the component that implements and exposes the operations functionality
needed for managing the Kubernetes cluster.

![cluster2][]

At the core of the `k8sd` functionality we have the cluster manager that is
responsible for configuring the services, workloads and features we deem
important for a Kubernetes cluster. Namely:

- Kubernetes systemd services
- DNS
- CNI
- Ingress
- Gateway API
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

Canonical `k8s` charms encompass two primary components: the [`k8s` charm][K8s
charm] and the [`k8s-worker` charm][K8s-worker charm].

![cluster4][]

Charms are instantiated on a machine as a Juju unit, and a collection of units
constitutes an application. Both `k8s` and `k8s-worker` units are responsible
for installing and managing its machine's `k8s` snap, however the charm type
determines the node's role in the Kubernetes cluster. The `k8s` charm manages
`control-plane` nodes, whereas the `k8s-worker` charm manages Kubernetes
`worker` nodes. The administrator manages the cluster via the `juju` client,
directing the `juju` controller to reach the model's eventually consistent
state. For more detail on Juju's concepts, see the [Juju docs][].

The administrator may choose any supported cloud-types (OpenStack, MAAS, AWS,
GCP, Azure...) on which to manage the machines making up the Kubernetes
cluster. Juju selects a single leader unit per application to act as a
centralized figure with the model. The `k8s` leader oversees Kubernetes
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

[cluster1]: https://assets.ubuntu.com/v1/60234b03-snap-14-08.svg
[cluster2]: https://assets.ubuntu.com/v1/b0ae732e-k8sd-13-08.svg
[cluster4]: https://assets.ubuntu.com/v1/53a083a9-charms.svg
[cluster5]: https://assets.ubuntu.com/v1/bcfe150f-overview.svg

<!-- LINKS -->
[C4 model]:           https://c4model.com/
[K8s charm]:          https://charmhub.io/k8s
[K8s-Worker charm]:   https://charmhub.io/k8s-worker
[Juju docs]:          https://juju.is/docs/juju
[COS docs]:           https://ubuntu.com/observability
[k8s-dqlite]:             https://github.com/canonical/k8s-dqlite
[Microcluster]:       https://github.com/canonical/microcluster
