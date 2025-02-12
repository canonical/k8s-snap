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
device. The bridge interface must be added to the list of `devices` using
cluster annotations so that `direct-routing-device` will not skip the bridge
interface.

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

<!-- LINKS -->
[reported here]: https://github.com/cilium/cilium/issues/30889
