# Integrating with OpenStack

This guide explains how to integrate {{product}} with the OpenStack cloud
platform.

## Prerequisites

To follow this guide, you will need:

- An [OpenStack][openstack] cloud environment.
- Juju set-up as the deployment tool for Openstack. Please refer to
  [Juju and Openstack][juju-openstack] documentation for more information.
- An OpenStack [project][project] with the necessary permissions to deploy
  {{product}}.
- A juju [controller][controller] with [access][credentials] to the OpenStack 
  cloud environment.
- A juju [model][model] for deploying {{product}} on OpenStack.
- Configured [proxy configuration][proxy] in constrained environments.

## Deploying {{product}} on OpenStack


To deploy the {{product}} [bundle][bundle] on OpenStack you need an overlay
bundle which serves as an extension of the core bundle. The overlay bundle
contains the necessary configuration to deploy {{product}} on OpenStack.

Refer to the base overlay [openstack-lb-overlay.yaml][openstack-overlay] and
modify it as needed.

### OpenStack overlay configurations:

Run `openstack project list` to retrieve your project id and include the
[project id][project] in the overlay template:

```yaml
applications:
  openstack-integrator:
    options:
      project-id: <my-openstack-project-id>
```


Adjust [easyrsa][easyrsa] to avoid creating an LXD machine:

```yaml
applications:
  easyrsa:
    to:
      - 0
```

The modified overlay template should look like this:

```yaml
applications:
  kubeapi-load-balancer: null
  kubernetes-control-plane:
    options:
      allow-privileged: "true"
  openstack-integrator:
    charm: openstack-integrator
    num_units: 1
    trust: true
    options:
      project-id: <my-openstack-project-id>
  openstack-cloud-controller:
    charm: openstack-cloud-controller
  cinder-csi:
    charm: cinder-csi
  easyrsa: 
    to:   
      - 0 
relations:
  - [openstack-cloud-controller:certificates,            easyrsa:client]
  - [openstack-cloud-controller:kube-control,            kubernetes-control-plane:kube-control]
  - [openstack-cloud-controller:external-cloud-provider, kubernetes-control-plane:external-cloud-provider]
  - [openstack-cloud-controller:openstack,               openstack-integrator:clients]
  - [easyrsa:client,                                     cinder-csi:certificates]
  - [kubernetes-control-plane:kube-control,              cinder-csi:kube-control]
  - [openstack-integrator:clients,                       cinder-csi:openstack]
  - [kubernetes-control-plane:loadbalancer-external,     openstack-integrator:lb-consumers]
```

Deploy the {{product}} bundle on OpenStack using the modified overlay:

```bash
juju deploy canonical-kubernetes --overlay openstack-lb-overlay.yaml --trust
```

The {{product}} bundle is now deployed and integrated with OpenStack. Run 
`juju status --watch 1s` to monitor the deployment. It is possible that your
deployment will take a few minutes until all the components are up and running.

<!-- LINKS -->
[openstack]: https://www.openstack.org/
[project]: https://docs.openstack.org/python-openstackclient/queens/cli/command-objects/project.html
[juju-openstack]: https://juju.is/docs/juju/openstack
[controller]: https://juju.is/docs/juju/manage-controllers
[model]: https://juju.is/docs/juju/manage-models
[proxy]: https://documentation.ubuntu.com/canonical-kubernetes/main/src/charm/howto/proxy/
[credentials]: https://juju.is/docs/juju/manage-credentials
[bundle]: https://juju.is/docs/juju/bundle
[openstack-overlay]: https://github.com/charmed-kubernetes/bundle/blob/main/overlays/openstack-lb-overlay.yaml
[easyrsa]: https://easy-rsa.readthedocs.io/en/latest/
