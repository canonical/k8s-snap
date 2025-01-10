# Getting started

The {{product}} `k8s` charm takes care of installing and configuring
Kubernetes on cloud instances managed by Juju. Operating Kubernetes through
this charm makes it significantly easier to manage at scale, on remote cloud
instances and also to integrate other operators to enhance or customise your
Kubernetes deployment. This tutorial will take you through installing
Kubernetes and some common first steps.

## What will be covered

- How to install {{product}}
- Making a cluster
- Deploying extra workers

## What you will need

- The [Juju client][]
- Access to a Juju-supported cloud for creating the required instances


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

```{literalinclude} ../../_parts/install.md
:start-after: <!-- juju control constraints start -->
:end-before: <!-- juju control constraints end -->
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

## 3. Deploy a worker

Before we start doing things in Kubernetes, we should consider adding a worker.
The K8s worker is an additional node for the cluster which focuses on running
workloads without running any control-plane services. This means it needs a
connection to a control-plane node to tell it what to do, but it also means
more of its resources are available for running workloads. We can deploy a
worker node in a similar way to the original K8s node:

```{literalinclude} ../../_parts/install.md
:start-after: <!-- juju worker constraints start -->
:end-before: <!-- juju worker constraints end -->
```

Once again, this will take a few minutes. In this case though, the `k8s-worker`
application won't settle into a 'Ready' status, because it requires a
connection to the control plane. This is handled in Juju by integrating the
charms so they can communicate using a standard interface. The charm info we
fetched earlier also includes a list of the relations possible, and from this
we can see that the k8s-worker requires "cluster: k8s-cluster".

To connect these charms and effectively add the worker to our cluster, we use
the `integrate` command, adding the interface we wish to connect.

```
juju integrate k8s k8s-worker:cluster
juju integrate k8s k8s-worker:containerd
juju integrate k8s k8s-worker:cos-tokens
```

After a short time, the worker node will share information with the control plane
and be joined to the cluster.

## Next steps

Congratulations — you now have a functional {{ product }} cluster! You can
start exploring the [basic operations] with the charm. In the near future more
charms are on the way to simplify usage and extend the base functionality of
{{product}}. Bookmark the [releases page] to keep informed of updates.

<!-- LINKS -->

[Juju client]: https://juju.is/docs/juju/install-and-manage-the-client
[Juju tutorial]: https://juju.is/docs/juju/tutorial
[the channel explanation page]: ../../snap/explanation/channels
[releases page]: ../reference/releases
[basic operations]: ./basic-operations.md
