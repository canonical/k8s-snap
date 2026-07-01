---
myst:
  html_meta:
    description: Troubleshoot and resolve common issues that come from running the Canonical Kubernetes snap in a dev environment.
---
# Install {{product}} in development environments

<!-- SPREAD SUITE: snap_clean -->

<!-- SPREAD
sudo snap install docker
sudo snap install k8s --classic --channel=1.35-classic/stable
# Tear down docker on exit
trap 'sudo snap remove docker --purge' EXIT
# Start doc test
-->

We recommend testing {{product}} in an isolated environment such as a clean
virtual machine.

You **can** install {{product}} directly on your development machine. But if
you choose to do so, please take note of the following considerations.

## Containerd

### Conflicts

{{product}} runs its own containerd service, which will use the standard
containerd-related paths by default (`/run/containerd`, `/var/lib/containerd`,
`/etc/containerd`). If containerd is already installed at these paths by
another application (e.g. Docker), the bootstrap will fail. To resolve this,
provide a base directory for the files to be installed at by setting
`containerd-base-dir` in the bootstrap config YAML:

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

By doing this, all containerd files will be stored under the parent
directory specified by `containerd-base-dir`. For example, if
`containerd-base-dir` is set to `/ck8s`, containerd files will be stored at
paths such as `/ck8s/etc/containerd` and
`/ck8s/run/containerd/containerd.sock`.

```{note}
It is strongly recommended that a non-temporary directory is chosen for
`containerd-base-dir`, or the cluster will break on reboot when these
files are cleared. The path provided should be an absolute path
to a directory dedicated to just these files.
```

### State directory on tmpfs

If you choose to use a tmpfs base directory for containerd,
make sure that it has sufficient space for operations like
image layer unpacking. Insufficient space can cause:

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

However, these changes will be cleared on reboot.

### Home directory usage

It is strongly recommended that you use a system-level directory like `\opt\{path}` 
or `\sys\{path}` instead of a user-home directory like `\home\{path}`. Although the
cluster will likely continue to work, using a user-home directory will cause lots of
temporary bind mount files to populate the folder that are ordinarily hidden. There
may also be issues with user-level disk encryption or restricted file permissions 
that can also cause a failed bootstrap.

### External consumption

When changing the containerd install path, make sure that the configurations of
external consumers of {{product}} such as operators are also updated. For example,
in the GPU operator, you will have to update the Helm chart to include the new
containerd paths.

```
helm install gpu-operator nvidia/gpu-operator -n gpu-operator \
 --set operator.defaultRuntime=containerd \
 --set toolkit.env[0].name=CONTAINERD_CONFIG \
 --set toolkit.env[0].value={custom_containerd_dir}/etc/containerd/config.toml \
 --set toolkit.env[1].name=CONTAINERD_SOCKET \
 --set toolkit.env[1].value={custom_containerd_dir}/run/containerd/containerd.sock
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
