# Canonical Kubernetes Snap
[![End to End Tests](https://github.com/canonical/k8s-snap/actions/workflows/e2e.yaml/badge.svg)](https://github.com/canonical/k8s-snap/actions/workflows/e2e.yaml)

Canonical Kubernetes is an opinionated Kubernetes delivered by snaps. The focus is on:
- simplified almost zero operations and
- enhanced security posture on any infrastructure


## What are the built-in components?
The built-in components (network, DNS, storage, RBAC, ingress and gateway) provide most basic functionalities crucial for K8s clusters.

## Quickstart

Install K8s with:

```
snap install k8s
```

`kubectl` is included as a command:

```
sudo k8s kubectl get nodes
sudo k8s kubectl get services
```
