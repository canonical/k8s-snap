# Clustering

Kubernetes clustering allows you to manage a group of hosts as a single entity.
This enables applications to be deployed across a cluster of machines without
tying them specifically to one host, providing high availability and
scalability. In Canonical Kubernetes the addition of `k8sd` to the Kubernetes
ecosystem introduces enhanced capabilities for cluster coordination and
management.

## Kubernetes Cluster Topology

A Kubernetes cluster consists of at least one control plane node and multiple
worker nodes. Each node is a server (physical or virtual) that runs
[Kubernetes components]. In Canonical Kubernetes, the components are bundled
inside the `k8s-snap`. The cluster's topology divides reponsibilities between
the control plane node(s) and the worker nodes, ensuring efficient management
and scheduling of workloads.

This is the overview of a Canonical Kubernetes cluster:

```{kroki} ../assets/ck-cluster.puml
```

## The Role of `k8sd` in Kubernetes Clustering

`k8sd` plays a vital role in the Canonical Kubernetes architecture, enhancing
the functionality of both the Control Plane and Worker nodes through the use
of [microcluster]. This component simplifies cluster management tasks, such as
adding or removing nodes and integrating them into the cluster. It also
manages essential features like DNS and networking within the cluster,
streamlining the entire process for a more efficient operation.

## Integration into the Kubernetes Cluster Topology

### Control Plane Node
The control plane node orchestrates the cluster, making decisions about
scheduling, deployment and management. With the addition of `k8sd`, the control
plane node's components include:
- **API Server (kube-apiserver)**: Acts as the front-end for the Kubernetes
    control plane. It exposes the Kubernetes API and is the central management
    entity through which all components and external users interact.
- **Scheduler (kube-scheduler)**: Responsible for allocating pods to nodes
    based on various criteria such as resource availability and constraints.
- **Controller Manager (kube-controller-manager)**: Runs controller processes
    that regulate the state of the cluster, ensuring the desired state matches
    the observed state.
- **k8s-dqlite**: A fast, embedded, persistent in-memory key-value store with 
    Raft consensus used to store all cluster data.
- **k8sd**: Implements and exposes the operations functionality needed for
    managing the Kubernetes cluster.

### Worker Node
Worker nodes are responsible for running the applications and workloads. Worker
nodes, can interact with the `k8sd` API, gaining capabilities to manage its
entire lifecycle. Their components include:

- **Local API Server Proxy**: This component forwards requests to the control
    plane nodes.
- **Kubelet**: Communicates with the control plane node and manages the
    containers running on the machine according to the configurations provided
    by the user.
- **Kube-Proxy (kube-proxy)**: Manages network communication within the cluster.
- **Container Runtime**: The software responsible for running containers. In
    Canonical Kubernetes the runtime is `containerd`.



<!-- LINKS -->

[Kubernetes Components]: https://kubernetes.io/docs/concepts/overview/components/
[microcluster]: https://github.com/canonical/microcluster
