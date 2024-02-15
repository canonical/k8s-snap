# Adding and Removing Nodes

Typical production clusters are hosted across multiple data centers and cloud
environments, enabling them to leverage geographical distribution for improved
availability and resilience.

This tutorial simplifies the concept by creating a cluster within a controlled
environment using two Multipass VMs. The approach here allows us to focus on
the foundational aspects of clustering using Canonical Kubernetes without the
complexities of a full-scale, production setup.

## Before starting

In this article, "**control plane**" refers to the Multipass VM that operates the control plane, while "**worker**" denotes the Multipass VM running the worker node.

## What you will need

- Multipass (See [Multipass Installation][Multipass Installation])

### 1. Create both VMs

The first step is to create the VMs.

```sh
multipass launch 22.04 --name control-plane
```

```sh
multipass launch 22.04 --name worker
```

This step can take a few minutes as Multipass creates the new virtual machines. It's normal and expected.

On both VMs, run the following command to install Canonical Kubernetes:

```sh
sudo snap install --classic --edge k8s
```

### 2. Bootstrap your control plane node

Let's bootstrap our control plane node:

```sh
sudo k8s bootstrap
```

Next, enable two components that are needed for communication between nodes.

```sh
sudo k8s enable dns
sudo k8s enable network
```

Canonical Kubernetes allows you to create two types of nodes: control plane and
worker nodes. In this example, we'll be creating a worker node.

Create the token that will be used by the worker node to join the cluster.

```sh
sudo k8s add-node worker --worker
```

A base64 token should be printed to your terminal. Keep it close as you will
need it for the next step.

> **Note**: It's best to give the new node the same name as the hostname of the
> worker node (in this case the VM's hostname is worker).

### 3. Join the cluster on the worker node

Run the following command to join the worker node to the cluster:

```sh
sudo k8s join-cluster <token>
```

After a couple of seconds, you should see: `Joined the cluster.`

### 4. View the status of your cluster

Let's see what we've achieved during this tutorial.

If you created a control plane node, you can check that it joined successfully by running:

```sh
sudo k8s status
```

If you created a worker node, you can check with this command:

```sh
sudo k8s kubectl get nodes
```

You should see that you've successfully added a worker or control plane node to
your cluster.

Congratulations!

### 4. Delete the VMs (Optional)

Two commands are needed to delete the expert-elf VM from your system:

```sh
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
- Bootstrap Kubernetes with your custom configurations [Bootstrap K8s][Bootstrap K8s]
- Learn how to set up a multi-node environment [Setting up a K8s
  cluster][Setting up K8s]
- Configure storage options [Storage][Storage]
- Master Kubernetes networking concepts [Networking][Networking]

<!-- LINKS -->

[Getting started]: getting-started.md
[Multipass Installation]: https://multipass.run/install
[Ingress]: #TODO
[Kubectl]: #TODO
[Command Reference]: #TODO
[Bootstrap K8s]: #TODO
[Setting up K8s]: #TODO
[Storage]: #TODO
[Networking]: #TODO
