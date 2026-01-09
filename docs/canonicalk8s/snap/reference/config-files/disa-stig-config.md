# DISA STIG example configuration files

During the [installation of a DISA STIG hardened cluster], {{product}} provides
default configuration files for cluster formation, control plane join and 
worker join that automatically apply the following DISA STIG recommendations.  

## Example control plane configuration files

`/var/snap/k8s/common/etc/configurations/disa-stig/bootstrap.yaml` is the
configuration file for bootstrapping the first node
of a cluster. 

`/var/snap/k8s/common/etc/configurations/disa-stig/control-plane.yaml` is the
control plane node join configuration file for joining additional control plane
nodes. 

Both of these configuration files
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

## Example worker node join configuration file

`/var/snap/k8s/common/etc/configurations/disa-stig/worker.yaml` is the
worker node join configuration file
for joining worker nodes.

It applies settings to align with the following recommendations:

| STIG       | Summary                                          |
| ---------- | ------------------------------------------------ |
| {ref}`242434` | Kubernetes Kubelet must enable kernel protection |
| {ref}`245541` | Kubernetes Kubelet must not disable timeouts     |

<!-- LINKS -->
[installation of a DISA STIG hardened cluster]: /snap/howto/install/disa-stig.md