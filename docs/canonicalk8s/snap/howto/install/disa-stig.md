# How to install {{ product }} with DISA STIG hardening

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
- Ubuntu Pro subscription 

## Configure the host

[DISA STIG host OS] compliance is achieved by running the [USG tool] that is 
part of the PRO tool set and running some additional manual steps.

### Configure the firewall

DISA STIG for the host recommends enabling the host firewall (UFW). This is not
done automatically through the USG tool and we recommend following our guide to 
[configure UFW]. This should be done *before* applying the host STIG steps and 
will help avoid connectivity issues that often happen when 
enabling UFW with the default configuration.

### Apply host STIG

To install the USG tool:

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
The rule [V-270714] will be applied in the following command. This prevents using accounts 
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

### Configure the kernel

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

{{product}} provides example [configuration files] to apply
DISA STIG specific settings on cluster formation and node-join. If you would 
like to apply the default settings, jump to 
[initializing the cluster](#initialize-the-cluster).


<!-- ### Review default configuration

Before bootstrapping or joining control plane nodes, review the
example [configuration files]. If you want to tailor the logging and the Pod Security Admission
control file see the [advanced configuration options](#advanced-configruation-options).
Once a node is configured, changing certain settings is more difficult
and may require re-deploying the node or cluster. -->


### Initialize the cluster

<!-- ```{attention}
Before bootstrapping or joining control plane nodes, review the
example [configuration files] and [advanced configuration options](#advanced-configruation-options).
Once a node is configured, changing certain settings is more difficult
and may require re-deploying the node or cluster.
``` -->

```{attention}
Before bootstrapping or joining control plane nodes, review the
example [configuration files].
Once a node is configured, changing certain settings is more difficult
and may require re-deploying the node or cluster.
``` 

Bootstrap the first control plane node using the
example bootstrap configuration file which will apply the relevant Kubernetes 
STIG recommendations:

```
sudo k8s bootstrap --file /var/snap/k8s/common/etc/templates/disa-stig/bootstrap.yaml
sudo k8s status --wait-ready
```

<!-- ### Join additional nodes -->

<!-- ### Apply Kubernetes control plane STIG -->
<!-- ### Set up control plane nodes -->

### Join control plane nodes

First retrieve a join token from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname>
```

Then join the new control plane node using the
example node-join configuration file which will apply the Kubernetes STIG 
recommendations:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/control-plane.yaml <join-token>
```

<!-- ### Apply Kubernetes worker STIG -->

### Join worker nodes

First retrieve a join token from an existing control plane node:

```
sudo k8s get-join-token <joining-node-hostname> --worker
```

Then join the new worker node using the example node-join configuration file 
which will apply the Kubernetes STIG recommendations:

```
sudo k8s join-cluster --file=/var/snap/k8s/common/etc/templates/disa-stig/worker.yaml <join-token>
```

If SSH is not needed to access the worker nodes it is recommended you disable
the SSH service:

```
sudo systemctl disable ssh.service ssh.socket
```

```{note}
According to Kubernetes STIG rule {ref}`242393` and {ref}`242394`, Kubernetes 
worker nodes must not have sshd service running or enabled. The host STIG 
rule [V-270665] on the other hand expects sshd to be installed on the host. 
To comply with both rules, leave SSH installed, but disable the service. 
Alternatively, SSH can be removed and the exception documented.
```

<!-- ## Post-deployment requirements -->
## Manage post-deployment Kubernetes STIG

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
- {ref}`242414`: User pods must only use non-privileged host ports
- {ref}`242415`: Secrets must not be stored as environment variables
- {ref}`242417`: User functionality must be separate from management functions
   meaning all user pods must be in user specific namespaces rather than system
   namespaces

## Post deployment configruation options 

The STIG configuration files provided to 
[initialize the cluster](#initialize-the-cluster) and 
[join control plane nodes](#join-control-plane-nodes) can be
adjusted to suit your specific needs. 

### Pod Security Admission control file

To comply with rule {ref}`254800`, you must configure a Pod Security Admission
control file for your Kubernetes cluster. This file defines the Pod Security
Standards (PSS) that are enforced at the namespace level. For more details, see 
the [Kubernetes Pod Security Admission documentation].

Inspect the default Pod Security Admission control file used when deploying 
control plane nodes with the example DISA STIG configuration files above:

```
sudo cat /var/snap/k8s/common/etc/configurations/pod-security-admission-baseline.yaml
```

This pod security policy is set to “baseline”, a minimally restrictive policy 
that prevents known privilege escalations. Edit this file as needed to meet your
cluseter's needs based on [upstream instructions]. 

<!-- This policy may be insufficient or 
impractical in some situations, in which case it needs to be adjusted.  -->

Alternavtively, {{product}} also provides a more restrictive configuration 
file:

```
sudo cat /var/snap/k8s/common/etc/configurations/pod-security-admission-restricted.yaml
```

If you would like to use this file instead, set the 
`--admission-control-config-file` path in the bootstrap and control plane 
configuration files to
`/var/snap/k8s/common/etc/configurations/pod-security-admission-restricted.yaml`.

If neither provided configurations meet your cluster's needs, create your own 
audit policy based on the [upstream instructions] and adjust the 
`--admission-control-config-file` path used in the configuration files.

### Kubernetes API Server audit log

To comply with rules {ref}`242402`, {ref}`242403`, {ref}`242461`, {ref}`242462`,
{ref}`242463`, {ref}`242464`, and {ref}`242465` you must configure the 
Kubernetes API Server audit log. 

Inspect the default audit policy used when deploying control plane nodes with 
the example DISA STIG configuration files above:

```
sudo cat /var/snap/k8s/common/etc/configurations/audit-policy.yaml
```

This configures logging of all (non-resource) events with request metadata, 
request body, and response body as recommended by {ref}`242403`. This level of 
logging may be impractical for some situations, in which case the settings would
need to be adjusted and an exception put in place. Edit this file to suit
your needs based on the [upstream audit instructions].

Alternavtively, {{product}} also provides another audting configuration that 
provides the same level of logging but only for events in the kube-system 
namespace:

```
sudo cat /var/snap/k8s/common/etc/configurations/audit-policy-kube-system.yaml
```

If you would like to use this file instead, set the `--audit-policy-file` path 
used when you bootstrap/join nodes to use 
`/var/snap/k8s/common/etc/configurations/audit-policy-kube-system.yaml`.

If neither provided configurations meet your clusters needs, create your own 
audit policy based on the [upstream audit instructions] and adjust the 
`--audit-policy-file` path used when you bootstrap/join nodes to use it.


<!-- ## Further reading  -->
## Reference material
<!-- ## Further DISA STIG material -->

### Configuration example files

See docs/canonicalk8s/snap/reference/config-files/disa-stig-config.md if you want to see the STIGs applied

### Full DISA STIG Kubernetes list 

See docs/canonicalk8s/snap/reference/disa-stig-audit.md for full audit

<!-- Links -->
[ports and services]: /snap/reference/ports-and-services/
[FIPS installation guide]: fips.md
[configure UFW]: /snap/howto/networking/ufw.md
[configuration files]: /snap/reference/config-files/disa-stig-config.md
[USG tool]: https://documentation.ubuntu.com/security/docs/compliance/usg/
[Ubuntu Pro]: https://documentation.ubuntu.com/pro/start-here/#start-here
[Kubernetes Pod Security Admission documentation]: https://kubernetes.io/docs/concepts/security/pod-security-admission/
[upstream instructions]: https://kubernetes.io/docs/tasks/configure-pod-container/enforce-standards-admission-controller/
[upstream audit instructions]: https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[V-270714]: https://www.stigviewer.com/stigs/canonical_ubuntu_2404_lts/2025-05-16/finding/V-270714
[V-270665]: https://www.stigviewer.com/stigs/canonical_ubuntu_2404_lts/2025-05-16/finding/V-270665
[DISA STIG host OS]: https://www.stigviewer.com/stigs/canonical_ubuntu_2404_lts

