# How to use default DNS

{{product}} includes a default DNS (Domain Name System) which is
essential for internal cluster communication. When enabled, the DNS facilitates
service discovery by assigning each service a DNS name. When disabled, you can
integrate a custom DNS solution into your cluster.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine.
- You have a bootstrapped {{product}} cluster (see the [Getting
  Started][getting-started-guide] guide).

## Check DNS status

Find out whether DNS is enabled or disabled with the following command:

```
sudo k8s status
```

The default state for the cluster is `dns enabled`.

## Enable DNS

To enable DNS, run:

```
sudo k8s enable dns
```

For more information on this command, run:

```
sudo k8s help enable
```

## Configure DNS

Discover your configuration options by running:

```
sudo k8s get dns
```

You should see three options:

- `upstream-nameservers`: DNS servers used to forward known entries
- `cluster-domain`: the cluster domain name
- `service-ip`: the cluster IP to be assigned to the DNS service

Set a new DNS server IP for forwarding known entries:

```
sudo k8s set dns.upstream-nameservers=<new-ips>
```

Change the cluster domain name:

```
sudo k8s set dns.cluster-domain=<new-domain-name>
```

Assign a new cluster IP to the DNS service:

```
sudo k8s set dns.service-ip=<new-cluster-ip>
```

Replace `<new-ip>`, `<new-domain-name>`, and `<new-cluster-ip>` with the
desired values for your DNS configuration.

## Disable DNS

{{product}} also allows you to disable the built-in DNS,
if you desire a custom solution:

``` {warning} Disabling DNS will disrupt internal cluster communication. Ensure
a suitable custom DNS solution is in place before disabling. You can re-enable
DNS at any point, and your cluster will return to normal functionality.```
```

```
sudo k8s disable dns
```

For more information on this command, execute:

```
sudo k8s help disable
```

<!-- LINKS -->

[getting-started-guide]: /snap/tutorial/getting-started
