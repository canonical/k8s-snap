# Default DNS

{{product}} includes a default DNS (Domain Name System) solution which is
essential for internal cluster communication. When enabled, the DNS facilitates
service discovery by assigning each service a DNS name.

## Scaling DNS

Default DNS is scaled automatically through a Horizontal Pod Autoscaler ([HPA])
which monitors DNS pods' CPU and memory resource usage and adjusts the number of
replicas accordingly.

## DNS scheduling

DNS pods are scheduled with [topology spread constraints] to spread them across
the cluster nodes and zones when applicable.
When {{ product }} detects that all pods are scheduled on the same node, it
will restart the DNS pods to re-balance their distribution.

A [priority class] is assigned to DNS pods to ensure their scheduling
before pods which are not node critical.

## DNS maintenance

When performing maintenance operations on the cluster, please be aware that
the pod disruption budget ([PDB]) will only allow one DNS pod to be taken down
at a time.

<!--LINKS -->
[PDB]: https://kubernetes.io/docs/tasks/run-application/configure-pdb/#specifying-a-poddisruptionbudget
[HPA]: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[topology spread constraints]: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
[priority class]: https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass