# Configuring proxy settings for K8s

Canonical Kubernetes packages a number of utilities (eg curl, helm) which need
to fetch resources they expect to find on the internet. In a constrained
network environment, such access is usually controlled through proxies.

On Ubuntu and other Linux operating systems, proxies are configured through
system-wide environment variables defined in the `/etc/environment` file.

## Adding proxy configuration

Edit the `/etc/environment` file and add the relevant URLs

```{note} It is important to add whatever address ranges are used by the
 cluster itself to the `NO_PROXY` and `no_proxy` variables.
```

For example, assume we have a proxy running at `http://squid.internal:3128` and
we are using the networks `10.0.0.0/8`,`192.168.0.0/16` and `172.16.0.0/12`. We
would edit the environment (`/etc/environment`) file to include these lines:

```
HTTPS_PROXY=http://squid.internal:3128
HTTP_PROXY=http://squid.internal:3128
NO_PROXY=10.0.0.0/8,192.168.0.0/16,127.0.0.1,172.16.0.0/12
https_proxy=http://squid.internal:3128
http_proxy=http://squid.internal:3128
no_proxy=10.0.0.0/8,192.168.0.0/16,127.0.0.1,172.16.0.0/12
```

Note that you may need to restart for these settings to take effect.
