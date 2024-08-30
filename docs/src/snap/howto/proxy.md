# Configure proxy settings for K8s

{{product}} packages a number of utilities (eg curl, helm) which need
to fetch resources they expect to find on the internet. In a constrained
network environment, such access is usually controlled through proxies.

To set up a proxy using squid follow the
[how-to-install-a-squid-server][squid] tutorial.

## Adding proxy configuration for the k8s snap

If necessary, create the `snap.k8s.containerd.service.d` directory:

```bash
sudo mkdir -p /etc/systemd/system/snap.k8s.containerd.service.d
```

```{note} It is important to add whatever address ranges are used by the
 cluster itself to the `NO_PROXY` and `no_proxy` variables.
```

For example, assume we have a proxy running at `http://squid.internal:3128` and
we are using the networks `10.0.0.0/8`,`192.168.0.0/16` and `172.16.0.0/12`.
We would add the configuration to the
(`/etc/systemd/system/snap.k8s.containerd.service.d/http-proxy.conf`) file:

```bash
# /etc/systemd/system/snap.k8s.containerd.service.d/http-proxy.conf
[Service]
Environment="HTTPS_PROXY=http://squid.internal:3128"
Environment="HTTP_PROXY=http://squid.internal:3128"
Environment="NO_PROXY=10.0.0.0/8,10.152.183.1,192.168.0.0/16,127.0.0.1,172.16.0.0/12"
Environment="https_proxy=http://squid.internal:3128"
Environment="http_proxy=http://squid.internal:3128"
Environment="no_proxy=10.0.0.0/8,10.152.183.1,192.168.0.0/16,127.0.0.1,172.16.0.0/12"
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
[squid]: https://ubuntu.com/server/docs/how-to-install-a-squid-server
