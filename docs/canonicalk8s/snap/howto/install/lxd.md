# How to install {{product}} in LXD

{{product}} can also be installed inside an LXD virtual machine. This is a
great way, for example, to test out clustered {{product}} without the
need for multiple physical hosts.

Why an LXD virtual machine and not a container? LXD is about to remove the
support for privileged containers and some Kubernetes services, such as the
Cilium CNI, cannot run inside unprivileged containers. Furthermore, by using
virtual machine we ensure that the Kubernetes environment is well isolated.

## Install LXD

You can install [LXD] via snaps:

```
sudo snap install lxd
sudo lxd init
```

## Start an LXD VM for {{product}}

We can now create the VM that {{product}} will run in.

```
lxc launch ubuntu:22.04 k8s --vm -c limits.cpu=2 -c limits.memory=4GB
```

## Install {{product}} in an LXD VM

First, we’ll need to install {{product}} within the VM.

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- lxd start -->
:end-before: <!-- lxd end -->
```

```{note}
Substitute your desired channel in the above command. Find the
available channels with `snap info k8s` and see the [channels][]
explanation page for more details on channels, tracks and versions.
```

## Access {{product}} services within LXD

Assuming you accepted the [default bridged
networking][default-bridged-networking] when you initially setup LXD, there is
minimal effort required to access {{product}} services inside the LXD VM.

Simply note the interface IP address from the command:

<!-- markdownlint-disable -->
```
lxc list k8s
```
```
+------+---------+------------------------+------------------------------------------------+-----------------+-----------+
| NAME |  STATE  |         IPV4           |                     IPV6                       |      TYPE       | SNAPSHOTS |
+------+---------+------------------------+------------------------------------------------+-----------------+-----------+
| k8s  | RUNNING | 10.122.174.30 (enp5s0) | fd42:80c6:c3e:445a:216:3eff:fe8d:add9 (enp5s0) | VIRTUAL-MACHINE | 0         |
+------+---------+------------------------+------------------------------------------------+-----------------+-----------+
```

<!-- markdownlint-restore -->

and use this to access services running inside the VM.

### Expose services to the VM

You’ll need to expose the deployment or service to the VM itself before
you can access it via the LXD VM’s IP address. This can be done using
`k8s kubectl expose`. This example will expose the deployment’s port 80 to a
port assigned by Kubernetes.

We will use [Microbot] as it provides a simple HTTP endpoint
to expose. These steps can be applied to any other deployment.

First, initialise the k8s cluster with

```
lxc exec k8s -- sudo k8s bootstrap
```

Now, let’s deploy Microbot (please note this image only works on `x86_64`).

```
lxc exec k8s -- sudo k8s kubectl create deployment \
  microbot --image=dontrebootme/microbot:v1
```

Then check that the deployment has come up.

```
lxc exec k8s -- sudo k8s kubectl get all
```

...should return an output similar to:

<!-- markdownlint-disable -->
```
NAME                            READY   STATUS    RESTARTS   AGE
pod/microbot-6d97548556-hchb7   1/1     Running   0          21m

NAME                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
service/kubernetes         ClusterIP   10.152.183.1     <none>        443/TCP        21m

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/microbot   1/1     1            1           21m

NAME                                  DESIRED   CURRENT   READY   AGE
replicaset.apps/microbot-6d97548556   1         1         1       21m
```
<!-- markdownlint-restore -->

Now that Microbot is up and running, let's make it accessible to the LXD
VM by using the `expose` command.

<!-- markdownlint-disable -->

```
lxc exec k8s -- sudo k8s kubectl expose deployment microbot --type=NodePort --port=80 --name=microbot-service
```

<!-- markdownlint-restore -->

We can now get the assigned port. In this example, it’s `32750`:

```
lxc exec k8s -- sudo k8s kubectl get service microbot-service
```

...returns output similar to:

```
NAME               TYPE       CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
microbot-service   NodePort   10.152.183.188   <none>        80:32750/TCP   27m
```

With this, we can access Microbot from our host but using the VM’s
address that we noted earlier.

```
curl 10.122.174.30:32750
```

## Stop/remove the VM

The `k8s` VM you created will keep running in the background until it is
either stopped or the host computer is shut down. You can stop the running
VM at any time by running:

```
lxc stop k8s
```

And it can be permanently removed with:

```
lxc delete k8s
```

[LXD]: https://canonical.com/lxd
[default-bridged-networking]: https://ubuntu.com/blog/lxd-networking-lxdbr0-explained
[Microbot]: https://github.com/dontrebootme/docker-microbot
[channels]: ../../explanation/channels
