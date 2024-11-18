# In-Place Upgrades

Regularly upgrading the Kubernetes version of the machines in a cluster 
is important. While rolling upgrades are a popular strategy, certain 
situations will require in-place upgrades:

- Resource constraints (i.e. cost of additional machines).
- Expensive manual setup process for nodes.

## Annotations

CAPI machines are considered immutable. Consequently, machines are replaced 
instead of reconfigured.
While CAPI doesn't support in-place upgrades, {{product}} CAPI does 
by leveraging annotations for the implementation.
For a deeper understanding of the CAPI design decisions, consider reading about 
[machine immutability in CAPI][1], and Kubernetes objects: [`labels`][2], 
[`spec` and `status`][3].

## Controllers

In {{product}} CAPI, there are two main types of controllers that handle the 
process of performing in-place upgrades:

- Single Machine In-Place Upgrade Controller
- Orchestrated In-Place Upgrade Controller

The core component of performing an in-place upgrade is the `Single Machine 
Upgrader`. The controller watches for annotations on machines and reconciles 
them to ensure the upgrades happen smoothly. 

The `Orchestrator` watches for certain annotations on 
machine owners, reconciles them and upgrades groups of owned machines. 
It’s responsible for ensuring that all the machines owned by the 
reconciled object get upgraded successfully.

The main annotations that drive the upgrade process are as follows:

- `v1beta2.k8sd.io/in-place-upgrade-to` --> `upgrade-to` : Instructs 
the controller to perform an upgrade with the specified option/method. 
- `v1beta2.k8sd.io/in-place-upgrade-status` --> `status` : As soon as the 
controller starts the upgrade process, the object will be marked with the 
`status` annotation which can either be `in-progress`, `failed` or `done`.
- `v1beta2.k8sd.io/in-place-upgrade-release` --> `release` : When the 
upgrade is performed successfully, this annotation will indicate the current 
Kubernetes release/version installed on the machine.

For a complete list of annotations and their values please 
refer to the [annotations reference page][4]. This explanation proceeds 
to use abbreviations of the mentioned labels.

### Single Machine In-Place Upgrade Controller

The Machine objects can be marked with the `upgrade-to` annotation to 
trigger an in-place upgrade for that machine. While watching for changes 
on the machines, the single machine upgrade controller notices this annotation  
and attempts to upgrade the Kubernetes version of that machine to the 
specified version.

Upgrade methods or options can be specified to upgrade to a snap channel, 
revision, or a local snap file already placed on the 
machine in air-gapped environments.

A successfully upgraded machine shows the following annotations:

```yaml
annotations:
  v1beta2.k8sd.io/in-place-upgrade-release: "channel=1.31/stable"
  v1beta2.k8sd.io/in-place-upgrade-status: "done"
```

If the upgrade fails, the controller will mark the machine and retry 
the upgrade immediately:

```yaml
annotations:
  # the `upgrade-to` causes the retry to happen
  v1beta2.k8sd.io/in-place-upgrade-to: "channel=1.31/stable"
  v1beta2.k8sd.io/in-place-upgrade-status: "failed"

  # orchestrator will notice this annotation and knows that the 
  # upgrade for this machine failed
  v1beta2.k8sd.io/in-place-upgrade-last-failed-attempt-at: "Sat, 7 Nov 
  2024 13:30:00 +0400"
```

By applying and removing annotations, the single machine 
upgrader determines the upgrade status of the machine it’s trying to 
reconcile and takes necessary actions to successfully complete an 
in-place upgrade. The following diagram shows the flow of the in-place 
upgrade of a single machine:

![Diagram][img-single-machine]

### Machine Upgrade Process

The {{product}}'s `k8sd` daemon exposes endpoints that can be used to 
interact with the cluster. The single machine upgrader calls the  
`/snap/refresh` endpoint on the machine to trigger the upgrade 
process while checking `/snap/refresh-status` periodically. 

![Diagram][img-k8sd-call]

### In-place upgrades on large workload clusters

While the “Single Machine In-Place Upgrade Controller” is responsible 
for upgrading individual machines, the "Orchestrated In-Place Upgrade 
Controller" ensures that groups of machines will get upgraded.
By applying the `upgrade-to` annotation on an object that owns machines 
(e.g. a `MachineDeployment`), this controller will mark the owned machines 
one by one which will cause the "Single Machine Upgrader" to pickup those 
annotations and upgrade the machines. To avoid undesirable situations
 like quorum loss or severe downtime, these upgrades happen in sequence. 

The failures and successes of individual machine upgrades will be reported back 
to the orchestrator by the single machine upgrader via annotations.

The illustrated flow of orchestrated in-place upgrades:

![Diagram][img-orchestrated]

<!-- IMAGES -->

[img-single-machine]: https://assets.ubuntu.com/v1/1200f040-single-machine.png
[img-k8sd-call]: https://assets.ubuntu.com/v1/518eb73a-k8sd-call.png
[img-orchestrated]: https://assets.ubuntu.com/v1/8f302a00-orchestrated.png

<!-- LINKS -->
[1]: https://cluster-api.sigs.k8s.io/user/concepts#machine-immutability-in-place-upgrade-vs-replace
[2]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
[3]: https://kubernetes.io/docs/concepts/overview/working-with-objects/#object-spec-and-status
[4]: ../reference/annotations.md
