# Add and remove nodes

Typical production clusters are hosted across multiple data centres and cloud
environments, enabling them to leverage geographical distribution for improved
availability and resilience.

This tutorial simplifies the concept by creating a cluster within a controlled
environment using two Multipass VMs. The approach here allows us to focus on
the foundational aspects of clustering using {{product}} without the
complexities of a full-scale, production setup. If your nodes are already
installed, you can skip the Multipass setup and go to [step 2](step2).

## Before starting

In this tutorial, "**control plane**" refers to the Multipass VM that operates
the control plane, while "**worker**" denotes the Multipass VM running the
worker node.

## What you will need

- Multipass (See [Multipass Installation][Multipass Installation])

### 1. Create both VMs

The first step is creating the VMs.

```
multipass launch 22.04 --name control-plane -m 4G -d 8G
```

```
multipass launch 22.04 --name worker -m 4G -c 4 -d 8G
```

This step can take a few minutes as Multipass creates the new virtual machines.
It's normal and expected.

Once the virtual machine has been created, you can run commands on it by
opening a shell. For example:

```
multipass shell control-plane
```

This will behave as a local terminal session on the virtual machine, so you can
run commands.

Install {{product}} on both VMs with the following command:

```{literalinclude} ../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

<!-- markdownlint-capture -->
<!-- markdownlint-disable -->
(step2)=
### 2. Bootstrap your control plane node

<!-- markdownlint-restore -->
Bootstrap the control plane node with default configuration:

```
sudo k8s bootstrap
```

{{product}} allows you to create two types of nodes: control plane and
worker nodes. In this example, we just initialised a control plane node, now
let's create a worker node.

Generate the token required for the worker node to join the cluster by executing
the following command on the control-plane node:

```
sudo k8s get-join-token worker --worker
```

`worker` refers to the name of the node we want to join. `--worker` is the type
of node we want to join.

A base64 token will be printed to your terminal. Keep it handy as you will need
it for the next step.

```{note} It's advisable to name the new node after the hostname of the
   worker node (in this case, the VM hostname is worker).
```

### 3. Join the cluster on the worker node

To join the worker node to the cluster, run on worker node:

```
sudo k8s join-cluster <join-token>
```

After a few seconds, you should see: `Joined the cluster.`

### 4. View the status of your cluster

Let's review what we've accomplished in this tutorial.

To see the control plane node created:

```
sudo k8s status
```

Verify the worker node joined successfully with this command
on control-plane node:

```
sudo k8s kubectl get nodes
```

You should see that you've successfully added a worker to your cluster.

Congratulations!

### 4. Remove nodes and delete the VMs (Optional)

It is important to clean-up your nodes before tearing down the VMs.

```{note}  Purging a VM does not remove the node from your cluster.
```

Keep in mind the consequences of removing nodes:

```{warning} Do not remove the leader node.
If you have less than 3 nodes and you remove any node you will lose availability
   of your cluster.
```

To tear down the entire cluster, execute:

```
sudo k8s remove-node worker
sudo k8s remove-node control-plane
```

To delete the VMs from your system use the following commands:

```
multipass delete control-plane
multipass delete worker
multipass purge
```

## Next steps

- Discover how to enable and configure Ingress resources [Ingress][Ingress]
- Learn more about {{product}} with kubectl [How to use
  kubectl][Kubectl]
- Explore Kubernetes commands with our [Command Reference
  Guide][Command Reference]
- Configure storage options [Storage][Storage]
- Master Kubernetes networking concepts [Networking][Networking]

<!-- LINKS -->

[Getting started]: getting-started
[Multipass Installation]: https://multipass.run/install
[Ingress]: ../howto/networking/default-ingress
[Kubectl]: kubectl
[Command Reference]: ../reference/commands
[Storage]: ../howto/storage/index
[Networking]: ../howto/networking/index.md
