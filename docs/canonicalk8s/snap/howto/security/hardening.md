---
myst:
  html_meta:
    description: "How to harden a Canonical Kubernetes cluster with post-deployment security steps aligned with CIS and DISA STIG benchmarks."
---

# How to harden your {{product}} cluster

<!-- SPREAD SUITE: snap_bootstrapped -->

The {{product}} hardening guide provides actionable steps to enhance the
security posture of your deployment. These steps are designed to help you align
with industry-standard frameworks such as CIS.

{{product}} aligns with many security recommendations by
default. However, since implementing all security recommendations
would come at the expense of compatibility and/or performance we expect
cluster administrators to follow post deployment hardening steps based on their
needs. Please evaluate the implications of each configuration before applying
it.

```{note}
For DISA STIG or FIPS 140-3 compliance, please see our [how to set up a FIPS
compliant Kubernetes cluster] and [how to install Canonical Kubernetes with
DISA STIG hardening] pages before continuing as they have more stringent
security recommendations that must be done at install.
```

<!-- Charm start here -->

## Platform hardening recommendations

These steps are common to the hardening process for not only CIS and DISA STIG
compliance, but also good suggestions if one is willing to incur the performance
costs for the benefit of an increased security posture.

### Control plane nodes

#### Encrypt secrets at rest

Encrypt key-value store secrets rather than leaving it as base64 encoded values
as described in the upstream Kubernetes documentation on
[encrypting secrets](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/).

Create the `EncryptionConfiguration` file under
`/var/snap/k8s/common/etc/encryption/`.

<!-- SPREAD
export BASE_64_ENCODED_SECRET=$(head -c 32 /dev/urandom | base64)
-->

```
sudo mkdir -p /var/snap/k8s/common/etc/encryption/
cat << EOL | sudo tee /var/snap/k8s/common/etc/encryption/enc.yaml > /dev/null
kind: "EncryptionConfiguration"
apiVersion: apiserver.config.k8s.io/v1
resources:
- resources: ["secrets"]
  providers:
  - aesgcm:
      keys:
      - name: key1
        secret: ${BASE_64_ENCODED_SECRET}
  - identity: {}
EOL
sudo chmod 600 /var/snap/k8s/common/etc/encryption/enc.yaml
```

<!-- SPREAD
source ${SPREAD_PATH}/docs/tools/repeat_checks.sh
sudo test -s /var/snap/k8s/common/etc/encryption/enc.yaml
sudo stat -c '%a' /var/snap/k8s/common/etc/encryption/enc.yaml | grep "600"
-->

Set the `--encryption-provider-config` file as an argument to the kubernetes
apiserver.

```
sudo sh -c '
cat >>/var/snap/k8s/common/args/kube-apiserver <<EOL
--encryption-provider-config=/var/snap/k8s/common/etc/encryption/enc.yaml
EOL'
```

<!-- SPREAD
sudo cat /var/snap/k8s/common/args/kube-apiserver | grep -e "--encryption-provider-config=/var/snap/k8s/common/etc/encryption/enc.yaml"
-->

Securing the contents of this key file is left as a separate exercise.

#### Configure authorization modes

Enforce RBAC (Role-Based Access Control) policies and confirm the value of the
apiserver [`authorization-mode`](https://kubernetes.io/docs/reference/access-authn-authz/authorization/#authorization-modules):

* includes `RBAC`
* doesn't include `AlwaysAllow`

```
sudo grep authorization-mode /var/snap/k8s/common/args/kube-apiserver | \
    grep -q "RBAC" && echo "okay" || echo "missing"
sudo grep authorization-mode /var/snap/k8s/common/args/kube-apiserver | \
    grep -q "AlwaysAllow" && echo "Remove AlwaysAllow" || echo "okay"
```

<!-- SPREAD
sudo grep authorization-mode /var/snap/k8s/common/args/kube-apiserver | grep -q "RBAC" 
! sudo grep -qE "authorization-mode.*AlwaysAllow" /var/snap/k8s/common/args/kube-apiserver
-->

By default, the value is `Node,RBAC`

* `Node`:
   A special-purpose authorization mode that grants permissions
   to kubelets based on the pods they are scheduled to run.

To apply RBAC to other cluster resources, see the upstream Kubernetes
[RBAC guide](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

#### Configure log auditing

```{note}
Configuring log auditing requires the cluster administrator's input and
may incur performance penalties in the form of disk I/O.
```

Create an audit-policy.yaml file under `/var/snap/k8s/common/etc/` and specify
the level of auditing you desire based on the [upstream instructions](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/).
Here is a minimal example of such a policy file.

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

<!-- SPREAD 
sudo test -s /var/snap/k8s/common/etc/audit-policy.yaml
-->

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

<!-- SPREAD
sudo cat /var/snap/k8s/common/args/kube-apiserver | grep -e "--audit-log-path=/var/log/kubernetes/audit.log"
-->

Restart the API server:

```
sudo systemctl restart snap.k8s.kube-apiserver
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get nodes" "Ready"
sudo test -f /var/log/kubernetes/audit.log
-->

#### Set event rate limits

```{note}
Configuring event rate limits requires the cluster administrator's input
in assessing the hardware and workload specifications/requirements.
```

Create a configuration file with the [rate limits](https://kubernetes.io/docs/reference/config-api/apiserver-eventratelimit.v1alpha1) and place it under
`/var/snap/k8s/common/etc/`.
For example:

```
sudo mkdir -p /var/snap/k8s/common/etc/
sudo sh -c 'cat >/var/snap/k8s/common/etc/eventconfig.yaml <<EOL
apiVersion: eventratelimit.admission.k8s.io/v1alpha1
kind: Configuration
limits:
  - type: Server
    qps: 5000
    burst: 20000
EOL'
```

<!-- SPREAD
sudo test -s /var/snap/k8s/common/etc/eventconfig.yaml
-->

Create an admissions control config file under `/var/k8s/snap/common/etc/` .

```
sudo mkdir -p /var/snap/k8s/common/etc/
sudo sh -c 'cat >/var/snap/k8s/common/etc/admission-control-config-file.yaml <<EOL
apiVersion: apiserver.config.k8s.io/v1
kind: AdmissionConfiguration
plugins:
  - name: EventRateLimit
    path: eventconfig.yaml
EOL'
```

<!-- SPREAD
sudo test -s /var/snap/k8s/common/etc/admission-control-config-file.yaml
-->

Make sure the EventRateLimit admission plugin is loaded in the
`/var/snap/k8s/common/args/kube-apiserver` .

<!-- SPREAD SKIP -->

```
--enable-admission-plugins=...,EventRateLimit,...
```

<!-- SPREAD SKIP END -->

<!-- SPREAD 
sudo sed -i 's/\(--enable-admission-plugins="[^"]*\)"/\1,EventRateLimit"/' /var/snap/k8s/common/args/kube-apiserver
sudo cat /var/snap/k8s/common/args/kube-apiserver | grep "enable-admission-plugins" | grep "EventRateLimit"
-->

Load the admission control config file.

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-apiserver <<EOL
--admission-control-config-file=/var/snap/k8s/common/etc/admission-control-config-file.yaml
EOL'
```

<!-- SPREAD
sudo cat /var/snap/k8s/common/args/kube-apiserver | grep -e "--admission-control-config-file=/var/snap/k8s/common/etc/admission-control-config-file.yaml"
-->

Restart the API server.

```
sudo systemctl restart snap.k8s.kube-apiserver
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get nodes" "Ready"
-->

#### Enable AlwaysPullImages admission control plugin

```{note}
Configuring the AlwaysPullImages admission control plugin may have performance
impact in the form of increased network traffic and may hamper offline deployments
that use image sideloading.
```

Make sure the AlwaysPullImages admission plugin is loaded in the
`/var/snap/k8s/common/args/kube-apiserver`

<!-- SPREAD SKIP -->

```
--enable-admission-plugins=...,AlwaysPullImages,...
```

<!-- SPREAD SKIP END -->

<!-- SPREAD
sudo sed -i 's/\(--enable-admission-plugins="[^"]*\)"/\1,AlwaysPullImages"/' /var/snap/k8s/common/args/kube-apiserver
sudo cat /var/snap/k8s/common/args/kube-apiserver | grep "enable-admission-plugins"| grep "AlwaysPullImages"
-->

Restart the API server.

```
sudo systemctl restart snap.k8s.kube-apiserver
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get nodes" "Ready"
-->

#### Set the Kubernetes scheduler and controller manager bind address

```{note}
This configuration may affect compatibility with workloads and metrics
collection.
```

Edit the Kubernetes scheduler arguments file
`/var/snap/k8s/common/args/kube-scheduler`
and set the `--bind-address` to be `127.0.0.1`.

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-scheduler <<EOL
--bind-address=127.0.0.1
EOL'
```

<!-- SPREAD
sudo cat /var/snap/k8s/common/args/kube-scheduler | grep -e "--bind-address=127.0.0.1"
-->

Do the same for the Kubernetes controller manager
(`/var/snap/k8s/common/args/kube-controller-manager`):

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-controller-manager <<EOL
--bind-address=127.0.0.1
EOL'
```

<!-- SPREAD
sudo cat /var/snap/k8s/common/args/kube-controller-manager | grep -e "--bind-address=127.0.0.1"
-->

Restart both services.

```
sudo systemctl restart snap.k8s.kube-scheduler
sudo systemctl restart snap.k8s.kube-controller-manager
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get nodes" "Ready"
-->

### Worker nodes

Run the following commands on nodes that host workloads. In the default
deployment the control plane nodes functions as workers and they may need
to be hardened.

#### Protect kernel defaults

```{note}
This configuration may affect compatibility of workloads.
```

Kubelet will not start if it finds kernel configurations incompatible with its
defaults.

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kubelet <<EOL
--protect-kernel-defaults=true
EOL'
```

<!-- SPREAD
sudo cat /var/snap/k8s/common/args/kubelet | grep -e "--protect-kernel-defaults=true"
-->

Restart `kubelet`.

```
sudo systemctl restart snap.k8s.kubelet
```

Reload the system daemons:

```
sudo systemctl daemon-reload
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get nodes" "Ready"
-->

#### Edit kubelet service file permissions

```{note}
Fully complying with the spirit of this hardening recommendation calls for
systemd configuration that is out of the scope of this documentation page.
```

Ensure that only the owner of `/etc/systemd/system/snap.k8s.kubelet.service`
has full read and write access to it. Setting the kubelet service file
permission needs to be performed every time the k8s snap refreshes.

```
chmod 600 /etc/systemd/system/snap.k8s.kubelet.service
```

<!-- SPREAD
stat -c '%a' /etc/systemd/system/snap.k8s.kubelet.service | grep "600"
-->

Restart `kubelet`.

```
sudo systemctl restart snap.k8s.kubelet
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get nodes" "Ready"
-->

#### Set the maximum time an idle session is permitted prior to disconnect

Idle connections from the Kubelet can be used by unauthorized users to
perform malicious activity to the nodes, pods, containers, and cluster within
the Kubernetes Control Plane.

Edit `/var/snap/k8s/common/args/kubelet` and set the argument
`--streaming-connection-idle-timeout` to `5m`.

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kubelet <<EOL
--streaming-connection-idle-timeout=5m
EOL'
```

<!-- SPREAD
sudo cat /var/snap/k8s/common/args/kubelet | grep -e "--streaming-connection-idle-timeout=5m"
-->

Restart `kubelet`.

```
sudo systemctl restart snap.k8s.kubelet
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get nodes" "Ready"
-->

<!-- Charm end here -->

## CIS hardening

To assess compliance to the CIS hardening guidelines, please see the [CIS
assessment page](cis-assessment.md).
