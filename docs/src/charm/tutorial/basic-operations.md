# Basic operations with the charm

This tutorial walks you through common management tasks for your {{ product }}
cluster using the `k8s` charm. You'll learn how to scale your cluster, manage
workers, and interact with it using `kubectl`.

## What you will need

- A running {{ product }} cluster deployed with the `k8s` charm
- The Juju [client][Juju client]
- [Kubectl] (installation instructions included below)

## Scaling Your Cluster

The `k8s` charm provides flexibility to scale your cluster as needed by adding
or removing control plane nodes or worker nodes.

To increase the control plane's capacity or ensure high availability, you can
add more units to the `k8s` application:

```
juju add-unit k8s -n 2
```

```{tip}
For high availability, we recommend deploying at least three `k8s` charm units.
```

Similarly, you can add more worker nodes when your workload demands increase:

```
juju add-unit k8s-worker -n 1
```

This command deploys an additional instance of the `k8s-worker` charm. Juju
manages all instances within the same application, so no extra configuration
is needed. After running this command, new units will appear in your cluster,
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

[kubectl][] is the standard upstream tool for interacting with a Kubernetes
cluster. This is the command that can be used to inspect and manage your
cluster.

If you don't already have `kubectl`, it can be installed from a snap:

```
sudo snap install kubectl --classic
```

If you have just installed it, you should also create a file to contain the configuration:

```
mkdir ~/.kube
```

To fetch the configuration information from the cluster we can run:

```
juju run k8s/0 get-kubeconfig
```

The Juju action is a piece of code which runs on a unit to perform a specific
task. In this case it collects the cluster information - the YAML formatted
details of the cluster and the keys required to connect to it.

```{warning}
If you already have `kubectl` and are using it to manage other clusters, you
should edit the relevant parts of the cluster yaml output and append them to
your current kubeconfig file.
```

We can use pipe to append your cluster's kubeconfig information directly to a
config file which will just require a bit of editing:

```
juju run k8s/0 get-kubeconfig >> ~/.kube/config
```

The output includes the root of the YAML, `kubeconfig: |`, so we can just use an
editor to remove that line:

```
vim ~/.kube/config
```

Please use the editor of your choice to delete the first line and save the file.

Alternatively, if you are a `yq` user, the same can be achieved with:

```
juju run k8s/0 get-kubeconfig | yq '.kubeconfig' -r >> ~/.kube/config
```

You can now confirm Kubectl can read the kubeconfig file:

```
kubectl config show
```

...which should output something like this:

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

You can then further confirm that it is possible to inspect the cluster by
running a simple command such as :

```
kubectl get pods -A
```

This should return some pods, confirming the command can reach the cluster.

## Next steps

Now that you're familiar with basic cluster operations, you might want to:

- Deploy applications to your cluster
- Configure storage solutions like [Ceph]
- Set up monitoring and observability with [Canonical Observability Stack][COS]

For more advanced operations and updates, keep an eye on the charm's
documentation and release [notes][release notes].

<!-- LINKS -->

[Ceph]: ../howto/ceph-csi
[COS]: ../howto/cos-lite
[Juju client]: https://juju.is/docs/juju/install-and-manage-the-client
[Kubectl]: https://kubernetes.io/docs/reference/kubectl/
[release notes]: ../reference/releases
