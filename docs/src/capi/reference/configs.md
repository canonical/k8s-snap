# Providers Configurations

{{product}} bootstrap and control plane providers (CABPCK and CACPCK) can be configured to aid the cluster admin in reaching the desired state for the workload cluster. In this section we will go through different configurations that each one of these providers expose.

## Common Configurations

The following configurations are available for both bootstrap and control plane providers.

### `version`
**Type:** `string`

**Required:** yes

`version` is used to specify the {{product}} version installed on the nodes.

```{note}
The {{product}} providers will install the latest patch in the `stable` risk level by default, e.g. `1.30/stable`. Patch versions specified in this configuration will be ignored.

To install a specific track or risk level, see [Install custom {{product}} on machines] guide.
```

**Example Usage:**
```yaml
spec:
    version: 1.30
```

### `files`
**Type:** `struct`

**Required:** no

`files` can be used to add new files to the machines or overwrite existing files.

**Fields:**

| Name | Type | Description | Default |
|------|------|-------------|---------|
| `path` | `string` | Where the file should be created | `""` |
| `content` | `string` | Content of the created file | `""` |
| `permissions` | `string` | Permissions of the file to create, e.g. "0600" | `""` |
| `owner` | `string` | Owner of the file to create, e.g. "root:root" | `""` |

**Example Usage:**
```yaml
spec:
    files:
        path: "/path/to/my-file"
        content: |
            #!/bin/bash -xe
            echo "hello from my-file
        permissions: "0500"
        owner: root:root
```

### `bootCommands`
**Type:** `[]string`

**Required:** no

`bootCommands` specifies extra commands to run in cloud-init early in the boot process.

**Example Usage:** 
```yaml
spec:
    bootCommands: 
        - echo "first-command"
        - echo "second-command"
```

(preruncommands)=
### `preRunCommands`
**Type:** `[]string`

**Required:** no

`preRunCommands` specifies extra commands to run in cloud-init before k8s-snap setup runs.

```{note}
`preRunCommands` can also be used to install custom {{product}} versions on machines. See [Install custom {{product}} on machines] guide for more info.
```

**Example Usage:** 
```yaml
spec:
    preRunCommands:
        - echo "first-command"
        - echo "second-command"
```

### `postRunCommands`
**Type:** `[]string`

**Required:** no

`postRunCommands` specifies extra commands to run in cloud-init after k8s-snap setup runs.

**Example Usage:** 
```yaml
spec:
    postRunCommands:
        - echo "first-command"
        - echo "second-command"
```

### `airGapped`
**Type:** `bool`

**Required:** no

`airGapped` is used to signal that we are deploying to an airgap environment. In this case, the provider will not attempt to install k8s-snap on the machine. The user is expected to install k8s-snap manually with [`preRunCommands`](#preruncommands), or provide an image with k8s-snap pre-installed.

**Example Usage:**
```yaml
spec:
    airGapped: true
```

### `initConfig`
**Type:** `struct`

**Required:** no

`initConfig` is configuration for the initializing the cluster features

**Fields:**

| Name | Type | Description | Default |
|------|------|-------------|---------|
| `annotations` | `map[string]string` | Are used to configure the behaviour of the built-in features. | `nil` |
| `enableDefaultDNS` | `bool` | Specifies whether to enable the default DNS configuration. | `true` |
| `enableDefaultLocalStorage` | `bool` | Specifies whether to enable the default local storage. | `true` |
| `enableDefaultMetricsServer` | `bool` | Specifies whether to enable the default metrics server. | `true` |
| `enableDefaultNetwork` | `bool` | Specifies whether to enable the default CNI. | `true` |


**Example Usage:**
```yaml
spec:
    initConfig:
        annotations:
            annotationKey: "annotationValue"
        enableDefaultDNS: false
        enableDefaultLocalStorage: true
        enableDefaultMetricsServer: false
        enableDefaultNetwork: true
```

### `nodeName`
**Type:** `string`

**Required:** no

`nodeName` is the name to use for the kubelet of this node. It is needed for clouds where the cloud-provider has specific pre-requisites about the node names. It is typically set in Jinja template form, e.g. `"{{ ds.meta_data.local_hostname }}"`.

**Example Usage:**
```yaml
spec:
    nodeName: "{{ ds.meta_data.local_hostname }}"
```

## Control plane provider (CACPCK)

The following configurations are only available for the control plane provider.

### `replicas`
**Type:** `int32`

**Required:** no

`replicas` is the number of desired machines. Defaults to 1. When stacked etcd is used only odd numbers are permitted, as per [etcd best practice].

**Example Usage:**
```yaml
spec:
    replicas: 2
```

### `controlPlane`
**Type:** `struct`

**Required:** no

`controlPlane` is configuration for control plane nodes.

**Fields:**

| Name | Type | Description | Default |
|------|------|-------------|---------|
| `extraSANs` | `[]string` | A list of SANs to include in the server certificates. | `[]` |
| `cloudProvider` | `string` | The cloud-provider configuration option to set. | `""` |
| `nodeTaints` | `[]string` | Taints to add to the control plane kubelet nodes. | `[]` |
| `datastoreType` | `string` | The type of datastore to use for the control plane. | `""` |
| `datastoreServersSecretRef` | `struct{name:str, key:str}` | A reference to a secret containing the datastore servers. | `{}` |
| `k8sDqlitePort` | `int` | The port to use for k8s-dqlite. If unset, 2379 (etcd) will be used. | `2379` |
| `microclusterAddress` | `string` | The address (or CIDR) to use for microcluster. If unset, the default node interface is chosen. | `""` |
| `microclusterPort` | `int` | The port to use for microcluster. If unset, ":2380" (etcd peer) will be used. | `":2380"` |
| `extraKubeAPIServerArgs` | `map[string]string` | Extra arguments to add to kube-apiserver. | `map[]` |

**Example Usage:**
```yaml
spec:
    controlPlane:
        extraSANs:
            - extra.san
        cloudProvider: external
        nodeTaints:
            - myTaint
        datastoreType: k8s-dqlite
        datastoreServersSecretRef:
            name: sfName
            key: sfKey
        k8sDqlitePort: 2379
        microclusterAddress: my.address
        microclusterPort: ":2380"
        extraKubeAPIServerArgs:
            argKey: argVal
```

<!-- LINKS -->
[Install custom {{product}} on machines]: ../howto/custom-ck8s.md
[etcd best practices]: https://etcd.io/docs/v3.5/faq/#why-an-odd-number-of-cluster-members



