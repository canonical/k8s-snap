# Basic {{ product }} charm operations

This tutorial walks you through common management tasks for your {{ product }}
cluster using the `k8s` control plane charm. You will learn how to scale your
cluster, manage workers, and interact with the cluster using `kubectl`.

## Prerequisites

- A running {{ product }} cluster deployed with the `k8s` charm
- The Juju [client][Juju client]
- [Kubectl] (installation instructions included below)

## Scaling Your Cluster

The `k8s` charm provides flexibility to scale your cluster as needed by adding
or removing control plane nodes or worker nodes.

To increase the control plane's capacity or ensure [high availability], you
can add more units of the `k8s` application:

```
juju add-unit k8s -n 1
```

Use `juju status` to view all the units in your cluster and monitor their
status.

Similarly, you can add more worker nodes when your workload demands increase:

```
juju add-unit k8s-worker -n 1
```

This command deploys an additional instance of the `k8s-worker` charm. No extra
configuration is needed as Juju manages all instances within the same
application. After running this command, new units will appear in your cluster,
such as `k8s-worker/0` and `k8s-worker/1`.

To scale up multiple units at once, adjust the unit count:

```
juju add-unit k8s-worker -n 3
```

If you need to scale down the cluster, you can remove units as follows:

```
juju remove-unit k8s-worker/1
```

Replace the unit name with the appropriate application name (e.g., `k8s` or
`k8s-worker`) and unit number.


## Set up `kubectl`

[kubectl] is the standard upstream tool for interacting with a Kubernetes
cluster. This is the command that can be used to inspect and manage your
cluster.

If necessary, `kubectl` can be installed from a snap:

```
sudo snap install kubectl --classic
```

Create a directory to house the kubeconfig:

```
mkdir ~/.kube
```

Fetch the configuration information from the cluster:

```
juju run k8s/0 get-kubeconfig
```

The Juju action is a piece of code which runs on a unit to perform a specific
task. In this case it collects the cluster information - the YAML formatted
details of the cluster and the keys required to connect to it.

```{warning}
If you already have `kubectl` installed and are using it to manage other
clusters, edit the relevant parts of the cluster yaml output and append them to
your current kubeconfig file.
```

Use `yq` to append your cluster's kubeconfig information directly to the
config file:

```
juju run k8s/0 get-kubeconfig | yq '.kubeconfig' >> ~/.kube/config
```

Confirm that `kubectl` can read the kubeconfig file:

```
kubectl config show
```

The output will be similar to this:

```
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: DATA+OMITTED
    server: https://10.158.52.236:6443
  name: k8s
contexts:
- context:
    cluster: k8s
    user: k8s-user
  name: k8s
current-context: k8s
kind: Config
preferences: {}
users:
- name: k8s-user
  user:
    token: REDACTED
```

Run a simple command to inspect your cluster:

```
kubectl get pods -A
```

This command returns a list of pods, confirming that `kubectl` can reach the
cluster.

## Next steps

Now that you are familiar with the basic cluster operations, learn to:

- Deploy applications to your cluster
- Configure storage solutions like [Ceph]
- Set up monitoring and observability with [Canonical Observability Stack][COS]

For more advanced operations and updates, keep an eye on the charm's
documentation and release [notes][release notes].

<!-- LINKS -->

[Ceph]: ../howto/ceph-csi
[COS]: ../howto/cos-lite
[high availability]: ../../snap/explanation/high-availability
[Juju client]: https://juju.is/docs/juju/install-and-manage-the-client
[Kubectl]: https://kubernetes.io/docs/reference/kubectl/
[release notes]: ../reference/releases
