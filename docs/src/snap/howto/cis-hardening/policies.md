
## Kubernetes Policies

### RBAC and Service Accounts

Control **5.1.1**
Description: `Ensure that the cluster-admin role is only used where required (Manual)`

Remediation:
```
Ensure that the cluster-admin role is only used where required (Manual)
```

Control **5.1.2**
Description: `Minimize access to secrets (Manual)`

Remediation:
```
Minimize access to secrets (Manual)
```

Control **5.1.3**
Description: `Minimize wildcard use in Roles and ClusterRoles (Manual)`

Remediation:
```
Minimize wildcard use in Roles and ClusterRoles (Manual)
```

Control **5.1.4**
Description: `Minimize access to create pods (Manual)`

Remediation:
```
Minimize access to create pods (Manual)
```

Control **5.1.5**
Description: `Ensure that default service accounts are not actively used. (Manual)`

Remediation:
```
Ensure that default service accounts are not actively used. (Manual)
```

Control **5.1.6**
Description: `Ensure that Service Account Tokens are only mounted where necessary (Manual)`

Remediation:
```
Ensure that Service Account Tokens are only mounted where necessary (Manual)
```

Control **5.1.7**
Description: `Avoid use of system:masters group (Manual)`

Remediation:
```
Avoid use of system:masters group (Manual)
```

Control **5.1.8**
Description: `Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual)`

Remediation:
```
Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual)
```


### Pod Security Standards

Control **5.2.1**
Description: `Ensure that the cluster has at least one active policy control mechanism in place (Manual)`

Remediation:
```
Ensure that the cluster has at least one active policy control mechanism in place (Manual)
```

Control **5.2.2**
Description: `Minimize the admission of privileged containers (Manual)`

Remediation:
```
Minimize the admission of privileged containers (Manual)
```

Control **5.2.3**
Description: `Minimize the admission of containers wishing to share the host process ID namespace (Automated)`

Remediation:
```
Minimize the admission of containers wishing to share the host process ID namespace (Automated)
```

Control **5.2.4**
Description: `Minimize the admission of containers wishing to share the host IPC namespace (Automated)`

Remediation:
```
Minimize the admission of containers wishing to share the host IPC namespace (Automated)
```

Control **5.2.5**
Description: `Minimize the admission of containers wishing to share the host network namespace (Automated)`

Remediation:
```
Minimize the admission of containers wishing to share the host network namespace (Automated)
```

Control **5.2.6**
Description: `Minimize the admission of containers with allowPrivilegeEscalation (Automated)`

Remediation:
```
Minimize the admission of containers with allowPrivilegeEscalation (Automated)
```

Control **5.2.7**
Description: `Minimize the admission of root containers (Automated)`

Remediation:
```
Minimize the admission of root containers (Automated)
```

Control **5.2.8**
Description: `Minimize the admission of containers with the NET_RAW capability (Automated)`

Remediation:
```
Minimize the admission of containers with the NET_RAW capability (Automated)
```

Control **5.2.9**
Description: `Minimize the admission of containers with added capabilities (Automated)`

Remediation:
```
Minimize the admission of containers with added capabilities (Automated)
```

Control **5.2.10**
Description: `Minimize the admission of containers with capabilities assigned (Manual)`

Remediation:
```
Minimize the admission of containers with capabilities assigned (Manual)
```

Control **5.2.11**
Description: `Minimize the admission of Windows HostProcess containers (Manual)`

Remediation:
```
Minimize the admission of Windows HostProcess containers (Manual)
```

Control **5.2.12**
Description: `Minimize the admission of HostPath volumes (Manual)`

Remediation:
```
Minimize the admission of HostPath volumes (Manual)
```

Control **5.2.13**
Description: `Minimize the admission of containers which use HostPorts (Manual)`

Remediation:
```
Minimize the admission of containers which use HostPorts (Manual)
```


### Network Policies and CNI

Control **5.3.1**
Description: `Ensure that the CNI in use supports NetworkPolicies (Manual)`

Remediation:
```
Ensure that the CNI in use supports NetworkPolicies (Manual)
```

Control **5.3.2**
Description: `Ensure that all Namespaces have NetworkPolicies defined (Manual)`

Remediation:
```
Ensure that all Namespaces have NetworkPolicies defined (Manual)
```


### Secrets Management

Control **5.4.1**
Description: `Prefer using Secrets as files over Secrets as environment variables (Manual)`

Remediation:
```
Prefer using Secrets as files over Secrets as environment variables (Manual)
```

Control **5.4.2**
Description: `Consider external secret storage (Manual)`

Remediation:
```
Consider external secret storage (Manual)
```


### Extensible Admission Control

Control **5.5.1**
Description: `Configure Image Provenance using ImagePolicyWebhook admission controller (Manual)`

Remediation:
```
Configure Image Provenance using ImagePolicyWebhook admission controller (Manual)
```


### General Policies

Control **5.7.1**
Description: `Create administrative boundaries between resources using namespaces (Manual)`

Remediation:
```
Create administrative boundaries between resources using namespaces (Manual)
```

Control **5.7.2**
Description: `Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual)`

Remediation:
```
Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual)
```

Control **5.7.3**
Description: `Apply SecurityContext to your Pods and Containers (Manual)`

Remediation:
```
Apply SecurityContext to your Pods and Containers (Manual)
```

Control **5.7.4**
Description: `The default namespace should not be used (Manual)`

Remediation:
```
The default namespace should not be used (Manual)
```

