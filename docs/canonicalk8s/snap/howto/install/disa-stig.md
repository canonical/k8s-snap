# How to install {{ product }} with DISA STIG hardening

```{versionadded} release-1.34
```

DISA Security Technical Implementation Guides (STIGs) provide hardening
guidelines for meeting regulations from the U.S. Government and Department of
Defense (DoD).

{{product}} aligns by default with many of these recommendations as they are
expected to benefit most users. This guide provides additional steps to meet
all DISA STIG guidelines for both Kubernetes and the host OS.

## Prerequisites

- FIPS compliance is required by the DISA STIG. This guide assumes you have
already followed our [FIPS installation guide], but stopped after installing
{{product}} without following the steps to bootstrap/join the cluster. Instead
continue here to first complete additional steps needed for DISA STIG
compliance.
- [Ubuntu Pro] subscription

## Configure the host

### Configure the firewall

DISA STIG for the host recommends enabling the host firewall. This is not
done automatically and we recommend following our guide to 
[configure Uncomplicated Firewall (UFW)]. This should be done *before* 
applying the host STIG steps and will help avoid connectivity issues that often 
happen when enabling UFW with the default configuration.

### Apply host STIG

The [USG tool] which is part of the PRO tool set can be run to automatically 
apply most other [DISA STIG host OS] recommendations. To install the USG tool:

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

The USG tool in the following command will apply host STIG password rules such 
as [V-270714] or [V-260570] that will prevent using accounts with empty 
passwords to access this machine. 

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

After rebooting, you can re-run `sudo usg audit disa_stig` to verify host
compliance. You may need to create a tailoring file to disable certain rules
and document exceptions
to reach your desired compliance state. See the USG
[tailoring guidance] for help with rule customization and manual remediation.

Some rules commonly require manual remediation or exceptions,
including but not limited to:

- **`content_rule_dir_perms_world_writable_sticky_bits`**: Upstream Kubernetes
violates this rule when creating workloads due to how it manages volume
permissions (see [Kubernetes issue #125876]). You will need an exception
for this rule.
- **`content_rule_only_allow_dod_certs`**: By default, {{product}} uses
self-signed [certificates] which may be
acceptable in your environment. If you require certificates signed by a DoD CA
for compliance requirements, you can configure custom
certificates via the [configuration files].

### Configure the kernel

DISA STIG recommends enabling `--protect-kernel-defaults=true` so that kubelet
will not modify kernel flags. This requires the kernel to be configured in
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
Ensure that the configuration in `/etc/sysctl.d/99-kubelet.conf` is not
overridden by another configuration file with higher precedence.
```

## Set configuration options 

{{product}} provides example configuration files to automatically apply
DISA STIG specific settings on cluster formation and node join. Once a node is 
configured, changing certain settings is more difficult and may require 
re-deploying the node or cluster. If you are happy to apply the default 
settings, jump to [initializing the cluster](#initialize-the-cluster). 
Otherwise, choose the configuration options that are best suited for your 
cluster. 

### Pod Security Admission control file

To comply with rule {ref}`254800`, you must configure a Pod Security Admission
control file for your Kubernetes cluster. This file defines the Pod Security
Standards (PSS) that are enforced at the namespace level. 

|                           |                                                                                                                                                                                                                   |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Current default           | `/var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml`. This pod security policy is set to “baseline”, a minimally restrictive policy that prevents known privilege escalations.          |
| Alternative configuration | `/var/snap/k8s/common/etc/configurations/pod-security-admission-restricted.yaml`. This pod security policy is set to "restricted", a heavily restricted policy that follows current pod hardening best practices. |

These policies can be edited based on [upstream instructions].

Set the `--admission-control-config-file` path in the bootstrap and control 
plane configuration files located at 
`/var/snap/k8s/common/etc/configurations/disa-stig/` to whichever policy best 
matches your cluster's needs. 

### Kubernetes API Server audit log

To comply with rules {ref}`242402`, {ref}`242403`, {ref}`242461`, {ref}`242462`,
{ref}`242463`, {ref}`242464`, and {ref}`242465` you must configure the 
Kubernetes API Server audit log. 

|                           |                                                                                                                                                                                                                                                                                                                                                             |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Current default           | `/var/snap/k8s/common/etc/configurations/audit-policy.yaml`. This configures logging of all (non-resource) events with request metadata, request body, and response body as recommended by {ref}`242403`. This level of logging may be impractical for some situations, in which case the settings would need to be adjusted and an exception put in place. |
| Alternative configuration | `/var/snap/k8s/common/etc/configurations/audit-policy-kube-system.yaml`. This provides the same level of logging, but only for events in the kube-system namespace.                                                                                                                                                                                         |

These policies can be edited based on [upstream audit instructions].

Set the `--audit-policy-file` path in the bootstrap and control plane 
configuration files located at 
`/var/snap/k8s/common/etc/configurations/disa-stig/` to use whichever policy 
best matches your cluster's needs.

### Default configuration files 

Review the remaining parameters in the example configuration YAML files located 
at `/var/snap/k8s/common/etc/configurations/disa-stig/` and ensure they are set 
according to your needs. The [DISA STIG configuration files] reference page 
details what hardening recommendations have been applied in the example 
configuration files. 

## Apply Kubernetes STIG  

### Initialize the cluster

Bootstrap the first control plane node using the
example bootstrap configuration file which will apply the relevant Kubernetes 
STIG recommendations:

```
sudo k8s bootstrap --file /var/snap/k8s/common/etc/configurations/disa-stig/bootstrap.yaml
sudo k8s status --wait-ready
```

### Join control plane nodes

First retrieve a join token from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname>
```

Then join the new control plane node using the
example control plane node join configuration file:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/configurations/disa-stig/control-plane.yaml <join-token>
```

### Join worker nodes

First retrieve a join token from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname> --worker
```

Then join the new worker node using the example worker node join configuration 
file:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/configurations/disa-stig/worker.yaml <join-token>
```

If SSH is not needed to access worker nodes, it is recommended you disable
the SSH service:

```
sudo systemctl disable ssh.service ssh.socket
```

```{note}
According to rule {ref}`242393` and {ref}`242394` Kubernetes worker nodes must not
have sshd service running or enabled. The host STIG on the
other hand expects sshd to be installed on the host (rule [V-270665] or [V-260523]). To comply with both
rules, leave SSH installed, but disable the service. Alternatively, SSH
can be removed and the exception documented.
```

## Post-deployment Kubernetes STIG requirements

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
registered. This step needs to be followed after the initial deployment and
anytime the list of ports, protocols, and services used by your cluster changes
(for instance each time a new service is exposed externally).
- {ref}`242393` and {ref}`242394`: SSH service must not be running or enabled on
   worker nodes (see [Join worker nodes](#join-worker-nodes))
- {ref}`242414`: User pods must only use non-privileged host ports
- {ref}`242415`: Secrets must not be stored as environment variables
- {ref}`242417`: User functionality must be separate from management functions
   meaning all user pods must be in user specific namespaces rather than system
   namespaces
- {ref}`242443`: Kubernetes components must be regularly updated to avoid
   vulnerabilities. We recommend using the latest revision of a<a href=
   "https://ubuntu.com/about/release-cycle?product=kubernetes&release=canonical+kubernetes&version=all">
   supported version</a> of {{product}}.

## Reference material

- If you would like to see what DISA STIG rules are applied in the example 
bootstrap, control plane and worker node configuration files provided, see 
the [DISA STIG configuration files] page.
- The [DISA STIG audit] page contains a list of all the DISA STIG 
recommendations and details how they apply to {{product}}.

<!-- Links -->
[ports and services]: /snap/reference/ports-and-services/
[FIPS installation guide]: fips.md
[configure Uncomplicated Firewall (UFW) ]: /snap/howto/networking/ufw.md
[USG tool]: https://documentation.ubuntu.com/security/docs/compliance/usg/
[Ubuntu Pro]: https://documentation.ubuntu.com/pro/start-here/#start-here
[certificates]: /snap/reference/certificates.md
[tailoring guidance]: https://documentation.ubuntu.com/security/compliance/usg/disa-customize/
[configuration files]: /snap/reference/config-files/index
[upstream instructions]: https://kubernetes.io/docs/tasks/configure-pod-container/enforce-standards-admission-controller/
[upstream audit instructions]: https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[V-270714]: https://www.stigviewer.com/stigs/canonical_ubuntu_2404_lts/2025-05-16/finding/V-270714
[V-260570]: https://www.stigviewer.com/stigs/canonical_ubuntu_2204_lts/2025-05-16/finding/V-260570
[V-270665]: https://www.stigviewer.com/stigs/canonical_ubuntu_2404_lts/2025-05-16/finding/V-270665
[V-260523]: https://www.stigviewer.com/stigs/canonical_ubuntu_2204_lts/2025-05-16/finding/V-260523
[Kubernetes issue #125876]: https://github.com/kubernetes/kubernetes/issues/125876
[DISA STIG host OS]: https://www.stigviewer.com/stigs/canonical_ubuntu_2404_lts
[DISA STIG configuration files]: /snap/reference/config-files/disa-stig-config.md
[DISA STIG audit]: /snap/reference/disa-stig-audit.md
[configuration yaml files]: https://github.com/canonical/k8s-snap/tree/main/k8s/resources/configurations/disa-stig
