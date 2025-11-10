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

## Configure Firewall

The host STIG will recommend enabling the host firewall (UFW), but does not do
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

```{warning}
The following command applies rule [V-270714] which will disallow using
accounts with an empty password to access your machine.
```

To automatically apply the recommended hardening changes:

```
sudo usg fix disa_stig
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

{{product}} provides [templates](#templates-information) to apply additional
configuration needed to fully align with DISA STIG requirements.

### Set up control plane nodes

```{attention}
Before bootstrapping or joining control-plane nodes, review the
[templates](#control-plane-templates) and
[alternative configurations](#alternative-control-plane-configurations).
Once a node is bootstrapped, changing certain settings is more difficult
and may require re-deploying the node or cluster.
```

#### Bootstrap the first control-plane node

To initialize the first control plane node using the template:

```
sudo k8s bootstrap --file /var/snap/k8s/common/etc/templates/disa-stig/bootstrap.yaml
sudo k8s status --wait-ready
```

#### Join control plane nodes

First retrieve a join token from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname>
```

Then join the new control plane node using the template:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/control-plane.yaml <join-token>
```

### Join worker nodes

First retrieve a join token from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname> --worker
```

Then join the new worker node using the template:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/worker.yaml <join-token>
```

If SSH is not needed to access the worker nodes it is recommended you disable
the SSH service:

```
sudo systemctl disable ssh.service ssh.socket
```

```{note}
According to rule [V-242393] and [V-242394] Kubernetes worker nodes must not
have sshd service running or enabled. The host STIG rule [V-270665] on the
other hand expects sshd to be installed on the host. To comply with both
rules, leave SSH installed, but disable the service. It is probably acceptable
however to remove SSH if it is not needed.
```

## Post-Deployment Requirements

In addition to the above deployment steps, there are some guidelines that must
be followed by users and administrators post-deployment and throughout the
life of the cluster in order to achieve and maintain DISA STIG compliance.

- [V-242383]: User-managed resources must be created in dedicated namespaces
- [V-242410], [V-242411], [V-242412], and [V-242413]: The Kubernetes
API Server, Scheduler, and Controllers as well as etcd must enforce ports,
protocols, and services (PPS) that adhere to the Ports, Protocols, and
Services Management Category Assurance List (PPSM CAL). The {{product}}
[ports and services] must be audited in accordance with this list. Those ports,
protocols, and services that fall outside the PPSM CAL must be blocked or
registered. This step needs followed after the initial deployment and anytime
the list of ports, protocols, and services used by your cluster changes (for
instance each time a new service is exposed externally).
- [V-242414]: User pods must only use non-privileged host ports
- [V-242415]: Secrets must not be stored as environment variables
- [V-242417]: User functionality must be separate from management functions
   meaning all user pods must be in user specific namespaces rather than system
   namespaces

## Appendix

### Templates Information

#### Control plane templates

`/var/snap/k8s/common/etc/templates/disa-stig/bootstrap.yaml` is the template
for [bootstrapping](#bootstrap-the-first-control-plane-node) and
`/var/snap/k8s/common/etc/templates/disa-stig/control-plane.yaml` is the
template for [joining additional control plane
nodes](#join-control-plane-nodes). Both of these templates apply configuration
to align with the following recommendations:

| STIG                                                                               | Summary                                                               |
| ---------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| [V-242384]                                                                         | The Kubernetes Scheduler must have secure binding                     |
| [V-242385]                                                                         | The Kubernetes Controller Manager must have secure binding            |
| [V-242400]                                                                         | The Kubernetes API server must have Alpha APIs disabled               |
| [V-242402], [V-242403], [V-242461], [V-242462], [V-242463], [V-242464], [V-242465] | The Kubernetes API Server must have an audit log configured           |
| [V-242434]                                                                         | Kubernetes Kubelet must enable kernel protection                      |
| [V-245541]                                                                         | Kubernetes Kubelet must not disable timeouts                          |
| [V-254800]                                                                         | Kubernetes must have a Pod Security Admission control file configured |

#### Worker node templates

`/var/snap/k8s/common/etc/templates/disa-stig/worker.yaml` is the template
for [joining worker nodes](#join-worker-nodes).
It applies configuration to align with the following recommendations:

| STIG       | Summary                                          |
| ---------- | ------------------------------------------------ |
| [V-242434] | Kubernetes Kubelet must enable kernel protection |
| [V-245541] | Kubernetes Kubelet must not disable timeouts     |

### Alternative control plane configurations

The STIG templates provided to [set up control plane
nodes](#set-up-control-plane-nodes) may need adjusted to suit your specific
needs.

#### Pod Security Admission control file

To comply with rule [V-254800], you must configure a Pod Security Admission
control file for your Kubernetes cluster. This file defines the Pod Security
Standards (PSS) that are enforced at the namespace level. By default, the
templates point to
`/var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml`,
which sets the pod security policy to “baseline”, a minimally restrictive
policy that prevents known privilege escalations.

This policy may be insufficient or impractical in some situations, in which
case the templates would need to be adjusted by doing one of the following:

1. Adjust the `--admission-control-config-file` path in the templates to
    `/var/snap/k8s/common/etc/configurations/pod-security-admission-restricted.yaml`
    rather than the file above. This sets a more restrictive policy.
2. Edit
    `/var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml`
    to suit your needs based on the [upstream instructions].
3. Create your own audit policy based on the [upstream instructions] and
    adjust the `--admission-control-config-file` path used in the templates.

For more details, see the [Kubernetes Pod Security Admission documentation],
which provides an overview of Pod Security Standards (PSS), their enforcement
levels, and configuration options.

#### Kubernetes API Server audit log

To comply with rules [V-242402], [V-242403], [V-242461], [V-242462],
[V-242463],[V-242464] and [V-242465] you must configure the Kubernetes API
Server audit log. The STIG templates we provide to bootstrap/join control
plane nodes configures the Kubernetes API servers audit settings and policy to
comply with these recommendations.

By default, the bootstrap configuration template will point to
`/var/snap/k8s/common/etc/configurations/audit-policy.yaml`, which configures
logging of all (non-resource) events with request metadata, request body, and
response body as recommended by [V-242403].

This level of logging may be impractical for some situations, in which case the
settings would need to be adjusted and an exception put in place. To adjust the
audit settings, do one of the following:

1. Adjust the `--audit-policy-file` path used when you bootstrap/join nodes to
    use `/var/snap/k8s/common/etc/configurations/audit-policy-kube-system.yaml`
    rather than the file above. This configures the same level of logging but
    only for events in the kube-system namespace.
2. Edit `/var/snap/k8s/common/etc/configurations/audit-policy.yaml` to suit
    your needs based on the [upstream audit instructions] for this policy file.
3. Create your own audit policy based on the [upstream audit instructions] and
    adjust the `--audit-policy-file` path used when you bootstrap/join nodes to
    use it.

<!-- Links -->
[ports and services]: /snap/reference/ports-and-services/
[FIPS installation guide]: fips.md
[configure UFW]: /snap/howto/networking/ufw.md
[USG tool]: https://documentation.ubuntu.com/security/docs/compliance/usg/
[Ubuntu Pro]: https://documentation.ubuntu.com/pro/start-here/#start-here
[Kubernetes Pod Security Admission documentation]: https://kubernetes.io/docs/concepts/security/pod-security-admission/
[upstream instructions]: https://kubernetes.io/docs/tasks/configure-pod-container/enforce-standards-admission-controller/
[upstream audit instructions]: https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[V-242384]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242384
[V-242385]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242385
[V-242383]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242383
[V-242393]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242393
[V-242394]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242394
[V-242400]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242400
[V-242402]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242402
[V-242403]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242403
[V-242410]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242410
[V-242411]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242411
[V-242412]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242412
[V-242413]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242413
[V-242414]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242414
[V-242415]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242415
[V-242417]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242417
[V-242434]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242434
[V-242461]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242461
[V-242462]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242462
[V-242463]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242463
[V-242464]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242464
[V-242465]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242465
[V-245541]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-245541
[V-254800]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-254800
[V-270665]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-270665
[V-270714]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-270714
