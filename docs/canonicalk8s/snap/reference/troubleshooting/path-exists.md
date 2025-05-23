# The path required for the containerd socket already exists

## Problem

{{product}} tries to create the containerd socket to manage containers,
but it fails because the socket file already exists, which indicates another
installation of containerd on the system.

## Explanation

In classic confinement mode, {{product}} uses the default containerd
paths. This means that a {{product}} installation will conflict with
any existing system configuration where containerd is already installed.
For example, if you have Docker installed, or another Kubernetes distribution
that uses containerd.

## Solution

We recommend running {{product}} in an isolated environment, for this purpose,
you can create a LXD container for your installation. See
[Install {{product}} in LXD][lxd-install] for instructions.

As an alternative, you may specify a custom containerd path like so:

```bash
cat <<EOF | sudo k8s bootstrap --file -
containerd-base-dir: $containerdBaseDir
EOF
```

<!-- LINKS -->

[lxd-install]: /snap/howto/install/lxd.md
