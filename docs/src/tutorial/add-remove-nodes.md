# Adding and Removing Nodes

## What you will need
- A bootstrapped cluster using Canonical K8s (See (getting-started.md)[Getting Started]
- The `dns` and `network` components enabled on your cluster.

### Verify the Status Of Your Cluster

Verify the status of your cluster with:

```
sudo k8s status
```

Make sure you can see `k8s is ready.` in the output of the command.

### Start a Multipass VM

Our second node will live on a Multipass VM. Let's launch it:

```
multipass launch --name expert-elf
```

**Note**: This step can take a few minutes. It's normal and expected.

### Generate a worker node token from the control plane node

In the shell connected to your control plane node, run the following command to get a token that'll be used by the worker node to join the cluster.

```
sudo k8s add-node expert-elf
```

**Note**: It's best to give the new node the same name as the hostname of the worker node.

A base64 string should be printed to your terminal. Copy it.

### Join the cluster from the worker node

Open a shell to the worker node:

```
multipass shell expert-elf
```

The VM doesn't come with the `k8s` snap, so let's install it.

```
sudo snap install --edge --classic k8s
```

From the worker node machine, use the `join-cluster` command.

```
sudo k8s kubectl join-cluster eyJu...XX0=
```

After a couple of seconds, you should see: `Joined the cluster.`

Disconnect from the VM and run the `status` command on your control plane node:

```
sudo k8s status
```



Congratulations!

### 10. Remove Canonical Kubernetes (Optional)
To uninstall the Canonical Kubernetes snap, execute:

```
sudo snap remove k8s
```

This command removes the `k8s` snap and automatically creates a snapshot of all data for future restoration.

If you wish to remove the snap without saving a snapshot of its data, add `--purge` to the command:

```
sudo snap remove k8s --purge
```
This option ensures complete removal of the snap and its associated data.

## Next Steps

- Keep mastering Canonical Kubernetes with kubectl: [How to use kubectl](#TODO)
- Explore Kubernetes commands with our [Command Reference Guide](#TODO)
- Bootstrap K8s with your custom configurations [Bootstrap K8s](#TODO)
- Learn how to set up a multi-node environment [Setting up a K8s cluster](#TODO)
- Configure storage options [Storage](#TODO)
- Master Kubernetes networking concepts: [Networking](#TODO)
- Discover how to enable and configure Ingress resources [Ingress](#TODO)