# High availability

High availability (HA) is a core feature of {{ product }}, ensuring that
a Kubernetes cluster remains operational and resilient, even when nodes or
critical components encounter failures. This capability is crucial for
maintaining continuous service for applications and workloads running in
production environments.

HA is automatically enabled in {{ product }} for clusters with three or
more nodes independent of the deployment method. By distributing key components
across multiple nodes, HA reduces the risk of downtime and service
interruptions, offering built-in redundancy and fault tolerance.

## Key components of a highly available cluster

A highly available Kubernetes cluster exhibits the following characteristics:

### 1. **Multiple nodes for redundancy**

Having multiple nodes in the cluster ensures workload distribution and
redundancy. If one node fails, workloads will be rescheduled automatically on
other available nodes without disrupting services. This node-level redundancy
minimizes the impact of hardware or network failures.

### 2. **Control plane redundancy**

The control plane manages the cluster’s state and operations. For high
availability, the control plane components—such as the API server, scheduler,
and controller-manager—are distributed across multiple nodes. This prevents a
single point of failure from rendering the cluster inoperable.

### 3. **Highly available datastore**

{{ product }} uses **etcd** to manage the Kubernetes
cluster state.
Etcd leverages the Raft consensus algorithm for leader
election and voting, ensuring reliable data replication and failover
capabilities. When a leader node fails, a new leader is elected seamlessly
without administrative intervention. This mechanism allows the cluster to
remain operational even in the event of node failures.

## Fault tolerance in a 2-node setup

```{warning}
Avoid using a two-node cluster in production environments, as it
can lead to availability issues if either node fails.
```

Quorum is the minimum number of nodes in a cluster that must agree
before making decisions, usually a majority. It's essential for
high availability because it ensures the system can keep running safely
even if some nodes fail.

In a two-node etcd cluster, the second node is added as a full voter right
away, effectively forming a quorum with just two nodes. This means that if
either node fails, the system becomes unavailable since both nodes are needed
to maintain quorum.

The [etcd documentation] explicitly warns against reconfiguration of a
two-member cluster by removing a member. Because quorum requires a
majority of nodes, and the majority in a two-node setup is also two,
any node failure during the removal process can render the
cluster inoperable.

<!-- LINKS -->
[etcd documentation]: https://etcd.io/docs/latest/faq/
