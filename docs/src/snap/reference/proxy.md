# Proxy environment variables

{{product}} uses the standard system-wide environment variables to
control access through proxies.

On Ubuntu and other Linux operating systems, proxies are configured through
system-wide environment variables defined in the `/etc/environment` file.

- **HTTPS_PROXY**
- **HTTP_PROXY**
- **NO_PROXY**
- **https_proxy**
- **http_proxy**
- **no_proxy**

## No-proxy CIDRS

When configuring proxies, it is important to note that there are always some
CIDRs which need to be excluded and added to the `no-proxy` lists. For
{{product}} these are:

- The range used by Kubernetes services (defaults to **10.152.183.0/24**)
- The range used by the Kubernetes pods (defaults to **10.1.0.0/16**)

And it is also important to exclude the local network to maintain access to any
local traffic.

## Configuring

For the `k8s` snap, proxy configuration is controlled by editing the
`etc/environment` file mentioned above. There is an example in the
[How to guide for configuring proxies for the k8s snap][].

For charms deployed by Juju, proxies are managed by configuring the model. See
the [How to guide for configuring proxies for k8s charms][] for an example of
how to set these.

<!-- LINKS -->

[How to guide for configuring proxies for the k8s snap]: ../howto/proxy
[How to guide for configuring proxies for k8s charms]: ../../charm/howto/proxy
