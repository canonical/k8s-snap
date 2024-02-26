# Architecture

A system architecture document is the starting point for many interested participants in a project, whether you intend contirbuting or simply want to understand how the software is structured. This documentation lays out the current design of Canonical Kubernetes, following the [C4 model]. 

##  System context 

This overview of Canononical Kubernetes demonstrates the interactions of Kubernetes with users and with other systems.

```{kroki} ../assets/overview.puml
```

Two actors interact with the kubernetes snap:

- **K8s admin**: The administrator of the cluster interacts directly with the Kubernetes API server. Out of the box our K8s distribution offers admin access to the cluster. That initial user is able to configure the cluster to match their needs and of course create other users that may or may not have the same privileges. The K8s admin is also able to deploy workloads running in the cluster.

- **K8s user**: A user consuming the services hosted in the cluster. Users do not have access to the Kubernetes API server. They need to access the cluster through the options (nodeport, ingress, load-balancer) offered by the administrator who deployed the workload they are interested in.

There are non-human users of the k8s snap. In this case that is the [K8s charm]. The K8s charm needs to drive the Kubernetes cluster and to orchestrate the multi-node clustering operations.

A set of external systems need to be easily integrated with our k8s distribution. We have identified the following:
 - **Loadbalancer**: Although the k8s snap distribution comes with a loadbalancer we expect the end customer environment to have a loadbalancer and thus we need to integrate with it.
- **Storage**: Kubernetes typically expects storage to be external to the cluster. The k8s snap comes with a local storage option but we still need to offer proper integration with any storage solution.
- **Identity management**: Out of the box the k8s snap offers credentials for an admin user. The admin user can complete the integration with any identity management system available or do user management manually.
- **External datastore**: By default, Kubernetes uses etcd to keep track of state. Our k8s snap comes with `dqlite` as its datastore. We should however be able to use any end client owned datastore installation. That should include an external `postgresql` or `etcd`.

## The k8s snap

Looking more closely at what is conatined within the K8s snap istelf:

```{kroki} ../assets/k8s-container.puml
```

The k8s snap distribution includes the following:

- **Kubectl**: through which users and other systems interact with Kubernetes and drive the cluster operations.
- **K8s upstream services**: These are Kubernetes binaries built from upstream and shipped in the snap.
- **Components** are the workloads and features we deem important to be available to our users and therefore are shipped in the snap and are enabled, configured and disabled in a guided way.
- State is backed up by **dqlite** by default, which keeps that state of the Kubernetes cluster as well as the the state we maintain for the needs of the cluster operations.
- **Runtime**: `containerd` and `runc` are the shipped container runtimes.
- **K8sd**: which implements the operations logic and exposes that functionality via CLIs and REST APIs.

## K8sd

K8sd is the component that implements and exposes the operations functionality needed for managing the Kubernetes cluster.

```{kroki} ../assets/k8sd-component.puml
```

At the core of the `k8sd` functionality we have the components and cluster managers:
The components manager is responsible for the workload features we deem important for a Kubernetes cluster. Namely:

- DNS
- CNI
- ingress
- gateway API
- load-balancer
- local storage
- observability

The cluster manager is responsible for implementing the formation of the cluster. This includes operations such as joining/removing nodes into the cluster and reporting status.

This functionality is exposed via the following interfaces:

- The **CLI**: The CLI is available to only the root user on the k8s snap and all CLI commands are mapped to respective REST calls.

- The **REST API**: apart from serving the CLI, is also used by the charm that has to programmatically drive the Kubernetes cluster. 



<!-- LINKS -->
[C4 model]: https://c4model.com/
[K8s charm]: https://charmhub.io/k8s
