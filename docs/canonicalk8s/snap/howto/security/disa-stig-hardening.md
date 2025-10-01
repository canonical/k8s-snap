# DISA STIG Hardening

This guide iterates over the steps necessary to apply DISA STIG hardening to a
{{ product }} cluster.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine
- You have Ubuntu Pro enabled on your system. For more information, see
  [Ubuntu Pro documentation].

## DISA STIG host compliance
<!-- TODO: Ask Niamh if we should include this section -->

DISA STIG host compliance is achieved by running the [usg tool]
(`usg fix disa_stig`) that is also part
of the PRO tool set. To install the usg tool, run:

```
sudo apt update
sudo apt install usg 
```

DISA STIG compliance for Ubuntu hosts can be achieved using the usg tool
installed above.

To generate a compliance audit report (without applying changes):

```
sudo usg audit disa_stig
```

```{warning}
The following command applies rule [V-270714] which will cause issues using
accounts with an empty password.
```

To automatically apply the recommended hardening changes:

```
sudo usg fix disa_stig
```

## Apply DISA Kubernetes STIG host rules

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

## Deploy and bootstrap stig-compliant nodes

For your convenience, we have provided template configuration files that can be
used to configure Kubernetes service arguments that align with DISA STIG
requirements.

### Bootstrap the first control-plane node

To initialize the first control plane node with the necessary arguments for
disa-stig compliance, run:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

```
sudo k8s bootstrap --file /var/snap/k8s/common/etc/templates/disa-stig/bootstrap.yaml
sudo k8s status --wait-ready
```

You can then optionally join additional control plane or worker nodes following
the sections below.

Through this configuration file, the following rules are applied:

- [V-242434] Kubernetes Kubelet must enable kernel protection
- [V-245541] Kubernetes Kubelet must not disable timeouts
- [V-242402] [V-242403] [V-242461] [V-242462] [V-242463] [V-242464]
  [V-242465] The Kubernetes API Server must have an audit log
  configured
- [V-254800] Kubernetes must have a Pod Security Admission control file
  configured
- [V-242400] The Kubernetes API server must have Alpha APIs disabled
- [V-254800] Kubernetes must have a Pod Security Admission control file
configured
- [V-242384] The Kubernetes Scheduler must have secure binding
- [V-242385] The Kubernetes Controller Manager must have secure binding

### Join additional control plane nodes

To join another control plane node to the cluster, first retrieve the join token
from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname>
```

On the joining control plane node, run:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/control-plane.yaml <join-token>
```

Through this configuration file, the following rules are applied:

- [V-242434] Kubernetes Kubelet must enable kernel protection
- [V-245541] Kubernetes Kubelet must not disable timeouts
- [V-242402] [V-242403] [V-242461] [V-242462] [V-242463] [V-242464] [V-242465] The
  Kubernetes API Server must have an audit log configured
- [V-254800] Kubernetes must have a Pod Security Admission control file configured
- [V-242400] The Kubernetes API server must have Alpha APIs disabled
- [V-254800] Kubernetes must have a Pod Security Admission control file
configured
- [V-242384] The Kubernetes Scheduler must have secure binding
- [V-242385] The Kubernetes Controller Manager must have secure binding

### Join worker nodes

To join a worker node to the cluster, first retrieve the join token from an
existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname> --worker
```

On the joining worker node, run:

```{literalinclude} ../../../_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/worker.yaml <join-token>
```

Through this configuration file, the following rules are applied:

- [V-242434] Kubernetes Kubelet must enable kernel protection
- [V-245541] Kubernetes Kubelet must not disable timeouts

### Disable SSH on the Worker Nodes

If ssh is not needed to access the worker nodes it is recommended you disable
the ssh:

```
sudo systemctl disable ssh.service ssh.socket
```

```{note}
According to rule [V-242393] and [V-242394] Kubernetes Worker Nodes must not have
sshd service running or enabled. The host STIG rule [V-270665] on the other hand
expects sshd to be installed on the host. To comply with both rules, leave SSH
installed, but disable the service.
```

## Control Plane Alternative Configurations

### Kubernetes must have a Pod Security Admission control file configured

To comply with rule [V-254800], you must configure a Pod Security Admission
control file for your Kubernetes cluster. This file defines the Pod Security
Standards (PSS) that are enforced at the namespace level.

The STIG templates we provide to bootstrap/join nodes configure the Pod Security
admission controller to comply with these recommendations.

By default, the bootstrap configuration template will point to
`/var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml`,
which sets the pod security policy to “baseline”, a minimally restrictive policy
that prevents known privilege escalations.

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
   the `--admission-control-config-file` path used when you bootstrap/join nodes.

For more details, see the [Kubernetes Pod Security Admission documentation],
which provides an overview of Pod Security Standards (PSS), their enforcement
levels, and configuration options.

## The Kubernetes API Server must have an audit log configured

This applies to rules:

- [V-242402]
- [V-242403]
- [V-242461]
- [V-242462]
- [V-242463]
- [V-242464]
- [V-242465]

The STIG templates we provide to bootstrap/join nodes configure the Kubernetes
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
3. Create your own audit policy based on the [upstream audit instructions]
   and adjust the `--audit-policy-file` path used when you bootstrap/join
   nodes to use it.

Canonical Kubernetes does not enable audit logging by default as it may incur
performance penalties in the form of increased disk I/O, which can lead to
slower response times and reduced overall cluster efficiency, especially under
heavy workloads.

<!-- Links -->
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