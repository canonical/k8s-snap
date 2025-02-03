# How to configure proxy settings for K8s

{{product}} packages a number of utilities (for example curl, Helm) which need
to fetch resources they expect to find on the internet. In a restricted
network environment, such access is usually controlled through proxies.

To set up a proxy using Squid follow the
[How to install a Squid server][squid] tutorial.

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
Environment="NO_PROXY=10.1.0.0/16,10.152.183.0/24,192.168.0.0/16,127.0.0.1,172.16.0.0/12"
Environment="https_proxy=http://squid.internal:3128"
Environment="http_proxy=http://squid.internal:3128"
Environment="no_proxy=10.1.0.0/16,10.152.183.0/24,192.168.0.0/16,127.0.0.1,172.16.0.0/12"
```

Note that you may need to restart for these settings to take effect.
You can restart with the following commands:

```
sudo systemctl daemon-reload
sudo systemctl restart snap.k8s.containerd.service
```


```{note}
Include the CIDRs **10.152.183.0/24** and **10.1.0.0/16** in both the
`no_proxy` and `NO_PROXY` environment variables. They are the default
Kubernetes service and pod CIDRs, respectively. If you are using different
service or pod CIDRs, update this setting accordingly.
```

## Adding proxy configuration for the k8s charms

Proxy configuration is handled by Juju when deploying the `k8s` charms. Please
see the [documentation for adding proxy configuration via Juju][juju-proxy].

<!-- LINKS -->

[juju-proxy]: ../../../charm/howto/proxy
[squid]: https://ubuntu.com/server/docs/how-to-install-a-squid-server
