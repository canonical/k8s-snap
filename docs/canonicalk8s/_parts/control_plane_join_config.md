### extra-sans
**Type:** `[]string`<br>

List of extra SANs to be added to certificates.

### front-proxy-client-crt
**Type:** `string`<br>

The client certificate to be used for the front proxy.
If omitted defaults to an auto generated certificate.

### front-proxy-client-key
**Type:** `string`<br>

The client key to be used for the front proxy.
If omitted defaults to an auto generated key.

### admin-client-crt
**Type:** `string`<br>

The admin client certificate to be used for Kubernetes services.
If omitted defaults to an auto generated certificate.

### admin-client-key
**Type:** `string`<br>

The admin client key to be used for Kubernetes services.
If omitted defaults to an auto generated key.

### kube-proxy-client-crt
**Type:** `string`<br>

The client certificate to be used by kubelet for communicating with the kube-apiserver.
If omitted defaults to an auto generated certificate.

### kube-proxy-client-key
**Type:** `string`<br>

The client key to be used by kubelet for communicating with the kube-apiserver.
If omitted defaults to an auto generated key.

### kube-scheduler-client-crt
**Type:** `string`<br>

The client certificate to be used for the kube-scheduler.
If omitted defaults to an auto generated certificate.

### kube-scheduler-client-key
**Type:** `string`<br>

The client key to be used for the kube-scheduler.
If omitted defaults to an auto generated key.

### kube-controller-manager-client-crt
**Type:** `string`<br>

The client certificate to be used for the Kubernetes controller manager.
If omitted defaults to an auto generated certificate.

### kube-controller-manager-client-key
**Type:** `string`<br>

The client key to be used for the Kubernetes controller manager.
If omitted defaults to an auto generated key.

### apiserver-crt
**Type:** `string`<br>

The certificate to be used for the kube-apiserver.
If omitted defaults to an auto generated certificate.

### apiserver-key
**Type:** `string`<br>

The key to be used for the kube-apiserver.
If omitted defaults to an auto generated key.

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

The client certificate to be used for the kubelet.
If omitted defaults to an auto generated certificate.

### kubelet-client-key
**Type:** `string`<br>

The client key to be used for the kubelet.
If omitted defaults to an auto generated key.

### disable-system-tuning
**Type:** `bool`<br>

Determines whether system tuning is allowed.
If omitted defaults to `true`.

### etcd-server-crt
**Type:** `string`<br>

The certificate to be used for the etcd server.

### etcd-server-key
**Type:** `string`<br>

The key to be used for the etcd server.

### etcd-peer-crt
**Type:** `string`<br>

The certificate to be used for the etcd server peer communication.

### etcd-peer-key
**Type:** `string`<br>

The key to be used for the etcd server peer communication.

### extra-node-config-files
**Type:** `map[string]string`<br>

Additional files that are uploaded `/var/snap/k8s/common/args/conf.d/<filename>`
to a node on bootstrap. These files can then be referenced by Kubernetes
service arguments.

The format is `map[<filename>]<filecontent>`.

### extra-node-kube-apiserver-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to the `kube-apiserver` only for that specific node.
A parameter that is explicitly set to `null` is deleted.
The format is `map[<--flag-name>]<value>`.

### extra-node-kube-controller-manager-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to the `kube-controller-manager` only for that specific node.
A parameter that is explicitly set to `null` is deleted.
The format is `map[<--flag-name>]<value>`.

### extra-node-kube-scheduler-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to the `kube-scheduler` only for that specific node.
A parameter that is explicitly set to `null` is deleted.
The format is `map[<--flag-name>]<value>`.

### extra-node-kube-proxy-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to the `kube-proxy` only for that specific node.
A parameter that is explicitly set to `null` is deleted.
The format is `map[<--flag-name>]<value>`.

### extra-node-kubelet-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to the `kubelet` only for that specific node.
A parameter that is explicitly set to `null` is deleted.
The format is `map[<--flag-name>]<value>`.

### extra-node-containerd-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to `containerd` only for that specific node.
A parameter that is explicitly set to `null` is deleted.
The format is `map[<--flag-name>]<value>`.

### extra-node-etcd-args
**Type:** `map[string]string`<br>

Additional arguments that are passed to `etcd` only for that specific node.
A parameter that is explicitly set to `null` is deleted.
The format is `map[<--flag-name>]<value>`.

### extra-node-containerd-config
**Type:** `api.MapStringAny`<br>

Extra configuration for the containerd config.toml

### containerd-base-dir
**Type:** `string`<br>

The base directory in which the containerd-related files are located.

