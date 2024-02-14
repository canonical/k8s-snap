# Adding and Removing Nodes

## What you will need

- Multipass (See [Multipass Installation](https://multipass.run/install))
- A bootstrapped cluster using Canonical K8s (See (getting-started.md)[Getting
  Started]
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
multipass launch --name expert-elf
```

> **Note**: This step can take a few minutes. It's normal and expected.

### 3. Generate a worker node token from the control plane node

> **Note:** Make sure your cluster is bootstrapped. See
> [Getting Started](getting-started.md) for instructions.

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

Open a shell to the worker node:

```sh
multipass shell expert-elf
```

The VM doesn't come with the `k8s` snap, so let's install it.

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

### 6. Remove Canonical Kubernetes (Optional)

To uninstall the Canonical Kubernetes snap, execute:

```sh
sudo snap remove k8s
```

This command removes the `k8s` snap and automatically creates a snapshot of all
data for future restoration.

If you wish to remove the snap without saving a snapshot of its data, add
`--purge` to the command:

```sh
sudo snap remove k8s --purge
```

This option ensures complete removal of the snap and its associated data.

## Next Steps

- Keep mastering Canonical Kubernetes with kubectl: [How to use kubectl](#TODO)
- Explore Kubernetes commands with our [Command Reference Guide](#TODO)
- Bootstrap K8s with your custom configurations [Bootstrap K8s](#TODO)
- Learn how to set up a multi-node environment
  [Setting up a K8s cluster](#TODO)
- Configure storage options [Storage](#TODO)
- Master Kubernetes networking concepts: [Networking](#TODO)
- Discover how to enable and configure Ingress resources [Ingress](#TODO)
