# Install custom {{product}} on machines

By default, the `version` field in the machine specifications will determine which {{product}} is downloaded from the `stable` risk level. While you can install different versions of the `stable` risk level by changing the `version` field, extra steps should be taken if you're willing to install a specific risk level.
This guide walks you through the process of installing custom {{product}} on workload cluster machines.

## Prerequisites

To follow this guide, you will need:

- A Kubernetes management cluster with Cluster API and providers installed and configured.
- A generated cluster spec manifest

Please refer to the [getting-started guide][getting-started] for further
details on the required setup.

In this guide we call the generated cluster spec manifest `cluster.yaml`.

## Overwrite the existing `install.sh` script

The installation of the {{product}} snap is done via running the `install.sh` script in the cloud-init.
While this file is automatically placed in every workload cluster machine which hard-coded content by {{product}} providers, you can overwrite this file to make sure your desired content is available in the script.

As an example, let's overwrite the `install.sh` for our control plane nodes. Inside the `cluster.yaml`, add the new file content:

```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta2
kind: CK8sControlPlane
...
spec:
  ...
  spec:
    files:
    - content: |
        #!/bin/bash -xe
        snap install k8s --classic --channel=1.31-classic/candidate
      owner: root:root
      path: /capi/scripts/install.sh
      permissions: "0500"
```

Now the new control plane nodes that are created using this manifest will have
the `1.31-classic/candidate` {{product}} snap installed on them!

## Use `preRunCommands`

As mentioned above, the `install.sh` script is responsible for installing {{product}} snap on machines. `preRunCommands` are executed before `install.sh`. You can also add an install command to the `preRunCommands` in order to install your desired {{product}} version.

```{note}
Installing the {{product}} snap via the `preRunCommands`, does not prevent the `install.sh` script from running. Instead, the installation process in the `install.sh` will fail with a message indicating that `k8s` is already installed.
This is not considered a standard way and overwriting the `install.sh` script is recommended.
```

Edit the `cluster.yaml` to add the installation command:

```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta2
kind: CK8sControlPlane
...
spec:
  ...
  spec:
    preRunCommands:
    - snap install k8s --classic --channel=1.31-classic/candidate
```

<!-- LINKS -->
[getting-started]: ../tutorial/getting-started.md
