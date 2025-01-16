# Integrating with OpenStack

This guide explains how to integrate {{product}} with the OpenStack cloud
platform. The `openstack-integrator` charm simplifies working with {{product}}
on OpenStack. Using the credentials provided to Juju, it acts as a proxy between
{{product}} and the underlying cloud, granting permissions to dynamically
create, for example, Cinder volumes.

## Prerequisites

To follow this guide, you will need:

- An [OpenStack][openstack] cloud environment.
- Octavia available both to support Kubernetes LoadBalancer services and to
  support the creation of a load balancer for the Kubernetes API.
- A valid [proxy configuration][proxy] in constrained environments.

## Installing {{product}} on OpenStack

To deploy the {{product}} [bundle][bundle] on OpenStack you need an overlay
bundle which serves as an extension of the core bundle. Through the overlay,
applications are deployed and relations are established between the
applications. These include the openstack integrator, cloud
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
```

### Deploying the overlay template

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



<!-- LINKS -->
[openstack]: https://www.openstack.org/
[proxy]: https://documentation.ubuntu.com/canonical-kubernetes/main/src/charm/howto/proxy/
[bundle]: https://github.com/canonical/k8s-bundles/blob/main/main/bundle.yaml
[openstack-overlay]: https://github.com/canonical/k8s-bundles/blob/main/overlays/openstack.yaml
