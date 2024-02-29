# How to use default DNS

Canonical Kubernetes includes a default DNS (Domain Name System) essential for internal communication within your cluster. 
When enabled, the DNS facilitates service discovery by assigning each service a DNS name. 
When disabled, you can integrate a custom DNS solution into your cluster.


## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstraped Canonical Kubernetes cluster (see the [getting-started-guide]).

## Check DNS status

Find out whether DNS is enabled or disabled with the following command:

```bash
sudo k8s status
```

The default state for the cluster is `dns enabled`.

## Enable DNS
To enable DNS, run:

```bash
sudo k8s enable dns
```

For more information on this command, run:

```bash
sudo k8s help enable
```

## Configure DNS
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
sudo k8s set dns.upstream-dns=<new-ip>
```
Change the cluster domain name:
```bash
sudo k8s set dns.cluster-domain=<new-domain-name>
```
Assign a new cluster IP to the DNS service.
```bash
sudo k8s set dns.dns-ip=<new-cluster-ip>
```
Replace `<new-ip>`, `<new-domain-name>`, and `<new-cluster-ip>` with the desired values for your DNS configuration.

## Disable DNS

Canonical Kubernetes also allows you to disable the built-in DNS, 
if you desire a custom solution:

``` {warning} Do not disable DNS unless you have a replacement configured. Disabling DNS will disrupt internal cluster communication. Ensure a suitable custom DNS solution is in place before disabling. You can re-enable dns at any point, and your cluster will return to normal functionality.```

```bash
sudo k8s disable dns
```

For more information on this command, execute:

```bash
sudo k8s help disable
```

<!-- LINKS -->

[getting-started-guide]: ../../../tutorial/getting-started

