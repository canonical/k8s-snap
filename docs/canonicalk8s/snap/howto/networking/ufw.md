# How to configure Uncomplicated Firewall (UFW)

In this how-to we present a set of firewall rules/guidelines you should
consider when setting up {{product}}. Be aware that these rules
may be incompatible with your network setup, you may find them too permissive
or too restrictive. We recommend you review them and tune them to match
your needs. Also, be aware that the firewall rules need to be reviewed
for each service hosted in Kubernetes as there might be special requirements.

## Prerequisites

This guide assumes the following:

- A machine with Ubuntu.
- You have root or sudo access to the machine.

## Install and enable UFW

To install Uncomplicated Firewall:

```sh
sudo apt update
sudo apt install ufw
```

To verify UFW is installed try:

```sh
sudo ufw status verbose
```

If you need to maintain ssh access to the machine, make sure you configure UFW to allow `OpenSSH` before enabling it:
configure UFW properly before enabling it:

```sh
sudo ufw allow OpenSSH
```

Now you are finally ready to enable:

```sh
sudo ufw enable
```

## Allow package forwarding

Forwarding is needed because containers typically live in isolated networks
and expect the host-to-route traffic between their internal network and the
outside world to be allowed.

First edit `/etc/default/ufw` and allow UFW to route/forward packets:

```sh
DEFAULT_FORWARD_POLICY="ACCEPT"
```

Enable IP forwarding by editing `/etc/sysctl.conf` (or use `sysctl` directly):

```sh
net.ipv4.ip_forward=1
```

Apply immediately:

```sh
sudo sysctl -w net.ipv4.ip_forward=1
```

Reload UFW:

```sh
sudo ufw reload
```

## Allow access to the Kubernetes services

Services such as for example CoreDNS require access to the Kubernetes API
server listening on port 6443.
 
Allow traffic on port 6443 with:

``` sh
sudo ufw allow 6443/tcp
```

Services such as the metrics-server need access to the kubelet,
controller manager and kube scheduler to query for metrics.
Kubelet runs on all nodes, while the kube-controller-manager and
kube-scheduler run only on the control plane nodes:

Allow traffic on port 10250 on all nodes:

```sh
sudo ufw allow 10250/tcp
```

Allow traffic on ports 10257 and 10259 on control plane nodes:

```sh
sudo ufw allow 10257/tcp
sudo ufw allow 10259/tcp
```

## Allow cluster formation

To form an HA cluster the datastore used by Kubernetes (dqlite/etcd) needs
to establish a direct connection among its peers. In dqlite this is done
through port 9000 while on etcd port 2380 is used.

Allow traffic on port 9000 on control plane nodes with dqlite:

```sh
sudo ufw allow 9000/tcp
```

Allow traffic on port 2380 on control plane nodes with etcd:

```sh
sudo ufw allow 2380/tcp
```

Cluster formation is overseen by a Kubernetes daemon running on all nodes
on port 6400.

Allow traffic on port 6400:

```sh
sudo ufw allow 6400/tcp
```

## Allow CNI specific communication

The default CNI used in {{product}} is Cilium.
Unless you are not disabling this network plugin and deploying your own,
you should consider the following firewall rules.

Allow cluster-wide Cilium agent health checks and VXLAN traffic:

```sh
sudo ufw allow 4240/tcp
sudo ufw allow 8472/udp
```

## UFW troubleshooting

The [ports-and-services] page has a list of all ports {{product}} uses.

To inspect a failing service you can enable logging:

```sh
sudo ufw logging on
```

Monitor the firewall logs with:

```sh
tail -f /var/log/ufw.log
```

At the end disable logging:

```sh
sudo ufw logging off
```


<!-- LINKS -->

[ports-and-services]: ../../reference/ports-and-services
