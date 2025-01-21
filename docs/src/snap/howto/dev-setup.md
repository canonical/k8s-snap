# Setting a local development environment

Follow this guide to prepare a local development environment that can be
used to build and test {{product}}.

Please use Ubuntu 22.04 or later, either locally or inside a virtual machine.

## Using a virtual machine

We strongly recommend using a clean virtual machine, which avoids conflicts
with other software and ensures that the environment can easily be cleaned up
or reproduced.

One option is to use Multipass, which supports a variety of operating systems.
Use the following document to [setup Multipass].

We recommend allocating at least 40GB of disk space, 4 vcpus and 8GB of RAM.

Note that by default [snapcraft] uses a VM when building the snap. Either make
sure that nested virtualization is enabled or configure it to use LXD instead,
as documented by the following sections.

## Getting the source code

The {{product}} source code is hosted on Github: https://github.com/canonical/k8s-snap.

Branches:

| branch                          | description                        |
|---------------------------------|------------------------------------|
| main                            | latest development                 |
| release/1.XX                    | stable version                     |
| autoupdate/moonray              | "Moonray" flavor (Calico based)    |
| autoupdate/strict               | hardened, strictly confined flavor |
| autoupdate/release-1.XX-moonray | stable "Moonray" version           |
| autoupdate/release-1.XX-strict  | stable "strict" flavor             |

For development purposes, you'll most probably want to use the "main" branch.

```bash
git clone https://github.com/canonical/k8s-snap
```

## Snapcraft

Before building the Kubernetes snap, we need to install Snapcraft:

```bash
sudo snap install snapcraft --classic
```

## Building the snap using Multipass

To initiate the build process, call ``snapcraft`` from the git clone directory.

Note that Snapcraft uses Multipass VMs to build the snap. If Multipass is not
already installed, it will prompt for its automatic installation.

```bash
cd k8s-snap
snapcraft
```

### Building the snap using LXD

Snapcraft can also be configured to use LXD, which significantly speeds up
the build process by avoiding the virtualization overhead and using more
resources than would normally be allocated through Multipass.

Use the following to install and initialize LXD:

```bash
sudo snap install lxd
sudo lxd init  # pass --auto for automatic configuration
```

Build the snap using LXD by issuing the following command:

```bash
snapcraft --use-lxd
```

Be aware that LXD may interfere with Docker installations, see the
[lxd network troubleshooting guide] for more details and possible
workarounds.

## Installing the snap

The default flavor of the snap expects ``classic`` confinement, so make sure to
specify the ``--classic`` flag when installing it. At the same time, since our
fresh build is unsigned, we also need to pass the ``--dangerous`` flag to allow
installation.

```bash
sudo snap install k8s_*.snap --classic --dangerous
```

Once the snap is installed, you can use ``k8s bootstrap`` to spin up a new
cluster or ``k8s join-cluster`` to join an existing one.

### Specifying the listening address

The local IP addresses may change, especially when installing ``k8s-snap``
directly on your development machine. For this reason, you may configure
it to use the ``localhost`` address:

```bash
sudo k8s boostrap --address=127.0.0.1
```

### Using LXD

We recommend running {{product}} in an isolated environment, such as a virtual
machine or LXD container.

Please see the [LXD installation guide] for more details on how to run
{{product}} inside LXD containers.

Also note that you can use the ``lxc file push`` command to copy your freshly
built snap to the LXD container.

### Specifying containerd path

In classic confinement mode, {{product}} uses the default containerd paths.
This means that a {{product}} installation will conflict with any existing
system configuration where containerd is already installed. For example,
if you have Docker installed, or another Kubernetes distribution that uses
containerd.

If using an isolated environment is not possible, you may specify a custom
containerd path like so:

```bash
cat <<EOF | sudo k8s bootstrap --file -
containerd-base-dir: $containerdBaseDir
EOF
```

### Conflicting Docker iptables routes

By default, Docker sets the ``FORWARD`` policy to drop, which can affect LXD
and Multipass connectivity.

See the [lxd network troubleshooting guide] for more details and possible
workarounds.

### Inspecting Dqlite databases

By default, {{product}} uses the [k8s-dqlite] datastore instead of etcd, which
is based on [Dqlite] and [kine].

At the same, the ``k8sd`` cluster management service stores its own internal
data in Dqlite.

See the [Dqlite configuration reference] for additional details.

<!--LINKS -->
[setup Multipass]: ./install/multipass.md
[snapcraft]: https://snapcraft.io/docs/snapcraft-setup
[LXD installation guide]: ./install/lxd.md
[lxd network troubleshooting guide]: https://documentation.ubuntu.com/lxd/en/latest/howto/network_bridge_firewalld/#prevent-connectivity-issues-with-lxd-and-docker
[k8s-dqlite]: https://github.com/canonical/k8s-dqlite
[Dqlite configuration reference]: ../reference/dqlite.md
[Dqlite]: https://dqlite.io/
[kine]: https://github.com/k3s-io/kine/

