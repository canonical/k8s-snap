---
myst:
  html_meta:
    description: Troubleshoot and resolve common issues with a Canonical Kubernetes (k8s) snap based cluster in this how-to guide.
---

# How to troubleshoot {{product}}

Identifying issues in a Kubernetes cluster can be difficult, especially to new
users. With {{product}} we aim to make deploying and managing your cluster as
easy as possible. This how-to guide will walk you through the steps to
troubleshoot your {{product}} cluster.

## Check the basics 

First ensure that all the cluster components are up and in a healthy state.

### Check the cluster status

Verify that the cluster status is ready by running the following command:

```
sudo k8s status
```

You should see a command output similar to the following:

```
cluster status:           ready
control plane nodes:      10.94.106.249:6400 (voter), 10.94.106.208:6400 (voter), 10.94.106.99:6400 (voter)
high availability:        yes
datastore:                etcd
network:                  enabled
dns:                      enabled at 10.152.183.106
ingress:                  disabled
load-balancer:            disabled
local-storage:            enabled at /var/snap/k8s/common/rawfile-storage
gateway                   enabled
```

### Test the API server health

Verify that the API server is healthy and reachable by running the following
command on a control plane node:

```
sudo k8s kubectl get all
```

This command lists resources that exist under the default namespace. You should
see a command output similar to the following if the API server is healthy:

```
NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.152.183.1   <none>        443/TCP   29m
```

A typical error message may look like this if the API server cannot be reached:

```
The connection to the server 127.0.0.1:6443 was refused - did you specify the right host or port?
```

A failure can mean that the API server on the particular node is unhealthy.
Check the status of the API server service:

```
sudo systemctl status snap.k8s.kube-apiserver
```

Access the logs of the API server service by running the following command:

```
sudo journalctl -xe -u snap.k8s.kube-apiserver
```

If you are trying to reach the API server from a host that is not a
control plane node, a failure could mean that:

* The API server is not reachable due to network issues or firewall limitations
* The API server is failing on the control plane node that's being reached
* The control plane node that's being reached is down

```{warning}
When running `sudo k8s config` on a control plane node you retrieve the kubeconfig file that uses this node's IP address.
```

Try reaching the API server on a different control plane node by updating the
IP address that's used in the kubeconfig file.

### Check the cluster nodes' health

Confirm that the nodes in the cluster are healthy by looking for the `Ready`
status:

```
sudo k8s kubectl get nodes
```

You should see a command output similar to the following:

```
NAME     STATUS   ROLES                  AGE     VERSION
node-1   Ready    control-plane,worker   10m     v1.32.0
node-2   Ready    control-plane,worker   6m51s   v1.32.0
node-3   Ready    control-plane,worker   6m21s   v1.32.0
```

### Troubleshoot an unhealthy node

Every healthy {{ product }} node has certain services up and running. The
required services depend on the type of node.

Services running on both control plane and worker nodes:

* `k8sd`
* `kubelet`
* `containerd`
* `kube-proxy`

Services running only on control plane nodes:

* `kube-apiserver`
* `kube-controller-manager`
* `kube-scheduler`
* `etcd`

Services running only on worker nodes:

* `k8s-apiserver-proxy`

Check the status of these services on the failing node by running the following
command:

```
sudo systemctl status snap.k8s.<service>
```

The logs of a failing service can be checked by running the following command:

```
sudo journalctl -xe -u snap.k8s.<service>
```

If the issue indicates a problem with the configuration of the services on the
node, examine the arguments used to run these services.

The arguments of a service on the failing node can be examined by reading the
file located at `/var/snap/k8s/common/args/<service>`.

### Investigate system pods' health

Check whether all of the cluster's pods are `Running` and `Ready`:

```
sudo k8s kubectl get pods -n kube-system
```

The pods in the `kube-system` namespace belong to {{product}} features such as
`network`. Unhealthy pods could be related to configuration issues or nodes not
meeting certain requirements.

### Troubleshoot a failing pod

Look at the events on a failing pod by running:

```
sudo k8s kubectl describe pod <pod-name> -n <namespace>
```

Check the logs on a failing pod by running the following command:

```
sudo k8s kubectl logs <pod-name> -n <namespace>
```

You can check out the upstream [debug pods documentation][] for more
information.

## Common issues and solutions

If you find any issue while working with {{product}} it is highly
likely that someone from the community has already faced the same problem. We
have documented some common issues users face and their workarounds.

### Kubectl error: `dial tcp 127.0.0.1:6443: connect: connection refused`

The [kubeconfig file] generated by the `k8s kubectl` CLI cannot be used to
access the cluster from an external machine. The following error is seen when
running `kubectl` with the invalid kubeconfig:

```
...
E0412 08:36:06.404499  517166 memcache.go:265] couldn't get current server API group list: Get "https://127.0.0.1:6443/api?timeout=32s": dial tcp 127.0.0.1:6443: connect: connection refused
The connection to the server 127.0.0.1:6443 was refused - did you specify the right host or port?
```

````{dropdown} Explanation

A common technique for viewing a cluster kubeconfig file is by using the
`kubectl config view` command.

The `k8s kubectl` command invokes an integrated `kubectl` client. Thus
`k8s kubectl config view` will output a seemingly valid kubeconfig file.
However, this will only be valid on cluster nodes where control plane services
are available on localhost endpoints.

````

````{dropdown} Solution

Use `k8s config` instead of `k8s kubectl config` to generate a kubeconfig file
that is valid for use on external machines.

````

### Kubelet Error: `failed to initialize top level QOS containers`

This is related to the `kubepods` cgroup not getting the cpuset controller up on
the kubelet. kubelet needs a feature from cgroup and the kernel may not be set
up appropriately to provide the cpuset feature.

```
E0125 00:20:56.003890    2172 kubelet.go:1466] "Failed to start ContainerManager" err="failed to initialise top level QOS containers: root container [kubepods] doesn't exist"
```

````{dropdown} Explanation

An excellent deep-dive of the issue exists at
[kubernetes/kubernetes #122955][kubernetes-122955].

Commenter [@haircommander][] [states][kubernetes-122955-2020403422]
>  basically: we've figured out that this issue happens because libcontainer
>  doesn't initialise the cpuset cgroup for the kubepods slice when the kubelet
>  initially calls into it to do so. This happens because there isn't a cpuset
>  defined on the top level of the cgroup. however, we fail to validate all of
>  the cgroup controllers we need are present. It's possible this is a
>  limitation in the dbus API: how do you ask systemd to create a cgroup that
>  is effectively empty?

>  if we delegate: we are telling systemd to leave our cgroups alone, and not
>  remove the "unneeded" cpuset cgroup.

````

````{dropdown} Solution

This is in the process of being fixed upstream via
[kubernetes/kubernetes #125923][kubernetes-125923].

In the meantime, the best solution is to create a `Delegate=yes` configuration
in systemd.

```bash
mkdir -p /etc/systemd/system/snap.k8s.kubelet.service.d
cat /etc/systemd/system/snap.k8s.kubelet.service.d/delegate.conf <<EOF
[Service]
Delegate=yes
EOF
reboot
```

````

### The path required for the containerd socket already exists

{{product}} tries to create the containerd socket to manage containers,
but it fails because the socket file already exists, which indicates another
installation of containerd on the system.

````{dropdown} Explanation

In classic confinement mode, {{product}} uses the default containerd
paths. This means that a {{product}} installation will conflict with
any existing system configuration where containerd is already installed.
For example, if you have Docker installed, or another Kubernetes distribution
that uses containerd.

````

````{dropdown} Solution

We recommend running {{product}} in an isolated environment, for this purpose,
you can create a LXD VM for your installation. See
[Install {{product}} in LXD][lxd-install] for instructions.

As an alternative, you may specify a custom containerd path like so:

```bash
cat <<EOF | sudo k8s bootstrap --file -
containerd-base-dir: $containerdBaseDir
EOF
```

````

### High disk usage for log files

When using {{product}} for a longer period of time, the disk usage for
log files can grow significantly.

````{dropdown} Solution

Check your system's journald configuration to limit log retention and disk usage
by looking at the following parameters in `/etc/systemd/journald.conf`:

- `SystemMaxUse` - Maximum disk space for all journal files
- `SystemMaxFileSize` - Maximum size per journal file
- `MaxRetentionSec` - Delete logs older than this time period
- `ForwardToSyslog` - Prevent double logging to syslog and journalctl

We recommend setting the following parameters either by editing the file directly
or by creating an override file:

```bash
# Limit disk usage
SystemMaxUse=500M
SystemMaxFileSize=100M
SystemMaxFiles=100
MaxRetentionSec=1week

# Don't forward to syslog (prevents double-logging)
ForwardToSyslog=no
```

After making changes, restart the journald service:

```
sudo systemctl restart systemd-journald
```

````

### Bootstrap issues on a host with custom routing policy rules

{{product}} bootstrap process might fail or face networking issues when
custom routing policy rules are defined, such as rules in a Netplan file.

````{dropdown} Explanation

Cilium, which is the current implementation for the `network` feature,
introduces and adjusts certain IP rules with
hard-coded priorities of `0` and `100`.

Adding IP rules with a priority lower than or equal to `100` might introduce
conflicts and cause networking issues.

````

````{dropdown} Solution

Adjust the custom defined `ip rule` to have a
priority value that is greater than `100`.

````

### Cilium pod fails to start as `cilium_vxlan: address already in use`

When deploying {{product}} the Cilium pods fail to start and reports the error:

```
failed to start: daemon creation failed: error while initializing daemon: failed
while reinitializing datapath: failed to setup vxlan tunnel device: setting up
vxlan device: creating vxlan device: setting up device cilium_vxlan: address
already in use
```

````{dropdown} Explanation

Fan networking is automatically enabled in some substrates. This causes
conflicts with some CNIs such as Cilium. This conflict of
`address already in use` causes Cilium to be unable to set up its VXLAN
tunneling network. There may also be other networking components on the system
attempting to use the default port for their own VXLAN interface that will
cause the same error.

````

````{dropdown} Solution

Configure Cilium to use another tunnel port. Set the annotation `tunnel-port`
to an appropriate value (the default is 8472).

```
sudo k8s set annotation="k8sd/v1alpha1/cilium/tunnel-port=<PORT-NUMBER>"
```

Since the Cilium pods are in a failing state, the recreation of the VXLAN
interface is automatically triggered. Verify the VXLAN interface has come
up:

```
ip link list type vxlan
```

It should be named `cilium_vxlan` or something similar.

Verify that Cilium is now in a running state:

```
sudo k8s kubectl get pods -n kube-system
```

````

<!-- markdownlint-disable -->

### Cilium pod `unable to determine direct routing device`

<!-- markdownlint-restore -->

When deploying {{product}}, the Cilium pods fail to start and reports
the error:

```
level=error msg="Start failed" error="daemon creation failed: unable to determine direct routing device. Use --direct-routing-device to specify it"
```

````{dropdown} Explanation
This issue was introduced in Cilium 1.15 and has been [reported here]. Both
`devices` and `direct-routing-device` lists must now be set in direct routing
mode. Direct routing mode is used by BPF, NodePort and BPF host routing.

If `direct-routing-device` is left undefined, it is automatically set to the
device with the k8s InternalIP/ExternalIP or the device with a default route.
However, bridge type devices are ignored in this automatic selection. In this
case, a bridge interface is used as the default route and
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

In this example of deploying {{product}}, the output is as follows:

```
Kernel IP routing table
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
default         _gateway        0.0.0.0         UG    0      0        0 br-ex
172.27.20.0     0.0.0.0         255.255.254.0   U     0      0        0 br-ex
```

The `br-ex` interface is the default interface used for this cluster. Apply
the annotation to the node adding bridge interfaces `br+` to the `devices` list:

```
sudo k8s set annotations="k8sd/v1alpha1/cilium/devices=br+"
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
````

(remove-permanently-lost-node)=

### Remove a permanently lost node from the cluster

A node that is permanently lost cannot be removed with `k8s remove-node`:

```sh
Error: Failed to remove node "t1" from the cluster.

The error was: failed after potential retry: wait check failed: failed to POST /k8sd/cluster/remove:
failed to delete cluster member t1: Post "https://10.23.245.80:6400/core/internal/hooks/pre-remove?target=t1":
Unable to connect to "10.23.245.80:6400": dial tcp 10.23.245.80:6400: connect: no route to host
```

````{dropdown} Explanation
By default, `k8s remove-node` attempts to contact the node being removed
to execute cleanup routines.
If the node is unreachable, the command fails with the error shown above.
This can also happen if the node membership tracked by `k8sd` becomes
inconsistent with the Kubernetes datastore.

````

````{dropdown} Solution

Use the `--force` flag to forcibly remove the node:

```
sudo k8s remove-node --force <node-name>
```
````

## Use the built-in inspection command

{{product}} ships with a command to compile a complete report on {{product}} and
its underlying system. This is an essential tool for bug reports and for
investigating whether a system is (or isn’t) working.

Run the inspection command, by entering the command (admin privileges are
required to collect all the data):

```
sudo k8s inspect
```

See the [inspection report reference page] for more details.

## Report a bug

If you cannot solve your issue and believe that the fault may lie in
{{product}}, please [file an issue on the project repository][].

Help us deal effectively with issues by including the report obtained from the
inspect script, any additional logs, and a summary of the issue.

You can check out the upstream [debug documentation][] for more details on
troubleshooting a Kubernetes cluster.

<!-- Links -->

[file an issue on the project repository]: https://github.com/canonical/k8s-snap/issues/new/choose
[debug pods documentation]: https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods
[debug documentation]: https://kubernetes.io/docs/tasks/debug
[inspection report reference page]: /snap/reference/inspection-reports.md
[kubeconfig file]: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
[kubernetes-122955]: https://github.com/kubernetes/kubernetes/issues/122955
[kubernetes-125923]: https://github.com/kubernetes/kubernetes/pull/125923
[kubernetes-122955-2020403422]: https://github.com/kubernetes/kubernetes/issues/122955#issuecomment-2020403422
[@haircommander]: https://github.com/haircommander
[lxd-install]: /snap/howto/install/lxd.md
[reported here]: https://github.com/cilium/cilium/issues/30889