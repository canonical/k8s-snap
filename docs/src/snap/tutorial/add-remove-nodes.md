# Adding and Removing Nodes

Typical production clusters are hosted across multiple data centres and cloud
environments, enabling them to leverage geographical distribution for improved
availability and resilience.

This tutorial simplifies the concept by creating a cluster within a controlled
environment using two Multipass VMs. The approach here allows us to focus on
the foundational aspects of clustering using Canonical Kubernetes without the
complexities of a full-scale, production setup. If your nodes are already
installed, you can skip the multipass setup and go to [step 2](step2).

## Before starting

In this article, "**control plane**" refers to the Multipass VM that operates
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

Install Canonical Kubernetes on both VMs with the following command:

```
sudo snap install --classic --edge k8s
```

(step2)=
### 2. Bootstrap your control plane node

Bootstrap the control plane node:

```
sudo k8s bootstrap
```

Canonical Kubernetes allows you to create two types of nodes: control plane and
worker nodes. In this example, we're creating a worker node.

Generate the token required for the worker node to join the cluster by executing
the following command on the control-plane node:

```
sudo k8s get-join-token worker --worker
```

A base64 token will be printed to your terminal. Keep it handy as you will need
it for the next step.

```{note} It's advisable to name the new node after the hostname of the
   worker node (in this case, the VM's hostname is worker).
```

### 3. Join the cluster on the worker node

To join the worker node to the cluster, run:

```
sudo k8s join-cluster <join-token>
```

After a few seconds, you should see: `Joined the cluster.`

### 4. View the status of your cluster

To see what we've accomplished in this tutorial:

If you created a control plane node, check that it joined successfully:

```
sudo k8s status
```

If you created a worker node, verify with this command:

```
sudo k8s kubectl get nodes
```

You should see that you've successfully added a worker or control plane node to
your cluster.

Congratulations!

### 4. Remove Nodes and delete the VMs (Optional)

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

To delete the VMs from your system, two commands are needed:

```
multipass remove control-plane
multipass remove worker
multipass purge
```

## Next Steps

- Discover how to enable and configure Ingress resources [Ingress][Ingress]
- Keep mastering Canonical Kubernetes with kubectl [How to use
  kubectl][Kubectl]
- Explore Kubernetes commands with our [Command Reference
  Guide][Command Reference]
- Configure storage options [Storage][Storage]
- Master Kubernetes networking concepts [Networking][Networking]

<!-- LINKS -->

[Getting started]: getting-started
[Multipass Installation]: https://multipass.run/install
[Ingress]: /snap/howto/networking/default-ingress
[Kubectl]: kubectl
[Command Reference]: /snap/reference/commands
[Storage]: /snap/howto/storage
[Networking]: /snap/howto/networking/index.md
