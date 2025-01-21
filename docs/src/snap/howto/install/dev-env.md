# Install {{product}} in development environments

We recommend testing {{product}} in an isolated environment such as a clean
virtual machine or LXD container.

If you choose to install {{product}} directly on your development machine,
please take note of the following considerations.

## Containerd conflicts

In classic confinement mode, {{product}} uses the default containerd paths.
This means that a {{product}} installation will conflict with any existing
system configuration where containerd is already installed. For example,
if you have Docker installed, or another Kubernetes distribution that uses
containerd.

You may specify a custom containerd path like so:

```bash
cat <<EOF | sudo k8s bootstrap --file -
containerd-base-dir: $containerdBaseDir
EOF
```

## Changing IP addresses

The local IP addresses of your development machine are likely to change,
for example after joining a different wi-fi network.

In this case, you may configure {{product}} to use the ``localhost`` address:

```bash
sudo k8s boostrap --address=127.0.0.1
```

## Conflicting Docker iptables rules

Docker can interfere with LXD and Multipass installations, setting the global
``FORWARD`` policy to drop.

See the [lxd network troubleshooting guide] for more details and possible
workarounds.

<!--LINKS -->
[lxd network troubleshooting guide]: https://documentation.ubuntu.com/lxd/en/latest/howto/network_bridge_firewalld/#prevent-connectivity-issues-with-lxd-and-docker
