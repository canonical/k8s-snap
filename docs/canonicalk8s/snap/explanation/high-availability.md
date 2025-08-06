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

By default, {{ product }} uses **dqlite** to manage the Kubernetes
cluster state. Dqlite leverages the Raft consensus algorithm for leader
election and voting, ensuring reliable data replication and failover
capabilities. When a leader node fails, a new leader is elected seamlessly
without administrative intervention. This mechanism allows the cluster to
remain operational even in the event of node failures. 
<!-- TODO: When Dqlite docs are back, uncomment this line
More details on
replication and leader elections can be found in
the [dqlite replication documentation][Dqlite-replication].
-->
<!-- LINKS -->
<!-- [Dqlite-replication]: https://dqlite.io/docs/explanation/replication --> 
