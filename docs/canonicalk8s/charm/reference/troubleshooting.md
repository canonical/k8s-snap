# Troubleshooting

This page provides techniques for troubleshooting common {{product}}
issues dealing specifically with the charm.


## Adjusting Kubernetes node labels

### Problem

Control-Plane or Worker nodes are automatically marked with a label that is
unwanted.

For example, the control-plane node may be marked with both control-plane and
worker roles

```
node-role.kubernetes.io/control-plane=
node-role.kubernetes.io/worker=
```

### Explanation

Each kubernetes node comes with a set of node labels enabled by default. The k8s
snap defaults with both control-plane and worker role labels, while the worker
node only has a role label.

For example, consider the following simple deployment with a worker and a
control-plane.

```sh
sudo k8s kubectl get nodes
```

Outputs

```
NAME            STATUS   ROLES                  AGE     VERSION
juju-c212aa-1   Ready    worker                 3h37m   v1.32.0
juju-c212aa-2   Ready    control-plane,worker   3h44m   v1.32.0
```

### Solution

Adjusting the roles (or any label) be executed by adjusting the application's
configuration of `node-labels`.

To add another node label:

```sh
current=$(juju config k8s node-labels)
if [[ $current == *" label-to-add="* ]]; then
   # replace an existing configured label
   updated=${current//label-to-add=*/}
   juju config k8s node-labels="${updated} label-to-add=and-its-value"
else
   # specifically configure a new label
   juju config k8s node-labels="${current} label-to-add=and-its-value"
fi
```

To remove a node label which was added by default

```sh
current=$(juju config k8s node-labels)
if [[ $current == *" label-to-remove="* ]]; then
   # remove an existing configured label
   updated=${current//label-to-remove=*/}
   juju config k8s node-labels="${updated}"
else
   # remove an automatically applied label
   juju config k8s node-labels="${current} label-to-remove=-"
fi
```

#### Node Role example

To remove the worker node-rule on a control-plane:

```sh
juju config k8s node-labels="node-role.kubernetes.io/worker=-"
```

<!-- markdownlint-disable -->
## Cilium pod `fails to detect devices: unable to determine direct routing devices`
<!-- markdownlint-restore -->

### Problem

When deploying {{product}} on MAAS, the Cilium pods fail to start and reports
the error:

```
level=fatal msg="failed to start: daemon creation failed: failed to detect devices: unable to determine direct routing device. Use --direct-routing-device to specify it\nfailed to stop: unable to find controller ipcache-inject-labels" subsys=daemon
```

### Explanation

This issue was introduced in Cilium 1.15 and has been [reported here]. Both
`devices` and `direct-routing-device` lists must now be set in direct routing
mode. Direct routing mode is used by BPF, NodePort and BPF host routing.

If `direct-routing-device` is left undefined, it is automatically set to the
device with the k8s InternalIP/ExternalIP or the device with a default route.
However, bridge type devices are ignored in this automatic selection. In the
case of deploying on MAAS, a bridge interface is used as the default route and
therefore Cilium enters a failed state being unable to find the direct routing
device.

The solution is to add the bridge interface to the list of `devices` using
cluster annotations. Once the bridge interface is included in `devices`,
Cilium will automatically populate `direct-routing-device` when the pod
restarts—there is no need to set `direct-routing-device` manually.

### Solution

Identify the default route used for the cluster. The `route` command is part
of the net-tools Debian package.

```
route
```

In this example of deploying {{product}} on MAAS, the output is as follows:

```
Kernel IP routing table
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
default         _gateway        0.0.0.0         UG    0      0        0 br-ex
172.27.20.0     0.0.0.0         255.255.254.0   U     0      0        0 br-ex
```

The `br-ex` interface is the default interface used for this cluster. Apply
the annotation to the node adding bridge interfaces `br+` to the `devices` list:

```
juju config k8s cluster-annotations="k8sd/v1alpha1/cilium/devices=br+"
```

The `+` acts as a wildcard operator to allow all bridge interfaces to be picked
up by Cilium.

Restart the Cilium pod so it is recreated with the updated annotation and
devices. Get the pod name which will be in the form `cilium-XXXX` where XXXX
is unique to each pod:

```
sudo k8s kubectl get pods -n kube-system
```

Delete the pod:

```
sudo k8s kubectl delete pod cilium-XXXX -n kube-system
```

Verify the Cilium pod has restarted and is now in the running state:

```
sudo k8s kubectl get pods -n kube-system
```

## Cilium pod fails to start as `cilum_vxlan: address already in use`

### Problem

When deploying {{product}} on a cloud provider such as OpenStack, the Cilium
pods fail to start and reports the error:

```
failed to start: daemon creation failed: error while initializing daemon: failed
while reinitializing datapath: failed to setup vxlan tunnel device: setting up
vxlan device: creating vxlan device: setting up device cilium_vxlan: address
already in use
```

### Explanation

Fan networking is automatically enabled in some substrates. This causes
conflicts with some CNIs such as Cilium. This conflict of
`address already in use` causes Cilium to be unable to set up it's VXLAN
tunneling network. There may also be other networking components on the system
attempting to use the default port for their own VXLAN interface that will
cause the same error.

### Solution

You can either disable fan networking or configure Cilium to use another tunnel
port.

#### Disable fan networking

```{note}
Only disable fan networking if it is not in use. Disabling fan networking may
have implications on your cluster where assets such as LXD VMs, are not
reachable if they rely on fan networking for communication.
```

Apply the following configuration to the Juju model:

```
juju model-config container-networking-method=local fan-config=
```

#### Change Cilium tunnel-port

Connect to the node and set the annotation `tunnel-port` to an appropriate value
(the default is 8472).

```
sudo k8s set annotation="k8sd/v1alpha1/cilium/tunnel-port=<PORT-NUMBER>"
```

Since the Cilium pods are in a failing state, the recreation of the VXLAN
interface is automatically triggered. Verify the VXLAN interface has come up:

```
ip link list type vxlan
```

It should be named `cilium_vxlan` or something similar.

Verify that Cilium is now in a running state:

```
sudo k8s kubectl get pods -n kube-system
```

## Bootstrap config change prevention

### Problem

When upgrading {{product}} or changing `bootstrap-*` configuration options,
the charm could block and produce a message on each unit:

```
k8s/0*  blocked  idle  0  10.246.154.22  6443/tcp  bootstrap-datastore is immutable; revert to 'managed-etcd'
```

or

```
k8s/0*  blocked  idle  0  10.246.154.22  6443/tcp  bootstrap-pod-cidr is immutable; revert to '10.0.0.0/8'
```

### Explanation

Juju allows for configuration to be fully mutable; however, some k8s options --
specifically those starting with `bootstrap-` are
[immutable](charm_configurations). Juju allows for these options to change over
time, but it will cause the charm to block if adjusted during day-2 operation
of the application.

```{note}
The only exception is `bootstrap-node-taints` which is allowed to be changed on
`k8s` or `k8s-worker` applications without triggering this blocked condition.
This configuration is only used when joining a new unit to the cluster; so
changing it does not affect the runtime taints of a node.
```

### Solution

`juju status` reflects the desired action. If the status message indicates

```
k8s/0*  blocked  idle  0  10.246.154.22  6443/tcp  bootstrap-datastore is immutable; revert to 'managed-etcd'
```

The appropriate adjustment is to update the configuration value:

```
juju config k8s bootstrap-datastore='managed-etcd'
```

<!-- LINKS -->
[reported here]: https://github.com/cilium/cilium/issues/30889
[charm_configuration]: ./charm-configuration
