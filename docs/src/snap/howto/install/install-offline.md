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



### Images Option B: private registry mirror

### Images Option C: Side-load images

## Deploy Canonical Kubernetes

### Step 1: Install Canonical Kubernetes

### Step 2: Configure registry mirrors

### Step 3: Container Runtime

<!-- LINKS -->

[Getting started]: getting-started
[Core20]: https://canonical.com/blog/ubuntu-core-20-secures-linux-for-iot
[snapstore]: https://snapcraft.io/k8s