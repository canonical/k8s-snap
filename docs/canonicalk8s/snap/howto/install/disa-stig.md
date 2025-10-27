# How to install {{ product }} with DISA STIG hardening

Security Technical Implementation Guides (STIGs) are developed by the Defense
Information System Agency (DISA) and are comprehensive frameworks of security
requirements designed to protect U.S. Department of Defense (DoD) systems and
networks from cybersecurity threats.

{{product}} aligns by default with many of these recommendations as they are
expected to benefit most users. The additional recommendations are covered in
this guide as well as the end-to-end procedure to be DISA STIG compliant at all
layers including FIPS 140-3 compliance and host hardening.

## Prerequisites

This guide assumes the following:

- Ubuntu machine with at least 4GB of RAM and 30 GB disk storage
- You have root or sudo access to the machine
- Internet access on the machine

## Firewall configuration

It is recommended to enable a host firewall (such as UFW) for DISA STIG
compliance, but this is not done automatically. If you choose to enable a
firewall, ensure you apply the required rules as described in the
[firewall configuration] guide to avoid connectivity issues.

## Install {{product}} in FIPS mode

DISA STIG compliance also incorporates FIPS compliance. Start by following our
[FIPS installation guide], but stop after installing {{product}} and do not
follow the steps to bootstrap the cluster. Instead continue here as there
are additional steps and configuration needed for DISA STIG compliance. The
guide also covers enabling [Ubuntu Pro], which is needed for some of the
steps below.

## DISA STIG host compliance

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
The following command applies rule [V-270714] which will cause issues using
accounts with an empty password to access your machine.
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
used to configure the additional Kubernetes service arguments that are needed to
fully align with DISA STIG requirements.

### Bootstrap the first control-plane node

```{attention}
Before bootstrapping the first control-plane node, review the available bootstrap
configuration templates to ensure you select the one that best fits your
requirements (for example, stricter or more permissive control-plane settings).
Once a node is bootstrapped, changing certain settings may require re-deploying
the node. See Alternative configurations below for more details.
```

Initialize the first control plane node with the necessary arguments for
DISA STIG compliance:

```
sudo k8s bootstrap --file /var/snap/k8s/common/etc/templates/disa-stig/bootstrap.yaml
sudo k8s status --wait-ready
```

This configuration file applies the following rules:  

| STIG                                                                               | Summary                                                               |
| ---------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| [V-242434]                                                                         | Kubernetes Kubelet must enable kernel protection                      |
| [V-245541]                                                                         | Kubernetes Kubelet must not disable timeouts                          |
| [V-242402], [V-242403], [V-242461], [V-242462], [V-242463], [V-242464], [V-242465] | The Kubernetes API Server must have an audit log configured           |
| [V-254800]                                                                         | Kubernetes must have a Pod Security Admission control file configured |
| [V-242400]                                                                         | The Kubernetes API server must have Alpha APIs disabled               |
| [V-242384]                                                                         | The Kubernetes Scheduler must have secure binding                     |
| [V-242385]                                                                         | The Kubernetes Controller Manager must have secure binding            |

### Join additional control plane nodes

On the additional control plane node, join the cluster with the DISA STIG
control plane configuration file:  

```
sudo k8s get-join-token <joining-node-hostname>
```

and join the cluster with the DISA STIG control plane configuration file:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/control-plane.yaml <join-token>
```

This configuration file applies the same rules shown above when bootstrapping
the first control plane node.

### Join worker nodes

To join a worker node to the cluster, first retrieve the join token from an
existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname> --worker
```

On the worker node, join the cluster with the DISA STIG worker configuration
file:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/worker.yaml <join-token>
```

This configuration file applies the following rules:  

| STIG       | Summary                                          |
| ---------- | ------------------------------------------------ |
| [V-242434] | Kubernetes Kubelet must enable kernel protection |
| [V-245541] | Kubernetes Kubelet must not disable timeouts     |

If SSH is not needed to access the worker nodes it is recommended you disable
the SSH service:

```
sudo systemctl disable ssh.service ssh.socket
```

```{note}
According to rule [V-242393] and [V-242394] Kubernetes Worker Nodes must not have
sshd service running or enabled. The host STIG rule [V-270665] on the other hand
expects sshd to be installed on the host. To comply with both rules, we
recommend leaving SSH installed, but disabling the service.  
```

## Alternative configurations

### Pod Security Admission control file

To comply with rule [V-254800], you must configure a Pod Security Admission
control file for your Kubernetes cluster. This file defines the Pod Security
Standards (PSS) that are enforced at the namespace level. The STIG templates we
provide to bootstrap/join control plane nodes configure the Pod Security
admission controller to comply with these recommendations.

By default, the bootstrap configuration template will point to
`/var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml`,
which sets the pod security policy to “baseline”, a minimally restrictive
policy that prevents known privilege escalations.

This policy may be insufficient or impractical in some situations, in which case
the settings would need to be adjusted by doing one of the following:

1. Adjust the `--admission-control-config-file` path used when you
   bootstrap/join nodes to
   `/var/snap/k8s/common/etc/configurations/pod-security-admission-restricted.yaml`
   rather than the file above. This sets a more restrictive policy.
2. Edit
   `/var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml`
   to suit your needs based on the [upstream instructions].
3. Create your own audit policy based on the [upstream instructions] and adjust
   the `--admission-control-config-file` path used when you bootstrap/join
   nodes.

For more details, see the [Kubernetes Pod Security Admission documentation],
which provides an overview of Pod Security Standards (PSS), their enforcement
levels, and configuration options.

### Kubernetes API Server audit log

To comply with rules [V-242402], [V-242403], [V-242461], [V-242462],
[V-242463],[V-242464] and [V-242465] you must configure the Kubernetes API
Server audit log. The STIG templates we provide to
bootstrap/join control plane nodes configures the Kubernetes
API servers audit settings and policy to comply with these recommendations.

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
   your needs based on the [upstream audit instructions] for this policy
   file.
3. Create your own audit policy based on the
   [upstream audit instructions] and adjust the `--audit-policy-file` path
   used when you bootstrap/join nodes to use it.

## Required Documentation

[V-242410], [V-242411], [V-242412], and [V-242413] requires that the Kubernetes
API Server, Scheduler, and Controllers as well as etcd enforce ports, protocols,
and services (PPS) that adhere to the Ports, Protocols, and Services Management
Category Assurance List (PPSM CAL). The {{product}} [ports and services] must be
manually audited in accordance with this policy. These ports, protocols, and
services will need to be added to your specific PPSM list and the list will need
to be updated anytime the list of ports, protocols, and services used by your
cluster changes. For instance, this list will need to be updated each time a
new service is exposed externally.

## Post-Deployment Requirements

In addition to the above deployment steps, there are some guidelines that must
be followed by users and administrators throughout the life of the cluster in
order to maintain DISA STIG compliance.

- [V-242383]  User-managed resources must be created in dedicated namespaces
- [V-242414] User pods must only use non-privileged host ports
- [V-242415] Secrets must not be stored as environment variables
- [V-242417] User functionality must be separate from management functions
  meaning all user pods must be in user specific namespaces rather than system
  namespaces

<!-- Links -->
[ports and services]: /snap/reference/ports-and-services/
[FIPS installation guide]: fips.md
[firewall configuration]: /snap/howto/networking/ufw.md
[USG tool]: https://documentation.ubuntu.com/security/docs/compliance/usg/
[Ubuntu Pro]: https://documentation.ubuntu.com/pro/start-here/#start-here
[Kubernetes Pod Security Admission documentation]: https://kubernetes.io/docs/concepts/security/pod-security-admission/
[upstream instructions]: https://kubernetes.io/docs/tasks/configure-pod-container/enforce-standards-admission-controller/
[upstream audit instructions]: https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
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
[V-242383]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242383
[V-242410]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242410
[V-242411]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242411
[V-242412]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242412
[V-242413]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242413
[V-242414]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242414
[V-242415]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242415
[V-242417]: https://stigviewer.com/stigs/kubernetes/2025-02-20/finding/V-242417
