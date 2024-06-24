# Configure proxy settings for K8s

Canonical Kubernetes packages a number of utilities (eg curl, helm) which need
to fetch resources they expect to find on the internet. In a constrained
network environment, such access is usually controlled through proxies.

On Ubuntu and other Linux operating systems, proxies are configured through
system-wide environment variables defined in the `/etc/environment` file.

To set up a proxy using squid follow the
[how-to-install-a-squid-server][squid] tutorial.

## Adding proxy configuration for the k8s snap

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

```{note} The **10.152.183.0/24** CIDR needs to be covered in the juju-no-proxy
   list as it is the Kubernetes service CIDR. Without this any pods will not be 
   able to reach the cluster's kubernetes-api. You should also exclude the range
   used by pods (which defaults to **10.1.0.0/16**) and any required
   local networks.
```

## Adding proxy configuration for the k8s charms

Proxy configuration is handled by Juju when deploying the `k8s` charms. Please
see the [documentation for adding proxy configuration via Juju].

<!-- LINKS -->

[documentation for adding proxy configuration via Juju]: /charm/howto/proxy
[squid]: https://ubuntu.com/server/docs/how-to-install-a-squid-servers
