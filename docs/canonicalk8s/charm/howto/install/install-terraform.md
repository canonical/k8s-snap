# How to install with Terraform

This guide walks you through the process of installing {{ product }} using
the [Terraform Juju Provider][juju-provider-tf].

## Prerequisites

This guide requires the following:

- A Juju controller
<!-- TODO remove Juju prerequisites once ground up module is available -->
- The Terraform CLI, which can be installed via the [snap store][Terraform]

## Authentication

As a first step, authenticate the Terraform Juju Provider with the Juju
controller. Choose one of the options outlined in the
[provider documentation][auth].

## Terraform module creation

The Terraform deployment is done using a root module that specifies the
Juju model to deploy the submodules to. The root module also references
the k8s-bundle module which helps to build the Juju model.

### Root module
<!-- TODO replace this section once we have a Juju ground up module -->

The root module ensures that Terraform is aware of the `juju_model`
dependency of the charm module. Additionally, it contains the path to the
k8s-bundle child modules:

Example `main.tf`:

```
module "k8s" {
  source  = "git::https://github.com/canonical/k8s-bundles//terraform?ref=main"
  model   = {
    name  = "my-canonical-k8s-model"
    cloud = "prod-k8s-openstack"
  }
  cloud_integration = "openstack"
  manifest_yaml = "/path/to/manifest.yaml"
}
```

Define your `manifest.yaml` based on the requirements for your deployment. We
 recommend at least 16GB of root-disk storage, 4GB of memory and 2 cores.
Specific charm configuration options can be found on Charmhub.io for charms
[k8s] and [k8s-worker].

Example `manifest.yaml`:

```
k8s:
  units: 3
  base: ubuntu@24.04
  constraints: arch=amd64 cores=2 mem=4096M root-disk=16384M
  channel: 1.35/stable
  config: {}
k8s-worker:
  units: 2
  base: ubuntu@24.04
  constraints: arch=amd64 cores=2 mem=8192M root-disk=16384M
  channel: 1.35/stable
  config: {}
```

Example `versions.tf`:

```
terraform {
  required_version = ">= 1.6"
  required_providers {
    juju = {
      source  = "juju/juju"
      version = ">= 0.14.0, < 1.0.0"
    }
  }
}
```

### Cloud Integrations

The base modules support various cloud integration charms to integrate
{{ product }} with the underlying cloud substrate. Rather than presume a cloud
integration, the main Terraform module requires one to select the cloud
integration desired. See [k8s-bundles] to see how to integrate with OpenStack or
other cloud-providers.

### CSI Integrations

The base modules default to a local node CSI, but come with charm support for
multiple CSI options. See [k8s-bundles] for details on how to configure to
integrate with Ceph or another CSI.

### Charm modules

Find the `k8s` module at `//charms/worker/k8s/terraform` and
the `k8s-worker` module at `//charms/worker/terraform`.

The charm module for the k8s charm offers the following
configuration options:

| Name | Type | Description | Required | Default |
| - | - | - | - | - |
| `app_name`| string | Application name | False | k8s |
| `base` | string | Ubuntu base to deploy the charm onto | False | ubuntu@24.04 |
| `channel`| string | Channel that the charm is deployed from | False | null |
| `config`| map(string) | Map of the charm configuration options | False | {} |
| `constraints` | string | Juju constraints to apply for this application | False | arch=amd64 |
| `model`| string | Name of the model that the charm is deployed on | True | null |
| `resources`| map(string) | Map of the charm resources | False | {} |
| `revision`| number | Revision number of the charm name | False | null |
| `units` | number | Number of units to deploy | False | 1 |

Upon application, the module exports the following outputs:

| Name | Description |
| - | - |
| `app_name`|  Application name |
| `provides`|  Map of `provides` endpoints |
| `requires`|  Map of `requires` endpoints |

## Deploying the charm model

Please navigate to the root module's directory and run the following
commands:

```bash
terraform init --upgrade
terraform plan
terraform apply
```

```{note}
Make sure the deployment [channel] is set within the `manifest.yaml`
```

The `terraform apply` command will deploy the k8s and k8s-worker charms to the
Juju model. Watch the deployment progress by running:

```bash
juju status --watch 5s
```

<!-- LINKS -->
[juju-provider-tf]: https://github.com/juju/terraform-provider-juju/
[auth]: https://registry.terraform.io/providers/juju/juju/latest/docs#authentication
[channel]: ../../explanation/channels.md
[Terraform]: https://snapcraft.io/terraform
[k8s]: https://charmhub.io/k8s/configurations
[k8s-worker]: https://charmhub.io/k8s-worker/configurations
[k8s-bundles]: https://github.com/canonical/k8s-bundles/tree/main/terraform
