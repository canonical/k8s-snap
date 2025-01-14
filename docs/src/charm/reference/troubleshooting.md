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
