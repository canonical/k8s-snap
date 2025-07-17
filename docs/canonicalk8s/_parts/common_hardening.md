### Control plane nodes

#### Encrypt secrets at rest
Encrypt key-value store secrets rather than leaving it as base64 encoded values
as described in the upstream Kubernetes documentation on
[encrypting secrets][encryption_at_rest].

Create the `EncryptionConfiguration` file under `/var/snap/k8s/common/etc/encryption/`.

```
sudo sh -c '
mkdir -p /var/snap/k8s/common/etc/encryption/
cat >/var/snap/k8s/common/etc/encryption/enc.yaml << EOL
kind: "EncryptionConfig"
apiVersion: apiserver.config.k8s.io/v1
resources:
- resources: ["secrets"]
  providers:
  - aesgcm:
      keys:
      - name: key1
        secret: ${BASE 64 ENCODED SECRET}
  - identity: {}
EOL
chmod 600 /var/snap/k8s/common/etc/encryption/enc.yaml
```

Set the `--encryption-provider-config` file as an argument to the kubernetes
apiserver.

```
sudo sh -c '
cat >>/var/snap/k8s/common/args/kube-apiserver <<EOL
--encryption-provider-config=/var/snap/k8s/common/etc/enc.yaml
EOL'
```

Securing the contents of this key file is left as a separate exercise.


#### Configure authorization modes
Enforce RBAC (Role-Based Access Control) policies and confirm the value of the
apiserver [`authorization-mode`][authorization_mode]:
* includes `RBAC`
* doesn't include `AlwaysAllow`

```
sudo grep authorization-mode /var/snap/k8s/common/args/kube-apiserver | \
    grep -q "RBAC" && echo "okay" || echo "missing"
sudo grep authorization-mode /var/snap/k8s/common/args/kube-apiserver | \
    grep -q "AlwaysAllow" && echo "Remove AlwaysAllow" || echo "okay"
```

By default, the value is `Node,RBAC`
* `Node`:
  A special-purpose authorization mode that grants permissions
  to kubelets based on the pods they are scheduled to run.

 To apply RBAC to other cluster resources, see the upstream Kubernetes
 [RBAC guide][access_authn_authz].


#### Configure log auditing

```{note}
Configuring log auditing requires the cluster administrator's input and
may incur performance penalties in the form of disk I/O.
```

Create an audit-policy.yaml file under `/var/snap/k8s/common/etc/` and specify
the level of auditing you desire based on the [upstream instructions].
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

#### Set event rate limits

```{note}
Configuring event rate limits requires the cluster administrator's input
in assessing the hardware and workload specifications/requirements.
```


Create a configuration file with the [rate limits] and place it under
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

Make sure the EventRateLimit admission plugin is loaded in the
`/var/snap/k8s/common/args/kube-apiserver` .

```
--enable-admission-plugins=...,EventRateLimit,...
```

Load the admission control config file.

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-apiserver <<EOL
--admission-control-config-file=/var/snap/k8s/common/etc/admission-control-config-file.yaml
EOL'
```

Restart the API server.

```
sudo systemctl restart snap.k8s.kube-apiserver
```

#### Enable AlwaysPullImages admission control plugin

```{note}
Configuring the AlwaysPullImages admission control plugin may have performance
impact in the form of increased network traffic and may hamper offline deployments
that use image sideloading.
```

Make sure the AlwaysPullImages admission plugin is loaded in the
`/var/snap/k8s/common/args/kube-apiserver`

```
--enable-admission-plugins=...,AlwaysPullImages,...
```

Restart the API server.

```
sudo systemctl restart snap.k8s.kube-apiserver
```


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

Do the same for the Kubernetes controller manager
(`/var/snap/k8s/common/args/kube-controller-manager`):

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-controller-manager <<EOL
--bind-address=127.0.0.1
EOL'
```

Restart both services.

```
sudo systemctl restart snap.k8s.kube-scheduler
sudo systemctl restart snap.k8s.kube-controller-manager
```

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

Restart `kubelet`.

```
sudo systemctl restart snap.k8s.kubelet
```

Reload the system daemons:

```
sudo systemctl daemon-reload
```

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

Restart `kubelet`.

```
sudo systemctl restart snap.k8s.kubelet
```

#### Set the maximum time an idle session is permitted prior to disconnect

Idle connections from the Kubelet can be used by unauthorized users to
perform malicious activity to the nodes, pods, containers, and cluster within
the Kubernetes Control Plane.

Edit `/var/snap/k8s/common/args/kubelet` and set the argument `--streaming-connection-idle-timeout` to `5m`.

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kubelet <<EOL
--streaming-connection-idle-timeout=5m
EOL'
```

Restart `kubelet`.

```
sudo systemctl restart snap.k8s.kubelet
```


<!-- Links -->
[upstream instructions]:https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[rate limits]:https://kubernetes.io/docs/reference/config-api/apiserver-eventratelimit.v1alpha1
[controlling_access]: https://kubernetes.io/docs/concepts/security/controlling-access/
[access_authn_authz]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[encryption_at_rest]: https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/
[authorization_mode]: https://kubernetes.io/docs/reference/access-authn-authz/authorization/#authorization-modules
