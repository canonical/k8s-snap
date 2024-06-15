<!-- markdownlint-disable -->
## Control Plane Security Configuration
### Control Plane Node Configuration Files
#### Control 1.1.1

Description: `Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c permissions=%a /var/snap/k8s/common/args/kube-apiserver; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the
control plane node.
For example, chmod 600 /var/snap/k8s/common/args/kube-apiserver
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 1.1.2

Description: `Ensure that the API server pod specification file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c %U:%G /var/snap/k8s/common/args/kube-apiserver; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example, chown root:root /var/snap/k8s/common/args/kube-apiserver
```

Expected output:
```
test_items:
- flag: root:root
```

#### Control 1.1.3

Description: `Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c permissions=%a /var/snap/k8s/common/args/kube-controller-manager; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example, chmod 600 /var/snap/k8s/common/args/kube-controller-manager
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 1.1.4

Description: `Ensure that the controller manager pod specification file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c %U:%G /var/snap/k8s/common/args/kube-controller-manager; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example, chown root:root /var/snap/k8s/common/args/kube-controller-manager
```

Expected output:
```
test_items:
- flag: root:root
```

#### Control 1.1.5

Description: `Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c permissions=%a /var/snap/k8s/common/args/kube-scheduler; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example, chmod 600 /var/snap/k8s/common/args/kube-scheduler
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 1.1.6

Description: `Ensure that the scheduler pod specification file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c %U:%G /var/snap/k8s/common/args/kube-scheduler; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example, chown root:root /var/snap/k8s/common/args/kube-scheduler
```

Expected output:
```
test_items:
- flag: root:root
```

#### Control 1.1.7

Description: `Ensure that the etcd pod specification file permissions are set to 644 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/default/etcd; then find /etc/default/etcd -name '*etcd*' | xargs stat -c permissions=%a; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chmod 644 /etc/default/etcd
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '644'
  flag: permissions
```

#### Control 1.1.8

Description: `Ensure that the etcd pod specification file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/default/etcd; then find /etc/default/etcd -name '*etcd*' | xargs stat -c %U:%G; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chown root:root /etc/default/etcd
```

Expected output:
```
test_items:
- flag: root:root
```

#### Control 1.1.9

Description: `Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual)`

Audit:
```
ps -ef | grep kubelet | grep -- --cni-conf-dir | sed 's%.*cni-conf-dir[= ]\([^ ]*\).*%\1%' | xargs -I{} find {} -mindepth 1 | xargs --no-run-if-empty stat -c permissions=%a
find /etc/cni/net.d/10-calico.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c permissions=%a
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example, chmod 644 <path/to/cni/files>
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '644'
  flag: permissions
```

#### Control 1.1.10

Description: `Ensure that the Container Network Interface file ownership is set to root:root (Manual)`

Audit:
```
ps -ef | grep kubelet | grep -- --cni-conf-dir | sed 's%.*cni-conf-dir[= ]\([^ ]*\).*%\1%' | xargs -I{} find {} -mindepth 1 | xargs --no-run-if-empty stat -c %U:%G
find /etc/cni/net.d/10-calico.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c %U:%G
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chown root:root <path/to/cni/files>
```

Expected output:
```
test_items:
- flag: root:root
```

#### Control 1.1.11

Description: `Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated)`

Audit:
```
DATA_DIR='/var/lib/etcd'
stat -c permissions=%a "$DATA_DIR"
```

Remediation:
```
On the etcd server node, get the etcd data directory, passed as an argument --data-dir,
from the command 'ps -ef | grep etcd'.
Run the below command (based on the etcd data directory found above). For example,
chmod 700 /var/lib/etcd
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '700'
  flag: permissions
```

#### Control 1.1.12

Description: `Ensure that the etcd data directory ownership is set to etcd:etcd (Automated)`

Audit:
```
DATA_DIR='/var/lib/etcd'
stat -c %U:%G "$DATA_DIR"
```

Remediation:
```
On the etcd server node, get the etcd data directory, passed as an argument --data-dir,
from the command 'ps -ef | grep etcd'.
Run the below command (based on the etcd data directory found above).
For example, chown root:root /var/lib/etcd
```

Expected output:
```
test_items:
- flag: etcd:etcd
```

#### Control 1.1.13

Description: `Ensure that the admin.conf file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c permissions=%a /etc/kubernetes/admin.conf; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example, chmod 600 /etc/kubernetes/admin.conf
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 1.1.14

Description: `Ensure that the admin.conf file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c %U:%G /etc/kubernetes/admin.conf; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example, chown root:root /etc/kubernetes/admin.conf
```

Expected output:
```
test_items:
- flag: root:root
```

#### Control 1.1.15

Description: `Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c permissions=%a /etc/kubernetes/scheduler.conf; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chmod 600 /etc/kubernetes/scheduler.conf
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 1.1.16

Description: `Ensure that the scheduler.conf file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c %U:%G /etc/kubernetes/scheduler.conf; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chown root:root /etc/kubernetes/scheduler.conf
```

Expected output:
```
test_items:
- flag: root:root
```

#### Control 1.1.17

Description: `Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c permissions=%a /etc/kubernetes/controller.conf; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chmod 600 /etc/kubernetes/controller.conf
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 1.1.18

Description: `Ensure that the controller-manager.conf file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c %U:%G /etc/kubernetes/controller.conf; fi'
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chown root:root /etc/kubernetes/controller.conf
```

Expected output:
```
test_items:
- flag: root:root
```

#### Control 1.1.19

Description: `Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated)`

Audit:
```
find /etc/kubernetes/pki/ | xargs stat -c %U:%G
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chown -R root:root /etc/kubernetes/pki/
```

Expected output:
```
test_items:
- flag: root:root
```

#### Control 1.1.20

Description: `Ensure that the Kubernetes PKI certificate file permissions are set to 600 or more restrictive (Manual)`

Audit:
```
find /etc/kubernetes/pki/ -name '*.crt' | xargs stat -c permissions=%a
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chmod -R 600 /etc/kubernetes/pki/*.crt
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 1.1.21

Description: `Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual)`

Audit:
```
find /etc/kubernetes/pki/ -name '*.key' | xargs stat -c permissions=%a
```

Remediation:
```
Run the below command (based on the file location on your system) on the control plane node.
For example,
chmod -R 600 /etc/kubernetes/pki/*.key
```

Expected output:
```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

### API Server
#### Control 1.2.1

Description: `Ensure that the --anonymous-auth argument is set to false (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the below parameter.
--anonymous-auth=false
```

Expected output:
```
test_items:
- compare:
    op: eq
    value: false
  flag: --anonymous-auth
```

#### Control 1.2.2

Description: `Ensure that the --token-auth-file parameter is not set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the documentation and configure alternate mechanisms for authentication. Then,
edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and remove the --token-auth-file=<filename> parameter.
```

Expected output:
```
test_items:
- flag: --token-auth-file
  set: false
```

#### Control 1.2.3

Description: `Ensure that the --DenyServiceExternalIPs is not set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and remove the `DenyServiceExternalIPs`
from enabled admission plugins.
```

Expected output:
```
bin_op: or
test_items:
- compare:
    op: nothave
    value: DenyServiceExternalIPs
  flag: --enable-admission-plugins
- flag: --enable-admission-plugins
  set: false
```

#### Control 1.2.4

Description: `Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the Kubernetes documentation and set up the TLS connection between the
apiserver and kubelets. Then, edit API server pod specification file
/var/snap/k8s/common/args/kube-apiserver on the control plane node and set the
kubelet client certificate and key parameters as below.
--kubelet-client-certificate=<path/to/client-certificate-file>
--kubelet-client-key=<path/to/client-key-file>
```

Expected output:
```
bin_op: and
test_items:
- flag: --kubelet-client-certificate
- flag: --kubelet-client-key
```

#### Control 1.2.5

Description: `Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the Kubernetes documentation and setup the TLS connection between
the apiserver and kubelets. Then, edit the API server pod specification file
/var/snap/k8s/common/args/kube-apiserver on the control plane node and set the
--kubelet-certificate-authority parameter to the path to the cert file for the certificate authority.
--kubelet-certificate-authority=<ca-string>
```

Expected output:
```
test_items:
- flag: --kubelet-certificate-authority
```

#### Control 1.2.6

Description: `Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode parameter to values other than AlwaysAllow.
One such example could be as below.
--authorization-mode=RBAC
```

Expected output:
```
test_items:
- compare:
    op: nothave
    value: AlwaysAllow
  flag: --authorization-mode
```

#### Control 1.2.7

Description: `Ensure that the --authorization-mode argument includes Node (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode parameter to a value that includes Node.
--authorization-mode=Node,RBAC
```

Expected output:
```
test_items:
- compare:
    op: has
    value: Node
  flag: --authorization-mode
```

#### Control 1.2.8

Description: `Ensure that the --authorization-mode argument includes RBAC (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode parameter to a value that includes RBAC,
for example `--authorization-mode=Node,RBAC`.
```

Expected output:
```
test_items:
- compare:
    op: has
    value: RBAC
  flag: --authorization-mode
```

#### Control 1.2.9

Description: `Ensure that the admission control plugin EventRateLimit is set (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the Kubernetes documentation and set the desired limits in a configuration file.
Then, edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
and set the below parameters.
--enable-admission-plugins=...,EventRateLimit,...
--admission-control-config-file=<path/to/configuration/file>
```

Expected output:
```
test_items:
- compare:
    op: has
    value: EventRateLimit
  flag: --enable-admission-plugins
```

#### Control 1.2.10

Description: `Ensure that the admission control plugin AlwaysAdmit is not set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and either remove the --enable-admission-plugins parameter, or set it to a
value that does not include AlwaysAdmit.
```

Expected output:
```
bin_op: or
test_items:
- compare:
    op: nothave
    value: AlwaysAdmit
  flag: --enable-admission-plugins
- flag: --enable-admission-plugins
  set: false
```

#### Control 1.2.11

Description: `Ensure that the admission control plugin AlwaysPullImages is set (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-plugins parameter to include
AlwaysPullImages.
--enable-admission-plugins=...,AlwaysPullImages,...
```

Expected output:
```
test_items:
- compare:
    op: has
    value: AlwaysPullImages
  flag: --enable-admission-plugins
```

#### Control 1.2.12

Description: `Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-plugins parameter to include
SecurityContextDeny, unless PodSecurityPolicy is already in place.
--enable-admission-plugins=...,SecurityContextDeny,...
```

Expected output:
```
bin_op: or
test_items:
- compare:
    op: has
    value: SecurityContextDeny
  flag: --enable-admission-plugins
- compare:
    op: has
    value: PodSecurityPolicy
  flag: --enable-admission-plugins
```

#### Control 1.2.13

Description: `Ensure that the admission control plugin ServiceAccount is set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the documentation and create ServiceAccount objects as per your environment.
Then, edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and ensure that the --disable-admission-plugins parameter is set to a
value that does not include ServiceAccount.
```

Expected output:
```
bin_op: or
test_items:
- compare:
    op: nothave
    value: ServiceAccount
  flag: --disable-admission-plugins
- flag: --disable-admission-plugins
  set: false
```

#### Control 1.2.14

Description: `Ensure that the admission control plugin NamespaceLifecycle is set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --disable-admission-plugins parameter to
ensure it does not include NamespaceLifecycle.
```

Expected output:
```
bin_op: or
test_items:
- compare:
    op: nothave
    value: NamespaceLifecycle
  flag: --disable-admission-plugins
- flag: --disable-admission-plugins
  set: false
```

#### Control 1.2.15

Description: `Ensure that the admission control plugin NodeRestriction is set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the Kubernetes documentation and configure NodeRestriction plug-in on kubelets.
Then, edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-plugins parameter to a
value that includes NodeRestriction.
--enable-admission-plugins=...,NodeRestriction,...
```

Expected output:
```
test_items:
- compare:
    op: has
    value: NodeRestriction
  flag: --enable-admission-plugins
```

#### Control 1.2.16

Description: `Ensure that the --secure-port argument is not set to 0 (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and either remove the --secure-port parameter or
set it to a different (non-zero) desired port.
```

Expected output:
```
bin_op: or
test_items:
- compare:
    op: gt
    value: 0
  flag: --secure-port
- flag: --secure-port
  set: false
```

#### Control 1.2.17

Description: `Ensure that the --profiling argument is set to false (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the below parameter.
--profiling=false
```

Expected output:
```
test_items:
- compare:
    op: eq
    value: false
  flag: --profiling
```

#### Control 1.2.18

Description: `Ensure that the --audit-log-path argument is set (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-path parameter to a suitable path and
file where you would like audit logs to be written, for example,
--audit-log-path=/var/log/apiserver/audit.log
```

Expected output:
```
test_items:
- flag: --audit-log-path
```

#### Control 1.2.19

Description: `Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxage parameter to 30
or as an appropriate number of days, for example,
--audit-log-maxage=30
```

Expected output:
```
test_items:
- compare:
    op: gte
    value: 30
  flag: --audit-log-maxage
```

#### Control 1.2.20

Description: `Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxbackup parameter to 10 or to an appropriate
value. For example,
--audit-log-maxbackup=10
```

Expected output:
```
test_items:
- compare:
    op: gte
    value: 10
  flag: --audit-log-maxbackup
```

#### Control 1.2.21

Description: `Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxsize parameter to an appropriate size in MB.
For example, to set it as 100 MB, --audit-log-maxsize=100
```

Expected output:
```
test_items:
- compare:
    op: gte
    value: 100
  flag: --audit-log-maxsize
```

#### Control 1.2.22

Description: `Ensure that the --request-timeout argument is set as appropriate (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
and set the below parameter as appropriate and if needed.
For example, --request-timeout=300s
```

#### Control 1.2.23

Description: `Ensure that the --service-account-lookup argument is set to true (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the below parameter.
--service-account-lookup=true
Alternatively, you can delete the --service-account-lookup parameter from this file so
that the default takes effect.
```

Expected output:
```
bin_op: or
test_items:
- flag: --service-account-lookup
  set: false
- compare:
    op: eq
    value: true
  flag: --service-account-lookup
```

#### Control 1.2.24

Description: `Ensure that the --service-account-key-file argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --service-account-key-file parameter
to the public key file for service accounts. For example,
--service-account-key-file=<filename>
```

Expected output:
```
test_items:
- flag: --service-account-key-file
```

#### Control 1.2.25

Description: `Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the Kubernetes documentation and set up the TLS connection between the apiserver and etcd.
Then, edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the etcd certificate and key file parameters.
--etcd-certfile=<path/to/client-certificate-file>
--etcd-keyfile=<path/to/client-key-file>
```

Expected output:
```
bin_op: and
test_items:
- flag: --etcd-certfile
- flag: --etcd-keyfile
```

#### Control 1.2.26

Description: `Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the Kubernetes documentation and set up the TLS connection on the apiserver.
Then, edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the TLS certificate and private key file parameters.
--tls-cert-file=<path/to/tls-certificate-file>
--tls-private-key-file=<path/to/tls-key-file>
```

Expected output:
```
bin_op: and
test_items:
- flag: --tls-cert-file
- flag: --tls-private-key-file
```

#### Control 1.2.27

Description: `Ensure that the --client-ca-file argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the Kubernetes documentation and set up the TLS connection on the apiserver.
Then, edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the client certificate authority file.
--client-ca-file=<path/to/client-ca-file>
```

Expected output:
```
test_items:
- flag: --client-ca-file
```

#### Control 1.2.28

Description: `Ensure that the --etcd-cafile argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the Kubernetes documentation and set up the TLS connection between the apiserver and etcd.
Then, edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the etcd certificate authority file parameter.
--etcd-cafile=<path/to/ca-file>
```

Expected output:
```
test_items:
- flag: --etcd-cafile
```

#### Control 1.2.29

Description: `Ensure that the --encryption-provider-config argument is set as appropriate (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Follow the Kubernetes documentation and configure a EncryptionConfig file.
Then, edit the API server pod specification file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --encryption-provider-config parameter to the path of that file.
For example, --encryption-provider-config=</path/to/EncryptionConfig/File>
```

Expected output:
```
test_items:
- flag: --encryption-provider-config
```

#### Control 1.2.30

Description: `Ensure that encryption providers are appropriately configured (Manual)`

Audit:
```
ENCRYPTION_PROVIDER_CONFIG=$(ps -ef | grep kube-apiserver | grep -- --encryption-provider-config | sed 's%.*encryption-provider-config[= ]\([^ ]*\).*%\1%')
if test -e $ENCRYPTION_PROVIDER_CONFIG; then grep -A1 'providers:' $ENCRYPTION_PROVIDER_CONFIG | tail -n1 | grep -o "[A-Za-z]*" | sed 's/^/provider=/'; fi
```

Remediation:
```
Follow the Kubernetes documentation and configure a EncryptionConfig file.
In this file, choose aescbc, kms or secretbox as the encryption provider.
```

Expected output:
```
test_items:
- compare:
    op: valid_elements
    value: aescbc,kms,secretbox
  flag: provider
```

#### Control 1.2.31

Description: `Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml
on the control plane node and set the below parameter.
--tls-cipher-suites=TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,
TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_RSA_WITH_3DES_EDE_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA,
TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_256_GCM_SHA384
```

Expected output:
```
test_items:
- compare:
    op: valid_elements
    value: TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_RSA_WITH_3DES_EDE_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_256_GCM_SHA384
  flag: --tls-cipher-suites
```

### Controller Manager
#### Control 1.3.1

Description: `Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Edit the Controller Manager pod specification file /var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --terminated-pod-gc-threshold to an appropriate threshold,
for example, --terminated-pod-gc-threshold=10
```

Expected output:
```
test_items:
- flag: --terminated-pod-gc-threshold
```

#### Control 1.3.2

Description: `Ensure that the --profiling argument is set to false (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Edit the Controller Manager pod specification file /var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the below parameter.
--profiling=false
```

Expected output:
```
test_items:
- compare:
    op: eq
    value: false
  flag: --profiling
```

#### Control 1.3.3

Description: `Ensure that the --use-service-account-credentials argument is set to true (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Edit the Controller Manager pod specification file /var/snap/k8s/common/args/kube-controller-manager
on the control plane node to set the below parameter.
--use-service-account-credentials=true
```

Expected output:
```
test_items:
- compare:
    op: noteq
    value: false
  flag: --use-service-account-credentials
```

#### Control 1.3.4

Description: `Ensure that the --service-account-private-key-file argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Edit the Controller Manager pod specification file /var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --service-account-private-key-file parameter
to the private key file for service accounts.
--service-account-private-key-file=<filename>
```

Expected output:
```
test_items:
- flag: --service-account-private-key-file
```

#### Control 1.3.5

Description: `Ensure that the --root-ca-file argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Edit the Controller Manager pod specification file /var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --root-ca-file parameter to the certificate bundle file`.
--root-ca-file=<path/to/file>
```

Expected output:
```
test_items:
- flag: --root-ca-file
```

#### Control 1.3.6

Description: `Ensure that the RotateKubeletServerCertificate argument is set to true (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Edit the Controller Manager pod specification file /var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --feature-gates parameter to include RotateKubeletServerCertificate=true.
--feature-gates=RotateKubeletServerCertificate=true
```

Expected output:
```
bin_op: or
test_items:
- compare:
    op: nothave
    value: RotateKubeletServerCertificate=false
  flag: --feature-gates
  set: true
- flag: --feature-gates
  set: false
```

#### Control 1.3.7

Description: `Ensure that the --bind-address argument is set to 127.0.0.1 (Automated)`

Audit:
```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

Remediation:
```
Edit the Controller Manager pod specification file /var/snap/k8s/common/args/kube-controller-manager
on the control plane node and ensure the correct value for the --bind-address parameter
```

Expected output:
```
bin_op: or
test_items:
- compare:
    op: eq
    value: 127.0.0.1
  flag: --bind-address
- flag: --bind-address
  set: false
```

### Scheduler
#### Control 1.4.1

Description: `Ensure that the --profiling argument is set to false (Automated)`

Audit:
```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

Remediation:
```
Edit the Scheduler pod specification file /var/snap/k8s/common/args/kube-scheduler file
on the control plane node and set the below parameter.
--profiling=false
```

Expected output:
```
test_items:
- compare:
    op: eq
    value: false
  flag: --profiling
```

#### Control 1.4.2

Description: `Ensure that the --bind-address argument is set to 127.0.0.1 (Automated)`

Audit:
```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

Remediation:
```
Edit the Scheduler pod specification file /var/snap/k8s/common/args/kube-scheduler
on the control plane node and ensure the correct value for the --bind-address parameter
```

Expected output:
```
bin_op: or
test_items:
- compare:
    op: eq
    value: 127.0.0.1
  flag: --bind-address
- flag: --bind-address
  set: false
```

