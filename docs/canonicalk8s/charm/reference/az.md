# Availability Zones

An availability zone determines the specific location where Juju
provisions hardware, enhancing redundancy and resilience in case
of an outage.

When using Juju with a cloud that supports availability zones, the [zone]
can be specified either via a [placement directive], a [constraint], or
will be automatically selected by Juju.

In the following example, the {{product}} charm is deployed on AWS in the
`us-east-1` region with the `us-east-1a` availability zone:

```
Model       Controller  Cloud/Region   Version  SLA          Timestamp
test-model  aws-ctrl    aws/us-east-1  3.6.3    unsupported  12:44:15+04:00

App  Version  Status   Scale  Charm  Channel      Rev  Exposed  Message
k8s  1.32.2   active   1      k8s    1.32/stable  347  no       Ready

Unit    Workload  Agent  Machine  Public address  Ports     Message
k8s/0*  active    idle   0        <REDACTED>      6443/tcp  Ready

Machine  State    Address     Inst id     Base          AZ          Message
0        started  <REDACTED>  <REDACTED>  ubuntu@22.04  us-east-1a  running
```

Depending on the AZ that is announced by Juju (in this case `us-east-1a`),
the nodes will be labeled with the [Kubernetes well-known topology label]:

```
local-machine$ juju ssh 0
aws-machine-0$ k8s kubectl get nodes -ojson | jq '.items[].metadata.labels'
{
  ...
  topology.kubernetes.io/zone: "us-east-1a"
}
```

This label is applied by the charm operator only if the node is not already
labeled with `topology.kubernetes.io/zone`. This means:

- If the node is labeled with `topology.kubernetes.io/zone` by a component or
controller other than the charm operator, the operator will not overwrite
this label.
- Even if the node is already labeled by the charm operator, in case the
underlying AZ changes from the Juju POV, the operator will not update the
label to reflect this new AZ.

<!-- LINKS -->
[zone]: https://documentation.ubuntu.com/juju/3.6/reference/zone/
[placement directive]: https://documentation.ubuntu.com/juju/3.6/reference/placement-directive/#zone-zone
[constraint]: https://documentation.ubuntu.com/juju/3.6/reference/constraint/#zones
[Kubernetes well-known topology label]: https://kubernetes.io/docs/reference/labels-annotations-taints/#topologykubernetesiozone
