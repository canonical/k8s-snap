# Adding and Removing Nodes

Typical production clusters are hosted across multiple data centers and cloud
environments, enabling them to leverage geographical distribution for improved
availability and resilience.

This tutorial simplifies the concept by creating a cluster within a controlled
environment using two Multipass VMs. The approach here allows us to focus on
the foundational aspects of clustering using Canonical Kubernetes without the
complexities of a full-scale, production setup.

## What you will need

- Multipass (See [Multipass Installation][Multipass Installation])
- A bootstrapped cluster using Canonical K8s (See the [Getting
  Started][Getting Started] tutorial)
- The `dns` and `network` components enabled on your cluster.

### 1. Verify the Status Of Your Cluster

Verify the status of your cluster with:

```sh
sudo k8s status
```

Make sure you can see `k8s is ready.` in the output of the command.

### 2. Start a Multipass VM

Our second node will live on Ubuntu 22.04 on a UMultipass VM. Let's launch it:

```sh
multipass launch 22.04 --name expert-elf
```

This step can take a few minutes as Multipass creates the new virtual machine. It's normal and expected.

### 3. Generate a worker node token from the control plane node

Canonical Kubernetes allows you to create two types of nodes: control plane and
worker nodes. In this example, we'll be creating a worker node.

Let's get the token that will allow our worker node to join our cluster.

Run this command from your control plane node:

```sh
sudo k8s add-node expert-elf --worker
```

**Note**: It's best to give the new node the same name as the hostname of the
worker node (in this case the VM).

A base64 string should be printed to your terminal. Keep it close as you will
need it in a few steps.

### 4. Join the cluster from the other node

Open a shell on the worker node:

```sh
multipass shell expert-elf
```

The VM doesn't come with the `k8s` snap, so let's install it, and make sure
the new node uses the same Canonical Kubernetes version as the existing node.

```sh
sudo snap install --edge --classic k8s
```

From the worker node machine, use the `join-cluster` command and pass your
token from step 3 as the last argument.

```sh
sudo k8s join-cluster eyJu...XX0=
```

After a couple of seconds, you should see: `Joined the cluster.`

Disconnect from the VM and run the `status` command on your control plane node:

```sh
sudo k8s status
```

You should see that you've successfully added a worker or control plane node to
your cluster.

Congratulations!

### 5. Delete the expert-elf VM (Optional)

Two commands are needed to delete the expert-elf VM from your system:

```sh
multipass remove expert-elf
multipass purge
```

## Next Steps

- Discover how to enable and configure Ingress resources [Ingress][Ingress]
- Keep mastering Canonical Kubernetes with kubectl [How to use
  kubectl][Kubectl]
- Explore Kubernetes commands with our [Command Reference
  Guide][Command Reference]
- Bootstrap K8s with your custom configurations [Bootstrap K8s][Bootstrap K8s]
- Learn how to set up a multi-node environment [Setting up a K8s
  cluster][Setting up K8s]
- Configure storage options [Storage][Storage]
- Master Kubernetes networking concepts [Networking][Networking]
- Discover how to enable and configure Ingress resources again
  [Ingress][Ingress]

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
