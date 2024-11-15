## Control Plane Security Configuration

### Control Plane Node Configuration Files

#### Control 1.1.1

##### Description:

Ensure that the API server configuration file permissions are
set to 600 (Automated)


##### Remediation:

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/kube-apiserver`


##### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c permissions=%a /var/snap/k8s/common/args/kube-apiserver; fi'
```

##### Expected output:

```
permissions=600
```

#### Control 1.1.2

##### Description:

Ensure that the API server configuration file ownership is set
to root:root (Automated)


##### Remediation:

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/kube-apiserver`


##### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c %U:%G /var/snap/k8s/common/args/kube-apiserver; fi'
```

##### Expected output:

```
root:root
```

#### Control 1.1.3

##### Description:

Ensure that the controller manager configuration file
permissions are set to 600 (Automated)


##### Remediation:

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/kube-controller-manager`


##### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c permissions=%a /var/snap/k8s/common/args/kube-controller-manager; fi'
```

##### Expected output:

```
permissions=600
```

#### Control 1.1.4

##### Description:

Ensure that the controller manager configuration file ownership
is set to root:root (Automated)


##### Remediation:

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/kube-controller-manager`


##### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c %U:%G /var/snap/k8s/common/args/kube-controller-manager; fi'
```

##### Expected output:

```
root:root
```

#### Control 1.1.5

##### Description:

Ensure that the scheduler configuration file permissions are set
to 600 (Automated)


##### Remediation:

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/kube-scheduler`


##### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c permissions=%a /var/snap/k8s/common/args/kube-scheduler; fi'
```

##### Expected output:

```
permissions=600
```

#### Control 1.1.6

##### Description:

Ensure that the scheduler configuration file ownership is set to
root:root (Automated)


##### Remediation:

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/kube-scheduler`


##### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c %U:%G /var/snap/k8s/common/args/kube-scheduler; fi'
```

##### Expected output:

```
root:root
```

#### Control 1.1.7

##### Description:

Ensure that the dqlite configuration file permissions are set to
644 or more restrictive (Automated)


##### Remediation:

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/k8s-dqlite`


##### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/k8s-dqlite; then find /var/snap/k8s/common/args/k8s-dqlite -name '*dqlite*' | xargs stat -c permissions=%a; fi'
```

##### Expected output:

```
permissions=600
```

#### Control 1.1.8

##### Description:

Ensure that the dqlite configuration file ownership is set to
root:root (Automated)


##### Remediation:

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/k8s-dqlite`


##### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/k8s-dqlite; then find /var/snap/k8s/common/args/k8s-dqlite -name '*dqlite*' | xargs stat -c %U:%G; fi'
```

##### Expected output:

```
root:root
```

#### Control 1.1.9

##### Description:

Ensure that the Container Network Interface file permissions are
set to 600 (Manual)


##### Remediation:

Run the following command on the control plane node.

`chmod 600 /etc/cni/net.d/05-cilium.conflist`


##### Audit (as root):

```
ps -ef | grep bin/kubelet | grep -- --cni-conf-dir | sed 's%.*cni-conf-dir[= ]\([^ ]*\).*%\1%' | xargs -I{} find {} -mindepth 1 | xargs --no-run-if-empty stat -c permissions=%a
find /etc/cni/net.d/05-cilium.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c permissions=%a
```

##### Expected output:

```
permissions=600
```

#### Control 1.1.10

##### Description:

Ensure that the Container Network Interface file ownership is
set to root:root (Manual)


##### Remediation:

Run the following command on the control plane node.

`chown root:root /etc/cni/net.d/05-cilium.conflist`


##### Audit (as root):

```
find /etc/cni/net.d/05-cilium.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c %U:%G
```

##### Expected output:

```
root:root
```

#### Control 1.1.11

##### Description:

Ensure that the dqlite data directory permissions are set to 700
or more restrictive (Automated)


##### Remediation:

Dqlite data are kept by default under
/var/snap/k8s/common/var/lib/k8s-dqlite.
To comply with the spirit of this CIS recommendation:

`chmod 700 /var/snap/k8s/common/var/lib/k8s-dqlite`


##### Audit (as root):

```
DATA_DIR='/var/snap/k8s/common/var/lib/k8s-dqlite'
stat -c permissions=%a "$DATA_DIR"
```

##### Expected output:

```
permissions=700
```

#### Control 1.1.12

##### Description:

Ensure that the dqlite data directory ownership is set to
root:root (Automated)


##### Remediation:

Dqlite data are kept by default under
/var/snap/k8s/common/var/lib/k8s-dqlite.
To comply with the spirit of this CIS recommendation:

`chown root:root /var/snap/k8s/common/var/lib/k8s-dqlite`


##### Audit (as root):

```
DATA_DIR='/var/snap/k8s/common/var/lib/k8s-dqlite'
stat -c %U:%G "$DATA_DIR"
```

##### Expected output:

```
root:root
```

#### Control 1.1.13

##### Description:

Ensure that the admin.conf file permissions are set to 600
(Automated)


##### Remediation:

Run the following command on the control plane node.

`chmod 600 /etc/kubernetes/admin.conf`


##### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c permissions=%a /etc/kubernetes/admin.conf; fi'
```

##### Expected output:

```
permissions=600
```

#### Control 1.1.14

##### Description:

Ensure that the admin.conf file ownership is set to root:root
(Automated)


##### Remediation:

Run the following command on the control plane node.

`chown root:root /etc/kubernetes/admin.conf`


##### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c %U:%G /etc/kubernetes/admin.conf; fi'
```

##### Expected output:

```
root:root
```

#### Control 1.1.15

##### Description:

Ensure that the scheduler.conf file permissions are set to 600
(Automated)


##### Remediation:

Run the following command on the control plane node.

`chmod 600 /etc/kubernetes/scheduler.conf`


##### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c permissions=%a /etc/kubernetes/scheduler.conf; fi'
```

##### Expected output:

```
permissions=600
```

#### Control 1.1.16

##### Description:

Ensure that the scheduler.conf file ownership is set to
root:root (Automated)


##### Remediation:

Run the following command on the control plane node.

`chown root:root /etc/kubernetes/scheduler.conf`


##### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c %U:%G /etc/kubernetes/scheduler.conf; fi'
```

##### Expected output:

```
root:root
```

#### Control 1.1.17

##### Description:

Ensure that the controller-manager.conf file permissions are set
to 600 (Automated)


##### Remediation:

Run the following command on the control plane node.

`chmod 600 /etc/kubernetes/controller.conf`


##### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c permissions=%a /etc/kubernetes/controller.conf; fi'
```

##### Expected output:

```
permissions=600
```

#### Control 1.1.18

##### Description:

Ensure that the controller-manager.conf file ownership is set to
root:root (Automated)


##### Remediation:

Run the following command on the control plane node.

`chown root:root /etc/kubernetes/controller.conf`


##### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c %U:%G /etc/kubernetes/controller.conf; fi'
```

##### Expected output:

```
root:root
```

#### Control 1.1.19

##### Description:

Ensure that the Kubernetes PKI directory and file ownership is
set to root:root (Automated)


##### Remediation:

Run the following command on the control plane node.

`chown -R root:root /etc/kubernetes/pki/`


##### Audit (as root):

```
find /etc/kubernetes/pki/ | xargs stat -c %U:%G
```

##### Expected output:

```
root:root
```

#### Control 1.1.20

##### Description:

Ensure that the Kubernetes PKI certificate file permissions are
set to 600 (Manual)


##### Remediation:

Run the following command on the control plane node.

`chmod -R 600 /etc/kubernetes/pki/*.crt`


##### Audit (as root):

```
find /etc/kubernetes/pki/ -name '*.crt' | xargs stat -c permissions=%a
```

##### Expected output:

```
permissions=600
```

#### Control 1.1.21

##### Description:

Ensure that the Kubernetes PKI key file permissions are set to
600 (Manual)


##### Remediation:

Run the following command on the control plane node.

`chmod -R 600 /etc/kubernetes/pki/*.key`


##### Audit (as root):

```
find /etc/kubernetes/pki/ -name '*.key' | xargs stat -c permissions=%a
```

##### Expected output:

```
permissions=600
```

### API Server

#### Control 1.2.1

##### Description:

Ensure that the --anonymous-auth argument is set to false
(Manual)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the following argument.

`--anonymous-auth=false`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--anonymous-auth=false
```

#### Control 1.2.2

##### Description:

Ensure that the --token-auth-file parameter is not set
(Automated)


##### Remediation:

Follow the documentation and configure alternate mechanisms for
authentication. Then,
edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and remove the --token-auth-file
argument.


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--token-auth-file is not set
```

#### Control 1.2.3

##### Description:

Ensure that the --DenyServiceExternalIPs is not set (Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and remove the
`DenyServiceExternalIPs`
from enabled admission plugins.


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--enable-admission-plugins does not contain
DenyServiceExternalIPs
```

#### Control 1.2.4

##### Description:

Ensure that the --kubelet-client-certificate and --kubelet-
client-key arguments are set as appropriate (Automated)


##### Remediation:

Follow the Kubernetes documentation and set up the TLS
connection between the
apiserver and kubelets. Then, edit API server configuration file
/var/snap/k8s/common/args/kube-apiserver on the control plane node and set the
kubelet client certificate and key parameters as follows.

```
--kubelet-client-certificate=<path/to/client-certificate-file>
--kubelet-client-key=<path/to/client-key-file>
```


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-
kubelet-client.crt and --kubelet-client-
key=/etc/kubernetes/pki/apiserver-kubelet-client.key
```

#### Control 1.2.5

##### Description:

Ensure that the --kubelet-certificate-authority argument is set
as appropriate (Automated)


##### Remediation:

Follow the Kubernetes documentation and setup the TLS connection
between
the apiserver and kubelets. Then, edit the API server
configuration file
/var/snap/k8s/common/args/kube-apiserver on the control plane node and set the
--kubelet-certificate-authority parameter to the path to the
cert file for the certificate authority.

`--kubelet-certificate-authority=<ca-string>`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--kubelet-certificate-authority=/etc/kubernetes/pki/ca.crt
```

#### Control 1.2.6

##### Description:

Ensure that the --authorization-mode argument is not set to
AlwaysAllow (Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode
parameter to values other than AlwaysAllow.
One such example could be as follows.

`--authorization-mode=Node,RBAC`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--authorization-mode=Node,RBAC
```

#### Control 1.2.7

##### Description:

Ensure that the --authorization-mode argument includes Node
(Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode
parameter to a value that includes Node.

`--authorization-mode=Node,RBAC`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--authorization-mode=Node,RBAC
```

#### Control 1.2.8

##### Description:

Ensure that the --authorization-mode argument includes RBAC
(Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode
parameter to a value that includes RBAC,

`--authorization-mode=Node,RBAC`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--authorization-mode=Node,RBAC
```

#### Control 1.2.9

##### Description:

Ensure that the admission control plugin EventRateLimit is set
(Manual)


##### Remediation:

Follow the Kubernetes documentation and set the desired limits
in a configuration file.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
and set the following arguments.

```
--enable-admission-plugins=...,EventRateLimit,...
--admission-control-config-file=<path/to/configuration/file>
```


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

#### Control 1.2.10

##### Description:

Ensure that the admission control plugin AlwaysAdmit is not set
(Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and either remove the --enable-
admission-plugins parameter, or set it to a
value that does not include AlwaysAdmit.


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

#### Control 1.2.11

##### Description:

Ensure that the admission control plugin AlwaysPullImages is set
(Manual)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-plugins
parameter to include
AlwaysPullImages.

`--enable-admission-plugins=...,AlwaysPullImages,...`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

#### Control 1.2.12

##### Description:

Ensure that the admission control plugin SecurityContextDeny is
set if PodSecurityPolicy is not used (Manual)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-plugins
parameter to include
SecurityContextDeny, unless PodSecurityPolicy is already in
place.

`--enable-admission-plugins=...,SecurityContextDeny,...`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

#### Control 1.2.13

##### Description:

Ensure that the admission control plugin ServiceAccount is set
(Automated)


##### Remediation:

Follow the documentation and create ServiceAccount objects as
per your environment.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and ensure that the --disable-
admission-plugins parameter is set to a
value that does not include ServiceAccount.


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--disable-admission-plugins is not set
```

#### Control 1.2.14

##### Description:

Ensure that the admission control plugin NamespaceLifecycle is
set (Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --disable-admission-
plugins parameter to
ensure it does not include NamespaceLifecycle.


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--disable-admission-plugins is not set
```

#### Control 1.2.15

##### Description:

Ensure that the admission control plugin NodeRestriction is set
(Automated)


##### Remediation:

Follow the Kubernetes documentation and configure
NodeRestriction plug-in on kubelets.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-plugins
parameter to a
value that includes NodeRestriction.

`--enable-admission-plugins=...,NodeRestriction,...`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

#### Control 1.2.16

##### Description:

Ensure that the --secure-port argument is not set to 0
(Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and either remove the --secure-port
parameter or
set it to a different (non-zero) desired port.


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--secure-port=6443
```

#### Control 1.2.17

##### Description:

Ensure that the --profiling argument is set to false (Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the following argument.

`--profiling=false`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--profiling=false
```

#### Control 1.2.18

##### Description:

Ensure that the --audit-log-path argument is set (Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-path parameter
to a suitable path and
file where you would like audit logs to be written.

`--audit-log-path=/var/log/apiserver/audit.log`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--audit-log-path=/var/log/apiserver/audit.log
```

#### Control 1.2.19

##### Description:

Ensure that the --audit-log-maxage argument is set to 30 or as
appropriate (Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxage
parameter to 30
or as an appropriate number of days.

`--audit-log-maxage=30`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--audit-log-maxage=30
```

#### Control 1.2.20

##### Description:

Ensure that the --audit-log-maxbackup argument is set to 10 or
as appropriate (Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxbackup
parameter to 10 or to an appropriate
value.

`--audit-log-maxbackup=10`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--audit-log-maxbackup=10
```

#### Control 1.2.21

##### Description:

Ensure that the --audit-log-maxsize argument is set to 100 or as
appropriate (Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxsize
parameter to an appropriate size in MB.

`--audit-log-maxsize=100`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--audit-log-maxsize=100
```

#### Control 1.2.22

##### Description:

Ensure that the --request-timeout argument is set as appropriate
(Manual)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
and set the following argument as appropriate and if needed.

`--request-timeout=300s`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--request-timeout=300s
```

#### Control 1.2.23

##### Description:

Ensure that the --service-account-lookup argument is set to true
(Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the following argument.

`--service-account-lookup=true`

Alternatively, you can delete the --service-account-lookup
argument from this file so
that the default takes effect.


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--service-account-lookup is not set
```

#### Control 1.2.24

##### Description:

Ensure that the --service-account-key-file argument is set as
appropriate (Automated)


##### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --service-account-key-file
parameter
to the public key file for service accounts.

`--service-account-key-file=<filename>`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--service-account-key-
file=/etc/kubernetes/pki/serviceaccount.key
```

#### Control 1.2.25

##### Description:

Ensure that the --etcd-certfile and --etcd-keyfile arguments are
set as appropriate (Automated)


##### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


#### Control 1.2.26

##### Description:

Ensure that the --tls-cert-file and --tls-private-key-file
arguments are set as appropriate (Automated)


##### Remediation:

Follow the Kubernetes documentation and set up the TLS
connection on the apiserver.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the TLS certificate and
private key file parameters.

```
--tls-cert-file=<path/to/tls-certificate-file>
--tls-private-key-file=<path/to/tls-key-file>
```


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--tls-cert-file=/etc/kubernetes/pki/apiserver.crt and --tls-
private-key-file=/etc/kubernetes/pki/apiserver.key
```

#### Control 1.2.27

##### Description:

Ensure that the --client-ca-file argument is set as appropriate
(Automated)


##### Remediation:

Follow the Kubernetes documentation and set up the TLS
connection on the apiserver.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the client certificate
authority file.

`--client-ca-file=<path/to/client-ca-file>`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--client-ca-file=/etc/kubernetes/pki/client-ca.crt
```

#### Control 1.2.28

##### Description:

Ensure that the --etcd-cafile argument is set as appropriate
(Automated)


##### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### Expected output:

```
--etcd-cafile=/etc/kubernetes/pki/etcd/ca.crt
```

#### Control 1.2.29

##### Description:

Ensure that the --encryption-provider-config argument is set as
appropriate (Manual)


##### Remediation:

Follow the Kubernetes documentation and configure a
EncryptionConfig file.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --encryption-provider-
config parameter to the path of that file.

`--encryption-provider-config=</path/to/EncryptionConfig/File>`


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--encryption-provider-config is set
```

#### Control 1.2.30

##### Description:

Ensure that encryption providers are appropriately configured
(Manual)


##### Remediation:

Follow the Kubernetes documentation and configure a
EncryptionConfig file.
In this file, choose aescbc, kms or secretbox as the encryption
provider.


##### Audit (as root):

```
ENCRYPTION_PROVIDER_CONFIG=$(ps -ef | grep kube-apiserver | grep -- --encryption-provider-config | sed 's%.*encryption-provider-config[= ]\([^ ]*\).*%\1%')
if test -e $ENCRYPTION_PROVIDER_CONFIG; then grep -A1 'providers:' $ENCRYPTION_PROVIDER_CONFIG | tail -n1 | grep -o "[A-Za-z]*" | sed 's/^/provider=/'; fi
```

##### Expected output:

```
--encryption-provider-config is one of or all of
aescbc,kms,secretbox
```

#### Control 1.2.31

##### Description:

Ensure that the API Server only makes use of Strong
Cryptographic Ciphers (Manual)


##### Remediation:

Edit the API server configuration file
/etc/kubernetes/manifests/kube-apiserver.yaml
on the control plane node and set the following argument.

```
--tls-cipher-suites=TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA38
4,TLS_CHACHA20_POLY1305_SHA256,
TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_12
8_GCM_SHA256,
TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_25
6_GCM_SHA384,
TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_CHAC
HA20_POLY1305_SHA256,
TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_C
BC_SHA,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_GC
M_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_RSA_WITH_3DES_ED
E_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA,
TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA,TLS
_RSA_WITH_AES_256_GCM_SHA384
```


##### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

##### Expected output:

```
--tls-cipher-suites=TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA38
4,TLS_CHACHA20_POLY1305_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_
SHA,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH
_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECD
HE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_CHACHA20_PO
LY1305_SHA256,TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_
WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_E
CDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA
384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_CHAC
HA20_POLY1305_SHA256,TLS_RSA_WITH_3DES_EDE_CBC_SHA,TLS_RSA_WITH_
AES_128_CBC_SHA,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES
_256_CBC_SHA,TLS_RSA_WITH_AES_256_GCM_SHA384
```

### Controller Manager

#### Control 1.3.1

##### Description:

Ensure that the --terminated-pod-gc-threshold argument is set as
appropriate (Manual)


##### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --terminated-pod-gc-
threshold to an appropriate threshold.

`--terminated-pod-gc-threshold=10`


##### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

##### Expected output:

```
--terminated-pod-gc-threshold=12500
```

#### Control 1.3.2

##### Description:

Ensure that the --profiling argument is set to false (Automated)


##### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the following argument.

`--profiling=false`


##### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

##### Expected output:

```
--profiling=false
```

#### Control 1.3.3

##### Description:

Ensure that the --use-service-account-credentials argument is
set to true (Automated)


##### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node to set the following argument.

`--use-service-account-credentials=true`


##### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

##### Expected output:

```
--user-service-account-credentials=true
```

#### Control 1.3.4

##### Description:

Ensure that the --service-account-private-key-file argument is
set as appropriate (Automated)


##### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --service-account-private-
key-file parameter
to the private key file for service accounts.

`--service-account-private-key-file=<filename>`


##### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

##### Expected output:

```
--service-account-private-key-
file=/etc/kubernetes/pki/serviceaccount.key
```

#### Control 1.3.5

##### Description:

Ensure that the --root-ca-file argument is set as appropriate
(Automated)


##### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --root-ca-file parameter
to the certificate bundle file.

`--root-ca-file=<path/to/file>`


##### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

##### Expected output:

```
--root-ca-file=/etc/kubernetes/pki/ca.crt
```

#### Control 1.3.6

##### Description:

Ensure that the RotateKubeletServerCertificate argument is set
to true (Automated)


##### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --feature-gates parameter
to include RotateKubeletServerCertificate=true.

`--feature-gates=RotateKubeletServerCertificate=true`


##### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

##### Expected output:

```
--feature-gates is not set
```

#### Control 1.3.7

##### Description:

Ensure that the --bind-address argument is set to 127.0.0.1
(Automated)


##### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and ensure the correct value for the
--bind-address parameter
and restart the controller manager service


##### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

##### Expected output:

```
---bind-address=127.0.0.1
```

### Scheduler

#### Control 1.4.1

##### Description:

Ensure that the --profiling argument is set to false (Automated)


##### Remediation:

Edit the Scheduler configuration file /var/snap/k8s/common/args/kube-scheduler file
on the control plane node and set the following argument.

`--profiling=false`


##### Audit (as root):

```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

##### Expected output:

```
--profiling=false
```

#### Control 1.4.2

##### Description:

Ensure that the --bind-address argument is set to 127.0.0.1
(Automated)


##### Remediation:

Edit the Scheduler configuration file /var/snap/k8s/common/args/kube-scheduler
on the control plane node and ensure the correct value for the
--bind-address parameter
and restart the kube-scheduler service


##### Audit (as root):

```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

##### Expected output:

```
---bind-address=127.0.0.1
```

