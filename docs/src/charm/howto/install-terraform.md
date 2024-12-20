# Installing with Terraform

This guide walks you through the process of installing {{ product }} using
the [Terraform Juju Provider][juju-provider-tf]. 

## Prerequisites

This guide requires the following:

- A Juju controller and model 
<!-- TODO remove Juju prerequisites once ground up module is available -->
- The Terraform cli, which can be installed via the [snap store][terraform]

## Authentication

As a first step, authenticate the Terraform Juju Provider with the Juju
controller. Choose one of the options outlined in the
[provider documentation][auth].

## Terraform Module creation

The Terraform deployment is done using a root module that specifies the
Juju model to deploy the submodules to. The root module also references
the k8s charm and the k8s-worker charm that each live in a separate child
module.

### Root module
<!-- TODO replace this section once we have a Juju ground up module -->

The root module ensures that Terraform is aware of the `juju_model`
dependency of the charm module. Additionally, it contains the path to the k8s
and k8s-worker child modules:

Example `main.tf`:

```hcl
data "juju_model" "testing" {
  name = "juju-myk8s"
}
module "k8s" {
  source = "path-to/k8s"
  juju_model = module.juju_model.testing.name
}

module "k8s-worker" {
  source = "path-to/k8s-worker"
  juju_model = module.juju_model.testing.name
}
```

```{note} 
Please ensure that the root module references the correct source path for the `k8s` and `k8s-worker` modules.
```

Example `versions.tf`:

```hcl
terraform {
  required_version = ">= 1.6"
  required_providers {
    juju = {
      source  = "juju/juju"
      version = "~> 0.14.0"
    }
  }
}
```

### Charm modules

Please download the charm modules from Github at:

```
git clone https://github.com/canonical/k8s-operator.git
```

Find the control-plane module at `k8s-operator/charms/worker/k8s/terraform` and
the k8s-worker module at `k8s-operator/tree/main/charms/worker/terraform`.

The charm module for the k8s charm offers the following
configuration options:

| Name | Type | Description | Required | Default |
| - | - | - | - | - |
| `app_name`| string | Application name | False | k8s |
| `base` | string | Ubuntu base to deploy the charm onto | False | ubuntu@24.04 |
| `channel`| string | Channel that the charm is deployed from | False | 1.30/edge |
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
| `provides`| Map of `provides` endpoints |
| `requires`|  Map of `requires` endpoints |

### Deploying the charms

Please navigate to the root module's directory and run the following
commands:

```bash
terraform init
terraform apply
```

The `terraform apply` command will deploy the k8s and k8s-worker charms to the
Juju model. Watch the deployment progress by running:

```bash
juju status --watch 5s
```

<!-- LINKS -->
[juju-provider-tf]: https://github.com/juju/terraform-provider-juju/
[auth]: https://registry.terraform.io/providers/juju/juju/latest/docs#authentication
[terraform]: https://snapcraft.io/terraform

