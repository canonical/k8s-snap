# Basic operations with Kubernetes using kubectl

Kubernetes provides a command line tool for communicating with a Kubernetes
cluster's control plane, using the Kubernetes API. This guide outlines how some
of the everyday operations of your Kubernetes cluster can be managed with this
tool.

## Prerequisites

Before you begin, make sure you have the following:

- A bootstrapped {{product}} cluster (See
  [Getting Started])

### 1. The kubectl command

The `kubectl` command communicates with the
[Kubernetes API server][kubernetes-api-server].

{{product}} includes a `kubectl` which is built from the
original upstream source into the `k8s` snap you have installed and is
configured to work with the cluster out of the box.

### 2. How to use kubectl

To access `kubectl`, run the following:

```
sudo k8s kubectl
```

This will display a list of commands possible with `kubectl`.

> **Note**: Only control plane nodes can use the `kubectl` command. Worker
> nodes do not have access to this command.

The format of `kubectl` commands are:

```
sudo k8s kubectl <command>
```

### 3. Configuration

In {{product}}, the `kubeconfig` file that is being read to display
the configuration when you run `kubectl config view` lives at
`/etc/kubernetes/admin.conf`. You can change this by setting a
`KUBECONFIG` environment variable or passing the `--kubeconfig` flag to a
command.

To find out more, you can visit
[the official kubeconfig documentation][kubeconfig-doc]

### 4. Viewing objects

Let's review what was created in the [Getting Started]
guide.

To see what pods were created when we enabled the `network` and `dns`
components during the cluster bootstrap:

```
sudo k8s kubectl get pods -o wide -n kube-system
```

You should be seeing the network operator, networking agent and CoreDNS pods.

> **Note**: If you see an error message here, it is likely that you forgot to
> bootstrap your cluster.

```
sudo k8s kubectl get services --all-namespaces
```

The `kubernetes` service in the `default` namespace is where the Kubernetes API
server resides, and it's the endpoint with which other nodes in your cluster
will communicate.

### 5. Creating and managing objects

Let's deploy an NGINX server using this command:

```
sudo k8s kubectl create deployment nginx --image=nginx:latest
```

To observe the NGINX pod running in the default namespace:

```
sudo k8s kubectl get pods
```

Let's now scale this deployment, which means increasing the number of pods it
manages.

```
sudo k8s kubectl scale deployment nginx --replicas=3
```

Execute `sudo k8s kubectl get pods` again and notice that you have 3 NGINX
pods.

Let's delete those 3 pods to demonstrate a deployment's ability to ensure the
declared state of the cluster is maintained.

First, open a new terminal so you can watch the changes as they happen. Run
this command in a new terminal:

```
sudo k8s kubectl get pods --all-namespace --watch
```

Now, go back to your original terminal and run:

```
sudo k8s kubectl delete pods -l app=nginx
```

The above command deletes all pods in the cluster that are labelled with
`app=nginx`.

You'll notice the original 3 pods will have a status of `Terminating` and 3 new
pods will have a status of `ContainerCreating`.

## Further information

- Explore Kubernetes commands with our
  [Command Reference Guide]
- See the official `kubectl` reference
  [kubectl-reference][kubectl-reference]

<!-- LINKS -->

[Command Reference Guide]: ../reference/commands
[Getting Started]: getting-started
[kubernetes-api-server]: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/
[kubeconfig-doc]: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
[kubectl-reference]: https://kubernetes.io/docs/reference/kubectl/
