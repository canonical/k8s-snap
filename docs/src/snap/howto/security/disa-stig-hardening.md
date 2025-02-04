# DISA STIG for {{product}}

Security Technical Implementation Guides (STIGs) are developed by the Defense
Information System Agency (DISA) for the U.S. Department of Defense (DoD).

The Kubernetes STIG contains guidelines on how to check and remediate various
potential security concerns for a Kubernetes deployment.

{{product}} aligns with many DISA STIG compliance recommendations by default.
However, additional hardening steps are required to fully meet the standard.

## What you'll need

This guide assumes the following:

- You have a bootstrapped {{product}} cluster (see the [getting started] guide)
- You have root or sudo access to the machine


## Critical post-deployment hardening steps

Follow these steps to apply critical security hardening steps to your cluster.
These steps address DISA STIG hardening recommendations
[V-242384](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242384),
[V-242385](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242385),
[V-242402](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242402),
[V-242403](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242403),
[V-242461](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242461),
[V-242462](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242462),
[V-242463](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242463),
[V-242464](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242464),
[V-242465](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242465),
and
[V-242434](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242434).

```{include} ../../../_parts/common_hardening.md
```

## DISA-STIG specific steps

The following steps are further security hardening steps recommended by DISA
 STIG. After addressing these recommendations correctly along with the critical
 post-deployment steps above, your cluster should be DISA STIG compliant.

#### [V-242383](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242383)

**User-managed resources must be created in dedicated namespaces**

**Upstream Finding Description**:

> Creating namespaces for user-managed resources is important when implementing
> Role-Based Access Controls (RBAC). RBAC allows for the authorization of users
> and helps support proper API server permissions separation and network micro
> segmentation. If user-managed resources are placed within the default
> namespaces, it becomes impossible to implement policies for RBAC permission,
> service account usage, network policies, and more.

**Remediation**:
>
> The Kubernetes System Administrators must manually inspect the services in
> all of the default namespaces to ensure there are no user-created resources
> within them:
>
>     kubectl -n default get all | grep -v "^(service|NAME)"
>     kubectl -n kube-public get all | grep -v "^(service|NAME)"
>     kubectl -n kube-node-lease get all | grep -v "^(service|NAME)"
>

#### [V-242410](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242410)

**The Kubernetes API Server must enforce ports, protocols, and services (PPS)
that adhere to the Ports, Protocols, and Services Management Category Assurance
List (PPSM CAL)**

**Upstream Finding Description**:

> Kubernetes API Server PPSs must be controlled and conform to the PPSM CAL.
> Those PPS that fall outside the PPSM CAL must be blocked. Instructions on the
> PPSM can be found in DoD Instruction 8551.01 Policy.

**Comments**:
>
> This STIG Finding relates to implementing PPSM CAL for kube-apiserver, and
> must be assessed manually by the Auditor.
>
> https://www.esd.whs.mil/portals/54/documents/dd/issuances/dodi/855101p.pdf
>

#### [V-242411](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242411)

**The Kubernetes Scheduler must enforce ports, protocols, and services (PPS)
that adhere to the Ports, Protocols, and Services Management Category Assurance
List (PPSM CAL)**

**Upstream Finding Description**:

> Kubernetes Scheduler PPS must be controlled and conform to the PPSM CAL.
> Those ports, protocols, and services that fall outside the PPSM CAL must be
> blocked. Instructions on the PPSM can be found in DoD Instruction 8551.01
> Policy.

**Comments**:
>
> This STIG Finding relates to implementing PPSM CAL for kube-scheduler, and
> must be assessed manually by the Auditor.
>
> https://www.esd.whs.mil/portals/54/documents/dd/issuances/dodi/855101p.pdf
>

#### [V-242412](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242412)

**The Kubernetes Controllers must enforce ports, protocols, and services (PPS)
that adhere to the Ports, Protocols, and Services Management Category Assurance
List (PPSM CAL)**

**Upstream Finding Description**:

> Kubernetes Controller ports, protocols, and services must be controlled and
> conform to the PPSM CAL. Those PPS that fall outside the PPSM CAL must be
> blocked. Instructions on the PPSM can be found in DoD Instruction 8551.01
> Policy.

**Comments**:

>
> This STIG Finding relates to implementing PPSM CAL for
> kube-controller-manager, and must be assessed manually by the Auditor.
>
> https://www.esd.whs.mil/portals/54/documents/dd/issuances/dodi/855101p.pdf
>

#### [V-242414](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242414)

**The Kubernetes cluster must use non-privileged host ports for user pods**

**Upstream Finding Description**:

> Privileged ports are those ports below 1024 and that require system
> privileges for their use. If containers can use these ports, the container
> must be run as a privileged user. Kubernetes must stop containers that try to
> map to these ports directly. Allowing non-privileged ports to be mapped to
> the container-privileged port is the allowable method when a certain port is
> needed. An example is mapping port 8080 externally to port 80 in the
> container.

**Comments**:
>
> The Kubernetes System Administrators must manually inspect the Pods in all of
> the default namespaces to ensure there are no user-created Pods with
> Containers exposing privileged port numbers (< 1024).
>
>     kubectl get pods --all-namespaces
>     kubectl -n NAMESPACE get pod PODNAME -o yaml | grep -i port
>

#### [V-242415](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242415)

**Secrets in Kubernetes must not be stored as environment variables**

**Upstream Finding Description**:

> Secrets, such as passwords, keys, tokens, and certificates should not be
> stored as environment variables. These environment variables are accessible
> inside Kubernetes by the "Get Pod" API call, and by any system, such as CI/CD
> pipeline, which has access to the definition file of the container. Secrets
> must be mounted from files or stored within password vaults.

**Comments**:
>
> The Kubernetes System Administrator must manually inspect the Environment of
> each user-created Pod to ensure there are no Pods passing information which
> the System Administrator may categorize as 'sensitive' (e.g. passwords,
> cryptographic keys, API tokens, etc).
>

#### [V-242417](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242417)

**Kubernetes must separate user functionality**

**Upstream Finding Description**:

> Separating user functionality from management functionality is a requirement
> for all the components within the Kubernetes Control Plane. Without the
> separation, users may have access to management functions that can degrade
> the Kubernetes architecture and the services being offered, and can offer a
> method to bypass testing and validation of functions before introduced into a
> production environment.

**Comments**:
>
> The Kubernetes System Administrators must manually inspect the Pods in all of
> the default namespaces to ensure there are no user-created Pods within them,
> and move them to dedicated user namespaces if present.
>
>     kubectl -n kube-system get pods
>     kubectl -n kube-public get pods
>     kubectl -n kube-node-lease get pods
>

#### [V-254800](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-254800)

**Kubernetes must have a Pod Security Admission control file configured**

**Upstream Finding Description**:

> An admission controller intercepts and processes requests to the Kubernetes
> API prior to persistence of the object, but after the request is
> authenticated and authorized.
>
> Kubernetes (> v1.23)offers a built-in Pod Security admission controller to
> enforce the Pod Security Standards. Pod security restrictions are applied at
> the namespace level when pods are created.
>
> The Kubernetes Pod Security Standards define different isolation levels for
> Pods. These standards define how to restrict the behavior of pods in a clear,
> consistent fashion.

**Comments**:
>
> This Finding stipulates the presence of a Pod Security Admission Control File
> which will need to be manually configured by the Kubernetes System
> Administrator on a per-organization basis.
>
> Instructions on how to configure an `--admission-control-config-file` for the
> Kube API Server of the k8s-snap can be found at:
>
> <!-- TODO: switch link to dedicated DISA Hardening page when published. -->
> https://documentation.ubuntu.com/canonical-kubernetes/latest/src/snap/howto/cis-hardening/#configure-auditing
>


## Manually audit DISA STIG hardening recommendations

For manual audits of DISA STIG hardening recommendations, please visit the
[Comprehensive Hardening Checklist][].


<!-- Links -->
[Hardening]:security/hardening.md
[Center for Internet Security (CIS)]:https://www.cisecurity.org/
[kube-bench]:https://aquasecurity.github.io/kube-bench/v0.6.15/
[CIS Kubernetes Benchmark]:https://www.cisecurity.org/benchmark/kubernetes
[getting started]: ../../tutorial/getting-started
[kube-bench release]: https://github.com/aquasecurity/kube-bench/releases
[Comprehensive Hardening Checklist]: auditing-steps.md#comprehensive-hardening-checklist
[upstream instructions]:https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
