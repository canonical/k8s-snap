# Install custom {{product}} on machines

By default, the `version` field in the machine specifications will determine 
which {{product}} version is downloaded from the `stable` risk level. While 
you can install different versions of the `stable` risk level by changing the 
`version` field, extra steps should be taken if you're willing to install 
a specific risk level, revision, or install the snap from local path.
This guide walks you through the process of installing custom {{product}} 
on workload cluster machines.

## Prerequisites

To follow this guide, you will need:

- A Kubernetes management cluster with Cluster API and providers installed 
and configured.
- A generated cluster spec manifest

Please refer to the [getting-started guide][getting-started] for further
details on the required setup.

This guide will call the generated cluster spec manifest `cluster.yaml`.

## Using the configuration specification

{{product}} can be installed on machines using a specific `channel`, 
`revision` or `localPath` by specifying the respective field in the spec 
of the machine.

```yaml
spec:
  ...
  spec:
    channel: 1.xx-classic/candidate
    # Or
    revision: 1234
    # Or
    localPath: /path/to/snap/on/machine
```

Note that for the `localPath` to work the snap must be available on the 
machine at the specified path on boot.

## Overwrite the existing `install.sh` script

The `install.sh` script can be overwritten to install a custom {{product}} 
snap on the machine. This can be done by adding a `files` field to the 
`spec` of the machine with a specific `path`.

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

```{note}
[Use the configuration specification](#using-config-spec), 
if you're only interested in installing a specific channel, revision, or 
form the local path.
```

<!-- LINKS -->
[getting-started]: ../tutorial/getting-started.md
