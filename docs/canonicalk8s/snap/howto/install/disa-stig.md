# How to install {{ product }} with DISA STIG hardening

DISA Security Technical Implementation Guides (STIGs) provide hardening
guidelines for meeting regulations from the U.S. Government and Department of
Defense (DoD).

{{product}} aligns by default with many of these recommendations as they are
expected to benefit most users. This guide provides additional steps to meet
all DISA STIG guidelines for both Kubernetes and the host OS.

## Prerequisites

FIPS compliance is required by the DISA STIG. This guide assumes you have
already followed our [FIPS installation guide], but stopped after installing
{{product}} without following the steps to bootstrap/join the cluster. Instead
continue here to first complete additional steps needed for DISA STIG
compliance.

## Configure UFW (Firewall)

The host STIG recommends enabling the host firewall (UFW), but does not do
so automatically. We recommend following our guide to [configure UFW]. This can
be done before applying the host STIG and will help avoid connectivity issues
that often happen when enabling UFW with the default configuration.

## Apply Host STIG

DISA STIG host compliance is achieved by running the [USG tool] that is also
part of the PRO tool set. To install the USG tool:

```
sudo pro enable usg
sudo apt update
sudo apt install usg
```

To generate a compliance audit report (without applying changes):

```
sudo usg audit disa_stig
```

```{attention}
The following command applies rule [V-270714], which will prevent using accounts
with empty passwords to access this machine.

You can check whether the current account has an empty password by running
`passwd --status` and looking for "NP" in the second field of the output.
```

To automatically apply the recommended hardening changes:

```
sudo usg fix disa_stig
```

Reboot to apply the changes:

```
sudo reboot
```

## Configure kernel

DISA STIG recommends enabling `--protect-kernel-defaults=true` so that kubelet
will not modify kernel flags. This requires that the kernel be configured in
advance as shown below:

```
sudo tee /etc/sysctl.d/99-kubelet.conf <<EOF
vm.overcommit_memory=1
vm.panic_on_oom=0
kernel.keys.root_maxbytes=25000000
kernel.keys.root_maxkeys=1000000
kernel.panic=10
kernel.panic_on_oops=1
EOF
sudo sysctl --system
```

```{note}
Please ensure that the configuration of `/etc/sysctl.d/99-kubelet.conf` is not
overridden by another higher order file.
```

## Apply Kubernetes STIG

{{product}} provides [configuration files] to apply
DISA STIG specific settings.

### Set up control plane nodes

```{attention}
Before bootstrapping or joining control plane nodes, review the
respective [configuration files](#configuration-files) as well as the
[audit logs and PSS](#audit-logs-and-pss-configuration)
alternative configuration options.
Once a node is configured, changing certain settings is more difficult
and may require re-deploying the node or cluster.
```

#### Bootstrap the first control-plane node

Initialize the first control plane node using the
bootstrap configuration file:

```
sudo k8s bootstrap --file /var/snap/k8s/common/etc/templates/disa-stig/bootstrap.yaml
sudo k8s status --wait-ready
```

#### Join control plane nodes

First retrieve a join token from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname>
```

Then join the new control plane node using the
respective node-join configuration file:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/control-plane.yaml <join-token>
```

### Join worker nodes

First retrieve a join token from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname> --worker
```

Then join the new worker node using the respective node-join configuration file:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/worker.yaml <join-token>
```

If SSH is not needed to access the worker nodes it is recommended you disable
the SSH service:

```
sudo systemctl disable ssh.service ssh.socket
```

```{note}
According to rule {ref}`242393` and {ref}`242394` Kubernetes worker nodes must not
have sshd service running or enabled. The host STIG rule [V-270665] on the
other hand expects sshd to be installed on the host. To comply with both
rules, leave SSH installed, but disable the service.
```

## Post-Deployment Requirements

In addition to the above deployment steps, there are some guidelines that must
be followed by users and administrators post-deployment and throughout the
life of the cluster in order to achieve and maintain DISA STIG compliance.

- {ref}`242383`: User-managed resources must be created in dedicated namespaces
- {ref}`242410`, {ref}`242411`, {ref}`242412`, and {ref}`242413`: The Kubernetes
API Server, Scheduler, and Controllers as well as etcd must enforce ports,
protocols, and services (PPS) that adhere to the Ports, Protocols, and
Services Management Category Assurance List (PPSM CAL). The {{product}}
[ports and services] must be audited in accordance with this list. Those ports,
protocols, and services that fall outside the PPSM CAL must be blocked or
registered. This step needs followed after the initial deployment and anytime
the list of ports, protocols, and services used by your cluster changes (for
instance each time a new service is exposed externally).
- {ref}`242393` and {ref}`242394`: SSH service must not be running or enabled on
   worker nodes (see [Join worker nodes](#join-worker-nodes))
- {ref}`242414`: User pods must only use non-privileged host ports
- {ref}`242415`: Secrets must not be stored as environment variables
- {ref}`242417`: User functionality must be separate from management functions
   meaning all user pods must be in user specific namespaces rather than system
   namespaces
- {ref}`242443`: Kubernetes components must be regularly updated to avoid
   vulnerabilities. We recommend using the latest revision of a
   [supported version] of {{product}}.

## Advance Configuration

### Configuration Files

#### Control plane & bootstrap configuration files

`/var/snap/k8s/common/etc/templates/disa-stig/bootstrap.yaml` is the
configuration file for 
[bootstrapping](#bootstrap-the-first-control-plane-node) the first node
of a cluster and
`/var/snap/k8s/common/etc/templates/disa-stig/control-plane.yaml` is the
control plane node join configuration file for [joining additional control plane
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

#### Worker node join configuration file

`/var/snap/k8s/common/etc/templates/disa-stig/worker.yaml` is the
worker node join configuration file 
for [joining worker nodes](#join-worker-nodes).
It applies settings to align with the following recommendations:

| STIG       | Summary                                          |
| ---------- | ------------------------------------------------ |
| {ref}`242434` | Kubernetes Kubelet must enable kernel protection |
| {ref}`245541` | Kubernetes Kubelet must not disable timeouts     |

### Audit Logs and PSS Configuration

The STIG configuration files provided
to [set up control plane nodes](#set-up-control-plane-nodes) can be
adjusted to suit your specific needs.

#### Pod Security Admission control file

To comply with rule {ref}`254800`, you must configure a Pod Security Admission
control file for your Kubernetes cluster. This file defines the Pod Security
Standards (PSS) that are enforced at the namespace level. By default, the
bootstrap and control plane configuration files point to
`/var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml`,
which sets the pod security policy to “baseline”, a minimally restrictive
policy that prevents known privilege escalations.

This policy may be insufficient or impractical in some situations, in which
case it needs to be adjusted by doing one of the following:

1. Set the `--admission-control-config-file` path in
    the bootstrap and control plane configuration files to
    `/var/snap/k8s/common/etc/configurations/pod-security-admission-restricted.yaml`
    rather than the baseline one. This sets a more restrictive policy.
2. Edit
    `/var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml`
    to suit your needs based on the [upstream instructions].
3. Create your own audit policy based on the [upstream instructions] and
    adjust the `--admission-control-config-file` path used in the
    configuration files.

For more details, see the [Kubernetes Pod Security Admission documentation],
which provides an overview of Pod Security Standards (PSS), their enforcement
levels, and configuration options.

#### Kubernetes API Server audit log

To comply with rules {ref}`242402`, {ref}`242403`, {ref}`242461`, {ref}`242462`,
{ref}`242463`, {ref}`242464`, and {ref}`242465` you must configure the 
Kubernetes API Server audit log. The STIG templates we provide to 
bootstrap/join control plane nodes configures the Kubernetes API servers audit 
settings and policy to comply with these recommendations.

By default, the configuration files will point to
`/var/snap/k8s/common/etc/configurations/audit-policy.yaml`, which configures
logging of all (non-resource) events with request metadata, request body, and
response body as recommended by {ref}`242403`.

This level of logging may be impractical for some situations, in which case the
settings would need to be adjusted and an exception put in place. To adjust the
audit settings, do one of the following:

1. Set the `--audit-policy-file` path used when you bootstrap/join nodes to
    use `/var/snap/k8s/common/etc/configurations/audit-policy-kube-system.yaml`
    rather than the file above. This configures the same level of logging but
    only for events in the kube-system namespace.
2. Edit `/var/snap/k8s/common/etc/configurations/audit-policy.yaml` to suit
    your needs based on the [upstream audit instructions] for this policy file.
3. Create your own audit policy based on the [upstream audit instructions] and
    adjust the `--audit-policy-file` path used when you bootstrap/join nodes to
    use it.

<!-- Links -->
[supported version]: https://ubuntu.com/about/release-cycle#canonical-kubernetes-release-cycle
[ports and services]: /snap/reference/ports-and-services/
[FIPS installation guide]: fips.md
[configure UFW]: /snap/howto/networking/ufw.md
[configuration files]: /snap/reference/config-files/index.md
[USG tool]: https://documentation.ubuntu.com/security/docs/compliance/usg/
[Kubernetes Pod Security Admission documentation]: https://kubernetes.io/docs/concepts/security/pod-security-admission/
[upstream instructions]: https://kubernetes.io/docs/tasks/configure-pod-container/enforce-standards-admission-controller/
[upstream audit instructions]: https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[V-270714]: https://www.stigviewer.com/stigs/canonical_ubuntu_2404_lts/2025-05-16/finding/V-270714
[V-270665]: https://www.stigviewer.com/stigs/canonical_ubuntu_2404_lts/2025-05-16/finding/V-270665