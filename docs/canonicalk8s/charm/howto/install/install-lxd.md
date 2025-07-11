# How to install to localhost/LXD

The main [install instructions][install] cover most situations for installing
{{product}} using a charm. However, using LXD requires special consideration.

Juju can leverage LXD by:

- deploying to the default 'localhost' cloud
- explicitly adding a Juju cloud that uses the ``lxd`` type
- deploying to a container on a machine (i.e. when installing a bundle or using
  the 'to:' directive to install to an existing machine)

```{warning}
LXD privileged containers are no longer supported and some Kubernetes services,
such as the Cilium CNI, cannot run inside unprivileged containers.
```

As such, we recommend using LXD virtual machines, which also ensure that the
Kubernetes environment is well isolated.

## Deploy an LXD VM

We can use Juju constraints to request virtual machines to be used instead
of containers and specify the amount of resources to allocate.

For example, we can pass the following constraints when deploying ``k8s``:

```
juju deploy k8s --channel=$channel \
  --constraints='cores=2 mem=4G root-disk=40G virt-type=virtual-machine'
```

The constraints can also be defined per model using
``juju set-model-constraints`` or per applications through
``juju set-constraints`` so that all the deployed units will use the specified
constraints.

<!-- LINKS -->
[install]: ./charm