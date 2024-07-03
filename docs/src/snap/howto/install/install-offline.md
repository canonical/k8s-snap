# Installing Canonical Kubernetes in air-gapped environments

There are situations where it is necessary or desirable to run Canonical
Kubernetes on a machine that is not connected to the internet.
Based on different degrees of separation from the network,
different solutions are offered to accomplish this goal.
This guide documents any necessary extra preparation for air-gap deployments,
as well the steps that are needed to successfully deploy Canonical Kubernetes
in such environments.

## Prepare for Deployment

In preparation for the offline deployment download the Canonical
Kubernetes snap, fulfill the networking requirements based on your scenario and
handle images for workloads and Canonical Kubernetes features.

### Download the Canonical Kubernetes snap

From a machine with access to the internet download the
`k8s` and `core20` snap with:

```
sudo snap download k8s --channel 1.30-classic/beta --basename k8s
sudo snap download core20 --basename core20
```

Besides the snaps, this will also download the corresponding assert files which
are necessary to verify the integrity of the packages.

```{note}
Update the version of k8s by adjusting the channel parameter.
For more information on channels visit the
[channels explanation](/snap/explanation/channels.md).
```

```{note}
With updates to the snap the base core is subject to change in the future.
```

### Network Requirements

Air-gap deployments are typically associated with a number of constraints and
restrictions when it comes to the networking connectivity of the machines.
Below we discuss the requirements that the deployment needs to fulfill.

#### Cluster node communication

<!-- TODO: Add Services and Ports Doc -->

Ensure that all cluster nodes are reachable from each other.

<!-- Refer to [Services and ports][svc-ports] used for a list of all network
ports used by Canonical Kubernetes.  -->

#### Default Gateway

In cases where the air-gap environment does not have a default gateway,
add a dummy default route on interface eth0 using the following command:

```bash
ip route add default dev eth0
```

```{note} 
Ensure that `eth0` is the name of the default network interface used for
pod-to-pod communication.
```

```{note} 
The dummy gateway will only be used by the Kubernetes services to 
know which interface to use, actual connectivity to the internet is not 
required. Ensure that the dummy gateway rule survives a node reboot.
```

#### Ensure proxy access

This section is only relevant if access to upstream image registries
(e.g. docker.io, quay.io, rocks.canonical.com, etc.)
is only allowed through an HTTP proxy (e.g. [squid][squid]).

Ensure that all nodes can use the proxy to access the image registry.
For example, if using `http://squid.internal:3128` to access docker.io,
an easy way to test connectivity is:

```
export https_proxy=http://squid.internal:3128
curl -v https://registry-1.docker.io/v2
```

### Images

All workloads in a Kubernetes cluster are running as an OCI image.
Kubernetes needs to be able to fetch these images and load them
into the container runtime, in order to run workloads.
For a Canonical Kubernetes deployment, it is necessary to fetch the images used
by its features (network, dns, etc) as well as any images that are
needed to run specific workloads.

```{note} 
The image options are presented in the order of
increasing complexity of implementation.
It may be helpful to combine these options for different scenarios.
```

#### List images

If the `k8s` snap is already installed,
list the images in use with the following command:

```bash
ubuntu@demo:~$ k8s list-images
ghcr.io/canonical/cilium-operator-generic:1.15.2-ck2
ghcr.io/canonical/cilium:1.15.2-ck2
ghcr.io/canonical/coredns:1.11.1-ck4
ghcr.io/canonical/k8s-snap/pause:3.10
ghcr.io/canonical/k8s-snap/sig-storage/csi-node-driver-registrar:v2.10.1
ghcr.io/canonical/k8s-snap/sig-storage/csi-provisioner:v5.0.1
ghcr.io/canonical/k8s-snap/sig-storage/csi-resizer:v1.11.1
ghcr.io/canonical/k8s-snap/sig-storage/csi-snapshotter:v8.0.1
ghcr.io/canonical/metrics-server:0.7.0-ck0
ghcr.io/canonical/rawfile-localpv:0.8.0-ck5
```

A list of images can be found in the `images.txt` file when unsquashing the
downloaded k8s snap.

Please ensure that the images used by workloads are tracked as well.

#### Images Option A: via an HTTP proxy

In many cases, the nodes of the air-gap deployment may not have direct access to
upstream registries, but can reach them through the
[use of an HTTP proxy][proxy].

The configuration of the proxy is out of the scope of this documentation.

#### Images Option B: private registry mirror

In case regulations and/or network constraints do not allow the cluster nodes
to access any upstream image registry,
it is typical to deploy a private registry mirror.
This is an image registry service that contains all the required OCI Images
(e.g. [registry](https://distribution.github.io/distribution/),
[harbor](https://goharbor.io/) or any other OCI registry) and
is reachable from all cluster nodes.

This requires three steps:

1. Deploy and secure the registry service.
   Please follow the instructions for the desired registry deployment.
2. Using [regsync][regsync], load all images from the upstream source and
   push to your registry mirror.
3. Configure the Canonical Kubernetes container runtime (`containerd`) to load
   images from
   the private registry mirror instead of the upstream source. This will be
   described in the [Configure registry mirrors](
      #Container-Runtime-Option-B:-Configure-registry-mirrors) section.

In order to load images into the private registry, a machine os needed with
access to both any upstream registries (e.g. `docker.io`)
and the private mirror.

##### Load images with regsync

We recommend using [regsync][regsync] to copy images
from the upstream registry to your private registry. Refer to the
[sync-images.yaml][sync-images-yaml] file that contains the configuration for
syncing images from the upstream registry to the private registry. Using the
output from `k8s list-images` update the images in the
[sync-images.yaml][sync-images-yaml] file if necessary. Update the file with the
appropriate mirror, and specify a mirror for ghcr.io that points to the
registry.

After creating the `sync-images.yaml` file, use [regsync][regsync] to sync the
images. Assuming your registry mirror is at http://10.10.10.10:5050, run:

```bash
USERNAME="$username" PASSWORD="$password" MIRROR="10.10.10.10:5050" \
./src/k8s/tools/regsync.sh once -c path/to/sync-images.yaml
```

An alternative to configuring a registry mirror is to download all necessary
OCI images, and then manually add them to all cluster nodes.
Instructions for this are described in
[Side-load images](#images-option-c-side-load-images).

#### Images Option C: Side-load images

Image side-loading is the process of loading all required OCI images directly
into the container runtime, so that they do not have to be fetched at runtime.

To create a bundle of images, use the [regctl][regctl] tool
or simply invoke the [regctl.sh][regctl.sh] script:

```bash
./src/k8s/tools/regctl.sh image export ghcr.io/canonical/k8s-snap/pause:3.10 \
--platform=local > pause.tar
```

Upon choosing this option, place all images under
`/var/snap/k8s/common/images` and they will be picked up by containerd.

## Deploy Canonical Kubernetes

After fulfilling all steps in preparation for your
air-gapped cluster, it is time to deploy it.

### Step 1: Install Canonical Kubernetes

Copy the `k8s.snap`, `k8s.assert`, `core20.snap` and `core20.assert` files into
the target node, then install the k8s snap by running:

```bash
sudo snap ack core20.assert && sudo snap install ./core20.snap
sudo snap ack k8s.assert && sudo snap install ./k8s.snap --classic
```

Repeat the above for all nodes of the cluster.

### Step 2: Container Runtime

The container runtime needs to be configured to be able to fetch images.

#### Container Runtime Option A: Configure HTTP proxy for registries

Edit `/etc/systemd/system/snap.k8s.containerd.service.d/http-proxy.conf`
on each node and set the appropriate http_proxy, https_proxy and
no_proxy variables as described in the
[adding proxy configuration section][proxy].

#### Container Runtime Option B: Configure registry mirrors

This requires having already set up a registry mirror,
as explained in the preparation section on the private registry mirror.
Complete the following instructions on all nodes.
For each upstream registry that needs mirroring, create a `hosts.toml` file.

This example configured `http://10.100.100.100:5000` as a mirror for
`ghcr.io`.

##### HTTP registry

In `/var/snap/k8s/common/etc/containerd/hosts.d/ghcr.io/hosts.toml`
add the configuration:

```
[host."http://10.100.100.100:5000"]
capabilities = ["pull", "resolve"]
```

##### HTTPS registry

HTTPS requires the additionally specification of the registry CA certificate.
Copy the certificate to
`/var/snap/k8s/common/etc/containerd/hosts.d/ghcr.io/ca.crt`.
Then add the configuration in 
`/var/snap/k8s/common/etc/containerd/hosts.d/ghcr.io/hosts.toml`:

```
[host."https://10.100.100.100:5000"]
capabilities = ["pull", "resolve"]
ca = "/var/snap/k8s/common/etc/containerd/hosts.d/ghcr.io/ca.crt"
```

#### Container Runtime Option C: Side-load images

This is only required if choosing to
[side-load images](#images-option-c-side-load-images).
Make sure that the directory `/var/snap/k8s/common/images` directory exists,
then copy all `$image.tar` to that directory, such that containerd automatically
picks them up and imports them when it starts.
Copy the `images.tar` file(s) to `/var/snap/k8s/common/images`
on each cluster node.

### Step 3: Bootstrap cluster

Now, bootstrap the cluster and replace `MY-NODE-IP` with the IP of the node
by running the command:

```bash
sudo k8s bootstrap --address MY-NODE-IP
```

Add and remove nodes as described in the
[add-and-remove-nodes tutorial][nodes].

After a while, confirm that all the cluster nodes show up in
the output of the `sudo k8s kubectl get node` command.

<!-- LINKS -->

[Core20]: https://canonical.com/blog/ubuntu-core-20-secures-linux-for-iot
[svc-ports]: /snap/explanation/services-and-ports.md
[proxy]: /snap/howto/proxy.md
[sync-images-yaml]: https://github.com/canonical/k8s-snap/blob/main/build-scripts/hack/sync-images.yaml
[regsync]: https://github.com/regclient/regclient/blob/main/docs/regsync.md
[regctl]: https://github.com/regclient/regclient/blob/main/docs/regctl.md
[regctl.sh]: https://github.com/canonical/k8s-snap/blob/main/src/k8s/tools/regctl.sh
[nodes]: /snap/tutorial/add-remove-nodes.md
[squid]: https://www.squid-cache.org/
