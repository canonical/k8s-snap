---
myst:
  html_meta:
    description: Learn how to deploy your first Canonical Kubernetes cluster using the k8s snap and run typical cluster operations in this tutorial.
relatedlinks: "[Snap&#32;documentation](https://snapcraft.io/docs/), [kubectl&#32;documentation](https://kubernetes.io/docs/reference/kubectl/), [Multipass&#32;documentation](https://canonical.com/multipass/docs/stable)"
---

# Getting started

<!-- SPREAD SUITE: snap_clean -->

{{product}} is a distribution of Kubernetes. It bundles upstream Kubernetes with selected tools and services such as networking, DNS and storage into a single package called a snap! 

In this tutorial you will deploy a cluster by installing the snap
in a Multipass virtual machine (VM). You will run some typical cluster operations such as
deploying an NGINX server as a sample workload, scaling up and down your workload, interacting with built-in features and adding a worker node.

## Prerequisites

- System Requirements: Your machine should have at least **40G disk space**
  and **16G of memory**
- This tutorial is run within a Multipass VM to ensure you have a clean working environment. See the [Multipass install page](https://canonical.com/multipass/install).

## Set up your test environment

<!-- SPREAD SKIP -->

The `k8s` snap will require a certain amount of resources, so the default settings for a Multipass VM aren't going to be suitable. Open a terminal and enter the following command:

```
multipass launch 24.04 --name control-plane -m 8G -c 4 -d 20G
```

This step can take a few minutes as Multipass creates the new virtual machine named `control-plane`.

To access the VM you just created, run:

```
multipass shell control-plane
```

This command will immediately open a shell on the instance, so further commands you enter will be run on the Ubuntu instance you created. You can tell you are within the VM because the terminal will show:

```console
ubuntu@control-plane:~$
```

<!-- SPREAD SKIP END -->

## Install {{product}}

Install the {{product}} `k8s` snap with:

<!-- SPREAD SKIP -->

```{literalinclude} ../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

<!-- SPREAD SKIP END -->

<!-- SPREAD
sudo snap install k8s --classic --channel=1.37-classic/stable
-->

This may take a few moments as the snap installs all the necessary Kubernetes
components for a fully functioning cluster such as the networking, storage, and so on.

## Bootstrap a Kubernetes cluster

The `bootstrap` command initializes your cluster and configures your host system
as a Kubernetes node. You only bootstrap the cluster once.

Let's bootstrap our {{product}} cluster with default configuration:

```
sudo k8s bootstrap
```

Once the `bootstrap` command has been successfully run, the output should list the
node address and confirm the CNI is being deployed.

## Check cluster status

It may take a few minutes for the cluster to be ready. To confirm the
installation was successful, use `k8s status` with the `wait-ready` flag
to wait for {{product}} to bring up the cluster:

```
sudo k8s status --wait-ready
```

<!-- SPREAD
sudo k8s kubectl wait --for=condition=Ready node --all --timeout=5m
-->

```{important}
This command waits for 90 seconds before timing out. If it ends with a "Context canceled" error, re-run it. The command will continue to wait for the cluster to be ready. 

If it keeps failing, you may have a very slow network connection or the VM might not have enough resources. Check that you launched the VM with the `multipass launch` command provided at the start of this tutorial so there are enough resources to bootstrap the cluster.
```

Congratulations, you have just deployed a single-node cluster with {{product}}!
Now let's see what you can do with it.

## Access Kubernetes

The standard tool for deploying and managing workloads on Kubernetes
is [kubectl](https://kubernetes.io/docs/reference/kubectl/).
For convenience, {{product}} bundles a version of
kubectl for you to use with no extra setup or configuration.

The format of `kubectl` commands are:

<!-- SPREAD SKIP -->

```
sudo k8s kubectl <COMMAND>
```

<!-- SPREAD SKIP END -->

For example, to view your node run the command:

```
sudo k8s kubectl get nodes
```

<!-- SPREAD
source ${SPREAD_PATH}/docs/tools/repeat_checks.sh
repeat_checks "sudo k8s kubectl get nodes" "Ready"
-->

…or to see the running services:

```
sudo k8s kubectl get services
```

<!-- SPREAD 
sudo k8s kubectl get services | grep kubernetes
-->


Run the following command to list all the pods in the `kube-system`
namespace:

```
sudo k8s kubectl get pods -n kube-system
```

A [pod](https://kubernetes.io/docs/concepts/workloads/pods/) is the smallest
deployable unit in Kubernetes. You will see at least five pods running.
The status of the pods may be in `ContainerCreating` while they are being
initialized. Run the command again after a few seconds and they should report
status as `Running`.

The functions of these pods are:

- DNS (`coredns`): lets pods find each other by name
- Network manager (`cilium-operator`): manages networking for the whole cluster
- Network agent (`cilium`): runs on every node and connects pods to each other
- Storage controller (`ck-storage-rawfile-csi-controller`): looks after local storage for the whole cluster
- Storage agent (`ck-storage-rawfile-csi-node`): runs on every node and creates volumes on that node's disk for its pods

<!-- SPREAD 
sudo k8s kubectl get pods -n kube-system | grep coredns
sudo k8s kubectl get pods -n kube-system | grep cilium-operator
sudo k8s kubectl get pods -n kube-system | grep cilium-
sudo k8s kubectl get pods -n kube-system | grep ck-storage-rawfile-csi-controller
sudo k8s kubectl get pods -n kube-system | grep ck-storage-rawfile-csi-node
-->

## Deploy an app

Kubernetes is meant for deploying and managing apps and services. We will use
the `kubectl` command to do this too.

Let's deploy a demo web server (NGINX):

```
sudo k8s kubectl create deployment nginx --image=nginx
```

This command launches a pod running the NGINX application within a
container.

Check the status of your pods by running:

```
sudo k8s kubectl get pods
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get pods" "Running"
-->

This command shows all pods in the default namespace. It may take a moment for the pod to be ready and running.

Now to check that the NGINX server in the pod works correctly, run the same command again but this time add
the `-owide` argument so we get more information about the pod:

<!-- SPREAD SKIP -->

```
sudo k8s kubectl get pods -owide
```

In the output, you will see more columns such as `IP` and `NODE`. We are interested in the IP address of the pod. Copy the IP of our NGINX pod and query it using `curl`:

```
curl <POD-IP>
```

<!-- SPREAD SKIP END -->

<!-- SPREAD 
POD_IP=$(sudo k8s kubectl get pods -l app=nginx -o jsonpath='{.items[0].status.podIP}')
curl $POD_IP | grep "Welcome to nginx"
-->

You should see the NGINX welcome page. Your app is up and running! 

Let's now scale this deployment, which means increasing the number of pods it deploys.

```
sudo k8s kubectl scale deployment nginx --replicas=3
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get deployment nginx" "3/3"
-->

Run `sudo k8s kubectl get pods` again and notice that you have three NGINX pods.

A neat feature of Kubernetes is that it tries to be self-healing where possible. It is always checking to see if the actual state of the cluster is the same as we requested.

If one of our NGINX pods gets deleted, it will spin up a new one in its place to bring the total back to three. Let's test that.

Delete the NGINX pods. This command deletes all pods with the label `app=nginx`:

```
sudo k8s kubectl delete pods -l app=nginx
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get deployment nginx" "3/3"
-->

Now watch (`sudo k8s kubectl get pods`) as the {{product}} cluster deletes the original three pods and creates three new ones with status of `ContainerCreating`. The pod name will still be `nginx-` but the unique identifier will change.

You have now deployed a {{product}} cluster with a workload application you can scale up and down!

## Remove an app

To remove the NGINX workload, run the following command:

```
sudo k8s kubectl delete deployment nginx
```

To verify that the pod has been removed, check the status of pods by
running:

```
sudo k8s kubectl get pods
```

There should be no resources listed.

<!-- SPREAD
repeat_checks "sudo k8s kubectl get pods" "No resources found"
-->

## Enable and disable local storage

As we learned earlier, {{product}} comes with services built into the snap such as network, storage, ingress and loadbalancer. The `enable` and `disable` commands turn
these services on and off. Local storage is on by default, so first check how it appears in the cluster
status:

```
sudo k8s status
```

<!-- SPREAD
repeat_checks "sudo k8s status | grep local-storage" "enabled"
-->

Look for the `local-storage` line which shows `enabled`.

Turn local storage off:

```
sudo k8s disable local-storage
```

Check the status again. This time you should see `local-storage` is set to `disabled`:

```
sudo k8s status
```

<!-- SPREAD
repeat_checks "sudo k8s status | grep local-storage" "disabled"
-->

Let's verify that all the local storage pods were taken down too.

```
sudo k8s kubectl get pods -n kube-system
```

The storage pods (`ck-storage-rawfile-csi-controller` and `ck-storage-rawfile-csi-node`) are removed from the
`kube-system` namespace while the feature is off.

<!-- SPREAD
repeat_checks "sudo k8s kubectl get pods -n kube-system -l app.kubernetes.io/name=rawfile-csi" "No resources found"
-->

Now turn local storage back on:

```
sudo k8s enable local-storage
```

<!-- SPREAD
sudo k8s get local-storage.enabled | grep true
repeat_checks "sudo k8s kubectl get pods -n kube-system -l app.kubernetes.io/name=rawfile-csi,component=controller" "Running"
repeat_checks "sudo k8s kubectl get pods -n kube-system -l app.kubernetes.io/name=rawfile-csi,component=node" "Running"
-->

## Use local storage

Local storage enables you to create persistent volumes, which is the
Kubernetes way to preserve application data beyond the life-cycle of a pod.
Create a `PersistentVolumeClaim` and use it in a `Pod`:

```
sudo k8s kubectl apply -f https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/assets/tutorial-pod-with-pvc.yaml
```

This command deploys the YAML configuration of a pod called `storage-writer-pod`
and a persistent volume claim called `myclaim` with a capacity of 1Gi.

Confirm that the persistent volume claim is up and running.

```
sudo k8s kubectl get pvc myclaim
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get pvc myclaim" "Bound"
-->

The persistent volume claim status should be `Bound`. Now let's inspect the
`storage-writer-pod`.

```
sudo k8s kubectl describe pod storage-writer-pod
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl describe pod storage-writer-pod" "Running"
-->

This output provides a detailed description of the `storage-writer-pod`. You
should see `myclaim` listed under Volumes showing that it has been assigned
correctly.

<!-- SPREAD SKIP -->

```console
Volumes:
  storage-volume:
    Type:       PersistentVolumeClaim
    ClaimName:  myclaim
```

<!-- SPREAD SKIP END -->

Remove the pod along with the persistent volume claim.

```
sudo k8s kubectl delete pod storage-writer-pod
sudo k8s kubectl delete pvc myclaim
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get pods" "No resources found"
-->

This may take a few moments as the cluster cleans up its resources.

<!-- SPREAD SKIP -->

## Add a worker node

Why just stick to one node? Kubernetes is capable of having multiple nodes where we can distribute our workloads.

In {{product}} we can have two types of nodes: control plane and worker nodes. In the tutorial so far, we just initialized a control plane node. Now let's add a worker.

Generate the token required for the worker node to join the cluster by running the following command on the control plane node.

```
sudo k8s get-join-token worker-node --worker
```

`worker-node` is the hostname of the node we want to join. The `--worker` flag sets the type of node to join.

A base64 token will be printed to your terminal. Keep it handy as you will need it for the next step.

We need to exit the control plane VM.

```
exit
```

You should be back to your original terminal. Now let's spin up a worker VM.

```
multipass launch 24.04 --name worker-node -m 8G -c 4 -d 20G
```

Shell into the worker node VM so we can run commands inside.

```
multipass shell worker-node
```

You will see the terminal now lists `worker-node`.

```console
ubuntu@worker-node:~$
```

We need to download the `k8s` snap on the worker node too, just like we did for the control plane node.

```{literalinclude} ../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

Setting up the worker node matches the control plane node up to this point. The next step is different. The control plane node ran `k8s bootstrap` to create the
cluster, but a worker node runs `k8s join-cluster` to join the cluster that already exists. Remember earlier we learned that bootstrapping only happens once!

On the worker node, run the `join-cluster` command with the join token we generated before.

```
sudo k8s join-cluster <JOIN-TOKEN>
```

After a few seconds, you should see: `Cluster services have started on "worker-node"`.

Now let's verify the node joined the cluster correctly. For this, go back to the control plane node.

Exit the worker node.

```
exit
```

Shell back into the control plane VM.

```
multipass shell control-plane
```

Now we're back in the control plane node, check the nodes again.

```
sudo k8s kubectl get nodes
```

This time you should see two nodes: the control plane node and the worker node.

```console
NAME            STATUS   ROLES                  AGE     VERSION
control-plane   Ready    control-plane,worker   27m     v1.37.0
worker-node     Ready    worker                 2m15s   v1.37.0
```

<!-- SPREAD SKIP END -->

## What you've achieved

You now have a working multi-node cluster!

Along the way you:

- Installed {{product}} and bootstrapped a cluster in a Multipass VM.
- Deployed an NGINX app with `kubectl`, scaled it, and saw the cluster replace deleted pods.
- Turned local storage on and off, and used it to run a pod with a persistent volume.
- Added a worker node to the cluster.

## Tear down the environment

<!-- SPREAD SKIP -->

Finally, if you want to undo everything you did in this tutorial, exit from the Multipass VM.

```
exit
```

…and you will be returned to the original terminal session.

The instances you created will keep running in the background until they are either stopped or the host computer is shut down. You can stop the running instances at any time by running:

```
multipass stop control-plane
multipass stop worker-node
```

And they can be permanently removed with:

```
multipass delete control-plane --purge
multipass delete worker-node --purge
```

<!-- SPREAD SKIP END -->

## Next steps

- Understand how {{product}} is put together in the {ref}`architecture guide <snap-architecture>`
- Configure {ref}`storage <snap-storage>` and {ref}`networking <snap-networking>` in the cluster
- Explore Kubernetes commands with the {ref}`command reference guide <snap-commands>`
