# Node roles

This document explains how {{product}} assigns Kubernetes node roles, how this
affects workload scheduling, and how this differs from standard Kubernetes
implementations.

When bootstrapping a cluster, {{product}} assigns the following roles to nodes:

| Node Type | Default Roles | Scheduling | Notes |
|-----------|--------------|------------|-------|
| Control Plane | `control-plane`, `worker` | Allowed | Differs from `kubeadm` which prevents workload scheduling on control plane nodes |
| Worker | `worker` | Allowed | Standard behavior |

## Rationale

Most kubernetes implementations (like kubeadm) set a `NoSchedule` taint on
control plane nodes as a security measure to isolate control planes from
workloads.

{{product}} intentionally allows workload scheduling on control plane nodes to
simplify initial cluster setup, especially for single-node deployments.
However, users should be aware that:

- Scheduling workloads on control plane nodes may introduce security risks
- For production multi-node clusters, isolating the control plane is
  recommended

## Enforcing control plane isolation

To apply standard Kubernetes security practices in a multi-node cluster, you
can set a taint on the control plane node using the `taint` command:

```
sudo k8s kubectl taint node node1 node-role.kubernetes.io/control-plane:NoSchedule
```
