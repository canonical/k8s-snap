# DISA STIG configuration files

text here 

## Control plane example configuration files

`/var/snap/k8s/common/etc/templates/disa-stig/bootstrap.yaml` is the
configuration file for
[bootstrapping](#bootstrap-the-first-control-plane-node) the first node
of a cluster and
`/var/snap/k8s/common/etc/templates/disa-stig/control-plane.yaml` is the
control plane node-join configuration file for [joining additional control plane
nodes](#join-control-plane-nodes). Both of these configuration files
apply settings to align with the following recommendations:

| STIG                                                                               | Summary                                                               |
| ---------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| {ref}`242384`                                                                         | The Kubernetes Scheduler must have secure binding                     |
| {ref}`242385`                                                                        | The Kubernetes Controller Manager must have secure binding            |
| {ref}`242400`                                                                         | The Kubernetes API server must have Alpha APIs disabled               |
| {ref}`242402`, {ref}`242403`, {ref}`242461`, {ref}`242462`, {ref}`242463`, {ref}`242464`, {ref}`242465` | The Kubernetes API Server must have an audit log configured           |
| {ref}`242434`                                                                         | Kubernetes Kubelet must enable kernel protection                      |
| {ref}`245541`                                                                         | Kubernetes Kubelet must not disable timeouts                          |
| {ref}`254800`                                                                         | Kubernetes must have a Pod Security Admission control file configured |

## Worker node-join example configuration file

`/var/snap/k8s/common/etc/templates/disa-stig/worker.yaml` is the
worker node-join configuration file
for [joining worker nodes](#join-worker-nodes).
It applies settings to align with the following recommendations:

| STIG       | Summary                                          |
| ---------- | ------------------------------------------------ |
| {ref}`242434` | Kubernetes Kubelet must enable kernel protection |
| {ref}`245541` | Kubernetes Kubelet must not disable timeouts     |