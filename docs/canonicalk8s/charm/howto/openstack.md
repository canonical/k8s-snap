# How to integrate with OpenStack

This guide explains how to integrate {{product}} with the OpenStack cloud
platform. The `openstack-integrator` charm simplifies working with {{product}}
on OpenStack. Using the credentials provided to Juju, it acts as a proxy between
{{product}} and the underlying cloud, granting permissions to dynamically
create, for example, Cinder volumes.

## Prerequisites

To follow this guide, you will need:

- An [OpenStack][OpenStack] cloud environment.
- Octavia available both to support Kubernetes LoadBalancer services and to
  support the creation of a load balancer for the Kubernetes API.
- A valid [proxy configuration][proxy] in constrained environments.

Before deploying {{product}}, make sure to apply the following
configuration on the Juju model:

```
juju model-config container-networking-method=local fan-config=
```

Otherwise you might notice the Cilium pods failing with the following message:

```
failed to start: daemon creation failed: error while initializing daemon: failed
while reinitializing datapath: failed to setup vxlan tunnel device: setting up
vxlan device: creating vxlan device: setting up device cilium_vxlan: address
already in use
```

## Install {{product}} on OpenStack

To deploy the {{product}} [bundle][bundle] on OpenStack you need an overlay
bundle which serves as an extension of the core bundle. Through the overlay,
applications are deployed and relations are established between the
applications. These include the OpenStack integrator, cloud
controller, and cinder-csi charm.

### OpenStack overlay configurations:

Refer to the base overlay [openstack-overlay.yaml][openstack-overlay] and
modify it as needed.

```yaml
applications:
  openstack-integrator:
    charm: openstack-integrator
    num_units: 1
    trust: true
    base: ubuntu@22.04
  openstack-cloud-controller:
    charm: openstack-cloud-controller
    base: ubuntu@22.04
  cinder-csi:
    charm: cinder-csi
    base: ubuntu@22.04
relations:
  - [openstack-cloud-controller:kube-control,            k8s:kube-control]
  - [cinder-csi:kube-control,                            k8s:kube-control]
  - [openstack-cloud-controller:external-cloud-provider, k8s:external-cloud-provider]
  - [openstack-cloud-controller:openstack,               openstack-integrator:clients]
  - [cinder-csi:openstack,                               openstack-integrator:clients]
  - [openstack-integrator:lb-consumers,                  k8s:external-load-balancer]
```

### Deploy the overlay template

Deploy the {{product}} bundle on OpenStack using the modified overlay:

```
juju deploy canonical-kubernetes --overlay ~/path/openstack-overlay.yaml --trust
```

...and remember to fetch the configuration file!

```
juju run k8s/leader get-kubeconfig | yq eval '.kubeconfig' > kubeconfig
```

The {{product}} bundle is now deployed and integrated with OpenStack. Run
`juju status --watch 1s` to monitor the deployment. It is possible that your
deployment will take a few minutes until all the components are up and running.

```{note}
Resources allocated by Kubernetes or the integrator are usually cleaned up automatically when no longer needed. However, it is recommended to periodically, and particularly after tearing down a cluster, use the OpenStack administration tools to make sure all unused resources have been successfully released.
```

```{note}
The OpenStack Octavia load balancer creates a `healthmonitor` of type `TLS-HELLO`
which simply ensures the back-end servers respond to SSLv3 client hello messages.
It will not check any other health metrics, like status code or body contents.
```

<!-- LINKS -->
[OpenStack]: https://www.openstack.org/
[proxy]: /charm/howto/proxy
[bundle]: https://github.com/canonical/k8s-bundles/blob/main/main/bundle.yaml
[openstack-overlay]: https://github.com/canonical/k8s-bundles/blob/main/overlays/openstack.yaml
