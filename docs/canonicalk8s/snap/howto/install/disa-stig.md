# How to install {{ product }} with DISA STIG hardening

Security Technical Implementation Guides (STIGs) are developed by the Defense
Information System Agency (DISA) and are comprehensive frameworks of security
requirements designed to protect U.S. Department of Defense (DoD) systems and
networks from cybersecurity threats.

The [Kubernetes STIG] contains guidelines on how to check and remediate various
potential security concerns for a Kubernetes deployment, both on the host and
within the cluster itself. This guide goes through the steps needed to deploy
{{product}} with our dedicated security configuration files to meet DISA STIG
hardening requirements.

## Prerequisites

This guide assumes the following:

- Ubuntu machine with at least 4GB of RAM and 30 GB disk storage
- You have root or sudo access to the machine
- Internet access on the machine
- You have Ubuntu Pro enabled on your system. For more information, see
  [Ubuntu Pro documentation]
<!-- - You have FIPS enabled on your machine. See our [FIPS installation guide]
 for guideance -->

## DISA STIG host compliance

DISA STIG host compliance is achieved by running the [usg tool] that is also
part of the PRO tool set. To install the usg tool:

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
The following command applies rule [V-270714] which will cause issues using
accounts with an empty password to access your machine. To avoid being locked
out use an RSA key to access your machine.
```

To automatically apply the recommended hardening changes:

```
sudo usg fix disa_stig
```

## Apply DISA STIG Kubernetes host rules

To comply with this guideline, the STIG templates we provide to bootstrap/join
nodes configure kubelet to run with the argument
`--protect-kernel-defaults=true`.

Configure the kernel as required for this setting by following the steps below:

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

## Deploy and bootstrap STIG-compliant nodes

For your convenience, we have provided template configuration files that can be
used to configure Kubernetes service arguments that align with DISA STIG
requirements.

### Bootstrap the first control-plane node

Install {{product}}:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

Initialize the first control plane node with the necessary arguments for
DISA STIG compliance:

```
sudo k8s bootstrap --file /var/snap/k8s/common/etc/templates/disa-stig/bootstrap.yaml
sudo k8s status --wait-ready
```

Through this configuration file, the following rules are applied:
| STIG                                                                               | Summary                                                               |
| ---------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| [V-242434]                                                                         | Kubernetes Kubelet must enable kernel protection                      |
| [V-245541]                                                                         | Kubernetes Kubelet must not disable timeouts                          |
| [V-242402], [V-242403], [V-242461], [V-242462], [V-242463], [V-242464], [V-242465] | The Kubernetes API Server must have an audit log configured           |
| [V-254800]                                                                         | Kubernetes must have a Pod Security Admission control file configured |
| [V-242400]                                                                         | The Kubernetes API server must have Alpha APIs disabled               |
| [V-254800]                                                                         | Kubernetes must have a Pod Security Admission control file configured |
| [V-242384]                                                                         | The Kubernetes Scheduler must have secure binding                     |
| [V-242385]                                                                         | The Kubernetes Controller Manager must have secure binding            |

### Join additional control plane nodes

To join another control plane node to the cluster, first retrieve the join token
from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname>
```

On the joining control plane node install {{product}}:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

and join the cluster with the DISA STIG control plane configuration file:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/control-plane.yaml <join-token>
```

Through this configuration file, the following rules are applied:

| STIG                                                                              | Summary                                                               |
| --------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| [V-242434]                                                                        | Kubernetes Kubelet must enable kernel protection                      |
| [V-245541]                                                                        | Kubernetes Kubelet must not disable timeouts                          |
| [V-242402], [V-242403], [V-242461], [V-242462], [V-242463], [V-242464],[V-242465] | The Kubernetes API Server must have an audit log configured           |
| [V-254800]                                                                        | Kubernetes must have a Pod Security Admission control file configured |
| [V-242400]                                                                        | The Kubernetes API server must have Alpha APIs disabled               |
| [V-254800]                                                                        | Kubernetes must have a Pod Security Admission control file configured |
| [V-242384]                                                                        | The Kubernetes Scheduler must have secure binding                     |
| [V-242385]                                                                        | The Kubernetes Controller Manager must have secure binding            |

### Join worker nodes

To join a worker node to the cluster, first retrieve the join token from an
existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname> --worker
```

On the joining worker node, install {{product}}:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

and join the cluster with the DISA STIG worker configuration file:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/worker.yaml <join-token>
```

Through this configuration file, the following rules are applied:

| STIG       | Summary                                          |
| ---------- | ------------------------------------------------ |
| [V-242434] | Kubernetes Kubelet must enable kernel protection |
| [V-245541] | Kubernetes Kubelet must not disable timeouts     |

If SSH is not needed to access the worker nodes it is recommended you disable
the SSH:

```
sudo systemctl disable ssh.service ssh.socket
```

```{note}
According to rule [V-242393] and [V-242394] Kubernetes Worker Nodes must not have
sshd service running or enabled. The host STIG rule [V-270665] on the other hand
expects sshd to be installed on the host. To comply with both rules, leave SSH
installed, but disable the service.
```

## Alternative control plane configurations

### Pod Security Admission control file

To comply with rule [V-254800], you must configure a Pod Security Admission
control file for your Kubernetes cluster. This file defines the Pod Security
Standards (PSS) that are enforced at the namespace level. The STIG templates we
provide to bootstrap/join nodes configure the Pod Security admission controller
to comply with these recommendations. For your own cluster, assess if:

- The default config at
`/var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml`
meets your needs or if it needs to be adjusted
- Using the more restrictive config we provide at
`/var/snap/k8s/common/etc/configurations/pod-security-admission-restricted.yaml`
is more suited to your needs
- You need to create your own audit policy based on the
[upstream instructions] and adjust the `--admission-control-config-file` path
used when you bootstrap/join nodes.

For more details, see the [Kubernetes Pod Security Admission documentation],
which provides an overview of Pod Security Standards (PSS), their enforcement
levels, and configuration options.

### Kubernetes API Server audit log

This applies to rules [V-242402], [V-242403], [V-242461], [V-242462],
[V-242463],[V-242464], [V-242465]. The STIG templates we provide to
bootstrap/join nodes configures the Kubernetes
API servers audit settings and policy to comply with these recommendations.

The default level of logging may be impractical for some situations, in which
case the settings would need to be adjusted and an exception put in place. For
your own cluster, assess if:

- The default logging configuration set at
`/var/snap/k8s/common/etc/configurations/audit-policy.yaml`, which configures
logging of all (non-resource) events with request metadata, request body, and
response body is suitable for your needs or if it needs to be adjusted
- Using the audit policy we provide at
`/var/snap/k8s/common/etc/configurations/audit-policy-kube-system.yaml` which
configures the same level of logging but only for events in the kube-system
namespace is more suited to your needs
- You need to create your own audit policy based on the
[upstream audit instructions] and adjust the `--audit-policy-file` path used
when you bootstrap/join nodes to use it.

Canonical Kubernetes does not enable audit logging by default as it may incur
performance penalties in the form of increased disk I/O, which can lead to
slower response times and reduced overall cluster efficiency, especially under
heavy workloads.

## Next steps

Please assess your cluster for compliance using the [DISA STIG auditing page].
Review all findings and apply any necessary remediations to be fully DISA STIG
compliant. Be aware that some rules need to be upheld when you add workloads
to your cluster.

<!-- Links -->
<!-- [FIPS installation guide]: fips.md -->
[Kubernetes STIG]: https://www.stigviewer.com/stig/kubernetes/
[DISA STIG auditing page]: ../security/disa-stig-assessment.md

[usg tool]: https://documentation.ubuntu.com/security/docs/compliance/usg/
[Ubuntu Pro documentation]: https://documentation.ubuntu.com/pro/start-here/#start-here
[Kubernetes Pod Security Admission documentation]: https://kubernetes.io/docs/concepts/security/pod-security-admission/
[upstream instructions]: https://kubernetes.io/docs/tasks/configure-pod-container/enforce-standards-admission-controller/
[upstream audit instructions]: https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[V-270714]: https://stigviewer.com/stigs/canonical_ubuntu_24.04_lts/2025-02-18/finding/V-270714
[V-270665]: https://stigviewer.com/stigs/canonical_ubuntu_24.04_lts/2025-02-18/finding/V-270665
[V-242403]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242403
[V-242434]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242434
[V-245541]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-245541
[V-254800]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-254800
[V-242402]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242402
[V-242461]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242461
[V-242462]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242462
[V-242463]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242463
[V-242464]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242464
[V-242465]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242465
[V-242400]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242400
[V-242384]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242384
[V-242385]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242385
[V-242393]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242393
[V-242394]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242394
