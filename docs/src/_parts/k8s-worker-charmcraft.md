### bootstrap-node-taints
**Type:** `string`<br>
**Default Value:** ``

Space-separated list of taints to apply to this node at registration time.

This config is only used at bootstrap time when Kubelet first registers the
node with Kubernetes. To change node taints after deploy time, use kubectl
instead.

For more information, see the upstream Kubernetes documentation about
taints:
https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/

### kube-proxy-extra-args
**Type:** `string`<br>
**Default Value:** ``

Space separated list of flags and key=value pairs that will be passed as arguments to
kube-proxy.

Notes:
  Options may only be set on charm deployment

For example a value like this:
  runtime-config=batch/v2alpha1=true profiling=true
will result in kube-proxy being run with the following options:
  --runtime-config=batch/v2alpha1=true --profiling=true

### kubelet-extra-args
**Type:** `string`<br>
**Default Value:** ``

Space separated list of flags and key=value pairs that will be passed as arguments to
kubelet.

Notes:
  Options may only be set on charm deployment

For example a value like this:
  runtime-config=batch/v2alpha1=true profiling=true
will result in kubelet being run with the following options:
  --runtime-config=batch/v2alpha1=true --profiling=true

### node-labels
**Type:** `string`<br>
**Default Value:** ``

Labels can be used to organize and to select subsets of nodes in the
cluster. Declare node labels in key=value format, separated by spaces.

Note: Due to NodeRestriction, workers are limited to how they can label themselves
https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#noderestriction

