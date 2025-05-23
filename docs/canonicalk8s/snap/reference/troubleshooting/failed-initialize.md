# Kubelet Error: `failed to initialize top level QOS containers`

## Problem


This is related to the `kubepods` cgroup not getting the cpuset controller up on
the kubelet. kubelet needs a feature from cgroup and the kernel may not be set
up appropriately to provide the cpuset feature.

```
E0125 00:20:56.003890    2172 kubelet.go:1466] "Failed to start ContainerManager" err="failed to initialise top level QOS containers: root container [kubepods] doesn't exist"
```

## Explanation

An excellent deep-dive of the issue exists at
[kubernetes/kubernetes #122955][kubernetes-122955].

Commenter [@haircommander][] [states][kubernetes-122955-2020403422]
>  basically: we've figured out that this issue happens because libcontainer
>  doesn't initialise the cpuset cgroup for the kubepods slice when the kubelet
>  initially calls into it to do so. This happens because there isn't a cpuset
>  defined on the top level of the cgroup. however, we fail to validate all of
>  the cgroup controllers we need are present. It's possible this is a
>  limitation in the dbus API: how do you ask systemd to create a cgroup that
>  is effectively empty?

>  if we delegate: we are telling systemd to leave our cgroups alone, and not
>  remove the "unneeded" cpuset cgroup.


## Solution

This is in the process of being fixed upstream via
[kubernetes/kubernetes #125923][kubernetes-125923].

In the meantime, the best solution is to create a `Delegate=yes` configuration
in systemd.

```bash
mkdir -p /etc/systemd/system/snap.k8s.kubelet.service.d
cat /etc/systemd/system/snap.k8s.kubelet.service.d/delegate.conf <<EOF
[Service]
Delegate=yes
EOF
reboot
```

<!-- LINKS -->

[kubernetes-122955]: https://github.com/kubernetes/kubernetes/issues/122955
[kubernetes-125923]: https://github.com/kubernetes/kubernetes/pull/125923
[kubernetes-122955-2020403422]: https://github.com/kubernetes/kubernetes/issues/122955#issuecomment-2020403422
[@haircommander]: https://github.com/haircommander