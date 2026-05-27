# How to configure Uncomplicated Firewall (UFW)

<!-- SPREAD SUITE: snap_clean -->

This how-to presents a set of firewall rules/guidelines that should be
considered when setting up {{product}}. These rules may be incompatible
with some network setups, so we recommend you review and tune them to
match your needs.

## Prerequisites

This guide assumes the following:

- An Ubuntu machine where {{product}} is or will be installed.
- Root or sudo access to the machine.

## Install UFW

Install Uncomplicated Firewall:

```
sudo apt update
sudo apt install ufw
```

Verify that UFW is installed:

```
sudo ufw status verbose
```

<!-- SPREAD
sudo ufw status verbose | grep "Status: inactive"
-->

To maintain SSH access to the machine, allow `OpenSSH` through UFW
before enabling the firewall:

```
sudo ufw allow OpenSSH
```

## Firewall rules for all nodes

Apply the following rules on all control plane and worker nodes.

### Allow packet forwarding

Packet forwarding is needed because containers typically live in
isolated networks and expect the host to route traffic between their
internal network and the outside world.

To enable IP forwarding:

```
sudo sed -i 's|^.*net.ipv4.ip_forward.*$|net.ipv4.ip_forward=1|' /etc/sysctl.conf
sudo sysctl -p
```

<!-- SPREAD 
sudo sysctl -p | grep "net.ipv4.ip_forward = 1"
-->

### Set forwarding rules

Set UFW forwarding rules using one of the following methods.

<!-- SPREAD SKIP -->

`````{tab-set}
````{tab-item} Allow system wide
Packet forwarding can be allowed system wide by editing `/etc/default/ufw`
and changing `DEFAULT_FORWARD_POLICY` to:

```
DEFAULT_FORWARD_POLICY="ACCEPT"
```
````

````{tab-item} By subnet
A less permissive approach would be to allow forwarding traffic only
between the subnets of the pods and the hosts. For example, assuming the
pods CIDR is `10.1.0.0/16` and the cluster nodes are in `10.0.20/24`, you
could:

```
sudo ufw route allow from 10.1.0.0/16 to 10.0.20.0/24
sudo ufw route allow from 10.1.0.0/16 to 10.1.0.0/16
```
````
`````

<!-- SPREAD SKIP END -->

<!-- SPREAD
sudo grep -qE '^\s*#?\s*DEFAULT_FORWARD_POLICY=' /etc/default/ufw \
  && sudo sed -i -E 's|^\s*#?\s*DEFAULT_FORWARD_POLICY=.*|DEFAULT_FORWARD_POLICY="ACCEPT"|' /etc/default/ufw \
  || echo 'DEFAULT_FORWARD_POLICY="ACCEPT"' | sudo tee -a /etc/default/ufw
-->

### Allow access to kubelet

```
sudo ufw allow 10250/tcp
```

### Allow access to the {{product}} daemon

Allow access to the {{product}} daemon (required for
cluster formation):

```
sudo ufw allow 6400/tcp
```

### Enable CNI communication

Allow the cluster-wide Cilium agent health checks and VXLAN traffic on
all nodes:

```
sudo ufw allow 4240/tcp
sudo ufw allow 8472/udp
```

## Firewall rules for control plane nodes only

Apply the following rules on all control plane nodes.

### Allow Kubernetes control plane services

Allow access to the API server:

```
sudo ufw allow 6443/tcp
```

Allow access to kube-controller-manager and kube-scheduler
(e.g. for metrics gathering):

```
sudo ufw allow 10257/tcp
sudo ufw allow 10259/tcp
```

### Allow datastore communication

To form a High Availability (HA) cluster, etcd
needs to establish direct connections among control plane nodes.
Allow access to the etcd peer and client port:

```
sudo ufw allow 2380/tcp
sudo ufw allow 2379/tcp
```

## Enable UFW

Now enable UFW:

<!-- SPREAD SKIP -->
```
sudo ufw enable
```
<!-- SPREAD SKIP END -->
<!-- SPREAD
echo "y" | sudo ufw enable
# Confirm all settings are correct
sudo ufw status verbose | grep "Status: active"
sysctl net.ipv4.ip_forward | grep "net.ipv4.ip_forward = 1"
sudo ufw status | grep "OpenSSH"
sudo ufw status | grep "10250/tcp"
sudo ufw status | grep "6400/tcp"
sudo ufw status | grep "4240/tcp"
sudo ufw status | grep "8472/udp"
sudo ufw status | grep "6443/tcp"
sudo ufw status | grep "10257/tcp"
sudo ufw status | grep "10259/tcp"
sudo ufw status | grep "2380/tcp"
sudo ufw status | grep "2379/tcp"
-->

## UFW troubleshooting

The [ports-and-services] page has a list of all ports {{product}} uses.

To inspect a failing service you can enable logging:

```
sudo ufw logging on
```

Monitor the firewall logs with:

<!-- SPREAD SKIP -->
```
tail -f /var/log/ufw.log
```

<!-- SPREAD SKIP END -->

The logs will show you which packets are dropped, their destination and
source as well as the protocol used and the destination port. This
information helps you identify any other ports or services you need to
enable within UFW.

After troubleshooting, keep the resources used by UFW to a minimum by
disabling logging:

```
sudo ufw logging off
```

<!-- LINKS -->

[ports-and-services]: ../../reference/ports-and-services
