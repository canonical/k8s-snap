# Guide to Basic Operations with Kubernetes using kubectl

## Introduction
This guide will walk you through performing basic operations on a Kubernetes cluster using kubectl. Kubernetes is an open-source platform designed to automate deploying, scaling, and managing containerized applications.

## Prerequisites
Before you begin, make sure you have the following:
- Snapd is running (`sudo systemctl status snapd`)
- Canonical K8s is installed (`which k8s`)
- Access to the kubeconfig file for your Kubernetes cluster

## Table of Contents
1. [Bootstrap a Cluster](#bootstrap-a-cluster)
1. [Checking Cluster Information](#checking-cluster-information)
2. [Managing Resources](#managing-resources)
3. [Pod Operations](#pod-operations)
4. [Service Operations](#service-operations)
5. [ConfigMap and Secret Operations](#configmap-and-secret-operations)
6. [Scaling](#scaling)
7. [Port Forwarding](#port-forwarding)
8. [Rolling Updates](#rolling-updates)
9. [Namespace Operations](#namespace-operations)
10. [Applying Labels](#applying-labels)
11. [Viewing and Editing Resources](#viewing-and-editing-resources)
12. [Accessing Kubernetes Dashboard](#accessing-kubernetes-dashboard)

## Bootstrap a Cluster
`$ sudo k8s bootstrap`

`$ sudo k8s status`
```sh
❯ sudo k8s status
k8s is not ready.
high-availability: no

control-plane nodes:
  dev: 192.168.0.159

components:
  dns        disabled
  gateway    disabled
  ingress    disabled
  loadbalancer disabled
  network    disabled
  storage    disabled
```

## Checking Cluster Information
The output of `k8s status` indicates that we have one node running. We can also check this information with `kubectl`.

`$ sudo k8s kubectl get nodes`

## Add a node to your cluster
Let's get ready to run some workloads in our cluster. The first step is to add a new node to the cluster. For this, we'll use `multipass`.

`$ sudo snap install multipass`

Create a VM called `expert-elf`

```sh
$ multipass shell expert-elf
ubuntu@expert-elf:~$
``` 

Install Canonical K8s inside the VM:

```sh
ubuntu@expert-elf:~$ sudo snap install --classic --edge k8s k8s (edge) v1.29.1 from Konstantinos Tsakalozos (kjackal) installed
```

Now, on your host machine, run the `add-node` command:

```sh
❯ sudo k8s add-node expert-elf
eyJuYW1lIjoiZXhwZXJ0LWVsZiIsInNlY3JldCI6IjhlN
zY4ZDkxYzRjZmY3MjZjZDdmMmNjODdkNGQ5OWEzNjkwMm
JhZDcwZDBhN2NiMGEzYmEyODJmNjRlMjk0ZGEiLCJmaW5
nZXJwcmludCI6IjE3ZDZkNTE2NmFkODhhNGY0YjdkMGE5
OTMyYzFlYmIzM2U3NGUyN2IwZmU1YWUxOWEwYmY3MzY5Z
WJkMTQ3ZTYiLCJqb2luX2FkZHJlc3NlcyI6WyIxOTIuMT
Y4LjAuMTU5OjY0MDAiXX0=
```


View cluster information
You can view more information about the control plane node by using `kubectl describe`.

`$ sudo k8s kubectl describe node dev`

## Managing Resources
- Creating resources from YAML files

Create a NGINX deployment.

`$ sudo k8s kubectl create deployment --image nginx:latest nginx`

- Viewing resources
- Describing resources
- Deleting resources

## Pod Operations
- Viewing pods
- Describing pods
- Getting logs from pods
- Executing commands in pods

## Service Operations
- Viewing services
- Describing services

## ConfigMap and Secret Operations
- Viewing ConfigMaps
- Viewing Secrets

## Scaling
- Scaling a deployment

## Port Forwarding
- Forwarding local ports to pods

## Rolling Updates
- Updating a deployment

## Namespace Operations
- Viewing namespaces
- Creating namespaces
- Switching namespaces

## Applying Labels
- Applying labels to resources

## Viewing and Editing Resources
- Editing resources

## Accessing Kubernetes Dashboard
- Accessing the Kubernetes dashboard

## Conclusion
This guide covers basic operations with Kubernetes using kubectl. With these commands, you can manage and interact with your Kubernetes cluster effectively.

## Additional Resources
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
