# How to manage images

{{product}} uses the containerd runtime to manage images, which in turn uses
runc to run containers. Some users may need to use containerd's capabilities
directly, for example if they do not have access to our default image registry,
they may wish import images manually.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have installed the {{product}} snap
  (see How-to [Install {{product}} from a snap][snap-install-howto]).
- You have a bootstrapped cluster

## Paths and confinement modes

Since {{product}} needs containerd to operate, we bundle the containerd binary
at `/snap/k8s/current/bin/ctr`. Although the containerd binary is in the snap
installation folder, the location of the containerd socket will depend on the
confinement mode of the snap.

| Description                | Strict path                                             | Classic path                      |
|----------------------------|----------------------------------------------------|------------------------------|
| Config Directory           | /var/snap/k8s/common/etc/containerd                | /etc/containerd              |
| Extra config directory     | /var/snap/k8s/common/etc/containerd/conf.d         | /etc/containerd/conf.d       |
| Registry config directory  | /var/snap/k8s/common/etc/containerd/hosts.d        | /etc/containerd/hosts.d      |
| Root directory             | /var/snap/k8s/common/var/lib/containerd            | /var/lib/containerd          |
| Socket directory           | /var/snap/k8s/common/run/containerd                | /run/containerd              |
| Socket path                | /var/snap/k8s/common/run/containerd/containerd.sock| /run/containerd/containerd.sock |
| State directory            | /var/snap/k8s/common/run/containerd                | /run/containerd              |

## Listing all images

{{product}} imports all images into the `k8s.io` namespace. When you're
interacting with containerd, make sure you always reference the `k8s.io`
namespace.

You can view a list of all images {{product}} has registered with containerd:

```
root@k8s:~# sudo /snap/k8s/current/bin/ctr --address /run/containerd/containerd.sock --namespace k8s.io images list
```


## Pulling images

You can import images manually using the following command:

```
sudo /snap/k8s/current/bin/ctr --address /run/containerd/containerd.sock --namespace k8s.io images pull docker.io/library/hello-world:latest
```

Verify it was pulled:

```
root@k8s:~# sudo /snap/k8s/current/bin/ctr --address /run/containerd/containerd.sock --namespace k8s.io images list | grep hello

docker.io/library/hello-world:latest
...
```

<!-- LINKS -->

[snap-install-howto]: ./install/snap
