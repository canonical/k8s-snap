# FIPS and DISA STIG

## Enable FIPS on kernel

ubuntu pro subscription attach

enable FIPS

```
sudo pro enable fips-updates
```

## Deploy K8s with FIPS

```
sudo snap install k8s --classic --channel=1.33-classic/stable
```

extra FIPS steps

Make sure that nodes are joined with their default hostname to
comply with V-242404

## DISA STIG

The Defense Information Systems Agency (DISA) Security Technical
Implementation Guides (STIGs) are comprehensive frameworks of
security requirements designed to protect U.S. Department of
Defense (DoD) systems and networks from cybersecurity threats.

STIGs are essentially recommendation guidelines for administrators
that aim to securely configure IT assets. One such STIG refers
to Kubernetes and here we show how {{product}} can be configured
to fully comply with it. As we discuss further below some of the
Kubernetes STIG guidelines have performance and functionality
implications while others need to be enforced by the cluster
administrations through user policies.


### Post deployment hardening steps

To configure your kubernetes cluster to fully comply with
DISA STIG the following post-deployment manual steps need
to be followed. Note that some of these steps
my affect the performance and functionality of the cluster.


#### [V-242434]

**Guideline:**  Kubernetes Kubelet must enable kernel protection

To comply with this guideline you need to do the following on every node:
Edit `/var/snap/k8s/common/args/kubelet` to add the argument 
`--protect-kernel-defaults=true`. Afterwards you need to restart kubelet with

```
sudo systemctl restart snap.k8s.kubelet
```

{{product}} does not set this argument by default as Kubelet may not start
if it finds kernel configurations incompatible with its expected defaults.
Not having this argument set by default offers greater compatibility.


#### [V-254800]

**Guideline:** Kubernetes must have a Pod Security Admission control
file configured

To comply with this guideline, you must configure a Pod Security Admission
control file for your Kubernetes cluster. This file defines the Pod Security
Standards (PSS) that are enforced at the namespace level.

Create a file named `pod-security-admission.yaml`
 under `/var/snap/k8s/common/etc/` with your desired policy.
You need to adjust the policy and exemptions as needed for your environment.
For more details, see the 
[Kubernetes Pod Security Admission documentation](https://kubernetes.io/docs/concepts/security/pod-security-admission/), 
which provides an overview of Pod Security Standards (PSS),
their enforcement levels, and configuration options for securing
Kubernetes namespaces. For example, to enforce the "restricted"
policy by default and allow "privileged" only in specific
namespaces:

```
sudo mkdir -p /var/snap/k8s/common/etc/
sudo sh -c 'cat >/var/snap/k8s/common/etc/pod-security-admission.yaml <<EOL
apiVersion: apiserver.config.k8s.io/v1
kind: AdmissionConfiguration
plugins:
  - name: PodSecurity
    configuration:
      apiVersion: pod-security.admission.config.k8s.io/v1alpha1
      kind: PodSecurityConfiguration
      defaults:
        enforce: "restricted"
        enforce-version: "latest"
        audit: "restricted"
        audit-version: "latest"
        warn: "restricted"
        warn-version: "latest"
      exemptions:
        namespaces: ["kube-system", "kube-public"]
EOL'
```

Edit `/var/snap/k8s/common/args/kube-apiserver` and add the argument 
`--admission-control-config-file=/var/snap/k8s/common/etc/pod-security-admission.yaml`
 so that the API server is aware of the policy configuration file.


Ensure the `PodSecurity` plugin is enabled in your API server arguments
found in `/var/snap/k8s/common/args/kube-apiserver`:

```
--enable-admission-plugins=...,PodSecurity,...
```

Finally, restart the API server

```
sudo systemctl restart snap.k8s.kube-apiserver
```

The setting a policy needs the input of the cluster administrator
and therefore {{product}} is not setting any.


#### [V-242384] and [V-242385]

**Guideline:** The Kubernetes Scheduler and Controller Manager must
have secure binding

To comply with these two guidelines edit the Kubernetes scheduler
arguments file `/var/snap/k8s/common/args/kube-scheduler`
and add the argument `--bind-address=127.0.0.1`. 
Do the same for the Kubernetes controller manager
(`/var/snap/k8s/common/args/kube-controller-manager`).

Afterwards, restart both services with:

```
sudo systemctl restart snap.k8s.kube-scheduler
sudo systemctl restart snap.k8s.kube-controller-manager
```

These steps needs to be performed on all control plane nodes.

{{product}} does not set these arguments by default as
this configuration may affect compatibility with workloads and metrics
collection.


#### [V-242400]

**Guideline:** The Kubernetes API server must have Alpha APIs disabled

To comply with this guideline edit
`/var/snap/k8s/common/args/kube-apiserver` in order to set the argument
`--feature-gate` for service `kube-apiserver` as appropriate.

It is possible to leave this argument unset completely.

If you'd like to explicitly set it, ensure it is set to one of:
`.*AllAlpha=false.*`, `.*AllAlpha=0.*`

Afterwards restart the `kube-apiserver` service with:

```
sudo systemctl restart snap.k8s.kube-apiserver
```

This change needs to be applied on all control plane nodes.


{{product}} by default does not disable any APIs as this may
cause certain workloads to fail and therefore harm workload
compatibility.


#### [V-242402] [V-242403] [V-242461] [V-242462] [V-242463] [V-242464] [V-242465]

**Guideline:** The Kubernetes API Server must have an audit log path set

**Guideline:** Kubernetes API Server must generate audit records that identify
what type of event has occurred, identify the source of the event, contain the
event results, identify any users, and identify any containers associated with
the event

**Guideline:** Kubernetes API Server audit logs must be enabled

**Guideline:** The Kubernetes API Server must be set to audit log max size

**Guideline:** The Kubernetes API Server must be set to audit log
maximum backup

**Guideline:** The Kubernetes API Server audit log retention must be set

**Guideline:** The Kubernetes API Server audit log path must be set


On every control plane node
create an `audit-policy.yaml` file under `/var/snap/k8s/common/etc/` and
specify the level of auditing you desire based on
the [upstream instructions]. Here is a minimal example of such
a policy file.

```
sudo mkdir -p /var/snap/k8s/common/etc/
sudo sh -c 'cat >/var/snap/k8s/common/etc/audit-policy.yaml <<EOL
# Log all requests at the Metadata level.
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: Metadata
EOL'
```

Enable auditing at the API server level by adding the following arguments.

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-apiserver <<EOL
--audit-log-path=/var/log/kubernetes/audit.log
--audit-log-maxage=30
--audit-log-maxbackup=10
--audit-log-maxsize=100
--audit-policy-file=/var/snap/k8s/common/etc/audit-policy.yaml
EOL'
```

Restart the API server:

```
sudo systemctl restart snap.k8s.kube-apiserver
```


{{product}} does not enable audit logging by default as it may
incur performance penalties in the form of increased disk I/O,
which can lead to slower response times and reduced overall
cluster efficiency, especially under heavy workloads.

#### [V-245541]

**Guideline:** Kubernetes Kubelet must not disable timeouts

Idle connections from the Kubelet can be used by unauthorized users to
perform malicious activity to the nodes, pods, containers, and cluster within
the Kubernetes Control Plane.

To comply with this guideline edit `/var/snap/k8s/common/args/kubelet`
and set the argument `--streaming-connection-idle-timeout` to `5m`.


Afterwards restart `kubelet`.

```
sudo systemctl restart snap.k8s.kubelet
```


### Runtime policies and good practices

In addition to the deployment specific configuration DISA STIG recommends
a few good practices that need to be followed throughout the life of the
cluster. It is up to the cluster administration to enforce such
practices any way they see fit.


#### [V-242415]

**Guideline:** Secrets in Kubernetes must not be stored as environment
variables

> The Kubernetes System Administrator must manually inspect the Environment of
> each user-created Pod to ensure there are no Pods passing information which
> the System Administrator may categorize as 'sensitive' (e.g. passwords,
> cryptographic keys, API tokens, etc).

Inspect the environment of each user-created pod to ensure there is no
sensitive information (e.g. passwords, cryptographic keys, API tokens, etc).

```
sudo k8s kubectl exec -it PODNAME -n kube-system -- env
```


#### [V-242393]

**Guideline:** Kubernetes Worker Nodes must not have sshd service running

> Worker Nodes are maintained and monitored by the Control Plane. Direct access
> and manipulation of the nodes should not take place by administrators. Worker
> nodes should be treated as immutable and updated via replacement rather than
> in-place upgrades.

#### [V-242395]

**Guideline:** Kubernetes dashboard must not be enabled

> While the Kubernetes dashboard is not inherently insecure on its own, it is
> often coupled with a misconfiguration of Role-Based Access control (RBAC)
> permissions that can unintentionally over-grant access. It is not commonly
> protected with "NetworkPolicies", preventing all pods from being able to
> reach it. In increasingly rare circumstances, the Kubernetes dashboard is
> exposed publicly to the internet.

#### [V-242396]

**Guideline:**  Kubernetes Kubectl cp command must give expected access and
results

> One of the tools heavily used to interact with containers in the Kubernetes
> cluster is kubectl. The command is the tool System Administrators used to
> create, modify, and delete resources. One of the capabilities of the tool is
> to copy files to and from running containers (i.e., kubectl cp). The command
> uses the "tar" command of the container to copy files from the container to
> the host executing the "kubectl cp" command. If the "tar" command on the
> container has been replaced by a malicious user, the command can copy files
> anywhere on the host machine. This flaw has been fixed in later versions of
> the tool. It is recommended to use kubectl versions newer than 1.12.9.

#### [V-242414]

**Guideline:** The Kubernetes cluster must use non-privileged host ports for
user pods

> Privileged ports are those ports below 1024 and that require system
> privileges for their use. If containers can use these ports, the container
> must be run as a privileged user. Kubernetes must stop containers that try to
> map to these ports directly. Allowing non-privileged ports to be mapped to
> the container-privileged port is the allowable method when a certain port is
> needed. An example is mapping port 8080 externally to port 80 in the
> container.

> The Kubernetes System Administrators must manually inspect the Pods in all of
> the default namespaces to ensure there are no user-created Pods with
> Containers exposing privileged port numbers (< 1024).
>
>     kubectl get pods --all-namespaces
>     kubectl -n NAMESPACE get pod PODNAME -o yaml | grep -i port
>


#### [V-242417]

**Guideline:** Kubernetes must separate user functionality

> The Kubernetes System Administrators must manually inspect the Pods in all of
> the default namespaces to ensure there are no user-created Pods within them,
> and move them to dedicated user namespaces if present.
>
>     kubectl -n kube-system get pods
>     kubectl -n kube-public get pods
>     kubectl -n kube-node-lease get pods
>

#### [V-242383]

**Guideline:** User-managed resources must be created in dedicated namespaces

> Creating namespaces for user-managed resources is important when implementing
> Role-Based Access Controls (RBAC). RBAC allows for the authorization of users
> and helps support proper API server permissions separation and network micro
> segmentation. If user-managed resources are placed within the default
> namespaces, it becomes impossible to implement policies for RBAC permission,
> service account usage, network policies, and more.

Manually inspect the services in all of the default namespaces to ensure there
are no user-created resources:

```
kubectl -n default get all | grep -v "^(service|NAME)"
kubectl -n kube-public get all | grep -v "^(service|NAME)"
kubectl -n kube-node-lease get all | grep -v "^(service|NAME)"
```


## Full DISA STIG audit

If you would like to manually audit any of the other
DISA (Defense Information Systems Agency)
STIG (Security Technical Implementation Guides) recommendations,
visit our [DISA STIG assessment page], where you will find detailed
instructions and tools to evaluate your Kubernetes cluster against
DISA STIG compliance requirements.


<!-- Links -->
[DISA STIG assessment page]: disa-stig-assessment.md
[upstream instructions]:https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[rate limits]:https://kubernetes.io/docs/reference/config-api/apiserver-eventratelimit.v1alpha1
[controlling_access]: https://kubernetes.io/docs/concepts/security/controlling-access/
[access_authn_authz]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[encryption_at_rest]: https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/
[authorization_mode]: https://kubernetes.io/docs/reference/access-authn-authz/authorization/#authorization-modules
