# Proxy environment variables

Canonical Kubernetes uses the standard system-wide environment variables to
controll access through proxies. For operation in a proxy environment, the
following should be set.

On Ubuntu and other Linux operating systems, proxies are configured through
system-wide environment variables defined in the `/etc/environment` file.

- **HTTPS_PROXY**
- **HTTP_PROXY**
- **NO_PROXY**
- **https_proxy**
- **http_proxy**
- **no_proxy**

See the [Proxy how to guide][] for an example of how to set these.

<!-- LINKS -->

[Proxy how to guide]: /snap/howto/proxy
