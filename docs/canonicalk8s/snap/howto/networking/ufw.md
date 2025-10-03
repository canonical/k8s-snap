# How to configure Uncomplicated Firewall (UFW)

This how-to presents a set of firewall rules/guidelines
that should be considered when setting up {{product}}.
These rules may be incompatible with your network setup,
so we recommend you review and tune them to match your needs.

## Prerequisites

This guide assumes the following:

- An ubuntu machine where {{product}} is installed or will be installed.
- Root or sudo access to the machine.

## Install UFW 

Uncomplicated Firewall needs to be configured on all nodes of {{product}}.
To do this, try:

```sh
sudo apt update
sudo apt install ufw
```

To verify UFW is installed try:

```sh
sudo ufw status verbose
```

To maintain SSH access to the machine, allow `OpenSSH` through
UFW before enabling the firewall:

```sh
sudo ufw allow OpenSSH
```

## Allow forwarding

Package forwarding is needed because containers typically live in isolated
networks and expect the host-to-route traffic between their internal network
and the outside world to be allowed.

### Enable IP forwarding

If you want the forwarding rules to persist through system reboots,
enable IP forwarding by editing `/etc/sysctl.conf`:

```sh
net.ipv4.ip_forward=1
```

Otherwise, use `sysctl` directly to apply the forwarding rules immediately
without rebooting the system:

```sh
sudo sysctl -w net.ipv4.ip_forward=1
```

### Set forwarding rules

Set UFW forwarding rules using one of the following methods.

`````{tabs}
````{group-tab} Allow system wide
Packet forwarding can be allowed system wide by editing `/etc/default/ufw`
and adding:

```sh
DEFAULT_FORWARD_POLICY="ACCEPT"
```
````

````{group-tab} By subnet
A less permissive approach would be to allow forward traffic only between
the subnets of the pods and the hosts.
For example, assuming the pods CIDR is `10.1.0.0/16` and the cluster nodes
are in `10.0.20/24`, you could:

```sh
sudo ufw route allow from 10.1.0.0/16 to 10.0.20.0/24
sudo ufw route allow from 10.1.0.0/16 to 10.1.0.0/16
```
````
`````

## Allow access to the Kubernetes services

Services such as CoreDNS require access to the Kubernetes API
server listening on port 6443.
 
Allow traffic on port 6443 with:

``` sh
sudo ufw allow 6443/tcp
```

Services such as the metrics-server need access to the kubelet,
controller manager and kube scheduler to query for metrics.

Kubelet runs on all nodes, so allow traffic on port 10250 on all nodes:

```sh
sudo ufw allow 10250/tcp
```

The kube-controller-manager and kube-scheduler only run on
the control plane, therefore permit traffic on ports 10257 and 10259
of the control plane nodes:

```sh
sudo ufw allow 10257/tcp
sudo ufw allow 10259/tcp
```

## Allow cluster formation

To form a High Availability (HA) cluster the datastore used by Kubernetes
(Dqlite/etcd) needs to establish a direct connection among its peers.

`````{tabs}
````{group-tab} etcd
Allow traffic on port 2380 on control plane nodes with etcd:

```sh
sudo ufw allow 2380/tcp
```
````

````{group-tab} Dqlite
Allow traffic on port 9000 on control plane nodes with Dqlite:

```sh
sudo ufw allow 9000/tcp
```
````
`````

Cluster formation is overseen by a Kubernetes daemon running on all nodes
on port 6400.

Open port 6400 to permit cluster formation traffic:

```sh
sudo ufw allow 6400/tcp
```

## Enable CNI specific communication

When using the default network plugin (Cilium),
consider the following firewall rules.

Allow cluster-wide Cilium agent health checks and VXLAN traffic:

```sh
sudo ufw allow 4240/tcp
sudo ufw allow 8472/udp
```

## Enable UFW

Now you are ready to enable UFW with:

```sh
sudo ufw enable
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

The logs will show you which packets are dropped, their destination
and source as well as the protocol used and the destination port.
This information helps you identify any other ports or services
you need to enable within UFW.

After troubleshooting, keep the resources used by UFW to a minimum
by disabling logging:

```sh
sudo ufw logging off
```


<!-- LINKS -->

[ports-and-services]: ../../reference/ports-and-services
