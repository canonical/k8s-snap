# How to install {{product}} using localhost/LXD

The main [install instructions][install] cover most situations for installing
{{product}} using a charm. However, using LXD requires special consideration.

Juju can leverage LXD by:

- deploying to the default 'localhost' cloud
- explicitly adding a Juju cloud that uses the ``lxd`` type
- deploying to a container on a machine (i.e. when installing a bundle or using
  the 'to:' directive to install to an existing machine)

```{warning}
Using LXD containers for Canonical Kubernetes deployments is not recommended.
Some Kubernetes services, such as Cilium, require privileged containers
to function properly. However, privileged containers allow the root user in the
container to be the root user on the host, creating security risks. Additionally,
newer versions of Ubuntu and systemd have compatibility issues with this setup.
```

As such, we recommend using LXD virtual machines, which also ensure that the
Kubernetes environment is well isolated.

## Deploy an LXD VM

We can use Juju constraints to request virtual machines to be used instead
of containers and specify the amount of resources to allocate.

For example, we can pass the following constraints when deploying ``k8s``:

```
juju deploy k8s --channel=$channel \
  --base="ubuntu@24.04" \
  --constraints='cores=2 mem=16G root-disk=40G virt-type=virtual-machine'
```

The constraints can also be defined per model using
``juju set-model-constraints`` or per applications through
``juju set-constraints`` so that all the deployed units will use the specified
constraints.

<!-- LINKS -->
[install]: ./charm
