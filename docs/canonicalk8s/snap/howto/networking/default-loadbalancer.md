---
myst:
  html_meta:
    description: "How to enable, disable and configure the Canonical Kubernetes load balancer to expose services with external IPs from a configured IP address pool."
---

# How to use the default load balancer

<!-- SPREAD SUITE: snap_bootstrapped -->

{{product}} includes a default load balancer. As this is not an
essential service for all deployments, it is not enabled by default. This guide
explains how to configure and enable the `load-balancer`.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine.
- You have a bootstrapped {{product}} cluster (see the [Getting
  Started][getting-started-guide] guide).

## Check the status and configuration

Find out whether load balancer is enabled or disabled with the following
command:

```
sudo k8s status
```

<!-- SPREAD
sudo k8s status | grep "load-balancer:            disabled"
-->

The load balancer is not enabled by default.

To check the current configuration of the `load-balancer`, run the following:

```
sudo k8s get load-balancer
```

<!-- SPREAD
sudo k8s get load-balancer | grep "enabled: false"
sudo k8s get load-balancer | grep "cidrs"
sudo k8s get load-balancer | grep "l2-mode"
sudo k8s get load-balancer | grep "l2-interfaces"
sudo k8s get load-balancer | grep "bgp-mode"
sudo k8s get load-balancer | grep "bgp-local-asn"
sudo k8s get load-balancer | grep "bgp-peer-address"
sudo k8s get load-balancer | grep "bgp-peer-asn"
sudo k8s get load-balancer | grep "bgp-peer-port"
-->

This should output a list of values like this:

- `enabled`: if set to true, load-balancer is enabled
- `cidrs` - a list containing [CIDR] or IP address ranges for the
load balancer's address pool
- `l2-mode` - whether L2 mode (failover) is turned on
- `l2-interfaces` - optional list of interfaces to announce services over
  (defaults to all)
- `bgp-mode` - whether BGP mode is active.
- `bgp-local-asn` - the local Autonomous System Number (ASN)
- `bgp-peer-address` - the peer address
- `bgp-peer-asn` - ASN of the peer network
- `bgp-peer-port` - port used on the BGP peer

These values are configured using the `k8s set` command, e.g.:

```
sudo k8s set load-balancer.l2-mode=true
```

<!-- SPREAD
sudo k8s get load-balancer | grep "l2-mode: true"
-->

Note that for the BGP mode, it is necessary to set ***all*** the values
simultaneously. E.g.

```
sudo k8s set load-balancer.bgp-mode=true load-balancer.bgp-local-asn=64512 load-balancer.bgp-peer-address=10.0.10.63 load-balancer.bgp-peer-asn=64512 load-balancer.bgp-peer-port=7012
```

<!-- SPREAD
sudo k8s get load-balancer | grep "bgp-mode: true"
sudo k8s get load-balancer | grep "bgp-local-asn: 64512"
sudo k8s get load-balancer | grep "bgp-peer-address: 10.0.10.63"
sudo k8s get load-balancer | grep "bgp-peer-asn: 64512"
sudo k8s get load-balancer | grep "bgp-peer-port: 7012"
-->

## Enable the load balancer

To enable the load balancer, run:

```
sudo k8s enable load-balancer
```

<!-- SPREAD
sudo k8s get load-balancer | grep "enabled: true"
-->

You can now confirm it is working by running:

```
sudo k8s status
```

<!-- SPREAD
sudo k8s get load-balancer | grep "enabled: true"
-->

```{important}
If you run `k8s status` soon after enabling the load balancer in BGP mode,
`k8s status` might report errors. Please wait a few moments for the load balancer to finish deploying and try again.
```

## Disable the load balancer

The default load balancer can be disabled again with:

```
sudo k8s disable load-balancer
```

<!-- SPREAD
sudo k8s get load-balancer | grep "enabled: false"
-->

## Next Step

- Learn more in the [Load-balancer explanation](/snap/explanation/networking.md#load-balancer) page.

<!-- LINKS -->
[CIDR]: https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing
[getting-started-guide]: /snap/tutorial/getting-started

