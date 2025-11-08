# How to configure Uncomplicated Firewall (UFW)

This how-to presents a set of firewall rules/guidelines
that should be considered when setting up {{product}}.
These rules may be incompatible with some network setups,
so we recommend you review and tune them to match your needs.

## Prerequisites

This guide assumes the following:

- An ubuntu machine where {{product}} is or will be installed.
- Root or sudo access to the machine.

## Install UFW 

Install Uncomplicated Firewall:

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

## Allow packet forwarding

Packet forwarding is needed because containers typically live in isolated
networks and expect the host to route traffic between their internal network
and the outside world.

To enable IP forwarding:

```sh
sudo sed -i 's|^.*net.ipv4.ip_forward.*$|net.ipv4.ip_forward=1|' /etc/sysctl.conf
sudo sysctl -p
```

## Set forwarding rules

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
A less permissive approach would be to allow forwarding traffic only between
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

Allow access the API server on control plane nodes:
 
``` sh
sudo ufw allow 6443/tcp
```

Allow access to kubelet on all nodes:

```sh
sudo ufw allow 10250/tcp
```

Allow access to kube-controller-manager and kube-scheduler on
control plane nodes (e.g. for metrics gathering):

```sh
sudo ufw allow 10257/tcp
sudo ufw allow 10259/tcp
```

## Allow cluster formation

To form a High Availability (HA) cluster the datastore used by Kubernetes
(etcd or k8s-dqlite) needs to establish a direct connection among its peers.

`````{tabs}
````{group-tab} etcd
Allow access to etcd on all control plane nodes:

```sh
sudo ufw allow 2380/tcp
```
````

````{group-tab} k8s-dqlite
Allow access to k8s-dqlite on all control plane nodes:

```sh
sudo ufw allow 9000/tcp
```
````
`````

Allow access to the {{product}} daemon running on all nodes (required for cluster formation):

```sh
sudo ufw allow 6400/tcp
```

## Enable CNI specific communication

Allow the cluster-wide Cilium agent health checks and VXLAN traffic on all nodes:

```sh
sudo ufw allow 4240/tcp
sudo ufw allow 8472/udp
```

## Enable UFW

Now enable UFW:

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
