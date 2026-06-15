# Install {{product}} in development environments

We recommend testing {{product}} in an isolated environment such as a clean
virtual machine.

You **can** install {{product}} directly on your development machine. But if
you choose to do so, please take note of the following considerations.

## Containerd 

### Conflicts

{{product}} runs its own containerd service, which will use the standard
containerd-related paths by default (`/run/containerd`, `/var/lib/containerd`,
`/etc/containerd`). If containerd is already installed at these paths by 
another application (e.g. Docker), the bootstrap will fail. To resolve this, provide
a base directory for the files to be installed at using the `--containerd-base-dir`
flag or by providing it in the bootstrap config yaml like below:

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
By doing this, all containerd files will be stored under the parent
directory specified by the flag (e.g. if `--containerd-base-dir=/ck8s`,
containerd files will be `/ck8s/etc/containerd`,
`/ck8s/var/run/containerd/containerd.sock`, etc.)

```{note}
It is strongly recommeneded that a non-temporary directory is chosen for 
`containerd-base-dir`, or the cluster will break on reboot when these
files are cleared.
```

### State Directory on tmpfs — Disk Pressure & ErrImagePull

If you choose to use a tmpfs base directory for containerd,
make sure that it has sufficient space for operations like 
image layer unpacking. Insufficient space can cause:

- Pod failures with `ErrImagePull`
- Node taints such as `node.kubernetes.io/disk-pressure`

To check the available space on the tmpfs:

```bash
df -h /run
```

If the space is low and you're experiencing these issues, you can temporarily 
increase the size of the tmpfs mount to see if it resolves the problem:

```bash
sudo mount -o remount,size=10G /run
```

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
