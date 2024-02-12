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

## The Kubectl Command

This commands interacts with the Kubernetes API server (kube-apiserver) and is the most commonly used command when working with Kubernetes, so let's take some time to familiarize ourselves with it.

It's syntax is as follows:

```
kubectl [command] [TYPE] [NAME] [flags]
```

- `command` is one of the commands defined [upstream](https://kubernetes.io/docs/reference/kubectl/generated/) like `get`, `delete` and `apply`.
- `TYPE` is the Kubernetes API resource type you want to interact with (Use `sudo k8s kubectl api-resources` to see all the available resources).
- `NAME` is the name of the instance of a resource you want to interact with, like the name of a pod. If it's omitted kube-apiserver will return all instances of that resource type.


## References
[https://kubernetes.io/docs/reference/kubectl/](https://kubernetes.io/docs/reference/kubectl/)