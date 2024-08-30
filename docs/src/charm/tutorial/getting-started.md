# Getting started

The {{product}} `k8s` charm takes care of installing and configuring
Kubernetes on cloud instances managed by Juju. Operating Kubernetes through
this charm makes it significantly easier to manage at scale, on remote cloud
instances and also to integrate other operators to enhance or customise your
Kubernetes deployment. This tutorial will take you through installing
Kubernetes and some common first steps.

## What you will learn

- How to install {{product}}
- Making a cluster
- Deploying extra workers
- Using Kubectl

## What you will need

- Ubuntu 22.04 LTS or 20.04 LTS
- The [Juju client][]
- Access to a Juju-supported cloud for creating the required instances
- [Kubectl] for interacting with the cluster (installation instructions are
  included in this tutorial)


## 1. Get prepared

Deploying charms with Juju requires a substrate or backing cloud to actually
run the instances. If you are unfamiliar with Juju, it would be useful to run
through the [Juju tutorial] first, and ensure you have a usable controller to
deploy with.

Before installing anything, we should first check what versions of the charm
are available. Charms are published to 'channels' which reflect both a specific
release version and the maturity or stability of that code. Sometimes we may
wish to run on the latest stable version, sometimes the goal is to test out
upcoming features or test migration to a new version. Channels are covered in
more detail in [the channel explanation page] if you want to learn more.
The currently available versions of the charm can be discovered by running:

```
juju info k8s
```
or
```
juju info k8s-worker
```

There are two distinct charms - one includes control-plane services for
administering the cluster, the other omits these for nodes which are to be
deployed purely for workloads. Both are published simultaneously from the same
source so the available channels should match. Running the commands will output
basic information about the charm, including a list of the available channels
at the end.

Charm deployments default to "latest/stable", but if you want to chose a
specific version it can be indicated when deploying with the `--channel=`
argument, for example `--channel=latest/edge`.

## 2. Deploy the K8s charm

To make sure that Juju creates an instance which has enough resources to
actually run Kubernetes, we will make use of 'constraints'. These specify the
minimums required. For the Kubernetes control plane (`k8s` charm), the
recommendation is two CPU cores, 16GB of memory and 40GB of disk space. Now we
can go ahead and create a cluster:

```
juju deploy k8s --channel=latest/edge --constraints='cores=2 mem=16G root-disk=40G'
```

At this point Juju will fetch the charm from Charmhub, create a new instance
according to your specification and configure and install the Kubernetes
components (i.e. the `k8s` snap ). This may take a few minutes depending on
your cloud. You can monitor progress by watching the Juju status output:

```
juju status --watch 2s
```

When the status reports that K8s is "idle/ready" you have successfully deployed
a {{product}} control-plane using Juju.

```{note} For High Availability you will need at least three units of the k8s 
   charm. Scaling the deployment is covered below.
```

## 3. Deploy a worker

Before we start doing things in Kubernetes, we should consider adding a worker.
The K8s worker is an additional node for the cluster which focuses on running
workloads without running any control-plane services. This means it needs a
connection to a control-plane node to tell it what to do, but it also means
more of its resources are available for running workloads. We can deploy a
worker node in a similar way to the original K8s node:

```
juju deploy k8s-worker --channel=latest/edge --constraints='cores=2 mem=16G root-disk=40G'
```

Once again, this will take a few minutes. In this case though, the `k8s-worker`
application won't settle into a 'Ready' status, because it requires a
connection to the control plane. This is handled in Juju by integrating the
charms so they can communicate using a standard interface. The charm info we
fetched earlier also includes a list of the relations possible, and from this
we can see that the k8s-worker requires "cluster: k8s-cluster".

To connect these charms and effectively add the worker to our cluster, we use
the 'integrate' command, adding the interface we wish to connect

```
juju integrate k8s k8s-worker:cluster
```

After a short time, the worker node will share information with the control plane and be joined to the cluster.

## 4. Scale the cluster (Optional)

If one worker doesn't seem like enough, we can easily add more:

```
juju add-unit k8s-worker -n 1
```

This will create an additional instance running the k8s-worker charm. Juju
manages all instances of the same application together, so there is no need to
add the integration again. If you check the Juju status, you should see that a
new unit is created, and now you have `k8s-worker/0` and `k8s-worker/1`


## 5. Use Kubectl

[Kubectl][] is the standard upstream tool for interacting with a Kubernetes
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

```{warning}  If you already have Kubectl and are using it to manage other clusters,
   you should edit the relevant parts of the cluster yaml output and append them to
   your current kubeconfig file.
```

We can use pipe to append your cluster's kubeconfig information directly to a
config file which will just require a bit of editing:

```
juju run k8s/0 get-kubeconfig >> ~/.kube/config
```

The output includes the root of the YAML, `kubeconfig: |`, so we can just use an editor to remove that line:

```
nano ~/.kube/config
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

This should return some pods, confirming the command can reach the cluster:

```
NAMESPACE     NAME                               READY   STATUS    RESTARTS   AGE
kube-system   cilium-4m5xj                       1/1     Running   0          35m
kube-system   cilium-operator-5ff9ddcfdb-b6qxm   1/1     Running   0          35m
kube-system   coredns-7d4dffcffd-tvs6v           1/1     Running   0          35m
kube-system   metrics-server-6f66c6cc48-wdxxk    1/1     Running   0          35m
```

## Next steps

Congratulations - you now have a functional Kubernetes cluster! In the near
future more charms are on the way to simplify usage and extend the base
functionality of {{product}}. Bookmark the [releases page] to keep
informed of updates.

<!-- LINKS -->

[Juju client]: https://juju.is/docs/juju/install-and-manage-the-client
[Juju tutorial]: https://juju.is/docs/juju/tutorial
[Kubectl]: https://kubernetes.io/docs/reference/kubectl/
[the channel explanation page]: /snap/explanation/channels
[releases page]: /charm/reference/releases