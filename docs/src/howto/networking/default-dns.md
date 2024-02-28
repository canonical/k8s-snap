# How to use default DNS

Canonical Kubernetes includes a default DNS (Domain Name System) essential for internal communication within your cluster. 
When enabled, the DNS facilitates service discovery
by assigning each service a DNS name. 

## What you'll need

This guide assumes the following:

- You are installing on Ubuntu 22.04 or later, **or** another OS which supports
  snap packages (see [snapd support])
- You have root or sudo access to the machine
- You have an internet connection
- The target machine has sufficient memory and disk space. To accommodate
  workloads, we recommend a system with ***at least*** 20G of disk space and 4G of
  memory.
- You have Canonical Kubernetes installed and a bootstraped cluster. (See: [getting-started-guide](#TODO))

## Is DNS enabled?

Find out wether you have enabled DNS with the following command:

```bash
sudo k8s status
```

The default state for the cluster is `dns enabled`.

## Enabling and disabling DNS
To enable DNS, run:

```bash
sudo k8s enable dns
```

Canonical Kubernetes also allows you to disable the built-in DNS, 
if you desire a custom solution:

```bash
sudo k8s disable dns
```

For more information on these two commands, execute:

```bash
sudo k8s help enable
```

Or for disabling:

```bash
sudo k8s help disable
```
To continue with the `Configuring DNS` section enable dns again.

## Configuring DNS
Discover your configuration options by running:
```bash
sudo k8s set dns –help
```
You should see three options:
- upstream-dns: the dns server used to forward known entries
- cluster-domain: the cluster domain name
- dns-ip: the cluster ip to be assigned to the dns service

Set a new DNS server IP for forwarding known entries:
```bash
sudo k8s set dns --upstream-dns=<new-ip>
```
Change the cluster domain name:
```bash
sudo k8s set dns --cluster-domain=<new-domain-name>
```
Assign a new cluster IP to the DNS service.
```bash
sudo k8s set dns --dns-ip=<new-cluster-ip>
```
Replace `<new-ip>`, `<new-domain-name>`, and `<new-cluster-ip>` with the desired values for your DNS configuration.


<!-- LINKS -->

[Component Upgrades]: #TODO
[getting-started-guide]: (#TODO)

