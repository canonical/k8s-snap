---
myst:
  html_meta:
    description: Troubleshoot your Canonical Kubernetes cluster using this how-to guide.
---

# How to troubleshoot {{product}}

Identifying issues in a Kubernetes cluster can be difficult, especially to new
users. With {{product}} we aim to make deploying and managing your cluster as
easy as possible. This how-to guide will walk you through the steps to
troubleshoot your {{product}} cluster.

## Check the cluster status

Verify that the cluster status is ready by running:

```
juju status
```

You should see a command output similar to the following:

```
Model        Controller           Cloud/Region         Version  SLA          Timestamp
k8s-testing  localhost-localhost  localhost/localhost  3.6.1    unsupported  09:06:50Z

App         Version  Status  Scale  Charm       Channel    Rev  Exposed  Message
k8s         1.32.0   active      1  k8s         1.32/beta  179  no       Ready
k8s-worker  1.32.0   active      1  k8s-worker  1.32/beta  180  no       Ready

Unit           Workload  Agent  Machine  Public address  Ports     Message
k8s-worker/0*  active    idle   1        10.94.106.154             Ready
k8s/0*         active    idle   0        10.94.106.136   6443/tcp  Ready

Machine  State    Address        Inst id        Base          AZ  Message
0        started  10.94.106.136  juju-380ff2-0  ubuntu@24.04      Running
1        started  10.94.106.154  juju-380ff2-1  ubuntu@24.04      Running
```

Interpreting the Output:

- The `Workload` column shows the status of a given service.
- The `Message` section details the health of a given service in the cluster.
- The `Agent` column reflects any activity of the Juju agent.

During deployment and maintenance the workload status will reflect the node's
activity. An example workload may display `maintenance` along with the message
details: `Ensuring snap installation`.

During normal cluster operation the `Workload` column reads `active`, the
`Agent` column shows `idle`, and the messages will either read `Ready` or
another descriptive term.

## Test the API server health

Fetch the kubeconfig file for a control-plane node in the cluster by running:

```
juju run k8s/leader get-kubeconfig | yq .kubeconfig > cluster-kubeconfig.yaml
```

```{warning}
When running `juju run k8s/leader get-kubeconfig` you retrieve the kubeconfig file that uses one of the unit's  public IP addresses in the Kubernetes endpoint. This endpoint ip can be overridden by providing a `server` argument if the API is exposed through a load balancer.
```

Verify that the API server is healthy and reachable by running:

```
kubectl --kubeconfig cluster-kubeconfig.yaml get all
```

This command lists resources that exist under the default namespace. If the API
server is healthy you should see a command output similar to the following:

```
NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.152.183.1   <none>        443/TCP   29m
```

A typical error message may look like this if the API server can not be reached:

```
The connection to the server 127.0.0.1:6443 was refused - did you specify the right host or port?
```

Check the status of the API server service:

```
juju exec --unit k8s/0 -- systemctl status snap.k8s.kube-apiserver
```

Access the logs of the API server service by running:

```
juju exec --unit k8s/0 -- journalctl -u snap.k8s.kube-apiserver
```

A failure can mean that:

* The API server is not reachable due to network issues or firewall limitations
* The API server on the particular node is unhealthy
* The control-plane node that's being reached is down

Try reaching the API server on a different unit by retrieving the kubeconfig
file with `juju run <k8s/unit#> get-kubeconfig`. Please replace `#` with the
desired unit's number.

## Check the cluster nodes' health

Confirm that the nodes in the cluster are healthy by looking for the `Ready`
status:

```
kubectl --kubeconfig cluster-kubeconfig.yaml get nodes
```

You should see a command output similar to the following:

```
NAME            STATUS   ROLES                  AGE     VERSION
juju-380ff2-0   Ready    control-plane,worker   9m30s   v1.32.0
juju-380ff2-1   Ready    worker                 77s     v1.32.0
```


## Troubleshoot an unhealthy node

Every healthy {{ product }} node has certain services up and running. The
required services depend on the type of node.

Services running on both the control plane and worker nodes:

* `k8sd`
* `kubelet`
* `containerd`
* `kube-proxy`

Services running only on the control-plane nodes:

* `kube-apiserver`
* `kube-controller-manager`
* `kube-scheduler`
* `etcd`

Services running only on the worker nodes:

* `k8s-apiserver-proxy`

SSH into the unhealthy node by running:

```
juju ssh <k8s/unit#>
```

Check the status of the services on the failing node by running:

```
sudo systemctl status snap.k8s.<service>
```

Check the logs of a failing service by executing:

```
sudo journalctl -xe -u snap.k8s.<service>
```

If the issue indicates a problem with the configuration of the services on the
node, examine the arguments used to run these services.

The arguments of a service on the failing node can be examined by reading the
file located at `/var/snap/k8s/common/args/<service>`.

## Investigate system pods' health

Check whether all of the cluster's pods are `Running` and `Ready`:

```
kubectl --kubeconfig cluster-kubeconfig.yaml get pods -n kube-system
```

The pods in the `kube-system` namespace belong to {{product}}' features such as
`network`. Unhealthy pods could be related to configuration issues or nodes not
meeting certain requirements.

## Troubleshoot a failing pod

Look at the events on a failing pod by running:

```
kubectl --kubeconfig cluster-kubeconfig.yaml describe pod <pod-name> -n <namespace>
```

Check the logs on a failing pod by executing:

```
kubectl --kubeconfig cluster-kubeconfig.yaml logs <pod-name> -n <namespace>
```

You can check out the upstream [debug pods documentation][] for more
information.

## Use the built-in inspection command

{{product}} ships with a command to compile a complete report on {{product}} and
its underlying system. This is an essential tool for bug reports and for
investigating whether a system is (or isn’t) working.

The inspection command can be executed on a specific unit by running the
following commands:

```
juju exec --unit <k8s/unit#> -- sudo k8s inspect /home/ubuntu/inspection-report.tar.gz
juju scp <k8s/unit#>:/home/ubuntu/inspection-report.tar.gz ./
```

See the [inspection report reference page] for more details.

## Collect debug information

To collect comprehensive debug output from your {{product}} cluster, install
and run [juju-crashdump][] on a computer that has the Juju client installed.
Please ensure that the current controller and model are pointing at your
{{product}} deployment.

```
sudo snap install juju-crashdump --classic --channel edge
juju-crashdump -a debug-layer -a config
```

Running the `juju-crashdump` script will generate a tarball of debug
information that includes [systemd][] unit status and logs, Juju logs, charm
unit data, and Kubernetes cluster information. Please include the generated
tarball when filing a bug.

## Common issues

If you find any issue while working with {{product}} it is highly likely that someone from the community has already faced the same problem. We have documented some common issues users face and their workarounds.


### Adjusting Kubernetes node labels

Control-Plane or Worker nodes are automatically marked with a label that is
unwanted. For example, the control-plane node may be marked with both 
control-plane and worker roles.

```
node-role.kubernetes.io/control-plane=
node-role.kubernetes.io/worker=
```

````{dropdown} Explanation

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
````

````{dropdown} Solution

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
````

<!-- markdownlint-disable -->
### Cilium pod `fails to detect devices: unable to determine direct routing devices`
<!-- markdownlint-restore -->

When deploying {{product}} on MAAS, the Cilium pods fail to start and reports
the error:

```
level=fatal msg="failed to start: daemon creation failed: failed to detect devices: unable to determine direct routing device. Use --direct-routing-device to specify it\nfailed to stop: unable to find controller ipcache-inject-labels" subsys=daemon
```

````{dropdown} Explanation

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
````

````{dropdown} Solution

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
````

### Cilium pod fails to start as `cilum_vxlan: address already in use`

When deploying {{product}} on a cloud provider such as OpenStack, the Cilium
pods fail to start and reports the error:

```
failed to start: daemon creation failed: error while initializing daemon: failed
while reinitializing datapath: failed to setup vxlan tunnel device: setting up
vxlan device: creating vxlan device: setting up device cilium_vxlan: address
already in use
```

````{dropdown} Explanation

Fan networking is automatically enabled in some substrates. This causes
conflicts with some CNIs such as Cilium. This conflict of
`address already in use` causes Cilium to be unable to set up it's VXLAN
tunneling network. There may also be other networking components on the system
attempting to use the default port for their own VXLAN interface that will
cause the same error.
````

````{dropdown} Solution

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
````

### Bootstrap config change prevention


When upgrading {{product}} or changing `bootstrap-*` configuration options,
the charm could block and produce a message on each unit:

```
k8s/0*  blocked  idle  0  10.246.154.22  6443/tcp  bootstrap-datastore is immutable; revert to 'managed-etcd'
```

or

```
k8s/0*  blocked  idle  0  10.246.154.22  6443/tcp  bootstrap-pod-cidr is immutable; revert to '10.0.0.0/8'
```

````{dropdown} Explanation

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
````

````{dropdown} Solution

`juju status` reflects the desired action. If the status message indicates

```
k8s/0*  blocked  idle  0  10.246.154.22  6443/tcp  bootstrap-datastore is immutable; revert to 'managed-etcd'
```

The appropriate adjustment is to update the configuration value:

```
juju config k8s bootstrap-datastore='managed-etcd'
```
````

## Report a bug

If you cannot solve your issue and believe that the fault may lie in
{{product}}, please [file an issue on the project repository][].

Help us deal effectively with issues by including the report obtained from the
inspect script, the tarball obtained from `juju-crashdump`, as well as any
additional logs, and a summary of the issue.

You can check out the upstream [debug documentation][] for more details on
troubleshooting a Kubernetes cluster.

<!-- Links -->

[file an issue on the project repository]: https://github.com/canonical/k8s-operator/issues/new/choose
[juju-crashdump]: https://github.com/juju/juju-crashdump
[systemd]: https://systemd.io
[debug pods documentation]: https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods
[debug documentation]: https://kubernetes.io/docs/tasks/debug
[inspection report reference page]: /snap/reference/inspection-reports.md
[reported here]: https://github.com/cilium/cilium/issues/30889