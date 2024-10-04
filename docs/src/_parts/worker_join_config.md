### kubelet-crt
**Type:** `string`<br>

The certificate to be used for the kubelet.
If omitted defaults to an auto generated certificate.

### kubelet-key
**Type:** `string`<br>

The key to be used for the kubelet.
If omitted defaults to an auto generated key.

### kubelet-client-crt
**Type:** `string`<br>


### kubelet-client-key
**Type:** `string`<br>


### kube-proxy-client-crt
**Type:** `string`<br>


### kube-proxy-client-key
**Type:** `string`<br>


### extra-node-config-files
**Type:** `map[string]string`<br>

Additional files that are uploaded `/var/snap/k8s/common/args/conf.d/<filename>`
to a node on bootstrap. These files can then be referenced by Kubernetes
service arguments.

The format is `map[<filename>]<filecontent>`.

### extra-node-kube-proxy-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to the `kube-proxy` only for that
specific node. Overwrites default configuration. A parameter that is explicitly
set to `null` is deleted. The format is `map[<--flag-name>]<value>`.

### extra-node-kubelet-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to the `kubelet` only for that
specific node. Overwrites default configuration. A parameter that is explicitly
set to `null` is deleted. The format is `map[<--flag-name>]<value>`.

### extra-node-containerd-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to `containerd` only for that
specific node. Overwrites default configuration. A parameter that is explicitly
set to `null` is deleted. The format is `map[<--flag-name>]<value>`.

### extra-node-k8s-apiserver-proxy-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to `k8s-api-server-proxy` only for that
specific node. Overwrites default configuration. A parameter that is explicitly
set to `null` is deleted. The format is `map[<--flag-name>]<value>`.

### extra-node-containerd-config
**Type:** `apiv1.MapStringAny`<br>

Extra configuration for the containerd config.toml

