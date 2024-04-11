# Architecture

A system architecture document is the starting point for many interested
participants in a project, whether you intend contributing or simply want to
understand how the software is structured. This documentation lays out the
current design of Canonical Kubernetes, following the [C4 model].

## System context

This overview of Canonical Kubernetes demonstrates the interactions of
Kubernetes with users and with other systems.

```{kroki} ../../assets/overview.puml
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

```{kroki} ../../assets/k8s-container.puml
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

```{kroki} ../../assets/k8sd-component.puml
```

At the core of the `k8sd` functionality we have the cluster manager that is
responsible for configuring the services, workload and features we deem
important for a Kubernetes cluster. Namely:

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

Canonical `k8s` Charms encompass two primary components: the [`k8s` charm][K8s charm] and the [`k8s-worker` charm][K8s-worker charm].

```{kroki} ../../assets/charms-architecture.puml
```

Charms are instantiated on a machine as a Juju unit, and a collection of units 
constitutes an application. Both `k8s` and `k8s-worker` units are responsible
for installing and managing its machine's `k8s` snap, however the charm type determines
the node's role in the Kubernetes cluster. The `k8s` charm manages `control-plane` nodes,
whereas the `k8s-worker` charm manages Kubernetes `worker` nodes. The administrator manages 
the cluster via the `juju` client, directing the `juju` controller to reach the model's 
eventually consistent state. For more detail on Juju's concepts, see the [Juju docs][].

The administrator may choose any supported cloud-types (Openstack, MAAS, AWS, GCP, Azure...) on
which to manage the machines making up the Kubernetes cluster. Juju selects a single leader unit
per application to act as a centralised figure with the model. The `k8s` leader oversees Kubernetes 
bootstrapping and enlistment of new nodes. Follower `k8s` units will join the cluster using
secrets shared through relation data from the leader. The entire lifecycle of the deployment
is orchestrated by the `k8s` charm, with tokens and cluster-related information being exchanged 
through Juju relation data.

Furthermore, the `k8s-worker` unit functions exclusively as a worker within the cluster, establishing
a relation with the `k8s` leader unit and requesting tokens and cluster-related information through
relation data. The `k8s` leader is responsible for issuing these tokens and revoking them if
a unit administratively departs the cluster.

The `k8s` charm also supports the integration of other compatible charms, enabling integrations 
such as connectivity with an external `etcd` datastore and the sharing of observability data with the 
[`Canonical Observability Stack (COS)`][COS docs]. This modular and integrated approach facilitates
a robust and flexible Canonical Kubernetes deployment managed through Juju.


<!-- LINKS -->
[C4 model]:           https://c4model.com/
[K8s charm]:          https://charmhub.io/k8s
[K8s-Worker charm]:   https://charmhub.io/k8s-worker
[Juju docs]:          https://juju.is/docs/juju
[COS docs]:           https://ubuntu.com/observability