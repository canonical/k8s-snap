## Kubernetes Policies
### RBAC and Service Accounts
Control **5.1.1**

Description: `Ensure that the cluster-admin role is only used where required (Manual)`

Remediation:
```
Identify all clusterrolebindings to the cluster-admin role. Check if they are used and
if they need this role or if they could use a role with fewer privileges.
Where possible, first bind users to a lower privileged role and then remove the
clusterrolebinding to the cluster-admin role :
kubectl delete clusterrolebinding [name]
```

Control **5.1.2**

Description: `Minimize access to secrets (Manual)`

Remediation:
```
Where possible, remove get, list and watch access to Secret objects in the cluster.
```

Control **5.1.3**

Description: `Minimize wildcard use in Roles and ClusterRoles (Manual)`

Remediation:
```
Where possible replace any use of wildcards in clusterroles and roles with specific
objects or actions.
```

Control **5.1.4**

Description: `Minimize access to create pods (Manual)`

Remediation:
```
Where possible, remove create access to pod objects in the cluster.
```

Control **5.1.5**

Description: `Ensure that default service accounts are not actively used. (Manual)`

Remediation:
```
Create explicit service accounts wherever a Kubernetes workload requires specific access
to the Kubernetes API server.
Modify the configuration of each default service account to include this value
automountServiceAccountToken: false
```

Control **5.1.6**

Description: `Ensure that Service Account Tokens are only mounted where necessary (Manual)`

Remediation:
```
Modify the definition of pods and service accounts which do not need to mount service
account tokens to disable it.
```

Control **5.1.7**

Description: `Avoid use of system:masters group (Manual)`

Remediation:
```
Remove the system:masters group from all users in the cluster.
```

Control **5.1.8**

Description: `Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual)`

Remediation:
```
Where possible, remove the impersonate, bind and escalate rights from subjects.
```

### Pod Security Standards
Control **5.2.1**

Description: `Ensure that the cluster has at least one active policy control mechanism in place (Manual)`

Remediation:
```
Ensure that either Pod Security Admission or an external policy control system is in place
for every namespace which contains user workloads.
```

Control **5.2.2**

Description: `Minimize the admission of privileged containers (Manual)`

Remediation:
```
Add policies to each namespace in the cluster which has user workloads to restrict the
admission of privileged containers.
```

Control **5.2.3**

Description: `Minimize the admission of containers wishing to share the host process ID namespace (Automated)`

Remediation:
```
Add policies to each namespace in the cluster which has user workloads to restrict the
admission of `hostPID` containers.
```

Control **5.2.4**

Description: `Minimize the admission of containers wishing to share the host IPC namespace (Automated)`

Remediation:
```
Add policies to each namespace in the cluster which has user workloads to restrict the
admission of `hostIPC` containers.
```

Control **5.2.5**

Description: `Minimize the admission of containers wishing to share the host network namespace (Automated)`

Remediation:
```
Add policies to each namespace in the cluster which has user workloads to restrict the
admission of `hostNetwork` containers.
```

Control **5.2.6**

Description: `Minimize the admission of containers with allowPrivilegeEscalation (Automated)`

Remediation:
```
Add policies to each namespace in the cluster which has user workloads to restrict the
admission of containers with `.spec.allowPrivilegeEscalation` set to `true`.
```

Control **5.2.7**

Description: `Minimize the admission of root containers (Automated)`

Remediation:
```
Create a policy for each namespace in the cluster, ensuring that either `MustRunAsNonRoot`
or `MustRunAs` with the range of UIDs not including 0, is set.
```

Control **5.2.8**

Description: `Minimize the admission of containers with the NET_RAW capability (Automated)`

Remediation:
```
Add policies to each namespace in the cluster which has user workloads to restrict the
admission of containers with the `NET_RAW` capability.
```

Control **5.2.9**

Description: `Minimize the admission of containers with added capabilities (Automated)`

Remediation:
```
Ensure that `allowedCapabilities` is not present in policies for the cluster unless
it is set to an empty array.
```

Control **5.2.10**

Description: `Minimize the admission of containers with capabilities assigned (Manual)`

Remediation:
```
Review the use of capabilites in applications running on your cluster. Where a namespace
contains applicaions which do not require any Linux capabities to operate consider adding
a PSP which forbids the admission of containers which do not drop all capabilities.
```

Control **5.2.11**

Description: `Minimize the admission of Windows HostProcess containers (Manual)`

Remediation:
```
Add policies to each namespace in the cluster which has user workloads to restrict the
admission of containers that have `.securityContext.windowsOptions.hostProcess` set to `true`.
```

Control **5.2.12**

Description: `Minimize the admission of HostPath volumes (Manual)`

Remediation:
```
Add policies to each namespace in the cluster which has user workloads to restrict the
admission of containers with `hostPath` volumes.
```

Control **5.2.13**

Description: `Minimize the admission of containers which use HostPorts (Manual)`

Remediation:
```
Add policies to each namespace in the cluster which has user workloads to restrict the
admission of containers which use `hostPort` sections.
```

### Network Policies and CNI
Control **5.3.1**

Description: `Ensure that the CNI in use supports NetworkPolicies (Manual)`

Remediation:
```
If the CNI plugin in use does not support network policies, consideration should be given to
making use of a different plugin, or finding an alternate mechanism for restricting traffic
in the Kubernetes cluster.
```

Control **5.3.2**

Description: `Ensure that all Namespaces have NetworkPolicies defined (Manual)`

Remediation:
```
Follow the documentation and create NetworkPolicy objects as you need them.
```

### Secrets Management
Control **5.4.1**

Description: `Prefer using Secrets as files over Secrets as environment variables (Manual)`

Remediation:
```
If possible, rewrite application code to read Secrets from mounted secret files, rather than
from environment variables.
```

Control **5.4.2**

Description: `Consider external secret storage (Manual)`

Remediation:
```
Refer to the Secrets management options offered by your cloud provider or a third-party
secrets management solution.
```

### Extensible Admission Control
Control **5.5.1**

Description: `Configure Image Provenance using ImagePolicyWebhook admission controller (Manual)`

Remediation:
```
Follow the Kubernetes documentation and setup image provenance.
```

### General Policies
Control **5.7.1**

Description: `Create administrative boundaries between resources using namespaces (Manual)`

Remediation:
```
Follow the documentation and create namespaces for objects in your deployment as you need
them.
```

Control **5.7.2**

Description: `Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual)`

Remediation:
```
Use `securityContext` to enable the docker/default seccomp profile in your pod definitions.
An example is as below:
  securityContext:
    seccompProfile:
      type: RuntimeDefault
```

Control **5.7.3**

Description: `Apply SecurityContext to your Pods and Containers (Manual)`

Remediation:
```
Follow the Kubernetes documentation and apply SecurityContexts to your Pods. For a
suggested list of SecurityContexts, you may refer to the CIS Security Benchmark for Docker
Containers.
```

Control **5.7.4**

Description: `The default namespace should not be used (Manual)`

Remediation:
```
Ensure that namespaces are created to allow for appropriate segregation of Kubernetes
resources and that all new resources are created in a specific namespace.
```

