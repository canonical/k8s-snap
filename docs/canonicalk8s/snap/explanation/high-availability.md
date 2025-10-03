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

By default, {{ product }} uses **etcd** to manage the Kubernetes
cluster state. Users can also switch to **Dqlite** if they wish. 
Both etcd and Dqlite leverage the Raft consensus algorithm for leader
election and voting, ensuring reliable data replication and failover
capabilities. When a leader node fails, a new leader is elected seamlessly
without administrative intervention. This mechanism allows the cluster to
remain operational even in the event of node failures. 
<!-- TODO: When Dqlite docs are back, uncomment this line
More details on
replication and leader elections can be found in
the [dqlite replication documentation][Dqlite-replication].
-->

## Fault tolerance in a 2-node setup

```{warning}
Avoid using a two-node cluster in production environments, as it 
can lead to availability issues if either node fails.
```

Quorum is the minimum number of nodes in a cluster that must agree 
before making decisions, usually a majority. It's essential for 
high availability because it ensures the system can keep running safely 
even if some nodes fail.

Dqlite and etcd handle quorum formation differently in a two-node 
cluster configuration. With Dqlite, the second node that joins the cluster 
initially acts as a follower, a node that doesn't participate in forming 
the quorum and is holding a replica of the state, and is only promoted to a 
voter once the cluster reaches three nodes. In contrast, etcd adds the 
second node as a full voter right away, effectively forming a quorum with 
just two nodes.

This design difference impacts cluster behavior during node failure. If one 
node fails in a Dqlite-backed cluster, the cluster can still operate as 
long as the remaining node is the leader. However, in an etcd-backed cluster, 
the system becomes unavailable regardless of which node goes down, since both 
nodes are needed to maintain quorum. Note that in the case of Dqlite, if the 
leader node is the one that goes down, the remaining follower cannot take over, 
and the cluster becomes unavailable. 

[etcd documentation] explicitly warns against reconfiguration of a two-member 
cluster by removing a member. Because quorum requires a majority of nodes, 
and the majority in a two-node setup is also two, any node failure during the 
removal process can render the cluster inoperable.

<!-- LINKS -->
<!-- [Dqlite-replication]: https://dqlite.io/docs/explanation/replication --> 
[etcd documentation]: https://etcd.io/docs/latest/faq/
