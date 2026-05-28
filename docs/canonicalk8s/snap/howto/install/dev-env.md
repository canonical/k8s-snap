# Install {{product}} in development environments

<!-- SPREAD SUITE: snap_clean -->

<!-- SPREAD
sudo snap install docker
sudo snap install k8s --classic --channel=1.35-classic/stable
-->

We recommend testing {{product}} in an isolated environment such as a clean
virtual machine.

You **can** install {{product}} directly on your development machine. But if
you choose to do so, please take note of the following considerations.

## Containerd 

### Conflicts

{{product}} runs its own containerd service, which will use the standard
containerd-related paths by default (`/run/containerd`, `/var/lib/containerd`,
`/etc/containerd`). Note that these default paths are important for various
upstream projects and operators (e.g.: GPU Operator).

If you already have Docker installed, or another Kubernetes instance that uses
containerd directly installed on the host, this can cause various conflicts
with {{product}}.

But, if necessary, {{product}} can be configured to use a custom containerd
path, like so:

<!-- SPREAD
export containerdBaseDir=/ck8s/etc/containerd
-->

```
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

<!-- SPREAD
sudo k8s status --wait-ready
ps -ef | grep kubelet | grep container-runtime-endpoint=/ck8s/etc/containerd/k8s-containerd
-->

### State Directory on tmpfs — Disk Pressure & ErrImagePull

When using a custom containerd, if it is configured to use a state directory on
`tmpfs` (e.g., `/run/containerd`), ensure that the `tmpfs` mount has sufficient 
space for operations like image layer unpacking. Insufficient space can cause:

- Pod failures with `ErrImagePull`
- Node taints such as `node.kubernetes.io/disk-pressure`

To check the available space on the tmpfs:

<!-- SPREAD SKIP -->

```
df -h /run
```

If the space is low and you're experiencing these issues, you can temporarily 
increase the size of the tmpfs mount to see if it resolves the problem:

```
sudo mount -o remount,size=10G /run
```

<!-- SPREAD SKIP END -->

```{note}
This change is not persistent and will reset on reboot.
```

## Changing IP addresses

The local IP addresses of your development machine are likely to change,
for example after joining a different Wi-Fi network.

In this case, you may configure {{product}} to use the ``localhost`` address:

<!-- SPREAD SKIP -->

```bash
sudo k8s bootstrap --address=127.0.0.1
```

<!-- SPREAD SKIP END -->

## Conflicting Docker iptables rules

Docker can interfere with LXD and Multipass installations, setting the global
``FORWARD`` policy to drop.

See the [LXD network troubleshooting guide] for more details and possible
workarounds.

<!--LINKS -->
[LXD network troubleshooting guide]: https://documentation.ubuntu.com/lxd/en/latest/howto/network_bridge_firewalld/#prevent-connectivity-issues-with-lxd-and-docker
