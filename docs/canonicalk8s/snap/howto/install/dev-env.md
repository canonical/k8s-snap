# Install {{product}} in development environments

We recommend testing {{product}} in an isolated environment such as a clean
virtual machine.

You **can** install {{product}} directly on your development machine. But if
you choose to do so, please take note of the following considerations.

## Containerd conflicts

{{product}} runs its own containerd service, which will use the standard
containerd-related paths by default (`/run/containerd`, `/var/lib/containerd`,
`/etc/containerd`). Note that these default paths are important for various
upstream projects and operators (e.g.: GPU Operator).

If you already have Docker installed, or another Kubernetes instance that uses
containerd directly installed on the host, this can cause various conflicts
with {{product}}.

But, if necessary, {{product}} can be configured to use a custom containerd
path, like so:

```bash
cat <<EOF | sudo k8s bootstrap --file -
containerd-base-dir: $containerdBaseDir
cluster-config:
  network:
    enabled: true
  dns:
    enabled: true
  local-storage:
    enabled: true
EOF
```

Any non-temporary directory can be chosen for `containerd-base-dir`
(e.g.: `/ck8s`). {{product}} will then use this base directory for the
containerd-related files (e.g.: `/ck8s/etc/containerd`,
`/ck8s/var/run/containerd/containerd.sock`, etc.).

## Changing IP addresses

The local IP addresses of your development machine are likely to change,
for example after joining a different Wi-Fi network.

In this case, you may configure {{product}} to use the ``localhost`` address:

```bash
sudo k8s bootstrap --address=127.0.0.1
```

## Conflicting Docker iptables rules

Docker can interfere with LXD and Multipass installations, setting the global
``FORWARD`` policy to drop.

See the [LXD network troubleshooting guide] for more details and possible
workarounds.

<!--LINKS -->
[LXD network troubleshooting guide]: https://documentation.ubuntu.com/lxd/en/latest/howto/network_bridge_firewalld/#prevent-connectivity-issues-with-lxd-and-docker
