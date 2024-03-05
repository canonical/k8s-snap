# Install Canonical Kubernetes in LXD

Canonical Kubernetes can also be installed inside an LXD container. This is a
great way, for example, to test out clustered Canonical Kubernetes without the
need for multiple physical hosts.

## Installing LXD 

You can install [LXD] via snaps:

```no-highlight
sudo snap install lxd
sudo lxd init
```

## Add the Canonical Kubernetes LXD profile

Canonical Kubernetes requires some specific settings to work within LXD (these
are explained in more detail below). These can be applied using a custom
profile. The first step is to create a new profile:
```no-highlight
lxc profile create k8s
```

Once created, we’ll need to add the rules. 
Get our pre-defined profile rules from github and save them as k8s.profile.
```no-highlight
wget https://raw.githubusercontent.com/canonical/k8s-snap/main/tests/e2e/lxd-profile.yaml -O k8s.profile
```

To pipe the content of the file into the k8s LXD profile, run:

```no-highlight
cat k8s.profile | lxc profile edit k8s
```

Remove the copied content from your directory:

```no-highlight
rm k8s.profile
```

## Start an LXD container for Canonical Kubernetes
We can now create the container that Canonical Kubernetes will run in.

```no-highlight
lxc launch -p default -p k8s ubuntu:22.04 k8s
```

```{note} This command uses the `default` profile, for any existing system
settings (networking, storage, etc.) before also applying the `k8s` profile -
the order is important.

## Install Canonical Kubernetes in an LXD container
First, we’ll need to install Canonical Kubernetes within the container.

```no-highlight
lxc exec k8s -- sudo snap install k8s --classic
```

## Access Canonical Kubernetes services within LXD
Assuming you left the [default bridged networking][default-bridged-networking] when you initially setup LXD,
there is minimal effort required to access Canonical Kubernetes services inside
the LXD container.

Simply note the `eth0` interface IP address from

```no-highlight
lxc list k8s

+------+---------+----------------------+----------------------------------------------+-----------+-----------+
| NAME |  STATE  |         IPV4         |                     IPV6                     |   TYPE    | SNAPSHOTS |
+------+---------+----------------------+----------------------------------------------+-----------+-----------+
| k8s  | RUNNING | 10.122.174.30 (eth0) | fd42:80c6:c3e:445a:216:3eff:fe8d:add9 (eth0) | CONTAINER | 0         |
+------+---------+----------------------+----------------------------------------------+-----------+-----------+
```

and use this to access services running inside the container.

### Expose services to the container

You’ll need to expose the deployment or service to the container itself before
you can access it via the LXD container’s IP address. This can be done using
`k8s kubectl expose`. This example will expose the deployment’s port 80 to a
port assigned by Kubernetes.

In this example, we will use [Microbot] as it provides a simple HTTP endpoint
to expose. These steps can be applied to any other deployment.

First, initialize the k8s cluster with 

```no-highlight
lxc exec k8s -- sudo k8s bootstrap
```

Now, let’s deploy Microbot (please note this image only works on `x86_64`).

```no-highlight
lxc exec k8s -- sudo k8s kubectl create deployment microbot --image=dontrebootme/microbot:v1
```

Then check that the deployment has come up.

```no-highlight
lxc exec k8s -- sudo k8s kubectl get all

NAME                            READY   STATUS    RESTARTS   AGE
pod/microbot-6d97548556-hchb7   1/1     Running   0          21m

NAME                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
service/kubernetes         ClusterIP   10.152.183.1     <none>        443/TCP        21m

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/microbot   1/1     1            1           21m

NAME                                  DESIRED   CURRENT   READY   AGE
replicaset.apps/microbot-6d97548556   1         1         1       21m
```

Now that Microbot is up and running, let's make it accessible to the LXD container by using the `expose` command.

```no-highlight
lxc exec k8s -- sudo k8s kubectl expose deployment microbot --type=NodePort --port=80 --name=microbot-service
```

We can now get the assigned port. In this example, it’s `32750`.

```no-highlight
lxc exec k8s -- sudo k8s kubectl get service microbot-service

NAME               TYPE       CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
microbot-service   NodePort   10.152.183.188   <none>        80:32750/TCP   27m
```

With this, we can access Microbot from our host but using the container’s
address that we noted earlier.

```no-highlight
curl 10.122.174.30:32750
```

## Stop/Remove the container

The `k8s` container you created will keep running in the background until it is either stopped or the host computer is shut down. You can stop the running container at any time by running:

```no-highlight
lxc stop k8s
```

And it can be permanently removed with:

```no-highlight
lxc delete k8s
```

## Explanation of custom LXD rules

**boot.autostart: “true”**: Always start the container when LXD starts. This is
needed to start the container when the host boots.

**linux.kernel_modules**: Comma separated list of kernel modules to load before
starting the container

**lxc.apparmor.profile=unconfined**: Disable [AppArmor]. Allow the container to
talk to a bunch of subsystems of the host (e.g. `/sys`). By default AppArmor
will block nested hosting of containers, however Kubernetes needs to host
Containers. Containers need to be confined based on their profiles thus we rely
on confining them and not the hosts. If you can account for the needs of the
Containers you could tighten the AppArmor profile instead of disabling it
completely, as suggested in [1].

**lxc.cap.drop=**: Do not drop any capabilities [2]. For justification see
above.

**lxc.mount.auto=proc:rw sys:rw**: Mount proc and sys rw [3]. For privileged
containers, lxc over-mounts part of /proc as read-only to avoid damage to the
host. Kubernetes will complain with messages like `Failed to start
ContainerManager open /proc/sys/kernel/panic: permission denied`

**lxc.cgroup.devices.allow=a**: "a" stands for "all." This means that the
container is allowed to access all devices. It's a wildcard character
indicating permission for all devices. For justification see above.

**security.nesting: “true”**: Support running LXD (nested) inside the
container.

**security.privileged: “true”**: Runs the container in privileged mode, not
using kernel namespaces [4, 5]. This is needed because hosted Containers may
need to access for example storage devices (See comment in [6]).

## Citations
[1] Container nesting: https://stgraber.org/2012/05/04/lxc-in-ubuntu-12-04-lts/.
[2] Capabilities https://stgraber.org/2014/01/01/lxc-1-0-security-features/.
[3] Mount proc and sys in privileged containers: https://github.com/lxc/lxd/issues/3042.
[4] Unprivileged containers: https://unix.stackexchange.com/questions/177030/what-is-an-unprivileged-lxc-container/177031#177031.
[5] Privileged containers: http://blog.benoitblanchon.fr/lxc-unprivileged-container/.
[6] Lxc Security https://wiki.ubuntu.com/LxcSecurity.

<!-- LINKS -->

[LXD]: https://canonical.com/lxd
[default-bridged-networking]: https://ubuntu.com/blog/lxd-networking-lxdbr0-explained
[Microbot]: https://github.com/dontrebootme/docker-microbot
[AppArmor]: https://apparmor.net/