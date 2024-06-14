
## Control Plane Security Configuration

### Control Plane Node Configuration Files

Control **1.1.1**
Description: `Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c permissions=%a /var/snap/k8s/common/args/kube-apiserver; fi'
```

Remediation:
```
Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated)
```

Control **1.1.2**
Description: `Ensure that the API server pod specification file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c %U:%G /var/snap/k8s/common/args/kube-apiserver; fi'
```

Remediation:
```
Ensure that the API server pod specification file ownership is set to root:root (Automated)
```

Control **1.1.3**
Description: `Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c permissions=%a /var/snap/k8s/common/args/kube-controller-manager; fi'
```

Remediation:
```
Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated)
```

Control **1.1.4**
Description: `Ensure that the controller manager pod specification file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c %U:%G /var/snap/k8s/common/args/kube-controller-manager; fi'
```

Remediation:
```
Ensure that the controller manager pod specification file ownership is set to root:root (Automated)
```

Control **1.1.5**
Description: `Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c permissions=%a /var/snap/k8s/common/args/kube-scheduler; fi'
```

Remediation:
```
Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated)
```

Control **1.1.6**
Description: `Ensure that the scheduler pod specification file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c %U:%G /var/snap/k8s/common/args/kube-scheduler; fi'
```

Remediation:
```
Ensure that the scheduler pod specification file ownership is set to root:root (Automated)
```

Control **1.1.7**
Description: `Ensure that the etcd pod specification file permissions are set to 644 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/default/etcd; then find /etc/default/etcd -name '*etcd*' | xargs stat -c permissions=%a; fi'
```

Remediation:
```
Ensure that the etcd pod specification file permissions are set to 644 or more restrictive (Automated)
```

Control **1.1.8**
Description: `Ensure that the etcd pod specification file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/default/etcd; then find /etc/default/etcd -name '*etcd*' | xargs stat -c %U:%G; fi'
```

Remediation:
```
Ensure that the etcd pod specification file ownership is set to root:root (Automated)
```

Control **1.1.9**
Description: `Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual)`

Audit:
```
ps -ef | grep kubelet | grep -- --cni-conf-dir | sed 's%.*cni-conf-dir[= ]\([^ ]*\).*%\1%' | xargs -I{} find {} -mindepth 1 | xargs --no-run-if-empty stat -c permissions=%a
find /etc/cni/net.d/10-calico.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c permissions=%a

```

Remediation:
```
Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual)
```

Control **1.1.10**
Description: `Ensure that the Container Network Interface file ownership is set to root:root (Manual)`

Audit:
```
ps -ef | grep kubelet | grep -- --cni-conf-dir | sed 's%.*cni-conf-dir[= ]\([^ ]*\).*%\1%' | xargs -I{} find {} -mindepth 1 | xargs --no-run-if-empty stat -c %U:%G
find /etc/cni/net.d/10-calico.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c %U:%G

```

Remediation:
```
Ensure that the Container Network Interface file ownership is set to root:root (Manual)
```

Control **1.1.11**
Description: `Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated)`

Audit:
```
DATA_DIR='/var/lib/etcd'
stat -c permissions=%a "/etc/default/etcd"

```

Remediation:
```
Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated)
```

Control **1.1.12**
Description: `Ensure that the etcd data directory ownership is set to etcd:etcd (Automated)`

Audit:
```
DATA_DIR='/var/lib/etcd'
stat -c %U:%G "/etc/default/etcd"

```

Remediation:
```
Ensure that the etcd data directory ownership is set to etcd:etcd (Automated)
```

Control **1.1.13**
Description: `Ensure that the admin.conf file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c permissions=%a /etc/kubernetes/admin.conf; fi'
```

Remediation:
```
Ensure that the admin.conf file permissions are set to 600 or more restrictive (Automated)
```

Control **1.1.14**
Description: `Ensure that the admin.conf file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c %U:%G /etc/kubernetes/admin.conf; fi'
```

Remediation:
```
Ensure that the admin.conf file ownership is set to root:root (Automated)
```

Control **1.1.15**
Description: `Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c permissions=%a /etc/kubernetes/scheduler.conf; fi'
```

Remediation:
```
Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated)
```

Control **1.1.16**
Description: `Ensure that the scheduler.conf file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c %U:%G /etc/kubernetes/scheduler.conf; fi'
```

Remediation:
```
Ensure that the scheduler.conf file ownership is set to root:root (Automated)
```

Control **1.1.17**
Description: `Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c permissions=%a /etc/kubernetes/controller.conf; fi'
```

Remediation:
```
Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated)
```

Control **1.1.18**
Description: `Ensure that the controller-manager.conf file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c %U:%G /etc/kubernetes/controller.conf; fi'
```

Remediation:
```
Ensure that the controller-manager.conf file ownership is set to root:root (Automated)
```

Control **1.1.19**
Description: `Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)`

Audit:
```
find /etc/kubernetes/pki/ | xargs stat -c %U:%G
```

Remediation:
```
Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)
```

Control **1.1.20**
Description: `Ensure that the Kubernetes PKI certificate file permissions are set to 600 or more restrictive (Manual)`

Audit:
```
find /etc/kubernetes/pki/ -name '*.crt' | xargs stat -c permissions=%a
```

Remediation:
```
Ensure that the Kubernetes PKI certificate file permissions are set to 600 or more restrictive (Manual)
```

Control **1.1.21**
Description: `Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual)`

Audit:
```
find /etc/kubernetes/pki/ -name '*.key' | xargs stat -c permissions=%a
```

Remediation:
```
Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual)
```


### API Server

Control **1.2.1**
Description: `Ensure that the --anonymous-auth argument is set to false (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --anonymous-auth argument is set to false (Manual)
```

Control **1.2.2**
Description: `Ensure that the --token-auth-file parameter is not set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --token-auth-file parameter is not set (Automated)
```

Control **1.2.3**
Description: `Ensure that the --DenyServiceExternalIPs is not set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --DenyServiceExternalIPs is not set (Automated)
```

Control **1.2.4**
Description: `Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated)
```

Control **1.2.5**
Description: `Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated)
```

Control **1.2.6**
Description: `Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)
```

Control **1.2.7**
Description: `Ensure that the --authorization-mode argument includes Node (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --authorization-mode argument includes Node (Automated)
```

Control **1.2.8**
Description: `Ensure that the --authorization-mode argument includes RBAC (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --authorization-mode argument includes RBAC (Automated)
```

Control **1.2.9**
Description: `Ensure that the admission control plugin EventRateLimit is set (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the admission control plugin EventRateLimit is set (Manual)
```

Control **1.2.10**
Description: `Ensure that the admission control plugin AlwaysAdmit is not set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the admission control plugin AlwaysAdmit is not set (Automated)
```

Control **1.2.11**
Description: `Ensure that the admission control plugin AlwaysPullImages is set (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the admission control plugin AlwaysPullImages is set (Manual)
```

Control **1.2.12**
Description: `Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual)
```

Control **1.2.13**
Description: `Ensure that the admission control plugin ServiceAccount is set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the admission control plugin ServiceAccount is set (Automated)
```

Control **1.2.14**
Description: `Ensure that the admission control plugin NamespaceLifecycle is set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the admission control plugin NamespaceLifecycle is set (Automated)
```

Control **1.2.15**
Description: `Ensure that the admission control plugin NodeRestriction is set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the admission control plugin NodeRestriction is set (Automated)
```

Control **1.2.16**
Description: `Ensure that the --secure-port argument is not set to 0 (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --secure-port argument is not set to 0 (Automated)
```

Control **1.2.17**
Description: `Ensure that the --profiling argument is set to false (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --profiling argument is set to false (Automated)
```

Control **1.2.18**
Description: `Ensure that the --audit-log-path argument is set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --audit-log-path argument is set (Automated)
```

Control **1.2.19**
Description: `Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated)
```

Control **1.2.20**
Description: `Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated)
```

Control **1.2.21**
Description: `Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated)
```

Control **1.2.22**
Description: `Ensure that the --request-timeout argument is set as appropriate (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --request-timeout argument is set as appropriate (Manual)
```

Control **1.2.23**
Description: `Ensure that the --service-account-lookup argument is set to true (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --service-account-lookup argument is set to true (Automated)
```

Control **1.2.24**
Description: `Ensure that the --service-account-key-file argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --service-account-key-file argument is set as appropriate (Automated)
```

Control **1.2.25**
Description: `Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated)
```

Control **1.2.26**
Description: `Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated)
```

Control **1.2.27**
Description: `Ensure that the --client-ca-file argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --client-ca-file argument is set as appropriate (Automated)
```

Control **1.2.28**
Description: `Ensure that the --etcd-cafile argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --etcd-cafile argument is set as appropriate (Automated)
```

Control **1.2.29**
Description: `Ensure that the --encryption-provider-config argument is set as appropriate (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the --encryption-provider-config argument is set as appropriate (Manual)
```

Control **1.2.30**
Description: `Ensure that encryption providers are appropriately configured (Manual)`

Audit:
```
ENCRYPTION_PROVIDER_CONFIG=$(ps -ef | grep kube-apiserver | grep -- --encryption-provider-config | sed 's%.*encryption-provider-config[= ]\([^ ]*\).*%\1%')
if test -e $ENCRYPTION_PROVIDER_CONFIG; then grep -A1 'providers:' $ENCRYPTION_PROVIDER_CONFIG | tail -n1 | grep -o "[A-Za-z]*" | sed 's/^/provider=/'; fi

```

Remediation:
```
Ensure that encryption providers are appropriately configured (Manual)
```

Control **1.2.31**
Description: `Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual)
```


### Controller Manager

Control **1.3.1**
Description: `Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual)
```

Control **1.3.2**
Description: `Ensure that the --profiling argument is set to false (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Ensure that the --profiling argument is set to false (Automated)
```

Control **1.3.3**
Description: `Ensure that the --use-service-account-credentials argument is set to true (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Ensure that the --use-service-account-credentials argument is set to true (Automated)
```

Control **1.3.4**
Description: `Ensure that the --service-account-private-key-file argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Ensure that the --service-account-private-key-file argument is set as appropriate (Automated)
```

Control **1.3.5**
Description: `Ensure that the --root-ca-file argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Ensure that the --root-ca-file argument is set as appropriate (Automated)
```

Control **1.3.6**
Description: `Ensure that the RotateKubeletServerCertificate argument is set to true (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Ensure that the RotateKubeletServerCertificate argument is set to true (Automated)
```

Control **1.3.7**
Description: `Ensure that the --bind-address argument is set to 127.0.0.1 (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Ensure that the --bind-address argument is set to 127.0.0.1 (Automated)
```


### Scheduler

Control **1.4.1**
Description: `Ensure that the --profiling argument is set to false (Automated)`

Audit:
```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

Remediation:
```
Ensure that the --profiling argument is set to false (Automated)
```

Control **1.4.2**
Description: `Ensure that the --bind-address argument is set to 127.0.0.1 (Automated)`

Audit:
```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

Remediation:
```
Ensure that the --bind-address argument is set to 127.0.0.1 (Automated)
```

