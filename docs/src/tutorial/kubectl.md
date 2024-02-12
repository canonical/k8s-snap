# Guide to Basic Operations with Kubernetes using kubectl

## Introduction

This guide will walk you through performing basic operations on a Kubernetes
cluster using kubectl. Kubernetes is an open-source platform designed to
automate deploying, scaling, and managing containerized applications.

## Prerequisites

Before you begin, make sure you have the following:

- A bootstrapped Canonical K8s cluster (See [insert link]...)

## Table of Contents

1. [The Kubectl Command](#check-kubernetes-version)
2. [Formatting Output](#formatting-output)

## The Kubectl Command

This commands interacts with the Kubernetes API server (kube-apiserver) and is the most commonly used command when working with Kubernetes, so let's take some time to familiarize ourselves with it.

It's syntax is as follows:

```
sudo k8s kubectl [command] [TYPE] [NAME] [flags]
```

- `command` is one of the commands defined [upstream](https://kubernetes.io/docs/reference/kubectl/generated/) like `get`, `delete` and `apply`.

- `TYPE` is the Kubernetes API resource type you want to interact with (Use `sudo k8s kubectl api-resources` to see all the available resources) like `node`, `pod`, `secret`. Remember you can use the singular, plural or abbreviated form of a type.
```
sudo k8s kubectl get deploy
sudo k8s kubectl get deployment
sudo k8s kubectl get deployments
```

- `NAME` is the name of the instance of a resource you want to interact with, like the name of a pod. If it's omitted kube-apiserver will return all instances of that resource type.

- `flags` is for optional flags and will change depending on which command you want to use `sudo k8s kubectl [command] -h` will show you available flags.

## Formatting Output

Kubectl offers powerful ways to consume it's output. Let's use the `-o` option to find the IP of the `storage-writer-pod` we created.

```
sudo k8s kubectl get pod storage-writer-pod -o yaml | yq '.status.podIP'
```

This is very useful if you want to create scripts that interact with a cluster and need to know it's status.

## References
[https://kubernetes.io/docs/reference/kubectl/](https://kubernetes.io/docs/reference/kubectl/)