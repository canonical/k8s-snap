---
myst:
  html_meta:
    description: "How to enable, disable, and configure the default DNS service in Canonical Kubernetes for cluster-internal service discovery."
---

# How to use default DNS

<!-- SPREAD SUITE: snap_bootstrapped -->

{{product}} includes a default DNS (Domain Name System) which is
essential for internal cluster communication. When enabled, the DNS facilitates
service discovery by assigning each service a DNS name. When disabled, you can
integrate a custom DNS solution into your cluster. Learn more about the
default DNS in the [DNS explanation](/snap/explanation/networking.md#dns).

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine.
- You have a bootstrapped {{product}} cluster (see the [Getting
  Started][getting-started-guide] guide).

## Check DNS status

Find out whether DNS is enabled or disabled with the following command:

```
sudo k8s status
```

<!-- SPREAD
sudo k8s status | grep "dns:                      enabled at"
-->

The default state for the cluster is `dns enabled`.

## Enable DNS

To enable DNS, run:

```
sudo k8s enable dns
```

<!-- SPREAD
sudo k8s get dns | grep "enabled: true"
-->

For more information on this command, run:

```
sudo k8s help enable
```

<!-- SPREAD
sudo k8s help enable | grep "Enable one of network, dns"
-->

## Configure DNS

Discover your configuration options by running:

```
sudo k8s get dns
```

<!-- SPREAD
sudo k8s get dns | grep "upstream-nameservers"
sudo k8s get dns | grep "cluster-domain"
sudo k8s get dns | grep "service-ip"
-->

You should see three options:

- `upstream-nameservers`: DNS servers used to forward known entries
- `cluster-domain`: the cluster domain name
- `service-ip`: the cluster IP to be assigned to the DNS service

Set a new DNS server IP for forwarding known entries:

<!-- SPREAD SKIP -->

```
sudo k8s set dns.upstream-nameservers=<new-ips>
```

<!-- SPREAD SKIP END -->

<!-- SPREAD 
sudo k8s set dns.upstream-nameservers=8.8.8.8
sudo k8s get dns | grep "8.8.8.8"
-->

Change the cluster domain name:

<!-- SPREAD SKIP -->

```
sudo k8s set dns.cluster-domain=<new-domain-name>
```

<!-- SPREAD SKIP END -->

<!-- SPREAD 
sudo k8s set dns.cluster-domain=k8s.testing
sudo k8s get dns | grep "cluster-domain: k8s.testing"
-->

Assign a new cluster IP to the DNS service (DNS must be disabled in order to 
do this):

<!-- SPREAD SKIP -->

```
sudo k8s set dns.service-ip=<new-cluster-ip>
```

<!-- SPREAD SKIP END -->

<!-- SPREAD 
sudo k8s disable dns
sudo k8s set dns.service-ip="10.152.183.254"
sudo k8s enable dns
sudo k8s get dns | grep "service-ip: 10.152.183.254"
-->

Replace `<new-ip>`, `<new-domain-name>`, and `<new-cluster-ip>` with the
desired values for your DNS configuration.

A cluster IP you set is never changed automatically. If another service takes
it while DNS is disabled, DNS stays disabled when you enable it again and its
status names the service holding the IP. Delete that service, or disable DNS
and set a different IP. Set an empty value to let Kubernetes choose the IP:

<!-- SPREAD SKIP -->

```
sudo k8s set dns.service-ip=""
```

<!-- SPREAD SKIP END -->

## Disable DNS

{{product}} also allows you to disable the built-in DNS,
if you desire a custom solution:

```{warning} Disabling DNS will disrupt internal cluster communication. Ensure
a suitable custom DNS solution is in place before disabling. You can re-enable
DNS at any point. If another service took the DNS cluster IP in the meantime,
the DNS service gets a new IP, and pods created before that keep using the old
one until they are restarted. If you set the IP yourself, it is kept instead,
see [Configure DNS](#configure-dns).
```

<!-- SPREAD SKIP -->

```
sudo k8s disable dns
```

<!-- SPREAD SKIP END -->

<!-- SPREAD
if [ "${SCENARIO_MODE}" != "true" ]; then
  sudo k8s disable dns
  sudo k8s get dns | grep "enabled: false"
fi
-->

For more information on this command, execute:

```
sudo k8s help disable
```

<!-- SPREAD
sudo k8s help disable | grep "Disable one of network, dns"
-->

<!-- LINKS -->

[getting-started-guide]: /snap/tutorial/getting-started