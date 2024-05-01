# Configuring proxy settings for K8s

Canonical Kubernetes packages a number of utilities (eg curl, helm) which need
to fetch resources they expect to find on the internet. In a constrained
network environment, such access is usually controlled through proxies.

## Adding proxy configuration for the k8s charms

For the charm deployments of Canonical Kubernetes, Juju manages proxy
configuration through the [Juju model][]

For example, assume we have a proxy running at `http://squid.internal:3128` and
we are using the networks `10.0.0.0/8`,`192.168.0.0/16` and `172.16.0.0/12`. In
this case we would configure the model in which the charms are to run with
Juju:

```
juju model-config \
    juju-http-proxy=http://squid.internal:3128 \
    juju-https-proxy=http://squid.internal:3128 \
    juju-no-proxy=10.0.8.0/24,192.168.0.0/16,127.0.0.1,10.152.183.0/24
```

```{note} The `10.152.183.0/24` CIDR needs to be in the `juju-no-proxy` list as 
it is the service-cidr. Without this any pods will not be able to reach the 
cluster's kubernetes-api.
```
