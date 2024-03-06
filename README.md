# Canonical Kubernetes Snap
[![End to End Tests](https://github.com/canonical/k8s-snap/actions/workflows/e2e.yaml/badge.svg)](https://github.com/canonical/k8s-snap/actions/workflows/e2e.yaml)

**Canonical Kubernetes** is the fastest, easiest way to deploy a fully-conformant Kubernetes cluster. Harnessing pure upstream Kubernetes, this distribution adds the missing pieces (e.g. ingress, dns, networking) for a zero-ops experience. For single-node clusters it can be deployed with two commands. Add new nodes with just two more. 

Get started in just two commands:
```bash
sudo snap install k8s --classic
```

And bring up the cluster:
```bash
sudo k8s bootstrap
```

For more information and instructions, please see the official documentation at: https://ubuntu.com/kubernetes

### What is included in the k8s distribution

In addition to the upstream Kubernetes services, Canonical Kubernetes also includes:

- a DNS service for the node
- a CNI for the node/cluster
- a simple local storage provider
- an ingress provider
- a load-balancer
- a gateway API controller
- a metrics server


## Quickstart

Install the Canonical Kubernetes and initialise the cluster with:

```bash
sudo snap install k8s --edge --classic
sudo k8s bootstrap
```

Confirm the installation was successful:

```bash
sudo k8s status
```

Use `kubectl` to interact with k8s:

```bash
sudo k8s kubectl get pods -A
```

Remove the snap with:

```bash
sudo snap remove k8s --purge
```


## Build the project from source

To build the Kubernetes snap on an Ubuntu machine you need Snapcraft.

```
sudo snap install snapcraft --classic
```

Building the project by running `snapcraft` in the root of this repository. Snapcraft spawns a VM managed by Multipass and builds the snap inside it. If you don’t have Multipass installed, snapcraft will first prompt for its automatic installation.

After snapcraft completes, you can install the newly compiled snap:

```
sudo snap install k8s_*.snap --classic --dangerous
```
