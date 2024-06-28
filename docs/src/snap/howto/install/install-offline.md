# Installing Canonical Kubernetes in air-gapped environments

There are situations where it is necessary or desirable to run Canonical
Kubernetes on a machine that is not connected to the internet.
Based on different degrees of separation from the network,
different solutions are offered to accomplish this goal.
This guide documents any necessary extra preparation for air-gap deployments,
as well the steps that are needed to successfully deploy Canonical Kubernetes
in such environments.

## Prepare for Deployment

In preparation for the offline deployment you will download the Canonical
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
For a Canonical Kubernetes deployment, you will need to fetch the images
used by its features (network, dns, etc) as well as any images that are
needed to run your workloads.

The following options are presented in the order of
increasing complexity of implementation.
You may also find it helpful to combine these options for your scenario.

If you already have the `k8s` snap installed,
you can list the images in use by running:

```bash
sudo k8s list-images
```

A list of images can also be found in the downloaded k8s snap for the
`images.txt` file.

Please remember to keep track of the images used by your workloads as well.

#### Images Option A: via an HTTP proxy

In many cases, the nodes of the airgap deployment may not have direct access to
upstream registries, but can reach them through the
[use of an HTTP proxy][proxy].

The configuration of the proxy is out of the scope of this documentation.

#### Images Option B: private registry mirror

In case regulations and/or network constraints do not allow the cluster nodes
to access any upstream image registry,
it is typical to deploy a private registry mirror.
This is an image registry service that contains all the required OCI Images
(e.g. [registry](https://distribution.github.io/distribution/),
[Harbor](https://goharbor.io/) or any other OCI registry) and
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

In order to load images into the private registry, you need a machine with
access to both any upstream registries (e.g. `docker.io`)
and the private mirror.

##### Load images with regsync

We recommend using [regsync][regsync] to copy images
from the upstream registry to your private registry.
For the images used in the k8s-snap we currently sync upstream images
to the `ghcr.io` repo.
Since you will need to do something similar you
will find it helpful to look at the [upstream-images.yaml][upstream-imgs] file
as well as the [sync-images][sync-images] script.

In [upstream-images.yaml][upstream-imgs] you will have to
change the sync target to your private registry mirror.

```yaml
sync:
  - source: ghcr.io/canonical/k8s-snap/pause:3.10
    target: '{{ env "MIRROR" }}/canonical/k8s-snap/pause:3.10'
    type: image
```

After you have updated the yaml file, you can run the [sync-images][sync-images]
script:

```bash
./src/k8s/tools/regctl.sh USERNAME="$username" PASSWORD="$password" \
MIRROR="$mirror"
```

An alternative to configuring a registry mirror is to download all necessary
OCI images, and then manually add them to all cluster nodes.
Instructions for this are described in
[Side-load images](#images-option-c-side-load-images).

#### Images Option C: Side-load images

Image side-loading is the process of loading all required OCI images directly
into the container runtime, so that they do not have to be fetched at runtime.

To create a bundle of images, you can use the [regctl][regctl] tool.

```bash
regctl image export --platform=local
```

Upon choosing this option, you place all images under
`/var/snap/k8s/common/images` and they will be picked up by containerd.

## Deploy Canonical Kubernetes

Now that you have fulfilled all steps in preparation for your
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

Edit `/etc/systemd/system/snap.k8s.containerd.conf.d/env.conf`
and set the appropriate http_proxy, https_proxy and
no_proxy variables as described in the
[adding proxy configuration section][proxy].

#### Container Runtime Option B: Configure registry mirrors

This requires that you have already setup a registry mirror,
as explained in the preparation section on the private registry mirror.
For each upstream registry that you want to mirror, create a `hosts.toml` file.

This example configured `http://10.100.100.100:5000` as a mirror for
`docker.io`.
Edit
`/var/snap/k8s/common/etc/containerd/hosts.d/docker.io/hosts.toml`
and make sure it looks like this:

##### HTTP registry

In `/var/snap/k8s/common/etc/containerd/hosts.d/docker.io/hosts.toml`
add the configuration:

```
[host."http://10.100.100.100:5000"]
capabilities = ["pull", "resolve"]
```

##### HTTPS registry

HTTPS requires that you additionally specify the registry CA certificate.
Copy the certificate to
`/var/snap/k8s/common/etc/containerd/hosts.d/docker.io/ca.crt`,

Then add your config in
`/var/snap/microk8s/current/args/certs.d/docker.io/hosts.toml`:

```
[host."https://10.100.100.100:5000"]
capabilities = ["pull", "resolve"]
ca = "/var/snap/k8s/common/etc/containerd/hosts.d/docker.io/ca.crt"
```

#### Container Runtime Option C: Side-load images

This is only required if you chose to
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

You can add and remove nodes as described in the
[add-and-remove-nodes tutorial][nodes].

After a while, confirm that all the cluster nodes show up in
the output of the `sudo k8s kubectl get node` command.

<!-- LINKS -->

[Core20]: https://canonical.com/blog/ubuntu-core-20-secures-linux-for-iot
[svc-ports]: /snap/explanation/services-and-ports.md
[proxy]: /snap/howto/proxy.md
[upstream-imgs]: https://github.com/canonical/k8s-snap/blob/main/build-scripts/hack/upstream-images.yaml
[sync-images]: https://github.com/canonical/k8s-snap/blob/main/build-scripts/hack/sync-images.sh
[regsync]: https://github.com/regclient/regclient
[regctl]: https://github.com/regclient/regclient/blob/main/docs/regctl.md
[nodes]: /snap/tutorial/add-remove-nodes.md
[squid]: https://www.squid-cache.org/
