# Installing Canonical Kubernetes Offline or in an air gapped environment

There are situations where it is necessary or desirable to run Canonical 
Kubernetes on a machine that is not connected to the internet. 
Based on different degrees of separation from the network,
different solutions are offered to accomplish this goal.
This guide explains the necessary preparation required for the
offline installation and walks you through the different potential scenarios.

## Prepare for Deployment

In preparation for the offline deployment you will download the Canonical
Kubernetes snap, fulfill the networking requirements based on your scenario and
handle images for workloads and Canonical Kubernetes features.

### Prep 1: Download the Canonical Kubernetes snap

From a machine with access to the internet download the following:
```
sudo snap download k8s --channel 1.30-classic/beta
sudo snap download core20
sudo mv k8s_*.snap k8s.snap
sudo mv k8s_*.assert k8s.assert
sudo mv core20_*.snap core20.snap
sudo mv core20_*.assert core20.assert
```

The [core20][Core20] and `k8s` snap are downloaded. The `core20.assert` and 
`k8s.assert` files, are necessary to verify the integrity of the snap packages.

 <!-- TODO: link in a note? -->
```{note} Update the version of k8s by adjusting the channel parameter.
    Find the version you desire in the [snapstore][snapstore].
```

```{note} With updates to the snap the base core is subject to change.```

### Prep 2: Network Requirements

Air-gap deployments are typically associated with a number of constraints and
restrictions with the networking connectivity of the machines.
Verify that your cluster nodes can communicate, machines have a default gateway
and optionally ensure proxy access.

#### Network Requirement: Cluster node communication
<!-- TODO: Services and Ports Doc -->
Ensure that all cluster nodes are reachable from each other.
Refer to Services and ports used for a list of all network ports
used by Canonical Kubernetes.

#### Network Requirement: Default Gateway

Kubernetes services use the default network interface of the machine
for the means of node discovery:

- kube-apiserver (part of kubelite) 
  - uses the default network interface to advertise this address to other nodes in the cluster.
  - Without a default route kube-apiserver does not start.
- kubelet (part of kubelite)
  - uses the default network interface to pick the node's InternalIP address.
  - A default gateway greatly simplifies the process of setting up the network feature.

In case your air gap environment does not have a default gateway,
you can add a dummy default route on interface eth0 using the following command:

```
ip route add default dev eth0
```
<!-- TODO: back-tick quoty thingies -->
```{note} Confirm the name of you default network interface used for pod-to-pod communication by running "ip a".
```
```{note} The dummy gateway will only be used by the Kubernetes services to know which interface to use, actual connectivity to the internet is not required. Ensure that the dummy gateway rule survives a node reboot.
```

#### (Optional) Network Requirement: Ensure proxy access
If you do not allow an HTTP proxy (e.g. squid) limited access to 
image registries (e.g. docker.io, quay.io, rocks.canonical.com, etc)
please skip this section.

Ensure that all nodes can use the proxy to access the image registry.
In this example we use squid as an http proxy.
This set up uses http://squid.internal:3128 to access docker.io.
Test the connectivity:
```
export https_proxy=http://squid.internal:3128
curl -v https://registry-1.docker.io
```

Please refer to the next section `images` on how to use the HTTP proxy
to allow limited access to image registries.

## Prep 3: Images

All workloads in a Kubernetes cluster are running as an OCI image.
Kubernetes needs to be able to fetch these images and load them
into the container runtime, in order to run workloads.
For a Canonical Kubernetes deployment, you will need to fetch the images
used by its features (network, dns, etc) as well as any images that are
needed to run your workloads.

The following options are presented in the order of
increasing complexity of implementation.
You may also find it helpful to combine these options for your scenario.

### Images Option A: via an HTTP proxy
In many cases, the nodes of the airgap deployment may not have direct access to upstream registries, but can reach them through the [use of an HTTP proxy][proxy].

### Images Option B: private registry mirror
In case regulations and/or network constraints do not allow the cluster nodes
to access any upstream image registry,
it is typical to deploy a private registry mirror.
This is an image registry service that contains all the required OCI Images
(e.g. [registry](https://docs.docker.com/registry/),
[Harbor](https://goharbor.io/) or any other OCI registry) and
is reachable from all cluster nodes.

This requires three steps:

1. Deploy and secure the registry service. This is out of scope for this
   document, please follow the instructions for the registry
   that you want to deploy.
2. Load all images from the upstream source and push to our registry mirror.
3. Configure the MicroK8s container runtime (`containerd`) to load images from
   the private registry mirror instead of the upstream source. This will be
   described in the
   [Configure registry mirrors](#option-b-configure-registry-mirrors) section.

In order to load images into the private registry, you need a machine with
access to both the upstream registry (e.g. `docker.io`) and the internal one.
Loading the images is possible with `docker` or `ctr`.

For the examples below we assume that a private registry mirror is running at `10.100.100.100:5000`.

#### Load images with ctr

On the machine with access to both registries, first install `ctr`.
For Ubuntu hosts, this can be done with:

```bash
sudo apt-get update
sudo apt-get install containerd
```

Then, pull an image:

```{note}  For DockerHub images, prefix with `docker.io/library`. ```

```bash
export IMAGE=library/nginx:latest
export FROM_REPOSITORY=docker.io
export TO_REPOSITORY=10.100.100.100:5000

# pull the image and tag
ctr image pull "$FROM_REPOSITORY/$IMAGE"
ctr image convert "$FROM_REPOSITORY/$IMAGE" "$TO_REPOSITORY/$IMAGE"
```

Finally, push the image (see `ctr image push --help` for a complete list of
supported arguments):

```bash
# push image
ctr image push "$TO_REPOSITORY/$IMAGE"
# OR, if using HTTP and basic auth
ctr image push "$TO_REPOSITORY/$IMAGE" --plain-http -u "$USER:$PASS"
# OR, if using HTTPS and a custom CA (assuming CA certificate is at `/path/to/ca.crt`)
ctr image push "$TO_REPOSITORY/$IMAGE" --ca /path/to/ca.crt
```

Make sure to repeat the steps above (pull, convert, push) for all the images that you need.

##### Load images with docker

On the machine with access to both registries, first install `docker`.
For Ubuntu hosts, this can be done with:

```bash
sudo apt-get update
sudo apt-get install docker.io
```

If needed, login to the private registry:

```bash
sudo docker login $TO_REGISTRY
```

Then pull, tag and push the image:

```bash
export IMAGE=library/nginx:latest
export FROM_REPOSITORY=docker.io
export TO_REPOSITORY=10.100.100.100:5000

sudo docker pull "$FROM_REPOSITORY/$IMAGE"
sudo docker tag "$FROM_REPOSITORY/$IMAGE" "$TO_REPOSITORY/$IMAGE"
sudo docker push "$TO_REPOSITORY/$IMAGE"
```

Repeat the pull, tag and push steps for all required images.

### Images Option C: Side-load images
Image side-loading is the process of loading all required OCI images directly
into the container runtime, so that they do not have to be fetched at runtime.
Upon choosing this option, you need to create a bundle of all the OCI images
that will be used by the cluster.

<!-- TODO: how to image side loading with CK8s -->

## Deploy Canonical Kubernetes

Now that you have fulfilled all steps in preparation for your
air gapped cluster, it is time to get it deployed.

### Step 1: Install Canonical Kubernetes

Copy the `k8s.snap`, `k8s.assert`, `core20.snap` and `core20.assert` files into
the target node, then install with:

```bash
sudo snap ack core20.assert && sudo snap install ./core20.snap
sudo snap ack k8s.assert && sudo snap install ./k8s.snap --classic
```

Repeat the above for all nodes of the cluster.

### Step 2: Form Canonical Kubernetes cluster

```{note}  Please skip this section for one node deployments. ```

You can add and remove nodes as described in the
[add-and-remove-nodes tutorial][nodes].

After a while, confirm that all the cluster nodes show up in
the output of the `sudo k8s kubectl get node`. 

The nodes will most likely be in `NotReady` state,
since we still need to ensure the container runtime can fetch images.

### Step 3: Container Runtime

<!-- LINKS -->

[Core20]: https://canonical.com/blog/ubuntu-core-20-secures-linux-for-iot
[snapstore]: https://snapcraft.io/k8s
[proxy]: /snap/howto/proxy.md
[nodes]: /snap/tutorial/add-remove-nodes.md