# How to install {{product}} in air-gapped environments

There are situations where it is necessary or desirable to run {{product}}
on a machine that is not connected to the internet. Based on different degrees
of separation from the network, different solutions are offered to accomplish
this goal. This guide documents any necessary extra preparation for air-gap
deployments, as well the steps that are needed to successfully deploy
{{product}} in such environments.

## Prepare for deployment

In preparation for the offline deployment download the {{product}} snap,
fulfill the networking requirements based on your scenario and handle images for
workloads and {{product}} features.

### Download the {{product}} snap

From a machine with access to the internet download the
`k8s` and `core22` snap with:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- offline start -->
:end-before: <!-- offline end -->
:append: sudo snap download core22 --basename core22
```

Besides the snaps, this will also download the corresponding assert files which
are necessary to verify the integrity of the packages.

### Check network requirements

Air-gap deployments are typically associated with a number of constraints and
restrictions when it comes to the networking connectivity of the machines.
Below we discuss the requirements that the deployment needs to fulfill.

#### Cluster node communication

<!-- TODO: Add Services and Ports Doc -->

Ensure that all cluster nodes are reachable from each other.

<!-- Refer to [Services and ports][svc-ports] used for a list of all network
ports used by {{product}}.  -->

#### Default Gateway

In cases where the air-gap environment does not have a default Gateway,
add a placeholder default route on the `eth0` interface using the following
command:

```
ip route add default dev eth0
```

```{note}
Ensure that `eth0` is the name of the default network interface used for
pod-to-pod communication.
```

The placeholder gateway will only be used by the Kubernetes services to
know which interface to use, actual connectivity to the internet is not
required. Ensure that the placeholder gateway rule survives a node reboot.

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

### Access images

All workloads in a Kubernetes cluster are run as an OCI image.
Kubernetes needs to be able to fetch these images and load them
into the container runtime.
For {{product}}, it is also necessary to fetch the images used
by its features (network, DNS, etc.) as well as any images that are
needed to run specific workloads.


#### List images

If the `k8s` snap is already installed,
list the images in use with the following command:

```
k8s list-images
```

The output will look similar to the following:

```
ghcr.io/canonical/cilium-operator-generic:1.15.2-ck2
ghcr.io/canonical/cilium:1.15.2-ck2
ghcr.io/canonical/coredns:1.11.1-ck4
ghcr.io/canonical/k8s-snap/pause:3.10
ghcr.io/canonical/k8s-snap/sig-storage/csi-node-driver-registrar:v2.10.1
ghcr.io/canonical/k8s-snap/sig-storage/csi-provisioner:v5.0.1
ghcr.io/canonical/k8s-snap/sig-storage/csi-resizer:v1.11.1
ghcr.io/canonical/k8s-snap/sig-storage/csi-snapshotter:v8.0.1
ghcr.io/canonical/metrics-server:0.7.0-ck2
ghcr.io/canonical/rawfile-localpv:0.8.2
```

A list of images can also be found in the `images.txt` file when the
downloaded `k8s` snap is unsquashed.

Please ensure that the images used by workloads are tracked as well.

#### Choose how to access images

You must select how the container runtime accesses OCI images in your air-gapped
installation:

```{note}
The image options are presented in the order of increasing complexity
of implementation.
It may be helpful to combine these options for different scenarios.
```

`````{tab-set}

````{tab-item} HTTP proxy

In many cases, the nodes of the air-gap deployment may not have direct access
to upstream registries, but can reach them through the use of an HTTP proxy.

Create or edit the
`/etc/systemd/system/snap.k8s.containerd.service.d/http-proxy.conf`
file on each node and set the appropriate `http_proxy`, `https_proxy` and
`no_proxy` variables as described in the
[adding proxy configuration section][proxy].

````

````{tab-item} Private registry mirrors
In case regulations and/or network constraints do not permit the cluster nodes
to access any upstream image registry, it is typical to deploy a private
registry mirror. This is an image registry service that contains all the
required OCI Images (e.g.
[registry](https://distribution.github.io/distribution/),
[Harbor](https://goharbor.io/) or any other OCI registry) and is reachable from
all cluster nodes.

This requires three steps:

1. Deploy and secure the registry service. Please follow the instructions for
   the desired registry deployment.
2. Using [regsync][regsync], load all images from the upstream source and
   push to your registry mirror.
3. Configure the {{product}} container runtime (`containerd`) to load
   images from the private registry mirror instead of the upstream source. This
   will be described in the [container runtime](
   #configure-container-runtime) section under configure registry mirrors.

 To load images into the private registry, a machine is needed with access to
any upstream registries (e.g. `docker.io`) and the private mirror.

##### Load images with regsync

We recommend using [`regsync`][regsync] to copy images from the upstream registry
to your private registry. First, install the tool:

```
go install github.com/regclient/regclient/cmd/regsync@latest
```

Create a `sync-images.yaml` file that maps the output from
`k8s list-images` to the private registry mirror and specify a mirror for
ghcr.io that points to the registry.

```
sync:
  - source: ghcr.io/canonical/k8s-snap/pause:3.10
    target: '{{ env "MIRROR" }}/canonical/k8s-snap/pause:3.10'
    type: image
  ...
```

After creating the `sync-images.yaml` file, use `regsync` to sync the
images. Assuming your registry mirror is at `http://10.10.10.10:5050`, run:

```
USERNAME="$username" PASSWORD="$password" MIRROR="10.10.10.10:5050" \
regsync once -c path/to/sync-images.yaml
```
````

````{tab-item} Side-load images
Image side-loading is the process of loading all required OCI images directly
into the container runtime, so they do not have to be fetched at runtime.

To create a bundle of images, use the [`regctl`][regctl] tool. First, install it:

```
go install github.com/regclient/regclient/cmd/regctl@latest
```

Then export the images:

```
regctl image export ghcr.io/canonical/k8s-snap/pause:3.10 \
--name ghcr.io/canonical/k8s-snap/pause:3.10 --platform=local > pause.tar
```

```{note}
The flag `--name` is essential. Without it, the exported image will be imported with a hash only,
and the image with the particular tag required by k8s will not be found.
```

   Upon choosing this option, place all images under
`/var/snap/k8s/common/images` and they will be picked up by containerd.
````
`````

## Deploy {{product}}

Once you've completed all the preparatory steps for your air-gapped cluster,
you can proceed with the deployment.

### Install {{product}}

Transfer the following files to the target node:

- `k8s.snap`
- `k8s.assert`
- `core22.snap`
- `core22.assert`

On the target node, run the following command to install the Kubernetes snap:

```
sudo snap ack core22.assert && sudo snap install ./core22.snap
sudo snap ack k8s.assert && sudo snap install ./k8s.snap --classic
```

Repeat the above for all nodes of the cluster.

### Configure container runtime

Based on the image access type you chose in the step
[choose how to access images](#choose-how-to-access-images), configure the
container runtime to fetch images properly:


`````{tab-set}

````{tab-item} HTTP proxy
The HTTP proxy should be set from the earlier access images via HTTP proxy step. Refer to the
[adding proxy configuration section][proxy] for more help if necessary.
````

````{tab-item} Private registry mirrors
You should have already set up a registry mirror, as explained in the
preparation section on the private registry mirror. Complete the following
instructions on all nodes. For each upstream registry that needs mirroring,
create a `hosts.toml` file. Here's an example that configures
`http://10.10.10.10:5050` as a mirror for `ghcr.io`:



#### HTTP registry mirror

In `/etc/containerd/hosts.d/ghcr.io/hosts.toml`
add the configuration:

```
[host."http://10.10.10.10:5050"]
capabilities = ["pull", "resolve"]
```


#### HTTPS registry mirror

HTTPS requires the additionally specification of the registry CA certificate.
Copy the certificate to
`/etc/containerd/hosts.d/ghcr.io/ca.crt`.
Then add the configuration in
`/etc/containerd/hosts.d/ghcr.io/hosts.toml`:

```
[host."https://10.10.10.10:5050"]
capabilities = ["pull", "resolve"]
ca = "/var/snap/k8s/common/etc/containerd/hosts.d/ghcr.io/ca.crt"
```

#### Nested path registry mirror

If the registry mirrors the images at a nested path, configure the hosts.toml
with an `override_path = true`.

```
# /etc/containerd/hosts.d/ghcr.io/hosts.toml
server = https://10.10.10.10:5050/v2/my/own/ghcr.io
[host."https://10.10.10.10:5050/v2/my/own/ghcr.io"]
capabilities = ["pull", "resolve"]
ca = "/var/snap/k8s/common/etc/containerd/hosts.d/ghcr.io/ca.crt"
override_path = true
```

The image `ghcr.io/canonical/coredns:1.12.0-ck1` would be fetched via https
from `my.mirror.io/my/own/ghcr.io/canonical/coredns:1.12.0-ck1`.

````

````{tab-item} Side-load images
This is only required if choosing to side-load images. Make sure
that the directory `/var/snap/k8s/common/images` directory exists, then copy
all `$image.tar` to that directory, such that containerd automatically picks
them up and imports them when it starts. Copy the `images.tar` file(s) to
`/var/snap/k8s/common/images`. Repeat this step for all cluster nodes.
````

`````

### Bootstrap cluster

Now, bootstrap the cluster and replace `MY-NODE-IP` with the IP of the node
by running the command:

```
sudo k8s bootstrap --address MY-NODE-IP
```

After a while, confirm that all the node show up in the output of the
`sudo k8s kubectl get node` command.

Adding nodes requires the same steps to be repeated but instead of
bootstrapping, you would need to [generate a join token] and [join the node]
to the cluster.

<!-- LINKS -->

[proxy]: /snap/howto/networking/proxy.md
[regsync]: https://github.com/regclient/regclient/blob/main/docs/regsync.md
[regctl]: https://github.com/regclient/regclient/blob/main/docs/regctl.md
[nodes]: /snap/tutorial/add-remove-nodes.md
[squid]: https://www.squid-cache.org/
[generate a join token]: /snap/reference/commands/#k8s-get-join-token
[join the node]:/snap/reference/commands/#k8s-join-cluster
