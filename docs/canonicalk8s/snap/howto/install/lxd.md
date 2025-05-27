# How to install {{product}} in LXD

{{product}} can also be installed inside an LXD virtual machine. This is a
great way, for example, to test out clustered {{product}} without the
need for multiple physical hosts.

Why an LXD virtual machine and not a container?
In order to run certain Kubernetes services, such as the Cilium CNI, the LXD
container would need to be a [privileged container]. While this is possible, it
is not the recommended pattern as it allows the root
user in the container to be the root user on the host. Also, newer versions of
Ubuntu and systemd require operations (such as mounting to the `/proc`
directory) that cannot be safely handled with privileged containers. By using
virtual machines, we ensure that the Kubernetes environment remains well
isolated.

## Install LXD

Install [LXD] via snaps:

```
sudo snap install lxd
sudo lxd init
```

## Start an LXD VM for {{product}}

Create the VM that {{product}} will run in.

```
lxc launch ubuntu:22.04 k8s-vm --vm -c limits.cpu=2 -c limits.memory=4GB
```

## Install {{product}} in an LXD VM

Install {{product}} within the VM.

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
lxc list k8s-vm
```
```
+--------+---------+------------------------+------------------------------------------------+-----------------+-----------+
|  NAME  |  STATE  |         IPV4           |                     IPV6                       |      TYPE       | SNAPSHOTS |
+--------+---------+------------------------+------------------------------------------------+-----------------+-----------+
| k8s-vm | RUNNING | 10.122.174.30 (enp5s0) | fd42:80c6:c3e:445a:216:3eff:fe8d:add9 (enp5s0) | VIRTUAL-MACHINE | 0         |
+--------+---------+------------------------+------------------------------------------------+-----------------+-----------+
```

<!-- markdownlint-restore -->

and use this to access services running inside the VM.

### Expose services to the VM

You’ll need to expose the deployment or service to the VM itself before
you can access it via the LXD VM's IP address. This can be done using
`k8s kubectl expose`. This example will expose the deployment’s port 80 to a
port assigned by Kubernetes.

We will use [Microbot] as it provides a simple HTTP endpoint
to expose. These steps can be applied to any other deployment.

First, initialize the k8s cluster with

```
lxc exec k8s-vm -- sudo k8s bootstrap
```

Now, let’s deploy Microbot (please note this image only works on `x86_64`).

```
lxc exec k8s-vm -- sudo k8s kubectl create deployment \
  microbot --image=dontrebootme/microbot:v1
```

Then check that the deployment has come up.

```
lxc exec k8s-vm -- sudo k8s kubectl get all
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
lxc exec k8s-vm -- sudo k8s kubectl expose deployment microbot --type=NodePort --port=80 --name=microbot-service
```

<!-- markdownlint-restore -->

Get the assigned port. In this example, it’s `32750`:

```
lxc exec k8s-vm -- sudo k8s kubectl get service microbot-service
```

...returns output similar to:

```
NAME               TYPE       CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
microbot-service   NodePort   10.152.183.188   <none>        80:32750/TCP   27m
```

With this, access Microbot from our host but using the VM's address that we
noted earlier.

```
curl 10.122.174.30:32750
```

## Stop/remove the VM

The `k8s-vm` VM you created will keep running in the background until it is
either stopped or the host computer is shut down. Stop the running VM at any
time by running:

```
lxc stop k8s-vm
```

And it can be permanently removed with:

```
lxc delete k8s-vm
```

[LXD]: https://canonical.com/lxd
[default-bridged-networking]: https://documentation.ubuntu.com/lxd/en/latest/reference/network_bridge/
[Microbot]: https://github.com/dontrebootme/docker-microbot
[channels]: ../../explanation/channels
[privileged container]: https://documentation.ubuntu.com/server/how-to/containers/lxd-containers/index.html#uid-mappings-and-privileged-containers
