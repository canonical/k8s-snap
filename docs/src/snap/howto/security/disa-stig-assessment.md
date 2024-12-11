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
- You have reviewed the [post-deployment hardening] guide and have applied the
  hardening steps that relevant to your use-case 


## Results overview

| Status | Findings |
| ------ | ----- |
| `Automated` (70) | V-242379, V-242380, V-242381, V-242382, V-242387, V-242388, V-242389, V-242391, V-242392, V-242397, V-242400, V-242405, V-242406, V-242407, V-242408, V-242409, V-242418, V-242419, V-242420, V-242421, V-242422, V-242423, V-242426, V-242427, V-242428, V-242429, V-242430, V-242431, V-242432, V-242433, V-242434, V-242436, V-242444, V-242445, V-242446, V-242447, V-242448, V-242449, V-242450, V-242451, V-242452, V-242453, V-242456, V-242457, V-242459, V-242460, V-242466, V-242467, V-245542, V-245543, V-245544, V-254801, V-242376, V-242377, V-242378, V-242384, V-242385, V-242390, V-242402, V-242403, V-242404, V-242424, V-242425, V-242438, V-242461, V-242462, V-242463, V-242464, V-242465, V-245541 |
| `Not Applicable` (13) | V-242386, V-242393, V-242394, V-242395, V-242396, V-242398, V-242399, V-242413, V-242437, V-242442, V-242443, V-242454, V-242455 |
| `Manual` (8) | V-242383, V-242410, V-242411, V-242412, V-242414, V-242415, V-242417, V-254800 |

**Automated**: An automated process can be used to validate that the system is in a conformant state.

**Not Applicable**: These Findings are not applicable to {{product}}. Some reasons for this include: a check on a Kubernetes feature that was removed prior to {{product}}'s first release, a check for a component that {{product}} does not package, etc.

**Manual**: These checks require manual intervention from a cluster administrator, so they cannot be automated.


### [V-242381](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242381): The Kubernetes Controller Manager must create unique service accounts for each work payload.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes Controller Manager is a background process that embeds core control loops regulating cluster system state through the API Server. Every process executed in a pod has an associated service account. By default, service accounts use the same credentials for authentication. Implementing the default settings poses a High risk to the Kubernetes Controller Manager. Setting the "--use-service-account-credential" value lowers the attack surface by generating unique service accounts settings for each controller instance.





#### Comments:
> The command line arguments of the Kubernetes Controller Manager
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-controller-manager
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-controller-manager` in order to set the argument `--use-service-account-credentials` for service `kube-controller-manager` as appropriate.

Ensure it is set to one of: `true`, `1`

Afterwards restart the `kube-controller-manager` service with:



    sudo systemctl restart snap.k8s.kube-controller-manager



#### Auditing (as root)

Ensure that the argument `--use-service-account-credentials` for service `kube-controller-manager` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-controller-manager`.

```bash
grep -E -q  '\-\-use-service-account-credentials=(true|1)' '/var/snap/k8s/common/args/kube-controller-manager'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242383](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242383): User-managed resources must be created in dedicated namespaces.

#### Severity: High

#### Status: Manual

#### Upstream Finding Description:

> Creating namespaces for user-managed resources is important when implementing Role-Based Access Controls (RBAC). RBAC allows for the authorization of users and helps support proper API server permissions separation and network micro segmentation. If user-managed resources are placed within the default namespaces, it becomes impossible to implement policies for RBAC permission, service account usage, network policies, and more.





#### Comments:
> The Kubernetes System Administrators must manually inspect the services
> in all of the default namespaces to ensure there are no
> user-created resources within them:
> 
>     kubectl -n default get all | grep -v "^(service|NAME)"
>     kubectl -n kube-public get all | grep -v "^(service|NAME)"
>     kubectl -n kube-node-lease get all | grep -v "^(service|NAME)"
> 



### [V-242386](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242386): The Kubernetes API server must have the insecure port flag disabled.

#### Severity: High

#### Status: Not Applicable

#### Upstream Finding Description:

> By default, the API server will listen on two ports. One port is the secure port and the other port is called the "localhost port". This port is also called the "insecure port", port 8080. Any requests to this port bypass authentication and authorization checks. If this port is left open, anyone who gains access to the host on which the Control Plane is running can bypass all authorization and authentication mechanisms put in place, and have full control over the entire cluster.
> 
> Close the insecure port by setting the API server's "--insecure-port" flag to "0", ensuring that the "--insecure-bind-address" is not set.





#### Comments:
> This Finding refers to the `--insecure-port` command line argument
> for the Kubernetes API Server service.
> 
> Support for the `--insecure-port` flag has been deprecated in
> Kubernetes 1.10, and completely removed in 1.21, so this Finding
> is Not Applicable to any versions of the k8s-snap.
> 
> https://github.com/kubernetes/kubernetes/issues/91506
> 



### [V-242387](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242387): The Kubernetes Kubelet must have the "readOnlyPort" flag disabled.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> Kubelet serves a small REST API with read access to port 10255. The read-only port for Kubernetes provides no authentication or authorization security control. Providing unrestricted access on port 10255 exposes Kubernetes pods and containers to malicious attacks or compromise. Port 10255 is deprecated and should be disabled.





#### Comments:
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service, but does explicitly pass
> `--read-only-port=0` as a command line argument.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--read-only-port` for service `kubelet` as appropriate.

Ensure it is set to: `0`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--read-only-port` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -q  '\-\-read-only-port=(0)' '/var/snap/k8s/common/args/kubelet' || echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242388](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242388): The Kubernetes API server must have the insecure bind address not set.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> By default, the API server will listen on two ports and addresses. One address is the secure address and the other address is called the "insecure bind" address and is set by default to localhost. Any requests to this address bypass authentication and authorization checks. If this insecure bind address is set to localhost, anyone who gains access to the host on which the Control Plane is running can bypass all authorization and authentication mechanisms put in place and have full control over the entire cluster.
> 
> Close or set the insecure bind address by setting the API server's "--insecure-bind-address" flag to an IP or leave it unset and ensure that the "--insecure-bind-port" is not set.





#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--insecure-bind-address` for service `kube-apiserver` as appropriate.

Ensure it is NOT set to any value.

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--insecure-bind-address` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-insecure-bind-address=(.*)' '/var/snap/k8s/common/args/kube-apiserver' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-242390](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242390): The Kubernetes API server must have anonymous authentication disabled.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes API Server controls Kubernetes via an API interface. A user who has access to the API essentially has root access to the entire Kubernetes cluster. To control access, users must be authenticated and authorized. By allowing anonymous connections, the controls put in place to secure the API can be bypassed.
> 
> Setting "--anonymous-auth" to "false" also disables unauthenticated requests from kubelets.
> 
> While there are instances where anonymous connections may be needed (e.g., health checks) and Role-Based Access Controls (RBACs) are in place to limit the anonymous access, this access should be disabled, and only enabled when necessary.





#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--anonymous-auth` for service `kube-apiserver` as appropriate.

Ensure it is set to one of: `false`, `0`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--anonymous-auth` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-anonymous-auth=(false|0)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242391](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242391): The Kubernetes Kubelet must have anonymous authentication disabled.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> A user who has access to the Kubelet essentially has root access to the nodes contained within the Kubernetes Control Plane. To control access, users must be authenticated and authorized. By allowing anonymous connections, the controls put in place to secure the Kubelet can be bypassed.
> 
> Setting anonymous authentication to "false" also disables unauthenticated requests from kubelets.
> 
> While there are instances where anonymous connections may be needed (e.g., health checks) and Role-Based Access Controls (RBAC) are in place to limit the anonymous access, this access must be disabled and only enabled when necessary.





#### Comments:
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service, but does explicitly pass
> `--anonymous-auth=0` as a command line argument.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--anonymous-auth` for service `kubelet` as appropriate.

Ensure it is set to one of: `false`, `0`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--anonymous-auth` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

```bash
grep -E -q  '\-\-anonymous-auth=(false|0)' '/var/snap/k8s/common/args/kubelet'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242392](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242392): The Kubernetes kubelet must enable explicit authorization.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> Kubelet is the primary agent on each node. The API server communicates with each kubelet to perform tasks such as starting/stopping pods. By default, kubelets allow all authenticated requests, even anonymous ones, without requiring any authorization checks from the API server. This default behavior bypasses any authorization controls put in place to limit what users may perform within the Kubernetes cluster. To change this behavior, the default setting of AlwaysAllow for the authorization mode must be set to "Webhook".





#### Comments:
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service, but does explicitly pass
> `--authorization-mode=Webhook` as a command line argument.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--authorization-mode` for service `kubelet` as appropriate.

Ensure it is set to: `Webhook`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--authorization-mode` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -q  '\-\-authorization-mode=(Webhook)' '/var/snap/k8s/common/args/kubelet' || echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-242397](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242397): The Kubernetes kubelet staticPodPath must not enable static pods.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> Allowing kubelet to set a staticPodPath gives containers with root access permissions to traverse the hosting filesystem. The danger comes when the container can create a manifest file within the /etc/kubernetes/manifests directory. When a manifest is created within this directory, containers are entirely governed by the Kubelet not the API Server. The container is not susceptible to admission control at all. Any containers or pods that are instantiated in this manner are called "static pods" and are meant to be used for pods such as the API server, scheduler, controller, etc., not workload pods that need to be governed by the API Server.





#### Comments:
> The Finding refers to checking the 'staticPodPath' in kubectl's `--config`
> file is not set.
> 
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service, nor does it pass
> `--pod-manifest-path` as a command line argument.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--pod-manifest-path` for service `kubelet` as appropriate.

Ensure it is NOT set to any value.

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--pod-manifest-path` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-pod-manifest-path=(.*)' '/var/snap/k8s/common/args/kubelet' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-242415](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242415): Secrets in Kubernetes must not be stored as environment variables.

#### Severity: High

#### Status: Manual

#### Upstream Finding Description:

> Secrets, such as passwords, keys, tokens, and certificates should not be stored as environment variables. These environment variables are accessible inside Kubernetes by the "Get Pod" API call, and by any system, such as CI/CD pipeline, which has access to the definition file of the container. Secrets must be mounted from files or stored within password vaults.





#### Comments:
> The Kubernetes System Administrator must manually inspect the Environment
> of each user-created Pod to ensure there are no Pods passing information
> which the System Administrator may categorize as 'sensitive'
> (e.g. passwords, cryptographic keys, API tokens, etc).
> 



### [V-242434](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242434): Kubernetes Kubelet must enable kernel protection.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> System kernel is responsible for memory, disk, and task management. The kernel provides a gateway between the system hardware and software. Kubernetes requires kernel access to allocate resources to the Control Plane. Threat actors that penetrate the system kernel can inject malicious code or hijack the Kubernetes architecture. It is vital to implement protections through Kubernetes components to reduce the attack surface.





#### Comments:
> The Finding stipulates that `--protect-kernel-defaults`
> must be set on the Kubelet service.
> 
> This flag is not set by default in the k8s-snap, as it
> may prevent kubelet from starting normally unless the
> kernel settings are as Kubelet expects.
> 
> Please review the hardening guide for information on how to
> properly configure the Node's Operating System for Kubelet.
> 
> <!-- TODO: link to dedicated K8s hardening doc. -->
> https://microk8s.io/docs/how-to-cis-harden#check-426
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--protect-kernel-defaults` for service `kubelet` as appropriate.

Ensure it is set to one of: `true`, `1`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--protect-kernel-defaults` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -q  '\-\-protect-kernel-defaults=(true|1)' '/var/snap/k8s/common/args/kubelet' || echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-242436](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242436): The Kubernetes API server must have the ValidatingAdmissionWebhook enabled.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> Enabling the admissions webhook allows for Kubernetes to apply policies against objects that are to be created, read, updated, or deleted. By applying a pod security policy, control can be given to not allow images to be instantiated that run as the root user. If pods run as the root user, the pod then has root privileges to the host system and all the resources it has. An attacker can use this to attack the Kubernetes cluster. By implementing a policy that does not allow root or privileged pods, the pod users are limited in what the pod can do and access.





#### Comments:
> This Finding stipulates that the `ValidatingAdmissionWebhook`
> Admission Plugin should be enabled.
> 
> The `ValidatingAdmissionWebhook` Admission Plugin is enabled
> by default in all modern versions of the k8s-snap.
> 
> The automated check associated with this Finding is thus meant
> to verify that `ValidatingAdmissionWebhook` is NOT disabled
> through the `--disable-admission-plugins` argument.
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--disable-admission-plugins` for service `kube-apiserver` as appropriate.

Ensure it is NOT set to one of: `.*ValidatingAdmissionWebhook.*`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--disable-admission-plugins` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-disable-admission-plugins=(.*ValidatingAdmissionWebhook.*)' '/var/snap/k8s/common/args/kube-apiserver' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-242437](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242437): Kubernetes must have a pod security policy set.

#### Severity: High

#### Status: Not Applicable

#### Upstream Finding Description:

> Enabling the admissions webhook allows for Kubernetes to apply policies against objects that are to be created, read, updated, or deleted. By applying a pod security policy, control can be given to not allow images to be instantiated that run as the root user. If pods run as the root user, the pod then has root privileges to the host system and all the resources it has. An attacker can use this to attack the Kubernetes cluster. By implementing a policy that does not allow root or privileged pods, the pod users are limited in what the pod can do and access.





#### Comments:
> This Finding stipulates some checks on the Pod Security Policy object
> which was deprecated in 1.21 and removed in 1.25, so it is Not Applicable
> to any versions of the k8s-snap.
> 
> https://kubernetes.io/docs/concepts/security/pod-security-policy/
> 



### [V-245542](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-245542): Kubernetes API Server must disable basic authentication to protect information in transit.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes basic authentication sends and receives request containing username, uid, groups, and other fields over a clear text HTTP communication. Basic authentication does not provide any security mechanisms using encryption standards. PKI certificate-based authentication must be set over a secure channel to ensure confidentiality and integrity. Basic authentication must not be set in the manifest file.





#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--basic-auth-file` for service `kube-apiserver` as appropriate.

Ensure it is NOT set to any value.

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--basic-auth-file` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-basic-auth-file=(.*)' '/var/snap/k8s/common/args/kube-apiserver' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-245543](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-245543): Kubernetes API Server must disable token authentication to protect information in transit.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes token authentication uses password known as secrets in a plaintext file. This file contains sensitive information such as token, username and user uid. This token is used by service accounts within pods to authenticate with the API Server. This information is very valuable for attackers with malicious intent if the service account is privileged having access to the token. With this token a threat actor can impersonate the service account gaining access to the Rest API service.





#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--token-auth-file` for service `kube-apiserver` as appropriate.

It is possible to leave this argument unset completely.

Ensure it is NOT set to any value.

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--token-auth-file` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-token-auth-file=(.*)' '/var/snap/k8s/common/args/kube-apiserver' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-245544](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-245544): Kubernetes endpoints must use approved organizational certificate and key pair to protect information in transit.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes control plane and external communication is managed by API Server. The main implementation of the API Server is to manage hardware resources for pods and container using horizontal or vertical scaling. Anyone who can gain access to the API Server can effectively control your Kubernetes architecture. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server with a means to be able to authenticate sessions and encrypt traffic.
> 
> By default, the API Server does not authenticate to the kubelet HTTPs endpoint. To enable secure communication for API Server, the parameter -kubelet-client-certificate and kubelet-client-key must be set. This parameter gives the location of the certificate and key pair used to secure API Server communication.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--kubelet-client-certificate` for service `kube-apiserver` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/apiserver-kubelet-client\.crt`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure that the argument `--kubelet-client-certificate` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-kubelet-client-certificate=(/etc/kubernetes/pki/apiserver-kubelet-client\.crt)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--kubelet-client-key` for service `kube-apiserver` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/apiserver-kubelet-client\.key`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--kubelet-client-key` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-kubelet-client-key=(/etc/kubernetes/pki/apiserver-kubelet-client\.key)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-254800](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-254800): Kubernetes must have a Pod Security Admission control file configured.

#### Severity: High

#### Status: Manual

#### Upstream Finding Description:

> An admission controller intercepts and processes requests to the Kubernetes API prior to persistence of the object, but after the request is authenticated and authorized.
> 
> Kubernetes (> v1.23)offers a built-in Pod Security admission controller to enforce the Pod Security Standards. Pod security restrictions are applied at the namespace level when pods are created. 
> 
> The Kubernetes Pod Security Standards define different isolation levels for Pods. These standards define how to restrict the behavior of pods in a clear, consistent fashion.





#### Comments:
> This Finding stipulates the presence of a Pod Security Admission
> Control File which will need to be manually configured by
> the Kubernetes System Administrator on a per-organization
> basis.
> 
> Instructions on how to configure an `--admission-control-config-file` for
> the Kube API Server of the k8s-snap can be found at:
> 
> <!-- TODO: switch link to dedicated DISA Hardening page when published. -->
> https://documentation.ubuntu.com/canonical-kubernetes/latest/src/snap/howto/cis-hardening/#configure-auditing
> 



### [V-254801](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-254801): Kubernetes must enable PodSecurity admission controller on static pods and Kubelets.

#### Severity: High

#### Status: Automated

#### Upstream Finding Description:

> PodSecurity admission controller is a component that validates and enforces security policies for pods running within a Kubernetes cluster. It is responsible for evaluating the security context and configuration of pods against defined policies. 
> 
> To enable PodSecurity admission controller on Static Pods (kube-apiserver, kube-controller-manager, or kube-schedule), the argument "--feature-gates=PodSecurity=true" must be set.
> 
> To enable PodSecurity admission controller on Kubelets, the featureGates PodSecurity=true argument must be set.
> 
> (Note: The PodSecurity feature gate is GA as of  v1.25.)





#### Comments:
> This Finding refers to setting the `--feature-gates=PodSecurity=true`
> feature gate for the Kubernetes API Server.
> 
> The `PodSecurity` feature gate has been GA and enabled by default
> since 1.25.
> 
> https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates-removed/
> 
> The automated check associated with this Finding is thus meant
> to verify that `PodSecurity` is NOT disabled.
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--feature-gates` for service `kubelet` as appropriate.

Ensure it is NOT set to one of: `.*PodSecurity=false.*`, `.*PodSecurity=0.*`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--feature-gates` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-feature-gates=(.*PodSecurity=false.*|.*PodSecurity=0.*)' '/var/snap/k8s/common/args/kubelet' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-242376](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242376): The Kubernetes Controller Manager must use TLS 1.2, at a minimum, to protect the confidentiality of sensitive data during electronic dissemination.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes Controller Manager will prohibit the use of SSL and unauthorized versions of TLS protocols to properly secure communication.
> 
> The use of unsupported protocol exposes vulnerabilities to the Kubernetes by rogue traffic interceptions, man-in-the-middle attacks, and impersonation of users or services from the container platform runtime, registry, and key store. To enable the minimum version of TLS to be used by the Kubernetes Controller Manager, the setting "tls-min-version" must be set.





#### Comments:
> The command line arguments of the Kubernetes Controller Manager
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-controller-manager
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-controller-manager` in order to set the argument `--tls-min-version` for service `kube-controller-manager` as appropriate.

Ensure it is set to one of: `VersionTLS12`, `VersionTLS13`

Afterwards restart the `kube-controller-manager` service with:



    sudo systemctl restart snap.k8s.kube-controller-manager



#### Auditing (as root)

Ensure that the argument `--tls-min-version` for service `kube-controller-manager` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-controller-manager`.

```bash
grep -E -q  '\-\-tls-min-version=(VersionTLS12|VersionTLS13)' '/var/snap/k8s/common/args/kube-controller-manager'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242377](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242377): The Kubernetes Scheduler must use TLS 1.2, at a minimum, to protect the confidentiality of sensitive data during electronic dissemination.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes Scheduler will prohibit the use of SSL and unauthorized versions of TLS protocols to properly secure communication.
> 
> The use of unsupported protocol exposes vulnerabilities to the Kubernetes by rogue traffic interceptions, man-in-the-middle attacks, and impersonation of users or services from the container platform runtime, registry, and keystore. To enable the minimum version of TLS to be used by the Kubernetes API Server, the setting "tls-min-version" must be set.





#### Comments:
> The command line arguments of the Kubernetes Scheduler
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-scheduler
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-scheduler` in order to set the argument `--tls-min-version` for service `kube-scheduler` as appropriate.

Ensure it is set to one of: `VersionTLS12`, `VersionTLS13`

Afterwards restart the `kube-scheduler` service with:



    sudo systemctl restart snap.k8s.kube-scheduler



#### Auditing (as root)

Ensure that the argument `--tls-min-version` for service `kube-scheduler` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-scheduler`.

```bash
grep -E -q  '\-\-tls-min-version=(VersionTLS12|VersionTLS13)' '/var/snap/k8s/common/args/kube-scheduler'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242378](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242378): The Kubernetes API Server must use TLS 1.2, at a minimum, to protect the confidentiality of sensitive data during electronic dissemination.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes API Server will prohibit the use of SSL and unauthorized versions of TLS protocols to properly secure communication.
> 
> The use of unsupported protocol exposes vulnerabilities to the Kubernetes by rogue traffic interceptions, man-in-the-middle attacks, and impersonation of users or services from the container platform runtime, registry, and keystore. To enable the minimum version of TLS to be used by the Kubernetes API Server, the setting "tls-min-version" must be set.





#### Comments:
> The command line arguments of the Kubernetes Scheduler
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-scheduler
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--tls-min-version` for service `kube-apiserver` as appropriate.

Ensure it is set to one of: `VersionTLS12`, `VersionTLS13`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--tls-min-version` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-tls-min-version=(VersionTLS12|VersionTLS13)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242379](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242379): The Kubernetes etcd must use TLS to protect the confidentiality of sensitive data during electronic dissemination.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes etcd will prohibit the use of SSL and unauthorized versions of TLS protocols to properly secure communication.
> 
> The use of unsupported protocol exposes vulnerabilities to the Kubernetes by rogue traffic interceptions, man-in-the-middle attacks, and impersonation of users or services from the container platform runtime, registry, and keystore. To enable the minimum version of TLS to be used by the Kubernetes API Server, the setting "--auto-tls" must be set.




<details>
<summary><h4>Step 1/3</h4></summary>
<br>


#### Comments:
> This finding refers to the `--auto-tls` command line argument for the
> etcd service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The k8s-snap configures the Kube API Server to connect to
> k8s-dqlite via local socket owned by root.
> 
> The Auditing section will describe how to check the ownership
> of the k8s-dqlite socket.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/3</h4></summary>
<br>


#### Comments:
> This check ensures the permissions on the k8s-dqlite socket.


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 600 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure all required files have permissions '600' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 3/3</h4></summary>
<br>


#### Comments:
> This check ensures the `--etcd-servers` argument of the Kube API Server
> is as expected.
> 


<details>
<summary><h4>Remediation for Step 3</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--etcd-servers` for service `kube-apiserver` as appropriate.

Ensure it is set to: `unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 3</h4></summary>
<br>

Ensure that the argument `--etcd-servers` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-etcd-servers=(unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242380](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242380): The Kubernetes etcd must use TLS to protect the confidentiality of sensitive data during electronic dissemination.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes API Server will prohibit the use of SSL and unauthorized versions of TLS protocols to properly secure communication.
> 
> The use of unsupported protocol exposes vulnerabilities to the Kubernetes by rogue traffic interceptions, man-in-the-middle attacks, and impersonation of users or services from the container platform runtime, registry, and keystore. To enable the minimum version of TLS to be used by the Kubernetes API Server, the setting "--peer-auto-tls" must be set.





#### Comments:
> This finding refers to the `--peer-auto-tls` command line argument for
> the etcd service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> k8s-dqlite communication defaults to using TLS unless the
> `--enable-tls` argument is set in k8s-dqlite argument configuration
> file located at:
> 
>     /var/snap/k8s/common/args/k8s-dqlite
> 


#### Remediation
Edit `/var/snap/k8s/common/args/k8s-dqlite` in order to set the argument `--enable-tls` for service `k8s-dqlite` as appropriate.

Ensure it is NOT set to one of: `false`, `0`

Afterwards restart the `k8s-dqlite` service with:



    sudo systemctl restart snap.k8s.k8s-dqlite



#### Auditing (as root)

Ensure that the argument `--enable-tls` for service `k8s-dqlite` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/k8s-dqlite`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-enable-tls=(false|0)' '/var/snap/k8s/common/args/k8s-dqlite' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-242382](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242382): The Kubernetes API Server must enable Node,RBAC as the authorization mode.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> To mitigate the risk of unauthorized access to sensitive information by entities that have been issued certificates by DOD-approved PKIs, all DOD systems (e.g., networks, web servers, and web portals) must be properly configured to incorporate access control methods that do not rely solely on the possession of a certificate for access. Successful authentication must not automatically give an entity access to an asset or security boundary. Authorization procedures and controls must be implemented to ensure each authenticated entity also has a validated and current authorization. Authorization is the process of determining whether an entity, once authenticated, is permitted to access a specific asset.
> 
> Node,RBAC is the method within Kubernetes to control access of users and applications. Kubernetes uses roles to grant authorization API requests made by kubelets.
> 
> Satisfies: SRG-APP-000340-CTR-000770, SRG-APP-000033-CTR-000095, SRG-APP-000378-CTR-000880, SRG-APP-000033-CTR-000090





#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 
> Note that the ordering of the values is mandatory.
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--authorization-mode` for service `kube-apiserver` as appropriate.

Ensure it is set to: `Node,RBAC`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--authorization-mode` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-authorization-mode=(Node,RBAC)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242384](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242384): The Kubernetes Scheduler must have secure binding.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Limiting the number of attack vectors and implementing authentication and encryption on the endpoints available to external sources is paramount when securing the overall Kubernetes cluster. The Scheduler API service exposes port 10251/TCP by default for health and metrics information use. This port does not encrypt or authenticate connections. If this port is exposed externally, an attacker can use this port to attack the entire Kubernetes cluster. By setting the bind address to localhost (i.e., 127.0.0.1), only those internal services that require health and metrics information can access the Scheduler API.





#### Comments:
> The command line arguments of the Kubernetes Scheduler
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-scheduler
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-scheduler` in order to set the argument `--bind-address` for service `kube-scheduler` as appropriate.

Ensure it is set to: `127.0.0.1`

Afterwards restart the `kube-scheduler` service with:



    sudo systemctl restart snap.k8s.kube-scheduler



#### Auditing (as root)

Ensure that the argument `--bind-address` for service `kube-scheduler` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-scheduler`.

```bash
grep -E -q  '\-\-bind-address=(127.0.0.1)' '/var/snap/k8s/common/args/kube-scheduler'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242385](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242385): The Kubernetes Controller Manager must have secure binding.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Limiting the number of attack vectors and implementing authentication and encryption on the endpoints available to external sources is paramount when securing the overall Kubernetes cluster. The Controller Manager API service exposes port 10252/TCP by default for health and metrics information use. This port does not encrypt or authenticate connections. If this port is exposed externally, an attacker can use this port to attack the entire Kubernetes cluster. By setting the bind address to only localhost (i.e., 127.0.0.1), only those internal services that require health and metrics information can access the Control Manager API.





#### Comments:
> The command line arguments of the Kubernetes Controller Manager
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-controller-manager
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-controller-manager` in order to set the argument `--bind-address` for service `kube-controller-manager` as appropriate.

Ensure it is set to: `127.0.0.1`

Afterwards restart the `kube-controller-manager` service with:



    sudo systemctl restart snap.k8s.kube-controller-manager



#### Auditing (as root)

Ensure that the argument `--bind-address` for service `kube-controller-manager` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-controller-manager`.

```bash
grep -E -q  '\-\-bind-address=(127.0.0.1)' '/var/snap/k8s/common/args/kube-controller-manager'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242389](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242389): The Kubernetes API server must have the secure port set.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> By default, the API server will listen on what is rightfully called the secure port, port 6443. Any requests to this port will perform authentication and authorization checks. If this port is disabled, anyone who gains access to the host on which the Control Plane is running has full control of the entire cluster over encrypted traffic.
> 
> Open the secure port by setting the API server's "--secure-port" flag to a value other than "0".





#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--secure-port` for service `kube-apiserver` as appropriate.

Ensure it is NOT set to one of: `0`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--secure-port` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-secure-port=(0)' '/var/snap/k8s/common/args/kube-apiserver' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242393](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242393): Kubernetes Worker Nodes must not have sshd service running.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> Worker Nodes are maintained and monitored by the Control Plane. Direct access and manipulation of the nodes should not take place by administrators. Worker nodes should be treated as immutable and updated via replacement rather than in-place upgrades.





#### Comments:
> This Finding aims to completely prohibit the *running*
> of SSHD on all worker Nodes, and must be assessed by the
> Kubernetes System Administrator as applicable.
> 
> It also mentions that:
> "If the worker nodes cannot be reached, this requirement is "not a finding"."
> 



### [V-242394](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242394): Kubernetes Worker Nodes must not have the sshd service enabled.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> Worker Nodes are maintained and monitored by the Control Plane. Direct access and manipulation of the nodes must not take place by administrators. Worker nodes must be treated as immutable and updated via replacement rather than in-place upgrades.





#### Comments:
> This Finding aims to prohibit the *enabling of the service*
> for SSHD on all worker Nodes, and must be assessed by the
> Kubernetes System Administrator as applicable.
> 
> It also mentions that:
> "If the worker nodes cannot be reached, this requirement is "not a finding"."
> 



### [V-242395](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242395): Kubernetes dashboard must not be enabled.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> While the Kubernetes dashboard is not inherently insecure on its own, it is often coupled with a misconfiguration of Role-Based Access control (RBAC) permissions that can unintentionally over-grant access. It is not commonly protected with "NetworkPolicies", preventing all pods from being able to reach it. In increasingly rare circumstances, the Kubernetes dashboard is exposed publicly to the internet.





#### Comments:
> The k8s-snap does not automatically deploy or configure the Kubernetes Dashboard,
> so this finding is Not Applicable.
> 
> You can check whether the Kubernetes Dashboard has been installed post-snap-setup
> by running:
> 
>     k8s kubectl get pods --all-namespaces -l k8s-app=kubernetes-dashboard
> 



### [V-242396](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242396): Kubernetes Kubectl cp command must give expected access and results.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> One of the tools heavily used to interact with containers in the Kubernetes cluster is kubectl. The command is the tool System Administrators used to create, modify, and delete resources. One of the capabilities of the tool is to copy files to and from running containers (i.e., kubectl cp). The command uses the "tar" command of the container to copy files from the container to the host executing the "kubectl cp" command. If the "tar" command on the container has been replaced by a malicious user, the command can copy files anywhere on the host machine. This flaw has been fixed in later versions of the tool. It is recommended to use kubectl versions newer than 1.12.9.





#### Comments:
> This Finding refers to checking the `kubectl version --client` to avoid
> a known security issue with `kubectl cp`.
> 
> This issue was fixed in 1.12.9, and thus is Not Applicable to any
> versions of the k8s-snap.
> 
> https://discuss.kubernetes.io/t/announce-security-release-of-kubernetes-kubectl-potential-directory-traversal-releases-1-11-9-1-12-7-1-13-5-and-1-14-0-cve-2019-1002101/5712
> 



### [V-242398](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242398): Kubernetes DynamicAuditing must not be enabled.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> Protecting the audit data from change or deletion is important when an attack occurs. One way an attacker can cover their tracks is to change or delete audit records. This will either make the attack unnoticeable or make it more difficult to investigate how the attack took place and what changes were made. The audit data can be protected through audit log file protections and user authorization.
> 
> One way for an attacker to thwart these measures is to send the audit logs to another source and filter the audited results before sending them on to the original target. This can be done in Kubernetes through the configuration of dynamic audit webhooks through the DynamicAuditing flag.





#### Comments:
> This finding relates to the `--feature-gate=DynamicAuditing` feature gate flag.
> 
> This Feature Gate was only available between Kubernetes versions 1.13-1.19,
> and is thus Not Applicable to any version of the k8s-snap.
> 
> https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates-removed/
> 



### [V-242399](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242399): Kubernetes DynamicKubeletConfig must not be enabled.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> Kubernetes allows a user to configure kubelets with dynamic configurations. When dynamic configuration is used, the kubelet will watch for changes to the configuration file. When changes are made, the kubelet will automatically restart. Allowing this capability bypasses access restrictions and authorizations. Using this capability, an attacker can lower the security posture of the kubelet, which includes allowing the ability to run arbitrary commands in any container running on that node.





#### Comments:
> Checks related to the `--feature-gate=DynamicKubeletConfig` feature gate flag.
> 
> This Feature Gate was only available between Kubernetes versions 1.4-1.25,
> and is thus Not Applicable to any version of the k8s-snap.
> 
> https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates-removed/
> 



### [V-242400](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242400): The Kubernetes API server must have Alpha APIs disabled.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes allows alpha API calls within the API server. The alpha features are disabled by default since they are not ready for production and likely to change without notice. These features may also contain security issues that are rectified as the feature matures. To keep the Kubernetes cluster secure and stable, these alpha features must not be used.





#### Comments:
> The k8s-snap does not set the `--feature-gate` flag on the `kube-apiserver`.
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--feature-gate` for service `kube-apiserver` as appropriate.

It is possible to leave this argument unset completely.

If you'd like to explicitly set it, ensure it is set to one of: `.*AllAlpha=false.*`, `.*AllAlpha=0.*`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--feature-gate` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -q  '\-\-feature-gate=(.*AllAlpha=false.*|.*AllAlpha=0.*)' '/var/snap/k8s/common/args/kube-apiserver' || echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-242402](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242402): The Kubernetes API Server must have an audit log path set.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> When Kubernetes is started, components and user services are started for auditing startup events, and events for components and services, it is important that auditing begin on startup. Within Kubernetes, audit data for all components is generated by the API server. To enable auditing to begin, an audit policy must be defined for the events and the information to be stored with each event. It is also necessary to give a secure location where the audit logs are to be stored. If an audit log path is not specified, all audit data is sent to studio.





#### Comments:
> This Finding refers to the `--audit-log-path` argument of the
> Kubernetes API Service.
> 
> The k8s-snap does not configure auditing by default.
> 
> Please review the below hardening document for a full guide on how
> to enable auditing for the kube-apiserver.
> 
> <!-- TODO: switch link to dedicated DISA Hardening page when published. -->
> https://documentation.ubuntu.com/canonical-kubernetes/latest/src/snap/howto/cis-hardening/#configure-auditing
> 
> This Finding is basically a duplicate of V-242465.
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--audit-log-path` for service `kube-apiserver` as appropriate.

Ensure it is set to any explicit value.

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--audit-log-path` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-audit-log-path=(.*)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242403](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242403): Kubernetes API Server must generate audit records that identify what type of event has occurred, identify the source of the event, contain the event results, identify any users, and identify any containers associated with the event.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Within Kubernetes, audit data for all components is generated by the API server. This audit data is important when there are issues, to include security incidents that must be investigated. To make the audit data worthwhile for the investigation of events, it is necessary to have the appropriate and required data logged. To fully understand the event, it is important to identify any users associated with the event. 
> 
> The API server policy file allows for the following levels of auditing:
>       None - Do not log events that match the rule.
>       Metadata - Log request metadata (requesting user, timestamp, resource, verb, etc.) but not request or response body.
>       Request - Log event metadata and request body but not response body.
>       RequestResponse - Log event metadata, request, and response bodies.
> 
> Satisfies: SRGID:SRG-APP-000092-CTR-000165, SRG-APP-000026-CTR-000070, SRG-APP-000027-CTR-000075, SRG-APP-000028-CTR-000080, SRG-APP-000101-CTR-000205, SRG-APP-000100-CTR-000200, SRG-APP-000100-CTR-000195, SRG-APP-000099-CTR-000190, SRG-APP-000098-CTR-000185, SRG-APP-000095-CTR-000170, SRG-APP-000096-CTR-000175, SRG-APP-000097-CTR-000180, SRG-APP-000507-CTR-001295, SRG-APP-000504-CTR-001280, SRG-APP-000503-CTR-001275, SRG-APP-000501-CTR-001265, SRG-APP-000500-CTR-001260, SRG-APP-000497-CTR-001245, SRG-APP-000496-CTR-001240, SRG-APP-000493-CTR-001225, SRG-APP-000492-CTR-001220, SRG-APP-000343-CTR-000780, SRG-APP-000381-CTR-000905





#### Comments:
> This Finding refers to the `--audit-policy-file` argument of the
> Kubernetes API Service.
> 
> The k8s-snap does not configure auditing by default.
> 
> Please review the below hardening document for a full guide on how
> to enable auditing for the kube-apiserver.
> 
> <!-- TODO: switch link to dedicated DISA Hardening page when published. -->
> https://documentation.ubuntu.com/canonical-kubernetes/latest/src/snap/howto/cis-hardening/#configure-auditing
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--audit-policy-file` for service `kube-apiserver` as appropriate.

Ensure it is set to any explicit value.

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--audit-policy-file` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-audit-policy-file=(.*)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242404](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242404): Kubernetes Kubelet must deny hostname override.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes allows for the overriding of hostnames. Allowing this feature to be implemented within the kubelets may break the TLS setup between the kubelet service and the API server. This setting also can make it difficult to associate logs with nodes if security analytics needs to take place. The better practice is to setup nodes with resolvable FQDNs and avoid overriding the hostnames.





#### Comments:
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--hostname-override` for service `kubelet` as appropriate.

Ensure it is NOT set to any value.

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--hostname-override` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-hostname-override=(.*)' '/var/snap/k8s/common/args/kubelet' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242405](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242405): The Kubernetes manifests must be owned by root.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The manifest files contain the runtime configuration of the API server, proxy, scheduler, controller, and etcd. If an attacker can gain access to these files, changes can be made to open vulnerabilities and bypass user authorizations inherit within Kubernetes with RBAC implemented.





#### Comments:
> The manifest files for the Kubernetes services in the k8s-snap are
> located in the following directories:
> 
>     /etc/kubernetes
>     /etc/containerd
> 


#### Remediation
Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /etc/containerd /etc/containerd/config.toml

#### Auditing (as root)

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/etc/kubernetes' | grep -q 0:0 && echo PASS /etc/kubernetes: 0:0 || echo FAIL /etc/kubernetes: 0:0
stat -c %u:%g '/etc/kubernetes/pki' | grep -q 0:0 && echo PASS /etc/kubernetes/pki: 0:0 || echo FAIL /etc/kubernetes/pki: 0:0
stat -c %u:%g '/etc/kubernetes/kubelet.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/kubelet.conf: 0:0 || echo FAIL /etc/kubernetes/kubelet.conf: 0:0
stat -c %u:%g '/etc/kubernetes/scheduler.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/scheduler.conf: 0:0 || echo FAIL /etc/kubernetes/scheduler.conf: 0:0
stat -c %u:%g '/etc/kubernetes/proxy.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/proxy.conf: 0:0 || echo FAIL /etc/kubernetes/proxy.conf: 0:0
stat -c %u:%g '/etc/kubernetes/admin.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/admin.conf: 0:0 || echo FAIL /etc/kubernetes/admin.conf: 0:0
stat -c %u:%g '/etc/kubernetes/controller.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/controller.conf: 0:0 || echo FAIL /etc/kubernetes/controller.conf: 0:0
stat -c %u:%g '/etc/kubernetes/pki/etcd' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/etcd: 0:0 || echo FAIL /etc/kubernetes/pki/etcd: 0:0
stat -c %u:%g '/etc/kubernetes/pki/client-ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/client-ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/client-ca.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-ca.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-ca.key: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-ca.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver.key: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver.crt: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver-kubelet-client.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver-kubelet-client.key: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver-kubelet-client.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-client.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-client.crt: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-client.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/serviceaccount.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/serviceaccount.key: 0:0 || echo FAIL /etc/kubernetes/pki/serviceaccount.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-client.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-client.key: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-client.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/kubelet.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/kubelet.crt: 0:0 || echo FAIL /etc/kubernetes/pki/kubelet.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/ca.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/ca.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/ca.key: 0:0 || echo FAIL /etc/kubernetes/pki/ca.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver-kubelet-client.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver-kubelet-client.crt: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver-kubelet-client.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-ca.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/kubelet.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/kubelet.key: 0:0 || echo FAIL /etc/kubernetes/pki/kubelet.key: 0:0
stat -c %u:%g '/etc/containerd' | grep -q 0:0 && echo PASS /etc/containerd: 0:0 || echo FAIL /etc/containerd: 0:0
stat -c %u:%g '/etc/containerd/config.toml' | grep -q 0:0 && echo PASS /etc/containerd/config.toml: 0:0 || echo FAIL /etc/containerd/config.toml: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.


### [V-242406](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242406): The Kubernetes KubeletConfiguration file must be owned by root.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The kubelet configuration file contains the runtime configuration of the kubelet service. If an attacker can gain access to this file, changes can be made to open vulnerabilities and bypass user authorizations inherent within Kubernetes with RBAC implemented.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> This Finding relates to the ownership of Kubelet's `--config` file.
> 
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 
> The Auditing section will advise on how to check the ownership
> of said file.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /var/snap/k8s/common/args/kubelet

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/var/snap/k8s/common/args/kubelet' | grep -q 0:0 && echo PASS /var/snap/k8s/common/args/kubelet: 0:0 || echo FAIL /var/snap/k8s/common/args/kubelet: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check is defined to ensure that Kubelet is not passed
> a `--config` file argument in the k8s-snap.
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--config` for service `kubelet` as appropriate.

Ensure it is NOT set to any value.

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--config` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-config=(.*)' '/var/snap/k8s/common/args/kubelet' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242407](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242407): The Kubernetes KubeletConfiguration files must have file permissions set to 644 or more restrictive.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The kubelet configuration file contains the runtime configuration of the kubelet service. If an attacker can gain access to this file, changes can be made to open vulnerabilities and bypass user authorizations inherit within Kubernetes with RBAC implemented.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> This Finding relates to the permissions on Kubelet's `--config` file.
> 
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 
> The Auditing section will advise on how to check the permissions
> of said file.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 644 /var/snap/k8s/common/args/kubelet

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/args/kubelet' | grep -q 600 && echo PASS /var/snap/k8s/common/args/kubelet: 600 || echo FAIL /var/snap/k8s/common/args/kubelet: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check is defined to ensure that Kubelet is not passed
> a `--config` file argument in the k8s-snap.
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--config` for service `kubelet` as appropriate.

Ensure it is NOT set to any value.

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--config` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-config=(.*)' '/var/snap/k8s/common/args/kubelet' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242408](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242408): The Kubernetes manifest files must have least privileges.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The manifest files contain the runtime configuration of the API server, scheduler, controller, and etcd. If an attacker can gain access to these files, changes can be made to open vulnerabilities and bypass user authorizations inherent within Kubernetes with RBAC implemented.
> 
> Satisfies: SRG-APP-000133-CTR-000310, SRG-APP-000133-CTR-000295, SRG-APP-000516-CTR-001335





#### Comments:
> The Finding requires checking the permissions of the files
> within the `/etc/kubernetes/manifests` directory, but the k8s-snap
> does not use it.
> 
> The usual manifest files for the k8s-snap are located under:
> 
>     /var/snap/k8s/common/args
> 


#### Remediation
Ensure all of the following paths have correct permissions by running: 



    chmod -R 644 /var/snap/k8s/common/args /var/snap/k8s/common/args/conf.d /var/snap/k8s/common/args/kube-apiserver /var/snap/k8s/common/args/kube-controller-manager /var/snap/k8s/common/args/k8sd /var/snap/k8s/common/args/kube-proxy /var/snap/k8s/common/args/kube-scheduler /var/snap/k8s/common/args/kubelet /var/snap/k8s/common/args/containerd /var/snap/k8s/common/args/k8s-dqlite /var/snap/k8s/common/args/conf.d/auth-token-webhook.conf

#### Auditing (as root)

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/args' | grep -q 700 && echo PASS /var/snap/k8s/common/args: 700 || echo FAIL /var/snap/k8s/common/args: 700
stat -c %a '/var/snap/k8s/common/args/conf.d' | grep -q 700 && echo PASS /var/snap/k8s/common/args/conf.d: 700 || echo FAIL /var/snap/k8s/common/args/conf.d: 700
stat -c %a '/var/snap/k8s/common/args/kube-apiserver' | grep -q 600 && echo PASS /var/snap/k8s/common/args/kube-apiserver: 600 || echo FAIL /var/snap/k8s/common/args/kube-apiserver: 600
stat -c %a '/var/snap/k8s/common/args/kube-controller-manager' | grep -q 600 && echo PASS /var/snap/k8s/common/args/kube-controller-manager: 600 || echo FAIL /var/snap/k8s/common/args/kube-controller-manager: 600
stat -c %a '/var/snap/k8s/common/args/k8sd' | grep -q 644 && echo PASS /var/snap/k8s/common/args/k8sd: 644 || echo FAIL /var/snap/k8s/common/args/k8sd: 644
stat -c %a '/var/snap/k8s/common/args/kube-proxy' | grep -q 600 && echo PASS /var/snap/k8s/common/args/kube-proxy: 600 || echo FAIL /var/snap/k8s/common/args/kube-proxy: 600
stat -c %a '/var/snap/k8s/common/args/kube-scheduler' | grep -q 600 && echo PASS /var/snap/k8s/common/args/kube-scheduler: 600 || echo FAIL /var/snap/k8s/common/args/kube-scheduler: 600
stat -c %a '/var/snap/k8s/common/args/kubelet' | grep -q 600 && echo PASS /var/snap/k8s/common/args/kubelet: 600 || echo FAIL /var/snap/k8s/common/args/kubelet: 600
stat -c %a '/var/snap/k8s/common/args/containerd' | grep -q 600 && echo PASS /var/snap/k8s/common/args/containerd: 600 || echo FAIL /var/snap/k8s/common/args/containerd: 600
stat -c %a '/var/snap/k8s/common/args/k8s-dqlite' | grep -q 600 && echo PASS /var/snap/k8s/common/args/k8s-dqlite: 600 || echo FAIL /var/snap/k8s/common/args/k8s-dqlite: 600
stat -c %a '/var/snap/k8s/common/args/conf.d/auth-token-webhook.conf' | grep -q 600 && echo PASS /var/snap/k8s/common/args/conf.d/auth-token-webhook.conf: 600 || echo FAIL /var/snap/k8s/common/args/conf.d/auth-token-webhook.conf: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

### [V-242409](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242409): Kubernetes Controller Manager must disable profiling.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes profiling provides the ability to analyze and troubleshoot Controller Manager events over a web interface on a host port. Enabling this service can expose details about the Kubernetes architecture. This service must not be enabled unless deemed necessary.





#### Comments:
> The command line arguments of the Kubernetes Controller Manager
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-controller-manager
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-controller-manager` in order to set the argument `--profiling` for service `kube-controller-manager` as appropriate.

Ensure it is set to one of: `false`, `0`

Afterwards restart the `kube-controller-manager` service with:



    sudo systemctl restart snap.k8s.kube-controller-manager



#### Auditing (as root)

Ensure that the argument `--profiling` for service `kube-controller-manager` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-controller-manager`.

```bash
grep -E -q  '\-\-profiling=(false|0)' '/var/snap/k8s/common/args/kube-controller-manager'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242410](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242410): The Kubernetes API Server must enforce ports, protocols, and services (PPS) that adhere to the Ports, Protocols, and Services Management Category Assurance List (PPSM CAL).

#### Severity: Medium

#### Status: Manual

#### Upstream Finding Description:

> Kubernetes API Server PPSs must be controlled and conform to the PPSM CAL. Those PPS that fall outside the PPSM CAL must be blocked. Instructions on the PPSM can be found in DoD Instruction 8551.01 Policy.





#### Comments:
> This STIG Finding relates to implementing PPSM CAL for kube-apiserver,
> and must be assessed manually by the Auditor.
> 
> https://www.esd.whs.mil/portals/54/documents/dd/issuances/dodi/855101p.pdf
> 



### [V-242411](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242411): The Kubernetes Scheduler must enforce ports, protocols, and services (PPS) that adhere to the Ports, Protocols, and Services Management Category Assurance List (PPSM CAL).

#### Severity: Medium

#### Status: Manual

#### Upstream Finding Description:

> Kubernetes Scheduler PPS must be controlled and conform to the PPSM CAL. Those ports, protocols, and services that fall outside the PPSM CAL must be blocked. Instructions on the PPSM can be found in DoD Instruction 8551.01 Policy.





#### Comments:
> This STIG Finding relates to implementing PPSM CAL for kube-scheduler,
> and must be assessed manually by the Auditor.
> 
> https://www.esd.whs.mil/portals/54/documents/dd/issuances/dodi/855101p.pdf
> 



### [V-242412](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242412): The Kubernetes Controllers must enforce ports, protocols, and services (PPS) that adhere to the Ports, Protocols, and Services Management Category Assurance List (PPSM CAL).

#### Severity: Medium

#### Status: Manual

#### Upstream Finding Description:

> Kubernetes Controller ports, protocols, and services must be controlled and conform to the PPSM CAL. Those PPS that fall outside the PPSM CAL must be blocked. Instructions on the PPSM can be found in DoD Instruction 8551.01 Policy.





#### Comments:
> This STIG Finding relates to implementing PPSM CAL for kube-controller-manager,
> and must be assessed manually by the Auditor.
> 
> https://www.esd.whs.mil/portals/54/documents/dd/issuances/dodi/855101p.pdf
> 



### [V-242413](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242413): The Kubernetes etcd must enforce ports, protocols, and services (PPS) that adhere to the Ports, Protocols, and Services Management Category Assurance List (PPSM CAL).

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> Kubernetes etcd PPS must be controlled and conform to the PPSM CAL. Those PPS that fall outside the PPSM CAL must be blocked. Instructions on the PPSM can be found in DoD Instruction 8551.01 Policy.





#### Comments:
> This STIG Finding relates to implementing PPSM CAL for etcd.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling, so this Finding is Not Applicable.
> 
> https://www.esd.whs.mil/portals/54/documents/dd/issuances/dodi/855101p.pdf
> 



### [V-242414](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242414): The Kubernetes cluster must use non-privileged host ports for user pods.

#### Severity: Medium

#### Status: Manual

#### Upstream Finding Description:

> Privileged ports are those ports below 1024 and that require system privileges for their use. If containers can use these ports, the container must be run as a privileged user. Kubernetes must stop containers that try to map to these ports directly. Allowing non-privileged ports to be mapped to the container-privileged port is the allowable method when a certain port is needed. An example is mapping port 8080 externally to port 80 in the container.





#### Comments:
> The Kubernetes System Administrators must manually inspect the Pods
> in all of the default namespaces to ensure there are no user-created
> Pods with Containers exposing priviledged port numbers (< 1024).
> 
>     kubectl get pods --all-namespaces
>     kubectl -n NAMESPACE get pod PODNAME -o yaml | grep -i port
> 



### [V-242417](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242417): Kubernetes must separate user functionality.

#### Severity: Medium

#### Status: Manual

#### Upstream Finding Description:

> Separating user functionality from management functionality is a requirement for all the components within the Kubernetes Control Plane. Without the separation, users may have access to management functions that can degrade the Kubernetes architecture and the services being offered, and can offer a method to bypass testing and validation of functions before introduced into a production environment.





#### Comments:
> The Kubernetes System Administrators must manually inspect the Pods
> in all of the default namespaces to ensure there are no
> user-created Pods within them, and move them to dedicated
> user namespaces if present.
> 
>     kubectl -n kube-system get pods
>     kubectl -n kube-public get pods
>     kubectl -n kube-node-lease get pods
> 



### [V-242418](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242418): The Kubernetes API server must use approved cipher suites.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes API server communicates to the kubelet service on the nodes to deploy, update, and delete resources. If an attacker were able to get between this communication and modify the request, the Kubernetes cluster could be compromised. Using approved cypher suites for the communication ensures the protection of the transmitted information, confidentiality, and integrity so that the attacker cannot read or alter this communication.





#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--tls-cipher-suites` for service `kube-apiserver` as appropriate.

Ensure it is set to one of: `.*TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256.*`, `.*TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256.*`, `.*TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384.*`, `.*TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384.*`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--tls-cipher-suites` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-tls-cipher-suites=(.*TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256.*|.*TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256.*|.*TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384.*|.*TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384.*)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242419](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242419): Kubernetes API Server must have the SSL Certificate Authority set.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes control plane and external communication are managed by API Server. The main implementation of the API Server is to manage hardware resources for pods and containers using horizontal or vertical scaling. Anyone who can access the API Server can effectively control the Kubernetes architecture. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols such as TLS. TLS provides the Kubernetes API Server with a means to authenticate sessions and encrypt traffic. 
> 
> To enable encrypted communication for API Server, the parameter client-ca-file must be set. This parameter gives the location of the SSL Certificate Authority file used to secure API Server communication.





#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--client-ca-file` for service `kube-apiserver` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/client-ca\.crt`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--client-ca-file` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-client-ca-file=(/etc/kubernetes/pki/client-ca\.crt)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242420](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242420): Kubernetes Kubelet must have the SSL Certificate Authority set.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes container and pod configuration are maintained by Kubelet. Kubelet agents register nodes with the API Server, mount volume storage, and perform health checks for containers and pods. Anyone who gains access to Kubelet agents can effectively control applications within the pods and containers. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols such as TLS. TLS provides the Kubernetes API Server with a means to authenticate sessions and encrypt traffic.
> 
> To enable encrypted communication for Kubelet, the clientCAFile must be set. This parameter gives the location of the SSL Certificate Authority file used to secure Kubelet communication.





#### Comments:
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service, but does explicitly pass the
> `--client-ca-file` argument as a command line argument.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--client-ca-file` for service `kubelet` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/client-ca\.crt`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--client-ca-file` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

```bash
grep -E -q  '\-\-client-ca-file=(/etc/kubernetes/pki/client-ca\.crt)' '/var/snap/k8s/common/args/kubelet'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242421](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242421): Kubernetes Controller Manager must have the SSL Certificate Authority set.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes Controller Manager is responsible for creating service accounts and tokens for the API Server, maintaining the correct number of pods for every replication controller and provides notifications when nodes are offline.  
> 
> Anyone who gains access to the Controller Manager can generate backdoor accounts, take possession of, or diminish system performance without detection by disabling system notification. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes Controller Manager with a means to be able to authenticate sessions and encrypt traffic.





#### Comments:
> The command line arguments of the Kubernetes Controller Manager
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-controller-manager
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-controller-manager` in order to set the argument `--root-ca-file` for service `kube-controller-manager` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/ca\.crt`

Afterwards restart the `kube-controller-manager` service with:



    sudo systemctl restart snap.k8s.kube-controller-manager



#### Auditing (as root)

Ensure that the argument `--root-ca-file` for service `kube-controller-manager` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-controller-manager`.

```bash
grep -E -q  '\-\-root-ca-file=(/etc/kubernetes/pki/ca\.crt)' '/var/snap/k8s/common/args/kube-controller-manager'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242422](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242422): Kubernetes API Server must have a certificate for communication.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes control plane and external communication is managed by API Server. The main implementation of the API Server is to manage hardware resources for pods and container using horizontal or vertical scaling. Anyone who can access the API Server can effectively control the Kubernetes architecture. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server with a means to be able to authenticate sessions and encrypt traffic. 
> 
> To enable encrypted communication for API Server, the parameter etcd-cafile must be set. This parameter gives the location of the SSL Certificate Authority file used to secure API Server communication.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--tls-cert-file` for service `kube-apiserver` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/apiserver\.crt`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure that the argument `--tls-cert-file` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-tls-cert-file=(/etc/kubernetes/pki/apiserver\.crt)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> The command line arguments of the Kubernetes API Server
> in the k8s-snap are defined in the following file:
> 
>     /var/snap/k8s/common/args/kube-apiserver
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--tls-private-key-file` for service `kube-apiserver` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/apiserver\.key`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--tls-private-key-file` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-tls-private-key-file=(/etc/kubernetes/pki/apiserver\.key)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242423](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242423): Kubernetes etcd must enable client authentication to secure service.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes container and pod configuration are maintained by Kubelet. Kubelet agents register nodes with the API Server, mount volume storage, and perform health checks for containers and pods. Anyone who gains access to Kubelet agents can effectively control applications within the pods and containers. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server with a means to be able to authenticate sessions and encrypt traffic.
> 
> To enable encrypted communication for Kubelet, the parameter client-cert-auth must be set. This parameter gives the location of the SSL Certificate Authority file used to secure Kubelet communication.




<details>
<summary><h4>Step 1/3</h4></summary>
<br>


#### Comments:
> This finding refers to the `--cert-file` command line argument for the
> etcd service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The k8s-snap configures the Kube API Server to connect to
> k8s-dqlite via local socket owned by root.
> 
> The Auditing section will describe how to check the ownership
> of the k8s-dqlite socket.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/3</h4></summary>
<br>


#### Comments:
> This check ensures the permissions on the k8s-dqlite socket.


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 600 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure all required files have permissions '600' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 3/3</h4></summary>
<br>


#### Comments:
> This check ensures the `--etcd-servers` argument of the Kube API Server
> is as expected.
> 


<details>
<summary><h4>Remediation for Step 3</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--etcd-servers` for service `kube-apiserver` as appropriate.

Ensure it is set to: `unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 3</h4></summary>
<br>

Ensure that the argument `--etcd-servers` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-etcd-servers=(unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242424](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242424): Kubernetes Kubelet must enable tlsPrivateKeyFile for client authentication to secure service.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes container and pod configuration are maintained by Kubelet. Kubelet agents register nodes with the API Server, mount volume storage, and perform health checks for containers and pods. Anyone who gains access to Kubelet agents can effectively control applications within the pods and containers. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols such as TLS. TLS provides the Kubernetes API Server with a means to authenticate sessions and encrypt traffic.
> 
> To enable encrypted communication for Kubelet, the tlsPrivateKeyFile must be set. This parameter gives the location of the SSL Certificate Authority file used to secure Kubelet communication.





#### Comments:
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service, but does explicitly pass
> `--tls-private-key-file=/etc/kubernetes/pki/kubelet.key`
> as a command line argument.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--tls-private-key-file` for service `kubelet` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/kubelet\.key`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--tls-private-key-file` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

```bash
grep -E -q  '\-\-tls-private-key-file=(/etc/kubernetes/pki/kubelet\.key)' '/var/snap/k8s/common/args/kubelet'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242425](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242425): Kubernetes Kubelet must enable tlsCertFile for client authentication to secure service.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes container and pod configuration are maintained by Kubelet. Kubelet agents register nodes with the API Server, mount volume storage, and perform health checks for containers and pods. Anyone who gains access to Kubelet agents can effectively control applications within the pods and containers. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols such as TLS. TLS provides the Kubernetes API Server with a means to authenticate sessions and encrypt traffic.
> 
> To enable encrypted communication for Kubelet, the parameter tlsCertFile must be set. This parameter gives the location of the SSL Certificate Authority file used to secure Kubelet communication.





#### Comments:
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service, but does explicitly pass
> `--tls-cert-file=/etc/kubernetes/pki/kubelet.crt`
> as a command line argument.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--tls-cert-file` for service `kubelet` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/kubelet\.crt`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--tls-cert-file` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

```bash
grep -E -q  '\-\-tls-cert-file=(/etc/kubernetes/pki/kubelet\.crt)' '/var/snap/k8s/common/args/kubelet'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242426](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242426): Kubernetes etcd must enable client authentication to secure service.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes container and pod configuration are maintained by Kubelet. Kubelet agents register nodes with the API Server, mount volume storage, and perform health checks for containers and pods. Anyone who gains access to Kubelet agents can effectively control applications within the pods and containers. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server with a means to be able to authenticate sessions and encrypt traffic.
> 
> Etcd is a highly-available key value store used by Kubernetes deployments for persistent storage of all of its REST API objects. These objects are sensitive and should be accessible only by authenticated etcd peers in the etcd cluster. The parameter "--peer-client-cert-auth" must be set for etcd to check all incoming peer requests from the cluster for valid client certificates.





#### Comments:
> This finding refers to the `--peer-client-cert-auth` command
> line argument for the etcd service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> k8s-dqlite peer communication defaults to using TLS unless the
> `--enable-tls` argument is set in k8s-dqlite argument configuration
> file located at:
> 
>     /var/snap/k8s/common/args/k8s-dqlite
> 


#### Remediation
Edit `/var/snap/k8s/common/args/k8s-dqlite` in order to set the argument `--enable-tls` for service `k8s-dqlite` as appropriate.

Ensure it is NOT set to one of: `false`, `0`

Afterwards restart the `k8s-dqlite` service with:



    sudo systemctl restart snap.k8s.k8s-dqlite



#### Auditing (as root)

Ensure that the argument `--enable-tls` for service `k8s-dqlite` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/k8s-dqlite`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-enable-tls=(false|0)' '/var/snap/k8s/common/args/k8s-dqlite' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

### [V-242427](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242427): Kubernetes etcd must have a key file for secure communication.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes stores configuration and state information in a distributed key-value store called etcd. Anyone who can write to etcd can effectively control the Kubernetes cluster. Even just reading the contents of etcd could easily provide helpful hints to a would-be attacker. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server and etcd with a means to be able to authenticate sessions and encrypt traffic. 
> 
> To enable encrypted communication for etcd, the parameter key-file must be set. This parameter gives the location of the key file used to secure etcd communication.




<details>
<summary><h4>Step 1/3</h4></summary>
<br>


#### Comments:
> This finding refers to the `--key-file` command line argument for the
> etcd service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The k8s-snap configures the Kube API Server to connect to
> k8s-dqlite via local socket owned by root.
> 
> The Auditing section will describe how to check the ownership
> of the k8s-dqlite socket.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/3</h4></summary>
<br>


#### Comments:
> This check ensures the permissions on the k8s-dqlite socket.


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 600 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure all required files have permissions '600' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 3/3</h4></summary>
<br>


#### Comments:
> This check ensures the `--etcd-servers` argument of the Kube API Server
> is as expected.
> 


<details>
<summary><h4>Remediation for Step 3</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--etcd-servers` for service `kube-apiserver` as appropriate.

Ensure it is set to: `unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 3</h4></summary>
<br>

Ensure that the argument `--etcd-servers` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-etcd-servers=(unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242428](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242428): Kubernetes etcd must have a certificate for communication.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes stores configuration and state information in a distributed key-value store called etcd. Anyone who can write to etcd can effectively control a Kubernetes cluster. Even just reading the contents of etcd could easily provide helpful hints to a would-be attacker. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server and etcd with a means to be able to authenticate sessions and encrypt traffic. 
> 
> To enable encrypted communication for etcd, the parameter cert-file must be set. This parameter gives the location of the SSL certification file used to secure etcd communication.




<details>
<summary><h4>Step 1/3</h4></summary>
<br>


#### Comments:
> This finding refers to the `--cert-file` command line argument for the
> etcd service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The k8s-snap configures the Kube API Server to connect to
> k8s-dqlite via local socket owned by root.
> 
> The Auditing section will describe how to check the ownership
> of the k8s-dqlite socket.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/3</h4></summary>
<br>


#### Comments:
> This check ensures the permissions on the k8s-dqlite socket.


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 600 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure all required files have permissions '600' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 3/3</h4></summary>
<br>


#### Comments:
> This check ensures the `--etcd-servers` argument of the Kube API Server
> is as expected.
> 


<details>
<summary><h4>Remediation for Step 3</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--etcd-servers` for service `kube-apiserver` as appropriate.

Ensure it is set to: `unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 3</h4></summary>
<br>

Ensure that the argument `--etcd-servers` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-etcd-servers=(unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242429](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242429): Kubernetes etcd must have the SSL Certificate Authority set.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes stores configuration and state information in a distributed key-value store called etcd. Anyone who can write to etcd can effectively control a Kubernetes cluster. Even just reading the contents of etcd could easily provide helpful hints to a would-be attacker. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server and etcd with a means to be able to authenticate sessions and encrypt traffic. 
> 
> To enable encrypted communication for etcd, the parameter "--etcd-cafile" must be set. This parameter gives the location of the SSL Certificate Authority file used to secure etcd communication.




<details>
<summary><h4>Step 1/3</h4></summary>
<br>


#### Comments:
> This finding refers to the `--etcd-cafile` command line argument for the
> Kube API Service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The k8s-snap configures the Kube API Server to connect to
> k8s-dqlite via local socket owned by root.
> 
> The Auditing section will describe how to check the ownership
> of the k8s-dqlite socket.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/3</h4></summary>
<br>


#### Comments:
> This check ensures the permissions on the k8s-dqlite socket.


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 600 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure all required files have permissions '600' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 3/3</h4></summary>
<br>


#### Comments:
> This check ensures the `--etcd-servers` argument of the Kube API Server
> is as expected.
> 


<details>
<summary><h4>Remediation for Step 3</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--etcd-servers` for service `kube-apiserver` as appropriate.

Ensure it is set to: `unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 3</h4></summary>
<br>

Ensure that the argument `--etcd-servers` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-etcd-servers=(unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242430](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242430): Kubernetes etcd must have a certificate for communication.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes stores configuration and state information in a distributed key-value store called etcd. Anyone who can write to etcd can effectively control the Kubernetes cluster. Even just reading the contents of etcd could easily provide helpful hints to a would-be attacker. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server and etcd with a means to be able to authenticate sessions and encrypt traffic. 
> 
> To enable encrypted communication for etcd, the parameter "--etcd-certfile" must be set. This parameter gives the location of the SSL certification file used to secure etcd communication.




<details>
<summary><h4>Step 1/3</h4></summary>
<br>


#### Comments:
> This finding refers to the `--etcd-certfile` command line argument for the
> Kube API Service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The k8s-snap configures the Kube API Server to connect to
> k8s-dqlite via local socket owned by root.
> 
> The Auditing section will describe how to check the ownership
> of the k8s-dqlite socket.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/3</h4></summary>
<br>


#### Comments:
> This check ensures the permissions on the k8s-dqlite socket.


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 600 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure all required files have permissions '600' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 3/3</h4></summary>
<br>


#### Comments:
> This check ensures the `--etcd-servers` argument of the Kube API Server
> is as expected.
> 


<details>
<summary><h4>Remediation for Step 3</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--etcd-servers` for service `kube-apiserver` as appropriate.

Ensure it is set to: `unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 3</h4></summary>
<br>

Ensure that the argument `--etcd-servers` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-etcd-servers=(unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242431](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242431): Kubernetes etcd must have a key file for secure communication.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes stores configuration and state information in a distributed key-value store called etcd. Anyone who can write to etcd can effectively control a Kubernetes cluster. Even just reading the contents of etcd could easily provide helpful hints to a would-be attacker. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server and etcd with a means to be able to authenticate sessions and encrypt traffic. 
> 
> To enable encrypted communication for etcd, the parameter "--etcd-keyfile" must be set. This parameter gives the location of the key file used to secure etcd communication.




<details>
<summary><h4>Step 1/3</h4></summary>
<br>


#### Comments:
> This finding refers to the `--etcd-keyfile` command line argument
> for the Kube API Service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The k8s-snap configures the Kube API Server to connect to
> k8s-dqlite via local socket owned by root.
> 
> The Auditing section will describe how to check the ownership
> of the k8s-dqlite socket.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/3</h4></summary>
<br>


#### Comments:
> This check ensures the permissions on the k8s-dqlite socket.


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 600 /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure all required files have permissions '600' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 3/3</h4></summary>
<br>


#### Comments:
> This check ensures the `--etcd-servers` argument of the Kube API Server
> is as expected.
> 


<details>
<summary><h4>Remediation for Step 3</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--etcd-servers` for service `kube-apiserver` as appropriate.

Ensure it is set to: `unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



</details>
<details>
<summary><h4>Auditing (as root) for Step 3</h4></summary>
<br>

Ensure that the argument `--etcd-servers` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-etcd-servers=(unix:///var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242432](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242432): Kubernetes etcd must have peer-cert-file set for secure communication.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes stores configuration and state information in a distributed key-value store called etcd. Anyone who can write to etcd can effectively control the Kubernetes cluster. Even just reading the contents of etcd could easily provide helpful hints to a would-be attacker. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server and etcd with a means to be able to authenticate sessions and encrypt traffic. 
> 
> To enable encrypted communication for etcd, the parameter peer-cert-file must be set. This parameter gives the location of the SSL certification file used to secure etcd communication.





#### Comments:
> This finding refers to the `--peer-cert-file` command
> line argument for the etcd service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The Peer Certificate File used by k8s-dqlite is located at:
> 
>     /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt
> 
> The directory of the certificate file is governed by the
> `--storage-dir` k8s-dqlite argument.
> 


#### Remediation
Edit `/var/snap/k8s/common/args/k8s-dqlite` in order to set the argument `--storage-dir` for service `k8s-dqlite` as appropriate.

Ensure it is set to: `/var/snap/k8s/common/var/lib/k8s-dqlite`

Afterwards restart the `k8s-dqlite` service with:



    sudo systemctl restart snap.k8s.k8s-dqlite



#### Auditing (as root)

Ensure that the argument `--storage-dir` for service `k8s-dqlite` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/k8s-dqlite`.

```bash
grep -E -q  '\-\-storage-dir=(/var/snap/k8s/common/var/lib/k8s-dqlite)' '/var/snap/k8s/common/args/k8s-dqlite'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242433](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242433): Kubernetes etcd must have a peer-key-file set for secure communication.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes stores configuration and state information in a distributed key-value store called etcd. Anyone who can write to etcd can effectively control a Kubernetes cluster. Even just reading the contents of etcd could easily provide helpful hints to a would-be attacker. Using authenticity protection, the communication can be protected against man-in-the-middle attacks/session hijacking and the insertion of false information into sessions.
> 
> The communication session is protected by utilizing transport encryption protocols, such as TLS. TLS provides the Kubernetes API Server and etcd with a means to be able to authenticate sessions and encrypt traffic. 
> 
> To enable encrypted communication for etcd, the parameter peer-key-file must be set. This parameter gives the location of the SSL certification file used to secure etcd communication.





#### Comments:
> This finding refers to the `--peer-key-file` command
> line argument for the etcd service.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The Peer Key File used by k8s-dqlite is located at:
> 
>     /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key
> 
> The directory of the key file is governed by the
> `--storage-dir` k8s-dqlite argument.
> 


#### Remediation
Edit `/var/snap/k8s/common/args/k8s-dqlite` in order to set the argument `--storage-dir` for service `k8s-dqlite` as appropriate.

Ensure it is set to: `/var/snap/k8s/common/var/lib/k8s-dqlite`

Afterwards restart the `k8s-dqlite` service with:



    sudo systemctl restart snap.k8s.k8s-dqlite



#### Auditing (as root)

Ensure that the argument `--storage-dir` for service `k8s-dqlite` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/k8s-dqlite`.

```bash
grep -E -q  '\-\-storage-dir=(/var/snap/k8s/common/var/lib/k8s-dqlite)' '/var/snap/k8s/common/args/k8s-dqlite'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

### [V-242438](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242438): Kubernetes API Server must configure timeouts to limit attack surface.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes API Server request timeouts sets the duration a request stays open before timing out. Since the API Server is the central component in the Kubernetes Control Plane, it is vital to protect this service. If request timeouts were not set, malicious attacks or unwanted activities might affect multiple deployments across different applications or environments. This might deplete all resources from the Kubernetes infrastructure causing the information system to go offline. The "--request-timeout" value must never be set to "0". This disables the request-timeout feature. (By default, the "--request-timeout" is set to "1 minute".)





#### Comments:
> The Finding also allows for setting a timeout larger than 300s.
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--request-timeout` for service `kube-apiserver` as appropriate.

Ensure it is set to: `300s`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--request-timeout` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-request-timeout=(300s)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242442](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242442): Kubernetes must remove old components after updated versions have been installed.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> Previous versions of Kubernetes components that are not removed after updates have been installed may be exploited by adversaries by allowing the vulnerabilities to still exist within the cluster. It is important for Kubernetes to remove old pods when newer pods are created using new images to always be at the desired security state.





#### Comments:
> This Finding recommends checking that no residual versions of Kubernetes
> components are left running following upgrades of the Kubernetes cluster.
> 
> Thanks to the k8s-snap's distribution and upgrade model, it is not
> possible for this to occur, so this Finding is Not Applicable.
> 



### [V-242443](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242443): Kubernetes must contain the latest updates as authorized by IAVMs, CTOs, DTMs, and STIGs.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> Kubernetes software must stay up to date with the latest patches, service packs, and hot fixes. Not updating the Kubernetes control plane will expose the organization to vulnerabilities.
> 
> Flaws discovered during security assessments, continuous monitoring, incident response activities, or information system error handling must also be addressed expeditiously. 
> 
> Organization-defined time periods for updating security-relevant container platform components may vary based on a variety of factors including, for example, the security category of the information system or the criticality of the update (i.e., severity of the vulnerability related to the discovered flaw). 
> 
> This requirement will apply to software patch management solutions that are used to install patches across the enclave and also to applications themselves that are not part of that patch management solution. For example, many browsers today provide the capability to install their own patch software. Patch criticality, as well as system criticality will vary. Therefore, the tactical situations regarding the patch management process will also vary. This means that the time period utilized must be a configurable parameter. Time frames for application of security-relevant software updates may be dependent upon the IAVM process.
> 
> The container platform components will be configured to check for and install security-relevant software updates within an identified time period from the availability of the update. The container platform registry will ensure the images are current. The specific time period will be defined by an authoritative source (e.g., IAVM, CTOs, DTMs, and STIGs).





#### Comments:
> This Finding recommends checking all Kubernetes component versions
> are actively supported.
> 
> https://kubernetes.io/releases/version-skew-policy/#supported-versions
> 
> Supported versions of the k8s-snap should always ship with supported
> versions of Kubernetes components within it, so this Finding is
> Not Applicable.
> 



### [V-242444](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242444): The Kubernetes component manifests must be owned by root.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes manifests are those files that contain the arguments and settings for the Control Plane services. These services are etcd, the api server, controller, proxy, and scheduler. If these files can be changed, the scheduler will be implementing the changes immediately. Many of the security settings within the document are implemented through these manifests.





#### Comments:
> The manifest files for the Kubernetes services in the k8s-snap are
> located in the following directories:
> 
>     /etc/kubernetes
> 


#### Remediation
Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /etc/kubernetes /etc/kubernetes/pki /etc/kubernetes/kubelet.conf /etc/kubernetes/scheduler.conf /etc/kubernetes/proxy.conf /etc/kubernetes/admin.conf /etc/kubernetes/controller.conf /etc/kubernetes/pki/etcd /etc/kubernetes/pki/client-ca.crt /etc/kubernetes/pki/front-proxy-ca.key /etc/kubernetes/pki/apiserver.key /etc/kubernetes/pki/apiserver.crt /etc/kubernetes/pki/apiserver-kubelet-client.key /etc/kubernetes/pki/front-proxy-client.crt /etc/kubernetes/pki/serviceaccount.key /etc/kubernetes/pki/front-proxy-client.key /etc/kubernetes/pki/kubelet.crt /etc/kubernetes/pki/ca.crt /etc/kubernetes/pki/ca.key /etc/kubernetes/pki/apiserver-kubelet-client.crt /etc/kubernetes/pki/front-proxy-ca.crt /etc/kubernetes/pki/kubelet.key

#### Auditing (as root)

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/etc/kubernetes' | grep -q 0:0 && echo PASS /etc/kubernetes: 0:0 || echo FAIL /etc/kubernetes: 0:0
stat -c %u:%g '/etc/kubernetes/pki' | grep -q 0:0 && echo PASS /etc/kubernetes/pki: 0:0 || echo FAIL /etc/kubernetes/pki: 0:0
stat -c %u:%g '/etc/kubernetes/kubelet.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/kubelet.conf: 0:0 || echo FAIL /etc/kubernetes/kubelet.conf: 0:0
stat -c %u:%g '/etc/kubernetes/scheduler.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/scheduler.conf: 0:0 || echo FAIL /etc/kubernetes/scheduler.conf: 0:0
stat -c %u:%g '/etc/kubernetes/proxy.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/proxy.conf: 0:0 || echo FAIL /etc/kubernetes/proxy.conf: 0:0
stat -c %u:%g '/etc/kubernetes/admin.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/admin.conf: 0:0 || echo FAIL /etc/kubernetes/admin.conf: 0:0
stat -c %u:%g '/etc/kubernetes/controller.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/controller.conf: 0:0 || echo FAIL /etc/kubernetes/controller.conf: 0:0
stat -c %u:%g '/etc/kubernetes/pki/etcd' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/etcd: 0:0 || echo FAIL /etc/kubernetes/pki/etcd: 0:0
stat -c %u:%g '/etc/kubernetes/pki/client-ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/client-ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/client-ca.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-ca.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-ca.key: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-ca.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver.key: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver.crt: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver-kubelet-client.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver-kubelet-client.key: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver-kubelet-client.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-client.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-client.crt: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-client.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/serviceaccount.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/serviceaccount.key: 0:0 || echo FAIL /etc/kubernetes/pki/serviceaccount.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-client.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-client.key: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-client.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/kubelet.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/kubelet.crt: 0:0 || echo FAIL /etc/kubernetes/pki/kubelet.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/ca.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/ca.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/ca.key: 0:0 || echo FAIL /etc/kubernetes/pki/ca.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver-kubelet-client.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver-kubelet-client.crt: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver-kubelet-client.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-ca.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/kubelet.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/kubelet.key: 0:0 || echo FAIL /etc/kubernetes/pki/kubelet.key: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

### [V-242445](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242445): The Kubernetes component etcd must be owned by etcd.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes etcd key-value store provides a way to store data to the Control Plane. If these files can be changed, data to API object and the Control Plane would be compromised. The scheduler will implement the changes immediately. Many of the security settings within the document are implemented through this file.





#### Comments:
> This Finding refers to checking the ownership of all etcd-related
> files under /var/lib/etcd/*.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The state directory for k8s-dqlite within the k8s-snap is located under:
> 
>     /var/snap/k8s/common/var/lib/k8s-dqlite
> 
> Related Finding V-242459 contains directives on the permissions of the files.
> 


#### Remediation
Ensure all of the following paths have correct ownership by running: 



    chown 0:0 /var/snap/k8s/common/var/lib/k8s-dqlite /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml /var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

#### Auditing (as root)

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite: 0:0
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml: 0:0
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml: 0:0
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key: 0:0
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt: 0:0
stat -c %u:%g '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 0:0 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

### [V-242446](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242446): The Kubernetes conf files must be owned by root.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes conf files contain the arguments and settings for the Control Plane services. These services are controller and scheduler. If these files can be changed, the scheduler will be implementing the changes immediately. Many of the security settings within the document are implemented through this file.





#### Comments:
> Note that the original Finding references 'controller-manager.conf',
> but the k8s-snap uses 'controller.conf'.
> 
> Finding V-242460 defines the permissions checks for these files.
> 


#### Remediation
Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /etc/kubernetes/admin.conf /etc/kubernetes/scheduler.conf /etc/kubernetes/controller.conf

#### Auditing (as root)

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/etc/kubernetes/admin.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/admin.conf: 0:0 || echo FAIL /etc/kubernetes/admin.conf: 0:0
stat -c %u:%g '/etc/kubernetes/scheduler.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/scheduler.conf: 0:0 || echo FAIL /etc/kubernetes/scheduler.conf: 0:0
stat -c %u:%g '/etc/kubernetes/controller.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/controller.conf: 0:0 || echo FAIL /etc/kubernetes/controller.conf: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

### [V-242447](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242447): The Kubernetes Kube Proxy kubeconfig must have file permissions set to 644 or more restrictive.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes Kube Proxy kubeconfig contain the argument and setting for the Control Planes. These settings contain network rules for restricting network communication between pods, clusters, and networks. If these files can be changed, data traversing between the Kubernetes Control Panel components would be compromised. Many of the security settings within the document are implemented through this file.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> Finding stipulates that permission mask should be at most 644,
> but they can also be set to be more restrictive.
> 
> Finding V-242448 defines the associated file ownership requirements.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 644 /etc/kubernetes/proxy.conf

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/etc/kubernetes/proxy.conf' | grep -q 600 && echo PASS /etc/kubernetes/proxy.conf: 600 || echo FAIL /etc/kubernetes/proxy.conf: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check was added to ensure the Kubernetes Proxy configuration
> file path is set as expected.
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-proxy` in order to set the argument `--kubeconfig` for service `kube-proxy` as appropriate.

Ensure it is set to: `/etc/kubernetes/proxy\.conf`

Afterwards restart the `kube-proxy` service with:



    sudo systemctl restart snap.k8s.kube-proxy



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--kubeconfig` for service `kube-proxy` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-proxy`.

```bash
grep -E -q  '\-\-kubeconfig=(/etc/kubernetes/proxy\.conf)' '/var/snap/k8s/common/args/kube-proxy'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242448](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242448): The Kubernetes Kube Proxy kubeconfig must be owned by root.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes Kube Proxy kubeconfig contain the argument and setting for the Control Planes. These settings contain network rules for restricting network communication between pods, clusters, and networks. If these files can be changed, data traversing between the Kubernetes Control Panel components would be compromised. Many of the security settings within the document are implemented through this file.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> Finding stipulates the file should be owned by the root user/group.
> 
> Finding V-242447 defines the associated file permissions requirements.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /etc/kubernetes/proxy.conf

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/etc/kubernetes/proxy.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/proxy.conf: 0:0 || echo FAIL /etc/kubernetes/proxy.conf: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check was added to ensure the proxy config is as expected.
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kube-proxy` in order to set the argument `--kubeconfig` for service `kube-proxy` as appropriate.

Ensure it is set to: `/etc/kubernetes/proxy\.conf`

Afterwards restart the `kube-proxy` service with:



    sudo systemctl restart snap.k8s.kube-proxy



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--kubeconfig` for service `kube-proxy` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-proxy`.

```bash
grep -E -q  '\-\-kubeconfig=(/etc/kubernetes/proxy\.conf)' '/var/snap/k8s/common/args/kube-proxy'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242449](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242449): The Kubernetes Kubelet certificate authority file must have file permissions set to 644 or more restrictive.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes kubelet certificate authority file contains settings for the Kubernetes Node TLS certificate authority. Any request presenting a client certificate signed by one of the authorities in the client-ca-file is authenticated with an identity corresponding to the CommonName of the client certificate. If this file can be changed, the Kubernetes architecture could be compromised. The scheduler will implement the changes immediately. Many of the security settings within the document are implemented through this file.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> Finding stipulates that permission mask should be at most 644,
> but they can also be set to be more restrictive.
> 
> Finding V-242450 defines the associated file ownership requirements.
> Finding V-242451 deines the associated directory ownership requirements.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod 644 /etc/kubernetes/pki/client-ca.crt

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/etc/kubernetes/pki/client-ca.crt' | grep -q 600 && echo PASS /etc/kubernetes/pki/client-ca.crt: 600 || echo FAIL /etc/kubernetes/pki/client-ca.crt: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check was added to ensure the `--client-ca-file` is as expected.
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--client-ca-file` for service `kubelet` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/client-ca\.crt`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--client-ca-file` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

```bash
grep -E -q  '\-\-client-ca-file=(/etc/kubernetes/pki/client-ca\.crt)' '/var/snap/k8s/common/args/kubelet'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242450](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242450): The Kubernetes Kubelet certificate authority must be owned by root.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes kube proxy kubeconfig contain the argument and setting for the Control Planes. These settings contain network rules for restricting network communication between pods, clusters, and networks. If these files can be changed, data traversing between the Kubernetes Control Panel components would be compromised. Many of the security settings within the document are implemented through this file.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> Finding stipulates the file should be owned by the root user/group.
> 
> Finding V-242449 defines the associated file permissions requirements.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /etc/kubernetes/pki/client-ca.crt

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/etc/kubernetes/pki/client-ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/client-ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/client-ca.crt: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check was added to ensure the `--client-ca-file` is as expected.
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--client-ca-file` for service `kubelet` as appropriate.

Ensure it is set to: `/etc/kubernetes/pki/client-ca\.crt`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--client-ca-file` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

```bash
grep -E -q  '\-\-client-ca-file=(/etc/kubernetes/pki/client-ca\.crt)' '/var/snap/k8s/common/args/kubelet'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242451](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242451): The Kubernetes component PKI must be owned by root.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes PKI directory contains all certificates (.crt files) supporting secure network communications in the Kubernetes Control Plane. If these files can be modified, data traversing within the architecture components would become unsecure and compromised. Many of the security settings within the document are implemented through this file.





#### Comments:
> The k8s-snap stores PKI-related files in the following directory:
> 
>     /etc/kubernetes/pki
> 
> Finding V-242466 stipulates the permissions of this directory.
> 


#### Remediation
Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /etc/kubernetes/pki /etc/kubernetes/pki/etcd /etc/kubernetes/pki/client-ca.crt /etc/kubernetes/pki/front-proxy-ca.key /etc/kubernetes/pki/apiserver.key /etc/kubernetes/pki/apiserver.crt /etc/kubernetes/pki/apiserver-kubelet-client.key /etc/kubernetes/pki/front-proxy-client.crt /etc/kubernetes/pki/serviceaccount.key /etc/kubernetes/pki/front-proxy-client.key /etc/kubernetes/pki/kubelet.crt /etc/kubernetes/pki/ca.crt /etc/kubernetes/pki/ca.key /etc/kubernetes/pki/apiserver-kubelet-client.crt /etc/kubernetes/pki/front-proxy-ca.crt /etc/kubernetes/pki/kubelet.key

#### Auditing (as root)

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/etc/kubernetes/pki' | grep -q 0:0 && echo PASS /etc/kubernetes/pki: 0:0 || echo FAIL /etc/kubernetes/pki: 0:0
stat -c %u:%g '/etc/kubernetes/pki/etcd' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/etcd: 0:0 || echo FAIL /etc/kubernetes/pki/etcd: 0:0
stat -c %u:%g '/etc/kubernetes/pki/client-ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/client-ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/client-ca.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-ca.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-ca.key: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-ca.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver.key: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver.crt: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver-kubelet-client.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver-kubelet-client.key: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver-kubelet-client.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-client.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-client.crt: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-client.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/serviceaccount.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/serviceaccount.key: 0:0 || echo FAIL /etc/kubernetes/pki/serviceaccount.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-client.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-client.key: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-client.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/kubelet.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/kubelet.crt: 0:0 || echo FAIL /etc/kubernetes/pki/kubelet.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/ca.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/ca.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/ca.key: 0:0 || echo FAIL /etc/kubernetes/pki/ca.key: 0:0
stat -c %u:%g '/etc/kubernetes/pki/apiserver-kubelet-client.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/apiserver-kubelet-client.crt: 0:0 || echo FAIL /etc/kubernetes/pki/apiserver-kubelet-client.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/front-proxy-ca.crt' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/front-proxy-ca.crt: 0:0 || echo FAIL /etc/kubernetes/pki/front-proxy-ca.crt: 0:0
stat -c %u:%g '/etc/kubernetes/pki/kubelet.key' | grep -q 0:0 && echo PASS /etc/kubernetes/pki/kubelet.key: 0:0 || echo FAIL /etc/kubernetes/pki/kubelet.key: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

### [V-242452](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242452): The Kubernetes kubelet KubeConfig must have file permissions set to 644 or more restrictive.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes kubelet agent registers nodes with the API Server, mounts volume storage for pods, and performs health checks to containers within pods. If these files can be modified, the information system would be unaware of pod or container degradation. Many of the security settings within the document are implemented through this file.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> Finding stipulates that permission mask should be at most 644,
> but they can also be set to be more restrictive.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 644 /etc/kubernetes/kubelet.conf

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/etc/kubernetes/kubelet.conf' | grep -q 600 && echo PASS /etc/kubernetes/kubelet.conf: 600 || echo FAIL /etc/kubernetes/kubelet.conf: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check was added to ensure Kubelet's `--kubeconfig` is as expected.


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--kubeconfig` for service `kubelet` as appropriate.

Ensure it is set to: `/etc/kubernetes/kubelet\.conf`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--kubeconfig` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

```bash
grep -E -q  '\-\-kubeconfig=(/etc/kubernetes/kubelet\.conf)' '/var/snap/k8s/common/args/kubelet'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242453](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242453): The Kubernetes kubelet KubeConfig file must be owned by root.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes kubelet agent registers nodes with the API server and performs health checks to containers within pods. If these files can be modified, the information system would be unaware of pod or container degradation. Many of the security settings within the document are implemented through this file.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> Finding stipulates the file should be owned by the root user/group.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct ownership by running: 



    chown -R 0:0 /etc/kubernetes/kubelet.conf

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all files exist and have the correct ownership.

```bash
stat -c %u:%g '/etc/kubernetes/kubelet.conf' | grep -q 0:0 && echo PASS /etc/kubernetes/kubelet.conf: 0:0 || echo FAIL /etc/kubernetes/kubelet.conf: 0:0
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check was added to ensure Kubelet's `--kubeconfig` is as expected.


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--kubeconfig` for service `kubelet` as appropriate.

Ensure it is set to: `/etc/kubernetes/kubelet\.conf`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--kubeconfig` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

```bash
grep -E -q  '\-\-kubeconfig=(/etc/kubernetes/kubelet\.conf)' '/var/snap/k8s/common/args/kubelet'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242454](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242454): The Kubernetes kubeadm.conf must be owned by root.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> The Kubernetes kubeeadm.conf contains sensitive information regarding the cluster nodes configuration. If this file can be modified, the Kubernetes Platform Plane would be degraded or compromised for malicious intent. Many of the security settings within the document are implemented through this file.





#### Comments:
> This Finding stipulates the file ownership of the kubeadm executable,
> which does not ship as part of the k8s-snap.
> 
> The Auditor may check whether the binary was installed separately and its
> permissions are correct by permforming:
> 
>     # Should print 'root:root' if the kubeadm binary exists in $PATH.
>     stat -c %U:%G $(which kubeadm)
> 



### [V-242455](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242455): The Kubernetes kubeadm.conf must have file permissions set to 644 or more restrictive.

#### Severity: Medium

#### Status: Not Applicable

#### Upstream Finding Description:

> The Kubernetes kubeadm.conf contains sensitive information regarding the cluster nodes configuration. If this file can be modified, the Kubernetes Platform Plane would be degraded or compromised for malicious intent. Many of the security settings within the document are implemented through this file.





#### Comments:
> This Finding stipulates the file ownership of the kubeadm executable,
> which does not ship as part of the k8s-snap.
> 
> The Auditor may check whether the binary was installed separately and its
> permissions are correct by permforming:
> 
>     # Should print 'root:root' if the kubeadm binary exists in $PATH.
>     stat -c %U:%G $(which kubeadm)
> 



### [V-242456](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242456): The Kubernetes kubelet config must have file permissions set to 644 or more restrictive.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes kubelet agent registers nodes with the API server and performs health checks to containers within pods. If this file can be modified, the information system would be unaware of pod or container degradation.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> This Finding relates to the permissions on the `/var/lib/kubelet/config.yaml`
> file.
> 
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 
> The Auditing section will advise on how to check the permissions
> of said file.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 644 /var/snap/k8s/common/args/kubelet

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/args/kubelet' | grep -q 600 && echo PASS /var/snap/k8s/common/args/kubelet: 600 || echo FAIL /var/snap/k8s/common/args/kubelet: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check is defined to ensure that Kubelet is not passed
> a `--config` file argument in the k8s-snap.
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--config` for service `kubelet` as appropriate.

Ensure it is NOT set to any value.

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--config` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-config=(.*)' '/var/snap/k8s/common/args/kubelet' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242457](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242457): The Kubernetes kubelet config must be owned by root.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes kubelet agent registers nodes with the API Server and performs health checks to containers within pods. If this file can be modified, the information system would be unaware of pod or container degradation.




<details>
<summary><h4>Step 1/2</h4></summary>
<br>


#### Comments:
> This Finding relates to the permissions on the `/var/lib/kubelet/config.yaml`
> file in relation to it being used by `kubeadm`.
> 
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service, nor does it ship with `kubeadm`.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 
> The Auditing section will advise on how to check the permissions
> of said file.
> 


<details>
<summary><h4>Remediation for Step 1</h4></summary>
<br>

Ensure all of the following paths have correct permissions by running: 



    chmod -R 644 /var/snap/k8s/common/args/kubelet

</details>
<details>
<summary><h4>Auditing (as root) for Step 1</h4></summary>
<br>

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/args/kubelet' | grep -q 600 && echo PASS /var/snap/k8s/common/args/kubelet: 600 || echo FAIL /var/snap/k8s/common/args/kubelet: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

</details>

<details>
<summary><h4>Step 2/2</h4></summary>
<br>


#### Comments:
> This check is defined to ensure that Kubelet is not passed
> a `--config` file argument in the k8s-snap.
> 


<details>
<summary><h4>Remediation for Step 2</h4></summary>
<br>

Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--config` for service `kubelet` as appropriate.

Ensure it is NOT set to any value.

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



</details>
<details>
<summary><h4>Auditing (as root) for Step 2</h4></summary>
<br>

Ensure that the argument `--config` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

Note: this Finding allows for this argument to be UNSET as well.

```bash
grep -E -qvz '\-\-config=(.*)' '/var/snap/k8s/common/args/kubelet' && echo UNSET
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `UNSET`.

The final line of the output will be `PASS`.

</details>

</details>

### [V-242459](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242459): The Kubernetes etcd must have file permissions set to 644 or more restrictive.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes etcd key-value store provides a way to store data to the Control Plane. If these files can be changed, data to API object and Control Plane would be compromised.





#### Comments:
> This Finding refers to checking the ownership of all etcd-related
> files under /var/lib/etcd/*.
> 
> The k8s-snap does not use etcd in any way, instead relying on
> [k8s-dqlite](https://github.com/canonical/k8s-dqlite) for its
> state handling.
> 
> The state directory for k8s-dqlite within the k8s-snap is located under:
> 
>     /var/snap/k8s/common/var/lib/k8s-dqlite
> 
> Related Finding V-242445 contains directives on the ownership of the files.
> 


#### Remediation
Ensure all of the following paths have correct permissions by running: 



    chmod 644 /var/snap/k8s/common/var/lib/k8s-dqlite /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml /var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock

#### Auditing (as root)

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite' | grep -q 700 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite: 700 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite: 700
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml: 600
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml: 600
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key: 600
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt: 600
stat -c %a '/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock' | grep -q 600 && echo PASS /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600 || echo FAIL /var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

### [V-242460](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242460): The Kubernetes admin kubeconfig must have file permissions set to 644 or more restrictive.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes admin kubeconfig files contain the arguments and settings for the Control Plane services. These services are controller and scheduler. If these files can be changed, the scheduler will be implementing the changes immediately.





#### Comments:
> Note that the original Finding references 'controller-manager.conf',
> but the k8s-snap uses 'controller.conf'.
> 
> Finding V-242446 defines the ownership checks for these files.
> 


#### Remediation
Ensure all of the following paths have correct permissions by running: 



    chmod -R 644 /etc/kubernetes/admin.conf /etc/kubernetes/scheduler.conf /etc/kubernetes/controller.conf

#### Auditing (as root)

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/etc/kubernetes/admin.conf' | grep -q 600 && echo PASS /etc/kubernetes/admin.conf: 600 || echo FAIL /etc/kubernetes/admin.conf: 600
stat -c %a '/etc/kubernetes/scheduler.conf' | grep -q 600 && echo PASS /etc/kubernetes/scheduler.conf: 600 || echo FAIL /etc/kubernetes/scheduler.conf: 600
stat -c %a '/etc/kubernetes/controller.conf' | grep -q 600 && echo PASS /etc/kubernetes/controller.conf: 600 || echo FAIL /etc/kubernetes/controller.conf: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

### [V-242461](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242461): Kubernetes API Server audit logs must be enabled.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes API Server validates and configures pods and services for the API object. The REST operation provides frontend functionality to the cluster share state. Enabling audit logs provides a way to monitor and identify security risk events or misuse of information. Audit logs are necessary to provide evidence in the case the Kubernetes API Server is compromised requiring a Cyber Security Investigation.





#### Comments:
> This Finding refers to the `--audit-policy-file` argument of the
> Kubernetes API Service.
> 
> The k8s-snap does not configure auditing by default.
> 
> Please review the below hardening document for a full guide on how
> to enable auditing for the kube-apiserver.
> 
> <!-- TODO: switch link to dedicated DISA Hardening page when published. -->
> https://documentation.ubuntu.com/canonical-kubernetes/latest/src/snap/howto/cis-hardening/#configure-auditing
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--audit-policy-file` for service `kube-apiserver` as appropriate.

Ensure it is set to any explicit value.

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--audit-policy-file` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-audit-policy-file=(.*)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242462](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242462): The Kubernetes API Server must be set to audit log max size.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes API Server must be set for enough storage to retain log information over the period required. When audit logs are large in size, the monitoring service for events becomes degraded. The function of the maximum log file size is to set these limits.





#### Comments:
> This Finding refers to the `--audit-log-maxsize` argument of the
> Kubernetes API Service.
> 
> The k8s-snap does not configure auditing by default.
> 
> Please review the below hardening document for a full guide on how
> to enable auditing for the kube-apiserver.
> 
> <!-- TODO: switch link to dedicated DISA Hardening page when published. -->
> https://documentation.ubuntu.com/canonical-kubernetes/latest/src/snap/howto/cis-hardening/#configure-auditing
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--audit-log-maxsize` for service `kube-apiserver` as appropriate.

Ensure it is set to: `\d{3,}`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--audit-log-maxsize` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-audit-log-maxsize=(\d{3,})' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242463](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242463): The Kubernetes API Server must be set to audit log maximum backup.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes API Server must set enough storage to retain logs for monitoring suspicious activity and system misconfiguration, and provide evidence for Cyber Security Investigations.





#### Comments:
> This Finding refers to the `--audit-log-maxbackup` argument of the
> Kubernetes API Service.
> 
> The k8s-snap does not configure auditing by default.
> 
> Please review the below hardening document for a full guide on how
> to enable auditing for the kube-apiserver.
> 
> <!-- TODO: switch link to dedicated DISA Hardening page when published. -->
> https://documentation.ubuntu.com/canonical-kubernetes/latest/src/snap/howto/cis-hardening/#configure-auditing
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--audit-log-maxbackup` for service `kube-apiserver` as appropriate.

Ensure it is set to: `\d{2,}`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--audit-log-maxbackup` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-audit-log-maxbackup=(\d{2,})' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242464](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242464): The Kubernetes API Server audit log retention must be set.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes API Server must set enough storage to retain logs for monitoring suspicious activity and system misconfiguration, and provide evidence for Cyber Security Investigations.





#### Comments:
> This Finding refers to the `--audit-log-maxage` argument of the
> Kubernetes API Service.
> 
> The k8s-snap does not configure auditing by default.
> 
> Please review the below hardening document for a full guide on how
> to enable auditing for the kube-apiserver.
> 
> <!-- TODO: switch link to dedicated DISA Hardening page when published. -->
> https://documentation.ubuntu.com/canonical-kubernetes/latest/src/snap/howto/cis-hardening/#configure-auditing
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--audit-log-maxage` for service `kube-apiserver` as appropriate.

Ensure it is set to: `\d{2,}`

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--audit-log-maxage` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-audit-log-maxage=(\d{2,})' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242465](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242465): The Kubernetes API Server audit log path must be set.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Kubernetes API Server validates and configures pods and services for the API object. The REST operation provides frontend functionality to the cluster share state. Audit logs are necessary to provide evidence in the case the Kubernetes API Server is compromised requiring Cyber Security Investigation. To record events in the audit log the log path value must be set.





#### Comments:
> This Finding refers to the `--audit-log-path` argument of the
> Kubernetes API Service.
> 
> The k8s-snap does not configure auditing by default.
> 
> Please review the below hardening document for a full guide on how
> to enable auditing for the kube-apiserver.
> 
> <!-- TODO: switch link to dedicated DISA Hardening page when published. -->
> https://documentation.ubuntu.com/canonical-kubernetes/latest/src/snap/howto/cis-hardening/#configure-auditing
> 
> This Finding is basically a duplicate of V-242402.
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kube-apiserver` in order to set the argument `--audit-log-path` for service `kube-apiserver` as appropriate.

Ensure it is set to any explicit value.

Afterwards restart the `kube-apiserver` service with:



    sudo systemctl restart snap.k8s.kube-apiserver



#### Auditing (as root)

Ensure that the argument `--audit-log-path` for service `kube-apiserver` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kube-apiserver`.

```bash
grep -E -q  '\-\-audit-log-path=(.*)' '/var/snap/k8s/common/args/kube-apiserver'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

</details>

### [V-242466](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242466): The Kubernetes PKI CRT must have file permissions set to 644 or more restrictive.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes PKI directory contains all certificates (.crt files) supporting secure network communications in the Kubernetes Control Plane. If these files can be modified, data traversing within the architecture components would become unsecure and compromised.





#### Comments:
> Finding stipulates that permission mask of all the '*.crt' files
> should be at most 644, but they can also be set to be more restrictive.
> 
> Finding V-242467 stipulates the permissions of the '*.key' files.
> 


#### Remediation
Ensure all of the following paths have correct permissions by running: 



    chmod -R 644 /etc/kubernetes/pki/apiserver-kubelet-client.crt /etc/kubernetes/pki/ca.crt /etc/kubernetes/pki/front-proxy-ca.crt /etc/kubernetes/pki/kubelet.crt /etc/kubernetes/pki/apiserver.crt /etc/kubernetes/pki/client-ca.crt /etc/kubernetes/pki/front-proxy-client.crt

#### Auditing (as root)

Ensure all required files have permissions '644' (or stricter):

```bash
stat -c %a '/etc/kubernetes/pki/apiserver-kubelet-client.crt' | grep -q 600 && echo PASS /etc/kubernetes/pki/apiserver-kubelet-client.crt: 600 || echo FAIL /etc/kubernetes/pki/apiserver-kubelet-client.crt: 600
stat -c %a '/etc/kubernetes/pki/ca.crt' | grep -q 600 && echo PASS /etc/kubernetes/pki/ca.crt: 600 || echo FAIL /etc/kubernetes/pki/ca.crt: 600
stat -c %a '/etc/kubernetes/pki/front-proxy-ca.crt' | grep -q 600 && echo PASS /etc/kubernetes/pki/front-proxy-ca.crt: 600 || echo FAIL /etc/kubernetes/pki/front-proxy-ca.crt: 600
stat -c %a '/etc/kubernetes/pki/kubelet.crt' | grep -q 600 && echo PASS /etc/kubernetes/pki/kubelet.crt: 600 || echo FAIL /etc/kubernetes/pki/kubelet.crt: 600
stat -c %a '/etc/kubernetes/pki/apiserver.crt' | grep -q 600 && echo PASS /etc/kubernetes/pki/apiserver.crt: 600 || echo FAIL /etc/kubernetes/pki/apiserver.crt: 600
stat -c %a '/etc/kubernetes/pki/client-ca.crt' | grep -q 600 && echo PASS /etc/kubernetes/pki/client-ca.crt: 600 || echo FAIL /etc/kubernetes/pki/client-ca.crt: 600
stat -c %a '/etc/kubernetes/pki/front-proxy-client.crt' | grep -q 600 && echo PASS /etc/kubernetes/pki/front-proxy-client.crt: 600 || echo FAIL /etc/kubernetes/pki/front-proxy-client.crt: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.


### [V-242467](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-242467): The Kubernetes PKI keys must have file permissions set to 600 or more restrictive.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> The Kubernetes PKI directory contains all certificate key files supporting secure network communications in the Kubernetes Control Plane. If these files can be modified, data traversing within the architecture components would become unsecure and compromised.





#### Comments:
> Finding stipulates that permission mask of all the '*.key' files
> should be 600.
> 
> Finding V-242467 stipulates the permissions of the '*.crt' files.
> 


#### Remediation
Ensure all of the following paths have correct permissions by running: 



    chmod -R 600 /etc/kubernetes/pki/apiserver-kubelet-client.key /etc/kubernetes/pki/ca.key /etc/kubernetes/pki/front-proxy-client.key /etc/kubernetes/pki/serviceaccount.key /etc/kubernetes/pki/apiserver.key /etc/kubernetes/pki/front-proxy-ca.key /etc/kubernetes/pki/kubelet.key

#### Auditing (as root)

Ensure all required files have permissions '600' (or stricter):

```bash
stat -c %a '/etc/kubernetes/pki/apiserver-kubelet-client.key' | grep -q 600 && echo PASS /etc/kubernetes/pki/apiserver-kubelet-client.key: 600 || echo FAIL /etc/kubernetes/pki/apiserver-kubelet-client.key: 600
stat -c %a '/etc/kubernetes/pki/ca.key' | grep -q 600 && echo PASS /etc/kubernetes/pki/ca.key: 600 || echo FAIL /etc/kubernetes/pki/ca.key: 600
stat -c %a '/etc/kubernetes/pki/front-proxy-client.key' | grep -q 600 && echo PASS /etc/kubernetes/pki/front-proxy-client.key: 600 || echo FAIL /etc/kubernetes/pki/front-proxy-client.key: 600
stat -c %a '/etc/kubernetes/pki/serviceaccount.key' | grep -q 600 && echo PASS /etc/kubernetes/pki/serviceaccount.key: 600 || echo FAIL /etc/kubernetes/pki/serviceaccount.key: 600
stat -c %a '/etc/kubernetes/pki/apiserver.key' | grep -q 600 && echo PASS /etc/kubernetes/pki/apiserver.key: 600 || echo FAIL /etc/kubernetes/pki/apiserver.key: 600
stat -c %a '/etc/kubernetes/pki/front-proxy-ca.key' | grep -q 600 && echo PASS /etc/kubernetes/pki/front-proxy-ca.key: 600 || echo FAIL /etc/kubernetes/pki/front-proxy-ca.key: 600
stat -c %a '/etc/kubernetes/pki/kubelet.key' | grep -q 600 && echo PASS /etc/kubernetes/pki/kubelet.key: 600 || echo FAIL /etc/kubernetes/pki/kubelet.key: 600
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `PASS`.

</details>

### [V-245541](https://www.stigviewer.com/stig/kubernetes/2024-06-10/finding/V-245541): Kubernetes Kubelet must not disable timeouts.

#### Severity: Medium

#### Status: Automated

#### Upstream Finding Description:

> Idle connections from the Kubelet can be used by unauthorized users to perform malicious activity to the nodes, pods, containers, and cluster within the Kubernetes Control Plane. Setting the streamingConnectionIdleTimeout defines the maximum time an idle session is permitted prior to disconnect. Setting the value to "0" never disconnects any idle sessions. Idle timeouts must never be set to "0" and should be defined at "5m" (the default is 4hr).





#### Comments:
> The k8s-snap does not pass a `--config` command line argument
> to the Kubelet service, nor does it explicitly pass
> `--streaming-connection-idle-timeout=5m` as a command line argument.
> 
> The command line arguments of Kubelet in the k8s-snap
> are defined in the following file:
> 
>     /var/snap/k8s/common/args/kubelet
> 


#### Remediation
Edit `/var/snap/k8s/common/args/kubelet` in order to set the argument `--streaming-connection-idle-timeout` for service `kubelet` as appropriate.

Ensure it is set to: `5m`

Afterwards restart the `kubelet` service with:



    sudo systemctl restart snap.k8s.kubelet



#### Auditing (as root)

Ensure that the argument `--streaming-connection-idle-timeout` for service `kubelet` is set as appropriate in the service's argument file `/var/snap/k8s/common/args/kubelet`.

```bash
grep -E -q  '\-\-streaming-connection-idle-timeout=(5m)' '/var/snap/k8s/common/args/kubelet'
test $? -eq 0 && echo PASS || echo FAIL
```

In the default configuration of the `k8s-snap`, resulting output lines will start with `FAIL`.

The final line of the output will be `FAIL`.

<!-- Links -->
[Hardening]:security/hardening.md
[Center for Internet Security (CIS)]:https://www.cisecurity.org/
[kube-bench]:https://aquasecurity.github.io/kube-bench/v0.6.15/
[CIS Kubernetes Benchmark]:https://www.cisecurity.org/benchmark/kubernetes
[getting started]: ../../tutorial/getting-started
[kube-bench release]: https://github.com/aquasecurity/kube-bench/releases
[Comprehensive Hardening Checklist]: auditing-steps.md#comprehensive-hardening-checklist
[upstream instructions]:https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[post-deployment hardening]: hardening.md
