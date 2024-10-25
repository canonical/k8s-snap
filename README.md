# Canonical Kubernetes Snap
[![End to End Tests](https://github.com/canonical/k8s-snap/actions/workflows/integration.yaml/badge.svg)](https://github.com/canonical/k8s-snap/actions/workflows/integration.yaml)
![](https://img.shields.io/badge/Kubernetes-1.31-326de6.svg)

TEST

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/k8s)



**Canonical Kubernetes** is the fastest, easiest way to deploy a fully-conformant Kubernetes cluster. Harnessing pure upstream Kubernetes, this distribution adds the missing pieces (e.g. ingress, dns, networking) for a zero-ops experience.

For more information and instructions, please see the official documentation at: [https://documentation.ubuntu.com/canonical-kubernetes](https://documentation.ubuntu.com/canonical-kubernetes/)

## Quickstart

Install Canonical Kubernetes and initialise the cluster with:

```bash
sudo snap install k8s --channel=1.30-classic/beta --classic
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

```bash
sudo snap install snapcraft --classic
```

Building the project by running `snapcraft` in the root of this repository. Snapcraft spawns a VM managed by Multipass and builds the snap inside it. If you don’t have Multipass installed, snapcraft will first prompt for its automatic installation.

After snapcraft completes, you can install the newly compiled snap:

```bash
sudo snap install k8s_*.snap --classic --dangerous
```
