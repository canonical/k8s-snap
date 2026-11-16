### front-proxy-client-crt
**Type:** `string`<br>


### front-proxy-client-key
**Type:** `string`<br>

The client key to be used for the front proxy.

### apiserver-kubelet-client-crt
**Type:** `string`<br>

The client certificate to be used by kubelet for communicating with the kube-apiserver.

### apiserver-kubelet-client-key
**Type:** `string`<br>

The client key to be used by kubelet for communicating with the kube-apiserver.

### admin-client-crt
**Type:** `string`<br>

The admin client certificate to be used for Kubernetes services.

### admin-client-key
**Type:** `string`<br>

The admin client key to be used for Kubernetes services.

### kube-proxy-client-crt
**Type:** `string`<br>

The client certificate to be used for the kube-proxy.

Deprecated: As of Canonical Kubernetes 1.36, this field is not used with the
default CNI (Cilium), which uses kube-proxy replacement instead of kube-proxy.
It remains accepted for backward compatibility and is used when the default
CNI is not used.

### kube-proxy-client-key
**Type:** `string`<br>

The client key to be used for the kube-proxy.

Deprecated: As of Canonical Kubernetes 1.36, this field is not used with the
default CNI (Cilium), which uses kube-proxy replacement instead of kube-proxy.
It remains accepted for backward compatibility and is used when the default
CNI is not used.

### kube-scheduler-client-crt
**Type:** `string`<br>

The client certificate to be used for the kube-scheduler.

### kube-scheduler-client-key
**Type:** `string`<br>

The client key to be used for the kube-scheduler.

### kube-controller-manager-client-crt
**Type:** `string`<br>

The client certificate to be used for the Kubernetes controller manager.

### kube-controller-manager-client-key
**Type:** `string`<br>

The client key to be used for the Kubernetes controller manager.

### apiserver-crt
**Type:** `string`<br>

The certificate to be used for the kube-apiserver.

### apiserver-key
**Type:** `string`<br>

The key to be used for the kube-apiserver.

### kubelet-crt
**Type:** `string`<br>

The certificate to be used for the kubelet.

### kubelet-key
**Type:** `string`<br>

The key to be used for the kubelet.

### kubelet-client-crt
**Type:** `string`<br>

The certificate to be used for the kubelet client.

### kubelet-client-key
**Type:** `string`<br>

The key to be used for the kubelet client.

