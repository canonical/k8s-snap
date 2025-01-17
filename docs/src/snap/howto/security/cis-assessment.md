# CIS compliance

CIS Hardening refers to the process of implementing security configurations that
align with the benchmarks set by the [Center for Internet Security (CIS)].
Out of the box {{product}} complies with the majority of the recommended
CIS security configurations. Since implementing all security recommendations
would come at the expense of compatibility and/or performance we expect
cluster administrators to follow post deployment hardening steps based on their
needs. This guide covers:

  * Post-deployment hardening steps you could consider for your {{product}}
  * Using [kube-bench] to automatically check whether your Kubernetes
    clusters are configured according to the [CIS Kubernetes Benchmark]
  * Manually configuring and auditing each CIS hardening recommendation


## Prerequisites

This guide assumes the following:

- You have a bootstrapped {{product}} cluster (see the [getting started] guide)
- You have root or sudo access to the machine
- You have reviewed the [post-deployment hardening] guide and have applied the
  hardening steps that relevant to your use-case


## Critical post-deployment hardening steps

By completing these steps, you can ensure your cluster achieves does not fail
any of the CIS hardening recommendations.

```{include} ../../../_parts/common_hardening.md
```

## Assess CIS hardening with kube-bench

Download the latest [kube-bench release] on your Kubernetes nodes. Make sure
to select the appropriate binary version.

For example, to download the Linux binary, use the following command. Replace
`KB` by the version listed in the releases page:

```
KB=8.0
mkdir kube-bench
cd kube-bench
curl -L https://github.com/aquasecurity/kube-bench/releases/download/v0.$KB/kube-bench_0.$KB\_linux_amd64.tar.gz -o kube-bench_0.$KB\_linux_amd64.tar.gz
```

Extract the downloaded tarball and move the binary to a directory in your PATH:

```
tar -xvf kube-bench_0.$KB\_linux_amd64.tar.gz
sudo mv kube-bench /usr/local/bin/
```

Verify kube-bench installation:

```
kube-bench version
```

The output should list the version installed.

Install `kubectl` and configure it to interact with the cluster:

```{warning}
This will override your ~/.kube/config if you already have kubectl installed in your cluster.
```

```
sudo snap install kubectl --classic
mkdir ~/.kube/
sudo k8s kubectl config view --raw > ~/.kube/config
export KUBECONFIG=~/.kube/config
```

Get CIS hardening checks applicable for {{product}}:

```
git clone -b ck8s-dqlite https://github.com/canonical/kube-bench.git kube-bench-ck8s-cfg
```

Test-run kube-bench against {{product}}:

```
sudo -E kube-bench --version ck8s-cis-1.24 --config-dir ./kube-bench-ck8s-cfg/cfg/ --config ./kube-bench-ck8s-cfg/cfg/config.yaml
```

Review the warnings detected and address any failing checks you see fit.

```
[INFO] 1 Control Plane Security Configuration
...
[PASS] 1.1.7 Ensure that the dqlite configuration file permissions are set to 644 or more restrictive (Automated)
[PASS] 1.1.8 Ensure that the dqlite configuration file ownership is set to root:root (Automated)
...
[PASS] 1.1.11 Ensure that the dqlite data directory permissions are set to 700 or more restrictive (Automated)
[PASS] 1.1.12 Ensure that the dqlite data directory ownership is set to root:root (Automated)
...
== Summary master ==
55 checks PASS
0 checks FAIL
4 checks WARN
0 checks INFO

[INFO] 3 Control Plane Configuration
...
== Summary controlplane ==
1 checks PASS
0 checks FAIL
2 checks WARN
0 checks INFO

[INFO] 4 Worker Node Security Configuration
...
== Summary node ==
23 checks PASS
0 checks FAIL
0 checks WARN
0 checks INFO

[INFO] 5 Kubernetes Policies
...
== Summary policies ==
0 checks PASS
0 checks FAIL
30 checks WARN
0 checks INFO

== Summary total ==
79 checks PASS
0 checks FAIL
36 checks WARN
0 checks INFO

```


## Manually audit CIS hardening recommendations

In what follows we iterate over all hardening recommendations
and, when possible, provide information on how to comply with each
one manually. This can be used for manually auditing the CIS
hardening state of a cluster.

### Control plane security configuration

#### Control plane node configuration files

##### CIS Control 1.1.1

**Description:**

Ensure that the API server configuration file permissions
are set to 600


**Remediation:**

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/kube-apiserver`


**Audit (as root):**

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c permissions=%a /var/snap/k8s/common/args/kube-apiserver; fi'
```

**Expected output:**

```
permissions=600
```

##### CIS Control 1.1.2

**Description:**

Ensure that the API server configuration file ownership is
set to root:root


**Remediation:**

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/kube-apiserver`


**Audit (as root):**

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c %U:%G /var/snap/k8s/common/args/kube-apiserver; fi'
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.3

**Description:**

Ensure that the controller manager configuration file
permissions are set to 600


**Remediation:**

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/kube-controller-manager`


**Audit (as root):**

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c permissions=%a /var/snap/k8s/common/args/kube-controller-manager; fi'
```

**Expected output:**

```
permissions=600
```

##### CIS Control 1.1.4

**Description:**

Ensure that the controller manager configuration file
ownership is set to root:root


**Remediation:**

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/kube-controller-manager`


**Audit (as root):**

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c %U:%G /var/snap/k8s/common/args/kube-controller-manager; fi'
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.5

**Description:**

Ensure that the scheduler configuration file permissions are
set to 600


**Remediation:**

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/kube-scheduler`


**Audit (as root):**

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c permissions=%a /var/snap/k8s/common/args/kube-scheduler; fi'
```

**Expected output:**

```
permissions=600
```

##### CIS Control 1.1.6

**Description:**

Ensure that the scheduler configuration file ownership is
set to root:root


**Remediation:**

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/kube-scheduler`


**Audit (as root):**

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c %U:%G /var/snap/k8s/common/args/kube-scheduler; fi'
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.7

**Description:**

Ensure that the dqlite configuration file permissions are
set to 644 or more restrictive


**Remediation:**

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/k8s-dqlite`


**Audit (as root):**

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/k8s-dqlite; then stat -c permissions=%a /var/snap/k8s/common/args/k8s-dqlite; fi'
```

**Expected output:**

```
permissions=600
```

##### CIS Control 1.1.8

**Description:**

Ensure that the dqlite configuration file ownership is set
to root:root


**Remediation:**

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/k8s-dqlite`


**Audit (as root):**

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/k8s-dqlite; then stat -c %U:%G /var/snap/k8s/common/args/k8s-dqlite; fi'
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.9

**Description:**

Ensure that the Container Network Interface file permissions
are set to 600


**Remediation:**

Run the following command on the control plane node.

`chmod 600 /etc/cni/net.d/05-cilium.conflist`


**Audit (as root):**

```
ps -ef | grep bin/kubelet | grep -- --cni-conf-dir | sed 's%.*cni-conf-dir[= ]\([^ ]*\).*%\1%' | xargs -I{} find {} -mindepth 1 | xargs --no-run-if-empty stat -c permissions=%a
find /etc/cni/net.d/05-cilium.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c permissions=%a
```

**Expected output:**

```
permissions=600
```

##### CIS Control 1.1.10

**Description:**

Ensure that the Container Network Interface file ownership
is set to root:root


**Remediation:**

Run the following command on the control plane node.

`chown root:root /etc/cni/net.d/05-cilium.conflist`


**Audit (as root):**

```
find /etc/cni/net.d/05-cilium.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c %U:%G
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.11

**Description:**

Ensure that the dqlite data directory permissions are set to
700 or more restrictive


**Remediation:**

Dqlite data are kept by default under
/var/snap/k8s/common/var/lib/k8s-dqlite.
To comply with the spirit of this CIS recommendation:

`chmod 700 /var/snap/k8s/common/var/lib/k8s-dqlite`


**Audit (as root):**

```
DATA_DIR='/var/snap/k8s/common/var/lib/k8s-dqlite'
stat -c permissions=%a "$DATA_DIR"
```

**Expected output:**

```
permissions=700
```

##### CIS Control 1.1.12

**Description:**

Ensure that the dqlite data directory ownership is set to
root:root


**Remediation:**

Dqlite data are kept by default under
/var/snap/k8s/common/var/lib/k8s-dqlite.
To comply with the spirit of this CIS recommendation:

`chown root:root /var/snap/k8s/common/var/lib/k8s-dqlite`


**Audit (as root):**

```
DATA_DIR='/var/snap/k8s/common/var/lib/k8s-dqlite'
stat -c %U:%G "$DATA_DIR"
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.13

**Description:**

Ensure that the admin.conf file permissions are set to 600


**Remediation:**

Run the following command on the control plane node.

`chmod 600 /etc/kubernetes/admin.conf`


**Audit (as root):**

```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c permissions=%a /etc/kubernetes/admin.conf; fi'
```

**Expected output:**

```
permissions=600
```

##### CIS Control 1.1.14

**Description:**

Ensure that the admin.conf file ownership is set to
root:root


**Remediation:**

Run the following command on the control plane node.

`chown root:root /etc/kubernetes/admin.conf`


**Audit (as root):**

```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c %U:%G /etc/kubernetes/admin.conf; fi'
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.15

**Description:**

Ensure that the scheduler.conf file permissions are set to
600


**Remediation:**

Run the following command on the control plane node.

`chmod 600 /etc/kubernetes/scheduler.conf`


**Audit (as root):**

```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c permissions=%a /etc/kubernetes/scheduler.conf; fi'
```

**Expected output:**

```
permissions=600
```

##### CIS Control 1.1.16

**Description:**

Ensure that the scheduler.conf file ownership is set to
root:root


**Remediation:**

Run the following command on the control plane node.

`chown root:root /etc/kubernetes/scheduler.conf`


**Audit (as root):**

```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c %U:%G /etc/kubernetes/scheduler.conf; fi'
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.17

**Description:**

Ensure that the controller-manager.conf file permissions are
set to 600


**Remediation:**

Run the following command on the control plane node.

`chmod 600 /etc/kubernetes/controller.conf`


**Audit (as root):**

```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c permissions=%a /etc/kubernetes/controller.conf; fi'
```

**Expected output:**

```
permissions=600
```

##### CIS Control 1.1.18

**Description:**

Ensure that the controller-manager.conf file ownership is
set to root:root


**Remediation:**

Run the following command on the control plane node.

`chown root:root /etc/kubernetes/controller.conf`


**Audit (as root):**

```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c %U:%G /etc/kubernetes/controller.conf; fi'
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.19

**Description:**

Ensure that the Kubernetes PKI directory and file ownership
is set to root:root


**Remediation:**

Run the following command on the control plane node.

`chown -R root:root /etc/kubernetes/pki/`


**Audit (as root):**

```
find /etc/kubernetes/pki/ | xargs stat -c %U:%G
```

**Expected output:**

```
root:root
```

##### CIS Control 1.1.20

**Description:**

Ensure that the Kubernetes PKI certificate file permissions
are set to 600


**Remediation:**

Run the following command on the control plane node.

`chmod -R 600 /etc/kubernetes/pki/*.crt`


**Audit (as root):**

```
find /etc/kubernetes/pki/ -name '*.crt' | xargs stat -c permissions=%a
```

**Expected output:**

```
permissions=600
```

##### CIS Control 1.1.21

**Description:**

Ensure that the Kubernetes PKI key file permissions are set
to 600


**Remediation:**

Run the following command on the control plane node.

`chmod -R 600 /etc/kubernetes/pki/*.key`


**Audit (as root):**

```
find /etc/kubernetes/pki/ -name '*.key' | xargs stat -c permissions=%a
```

**Expected output:**

```
permissions=600
```

#### API Server

##### CIS Control 1.2.1

**Description:**

Ensure that the --anonymous-auth argument is set to false


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the following argument.

`--anonymous-auth=false`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--anonymous-auth=false
```

##### CIS Control 1.2.2

**Description:**

Ensure that the --token-auth-file parameter is not set


**Remediation:**

Follow the Kubernetes documentation and configure alternate
mechanisms for
authentication. Then, edit the API server configuration file
/var/snap/k8s/common/args/kube-apiserver
on the control plane node and remove the --token-auth-file
argument.


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--token-auth-file is not set
```

##### CIS Control 1.2.3

**Description:**

Ensure that the --DenyServiceExternalIPs is not set


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and remove the
`DenyServiceExternalIPs`
from enabled admission plugins.


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--enable-admission-plugins does not contain
DenyServiceExternalIPs
```

##### CIS Control 1.2.4

**Description:**

Ensure that the --kubelet-client-certificate and --kubelet-
client-key arguments are set as appropriate


**Remediation:**

Follow the Kubernetes documentation and set up the TLS
connection between the
apiserver and kubelets. Then, edit API server configuration
file
/var/snap/k8s/common/args/kube-apiserver on the control plane node and set the
kubelet client certificate and key parameters as follows.

```
--kubelet-client-certificate=<path/to/client-certificate-file>
--kubelet-client-key=<path/to/client-key-file>
```


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt
and --kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key
```

##### CIS Control 1.2.5

**Description:**

Ensure that the --kubelet-certificate-authority argument is
set as appropriate


**Remediation:**

Follow the Kubernetes documentation and setup the TLS
connection between
the apiserver and kubelets. Then, edit the API server
configuration file
/var/snap/k8s/common/args/kube-apiserver on the control plane node and set the
--kubelet-certificate-authority parameter to the path to the
cert file for the certificate authority.

`--kubelet-certificate-authority=<ca-string>`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--kubelet-certificate-authority=/etc/kubernetes/pki/ca.crt
```

##### CIS Control 1.2.6

**Description:**

Ensure that the --authorization-mode argument is not set to
AlwaysAllow


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode
parameter to values other than AlwaysAllow.
One such example could be as follows.

`--authorization-mode=Node,RBAC`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--authorization-mode=Node,RBAC
```

##### CIS Control 1.2.7

**Description:**

Ensure that the --authorization-mode argument includes Node


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode
parameter to a value that includes Node.

`--authorization-mode=Node,RBAC`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--authorization-mode=Node,RBAC
```

##### CIS Control 1.2.8

**Description:**

Ensure that the --authorization-mode argument includes RBAC


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode
parameter to a value that includes RBAC,

`--authorization-mode=Node,RBAC`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--authorization-mode=Node,RBAC
```

##### CIS Control 1.2.9

**Description:**

Ensure that the admission control plugin EventRateLimit is
set


**Remediation:**

Follow the Kubernetes documentation and set the desired
limits in a configuration file.
Then, edit the API server configuration file
/var/snap/k8s/common/args/kube-apiserver
and set the following arguments.

```
--enable-admission-plugins=...,EventRateLimit,...
--admission-control-config-file=<path/to/configuration/file>
```


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### CIS Control 1.2.10

**Description:**

Ensure that the admission control plugin AlwaysAdmit is not
set


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and either remove the
--enable-admission-plugins parameter, or set it to a
value that does not include AlwaysAdmit.


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### CIS Control 1.2.11

**Description:**

Ensure that the admission control plugin AlwaysPullImages is
set


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the
--enable-admission-plugins parameter to include
AlwaysPullImages.

`--enable-admission-plugins=...,AlwaysPullImages,...`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### CIS Control 1.2.12

**Description:**

Ensure that the admission control plugin SecurityContextDeny
is set if PodSecurityPolicy is not used


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-
plugins parameter to include
SecurityContextDeny, unless PodSecurityPolicy is already in
place.

`--enable-admission-plugins=...,SecurityContextDeny,...`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### CIS Control 1.2.13

**Description:**

Ensure that the admission control plugin ServiceAccount is
set


**Remediation:**

Follow the documentation and create ServiceAccount objects
as per your environment. Then, edit the API server configuration
file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and ensure that the
--disable-admission-plugins parameter is set to a
value that does not include ServiceAccount.


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--disable-admission-plugins is not set
```

##### CIS Control 1.2.14

**Description:**

Ensure that the admission control plugin NamespaceLifecycle
is set


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --disable-admission-
plugins parameter to
ensure it does not include NamespaceLifecycle.


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--disable-admission-plugins is not set
```

##### CIS Control 1.2.15

**Description:**

Ensure that the admission control plugin NodeRestriction is
set


**Remediation:**

Follow the Kubernetes documentation and configure
NodeRestriction plug-in on kubelets. Then, edit the API server
configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-
plugins parameter to a
value that includes NodeRestriction.

`--enable-admission-plugins=...,NodeRestriction,...`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### CIS Control 1.2.16

**Description:**

Ensure that the --secure-port argument is not set to 0


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and either remove the --secure-
port parameter or
set it to a different (non-zero) desired port.


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--secure-port=6443
```

##### CIS Control 1.2.17

**Description:**

Ensure that the --profiling argument is set to false


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the following argument.

`--profiling=false`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--profiling=false
```

##### CIS Control 1.2.18 / DISA-STIG V-242402

**Description:**

Ensure that the --audit-log-path argument is set


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-path
parameter to a suitable path and
file where you would like audit logs to be written.

`--audit-log-path=/var/log/apiserver/audit.log`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--audit-log-path=/var/log/apiserver/audit.log
```

##### CIS Control 1.2.19

**Description:**

Ensure that the --audit-log-maxage argument is set to 30 or
as appropriate


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxage
parameter to 30
or as an appropriate number of days.

`--audit-log-maxage=30`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--audit-log-maxage=30
```

##### CIS Control 1.2.20 / DISA STIG V-242463

**Description:**

Ensure that the --audit-log-maxbackup argument is set to 10
or as appropriate


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxbackup
parameter to 10 or to an appropriate
value.

`--audit-log-maxbackup=10`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--audit-log-maxbackup=10
```

##### CIS Control 1.2.21 / DISA STIG V-242462

**Description:**

Ensure that the --audit-log-maxsize argument is set to 100
or as appropriate


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxsize
parameter to an appropriate size in MB.

`--audit-log-maxsize=100`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--audit-log-maxsize=100
```

##### CIS Control 1.2.22

**Description:**

Ensure that the --request-timeout argument is set as
appropriate


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
and set the following argument as appropriate and if needed.

`--request-timeout=300s`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--request-timeout=300s
```

##### CIS Control 1.2.23

**Description:**

Ensure that the --service-account-lookup argument is set to
true


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the following argument.

`--service-account-lookup=true`

Alternatively, you can delete the --service-account-lookup
argument from this file so
that the default takes effect.


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--service-account-lookup is not set
```

##### CIS Control 1.2.24

**Description:**

Ensure that the --service-account-key-file argument is set
as appropriate


**Remediation:**

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --service-account-key-
file parameter
to the public key file for service accounts.

`--service-account-key-file=<filename>`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--service-account-key-
file=/etc/kubernetes/pki/serviceaccount.key
```

##### CIS Control 1.2.25

**Description:**

Ensure that the --etcd-certfile and --etcd-keyfile arguments
are set as appropriate


**Remediation:**

Not applicable. Canonical K8s uses dqlite and the
communication to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### CIS Control 1.2.26

**Description:**

Ensure that the --tls-cert-file and --tls-private-key-file
arguments are set as appropriate


**Remediation:**

Follow the Kubernetes documentation and set up the TLS
connection on the apiserver.
Then, edit the API server configuration file
/var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the TLS certificate and
private key file parameters.

```
--tls-cert-file=<path/to/tls-certificate-file>
--tls-private-key-file=<path/to/tls-key-file>
```


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--tls-cert-file=/etc/kubernetes/pki/apiserver.crt and --tls-
private-key-file=/etc/kubernetes/pki/apiserver.key
```

##### CIS Control 1.2.27

**Description:**

Ensure that the --client-ca-file argument is set as
appropriate


**Remediation:**

Follow the Kubernetes documentation and set up the TLS
connection on the apiserver. Then, edit the API server configuration
file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the client certificate
authority file.

`--client-ca-file=<path/to/client-ca-file>`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--client-ca-file=/etc/kubernetes/pki/client-ca.crt
```

##### CIS Control 1.2.28

**Description:**

Ensure that the --etcd-cafile argument is set as appropriate


**Remediation:**

Not applicable. Canonical K8s uses dqlite and the
communication to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### CIS Control 1.2.29

**Description:**

Ensure that the --encryption-provider-config argument is set
as appropriate


**Remediation:**

Follow the Kubernetes documentation and configure a
EncryptionConfig file. Then, edit the API server configuration
file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --encryption-provider-
config parameter to the path of that file.

`--encryption-provider-
config=</path/to/EncryptionConfig/File>`


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--encryption-provider-config is set
```

##### CIS Control 1.2.30

**Description:**

Ensure that encryption providers are appropriately
configured


**Remediation:**

Follow the Kubernetes documentation and configure a
EncryptionConfig file.
In this file, choose aescbc, kms or secretbox as the
encryption provider.


**Audit (as root):**

```
ENCRYPTION_PROVIDER_CONFIG=$(ps -ef | grep kube-apiserver | grep -- --encryption-provider-config | sed 's%.*encryption-provider-config[= ]\([^ ]*\).*%\1%')
if test -e $ENCRYPTION_PROVIDER_CONFIG; then grep -A1 'providers:' $ENCRYPTION_PROVIDER_CONFIG | tail -n1 | grep -o "[A-Za-z]*" | sed 's/^/provider=/'; fi
```

**Expected output:**

```
--encryption-provider-config is one of or all of
aescbc,kms,secretbox
```

##### CIS Control 1.2.31

**Description:**

Ensure that the API Server only makes use of Strong
Cryptographic Ciphers


**Remediation:**

Edit the API server configuration file
/etc/kubernetes/manifests/kube-apiserver.yaml
on the control plane node and set the following argument.

```
--tls-cipher-suites=TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_S
HA384,TLS_CHACHA20_POLY1305_SHA256,
TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AE
S_128_GCM_SHA256,
TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AE
S_256_GCM_SHA384,
TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_
CHACHA20_POLY1305_SHA256,
TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_1
28_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_25
6_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_RSA_WITH_3DE
S_EDE_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA,
TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA
,TLS_RSA_WITH_AES_256_GCM_SHA384
```


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

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

#### Controller manager

##### CIS Control 1.3.1

**Description:**

Ensure that the --terminated-pod-gc-threshold argument is
set as appropriate


**Remediation:**

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --terminated-pod-gc-
threshold to an appropriate threshold.

`--terminated-pod-gc-threshold=12500`


**Audit (as root):**

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

**Expected output:**

```
--terminated-pod-gc-threshold=12500
```

##### CIS Control 1.3.2

**Description:**

Ensure that the --profiling argument is set to false


**Remediation:**

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the following argument.

`--profiling=false`


**Audit (as root):**

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

**Expected output:**

```
--profiling=false
```

##### CIS Control 1.3.3

**Description:**

Ensure that the --use-service-account-credentials argument
is set to true


**Remediation:**

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node to set the following argument.

`--use-service-account-credentials=true`


**Audit (as root):**

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

**Expected output:**

```
--use-service-account-credentials=true
```

##### CIS Control 1.3.4

**Description:**

Ensure that the --service-account-private-key-file argument
is set as appropriate


**Remediation:**

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --service-account-
private-key-file parameter
to the private key file for service accounts.

`--service-account-private-key-file=<filename>`


**Audit (as root):**

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

**Expected output:**

```
--service-account-private-key-
file=/etc/kubernetes/pki/serviceaccount.key
```

##### CIS Control 1.3.5

**Description:**

Ensure that the --root-ca-file argument is set as
appropriate


**Remediation:**

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --root-ca-file
parameter to the certificate bundle file.

`--root-ca-file=<path/to/file>`


**Audit (as root):**

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

**Expected output:**

```
--root-ca-file=/etc/kubernetes/pki/ca.crt
```

##### CIS Control 1.3.6

**Description:**

Ensure that the RotateKubeletServerCertificate argument is
set to true


**Remediation:**

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --feature-gates
parameter to include RotateKubeletServerCertificate=true.

`--feature-gates=RotateKubeletServerCertificate=true`


**Audit (as root):**

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

**Expected output:**

```
RotateKubeletServerCertificate feature gate is not set, or set
to true
```

##### CIS Control 1.3.7 / DISA STIG V-242385

**Description:**

Ensure that the --bind-address argument is set to 127.0.0.1


**Remediation:**

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and ensure the correct value for
the --bind-address parameter
and restart the controller manager service


**Audit (as root):**

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

**Expected output:**

```
---bind-address=127.0.0.1
```

#### Scheduler

##### CIS Control 1.4.1

**Description:**

Ensure that the --profiling argument is set to false


**Remediation:**

Edit the Scheduler configuration file /var/snap/k8s/common/args/kube-scheduler
on the control plane node and set the following argument.

`--profiling=false`


**Audit (as root):**

```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

**Expected output:**

```
--profiling=false
```

##### CIS Control 1.4.2 / DISA STIG V-242384

**Description:**

Ensure that the --bind-address argument is set to 127.0.0.1


**Remediation:**

Edit the Scheduler configuration file /var/snap/k8s/common/args/kube-scheduler
on the control plane node and ensure the correct value for
the --bind-address parameter
and restart the kube-scheduler service


**Audit (as root):**

```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

**Expected output:**

```
---bind-address=127.0.0.1
```

### Datastore node configuration

#### Datastore node configuration

##### CIS Control 2.1

**Description:**

Ensure that the --cert-file and --key-file arguments are set
as appropriate


**Remediation:**

Not applicable. Canonical K8s uses dqlite and the
communication to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### CIS Control 2.2

**Description:**

Ensure that the --client-cert-auth argument is set to true


**Remediation:**

Not applicable. Canonical K8s uses dqlite and the
communication to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### CIS Control 2.3

**Description:**

Ensure that the --auto-tls argument is not set to true


**Remediation:**

Not applicable. Canonical K8s uses dqlite and the
communication to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### CIS Control 2.4

**Description:**

Ensure that the --peer-cert-file and --peer-key-file
arguments are set as appropriate


**Remediation:**

The certificate pair for dqlite and tls peer communication
is
/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt and
/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key.


**Audit (as root):**

```
if test -e /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt && test -e /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key; then echo 'certs-found'; fi
```

**Expected output:**

```
certs-found
```

##### CIS Control 2.5

**Description:**

Ensure that the --peer-client-cert-auth argument is set to
true


**Remediation:**

Dqlite peer communication uses TLS unless the --enable-tls
is set to false in
/var/snap/k8s/common/args/k8s-dqlite.


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/k8s-dqlite | /bin/grep enable-tls || true; echo $?
```

**Expected output:**

```
0
```

##### CIS Control 2.6

**Description:**

Ensure that the --peer-auto-tls argument is not set to true


**Remediation:**

Not applicable. Canonical K8s uses dqlite and tls peer
communication uses the certificates
created upon the snap creation.


##### CIS Control 2.7

**Description:**

Ensure that a unique Certificate Authority is used for the
datastore


**Remediation:**

Not applicable. Canonical K8s uses dqlite and tls peer
communication uses certificates
created upon cluster setup.


### Control plane configuration

#### Authentication and authorization

##### CIS Control 3.1.1

**Description:**

Client certificate authentication should not be used for
users


**Remediation:**

Alternative mechanisms provided by Kubernetes such as the
use of OIDC should be
implemented in place of client certificates.


#### Logging

##### CIS Control 3.2.1

**Description:**

Ensure that a minimal audit policy is created


**Remediation:**

Create an audit policy file for your cluster.


**Audit (as root):**

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

**Expected output:**

```
--audit-policy-file=/var/snap/k8s/common/etc/audit-policy.yaml
```

##### CIS Control 3.2.2 / DISA STIG V-242403

**Description:**

Ensure that the audit policy covers key security concerns


**Remediation:**

Review the audit policy provided for the cluster and ensure
that it covers
at least the following areas,

- Access to Secrets managed by the cluster. Care should be
taken to only
  log Metadata for requests to Secrets, ConfigMaps, and
TokenReviews, in
  order to avoid risk of logging sensitive data.
- Modification of Pod and Deployment objects.
- Use of `pods/exec`, `pods/portforward`, `pods/proxy` and
`services/proxy`.

For most requests, minimally logging at the Metadata level
is recommended
(the most basic level of logging).


### Worker node security configuration

#### Worker node configuration files

##### CIS Control 4.1.1

**Description:**

Ensure that the kubelet service file permissions are set to
600


**Remediation:**

Run the following command on each worker node.


`chmod 600 /etc/systemd/system/snap.k8s.kubelet.service`


**Audit (as root):**

```
/bin/sh -c "if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c permissions=%a /etc/systemd/system/snap.k8s.kubelet.service; fi"
```

**Expected output:**

```
permissions=600
```

##### CIS Control 4.1.2

**Description:**

Ensure that the kubelet service file ownership is set to
root:root


**Remediation:**

Run the following command on each worker node.


`chown root:root /etc/systemd/system/snap.k8s.kubelet.service`


**Audit (as root):**

```
/bin/sh -c "if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c %U:%G /etc/systemd/system/snap.k8s.kubelet.service; else echo \"File not found\"; fi"
```

**Expected output:**

```
root:root
```

##### CIS Control 4.1.3

**Description:**

If proxy kubeconfig file exists ensure permissions are set
to 600


**Remediation:**

Run the following command on each worker node.


`chmod 600 /etc/kubernetes/proxy.conf`


**Audit (as root):**

```
/bin/sh -c "if test -e /etc/kubernetes/proxy.conf; then stat -c permissions=%a /etc/kubernetes/proxy.conf; fi"
```

**Expected output:**

```
permissions=600
```

##### CIS Control 4.1.4

**Description:**

If proxy kubeconfig file exists ensure ownership is set to
root:root


**Remediation:**

Run the following command on each worker node.


`chown root:root /etc/kubernetes/proxy.conf`


**Audit (as root):**

```
/bin/sh -c "if test -e /etc/kubernetes/proxy.conf; then stat -c %U:%G /etc/kubernetes/proxy.conf; fi"
```

**Expected output:**

```
root:root
```

##### CIS Control 4.1.5

**Description:**

Ensure that the --kubeconfig kubelet.conf file permissions
are set to 600


**Remediation:**

Run the following command on each worker node.


`chmod 600 /etc/kubernetes/kubelet.conf`


**Audit (as root):**

```
/bin/sh -c "if test -e /etc/kubernetes/kubelet.conf; then stat -c permissions=%a /etc/kubernetes/kubelet.conf; fi"
```

**Expected output:**

```
permissions=600
```

##### CIS Control 4.1.6

**Description:**

Ensure that the --kubeconfig kubelet.conf file ownership is
set to root:root


**Remediation:**

Run the following command on each worker node.


`chown root:root /etc/kubernetes/kubelet.conf`


**Audit (as root):**

```
/bin/sh -c "if test -e /etc/kubernetes/kubelet.conf; then stat -c %U:%G /etc/kubernetes/kubelet.conf; fi"
```

**Expected output:**

```
root:root
```

##### CIS Control 4.1.7

**Description:**

Ensure that the certificate authorities file permissions are
set to 600


**Remediation:**

Run the following command to modify the file permissions of
the
--client-ca-file.

`chmod 600 /etc/kubernetes/pki/client-ca.crt`


**Audit (as root):**

```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c permissions=%a $CAFILE; fi
```

**Expected output:**

```
permissions=600
```

##### CIS Control 4.1.8

**Description:**

Ensure that the client certificate authorities file
ownership is set to root:root


**Remediation:**

Run the following command to modify the ownership of the
--client-ca-file.

`chown root:root /etc/kubernetes/pki/client-ca.crt`


**Audit (as root):**

```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c %U:%G $CAFILE; fi
```

**Expected output:**

```
root:root
```

##### CIS Control 4.1.9

**Description:**

If the kubelet config.yaml configuration file is being used
validate permissions set to 600


**Remediation:**

Run the following command (using the config file location
identified in the Audit step)


`chmod 600 /var/snap/k8s/common/args/kubelet`


**Audit (as root):**

```
/bin/sh -c "if test -e /var/snap/k8s/common/args/kubelet; then stat -c permissions=%a /var/snap/k8s/common/args/kubelet; fi"
```

**Expected output:**

```
permissions=600
```

##### CIS Control 4.1.10

**Description:**

If the kubelet config.yaml configuration file is being used
validate file ownership is set to root:root


**Remediation:**

Run the following command (using the config file location
identified in the Audit step)


`chown root:root /var/snap/k8s/common/args/kubelet`


**Audit (as root):**

```
/bin/sh -c "if test -e /var/snap/k8s/common/args/kubelet; then stat -c %U:%G /var/snap/k8s/common/args/kubelet; fi"
```

**Expected output:**

```
root:root
```

#### Kubelet

##### CIS Control 4.2.1

**Description:**

Ensure that the --anonymous-auth argument is set to false


**Remediation:**

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following
argument.

`--anonymous-auth=false`

Restart the kubelet service.

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--anonymous-auth=false
```

##### CIS Control 4.2.2

**Description:**

Ensure that the --authorization-mode argument is not set to
AlwaysAllow


**Remediation:**

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following
argument.

`--authorization-mode=Webhook`

Restart the kubelet service:

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--authorization-mode=Webhook
```

##### CIS Control 4.2.3

**Description:**

Ensure that the --client-ca-file argument is set as
appropriate


**Remediation:**

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following
argument.

`--client-ca-file=/etc/kubernetes/pki/client-ca.crt`

Restart the kubelet service:

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--client-ca-file=/etc/kubernetes/pki/client-ca.crt
```

##### CIS Control 4.2.4

**Description:**

Verify that the --read-only-port argument is set to 0


**Remediation:**

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following
argument.

`--read-only-port=0`

Restart the kubelet service:

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--read-only-port=0
```

##### CIS Control 4.2.5

**Description:**

Ensure that the --streaming-connection-idle-timeout argument
is not set to 0


**Remediation:**

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following
argument.

`--streaming-connection-idle-timeout=5m`

Restart the kubelet service:

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--streaming-connection-idle-timeout is not set, or set to a
value greater or equal to 5m
```

##### CIS Control 4.2.6 / DISA STIG V-242434

**Description:**

Ensure that the --protect-kernel-defaults argument is set to
true


**Remediation:**

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following
argument:

`--protect-kernel-defaults=true`

Restart the kubelet service:

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--protect-kernel-defaults=true
```

##### CIS Control 4.2.7

**Description:**

Ensure that the --make-iptables-util-chains argument is set
to true


**Remediation:**

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and
set the following argument:

`--make-iptables-util-chains=true`

Restart the kubelet service.

For example: `snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--make-iptables-util-chains is not set or set to true
```

##### CIS Control 4.2.8

**Description:**

Ensure that the --hostname-override argument is not set


**Remediation:**

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet
on each worker node and remove the --hostname-override
argument.

Restart the kubelet service.

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--hostname-override is set to false
```

##### CIS Control 4.2.9

**Description:**

Ensure that the --event-qps argument is set to a level which
ensures appropriate event capture


**Remediation:**

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each
worker node and
set the --event-qps parameter as appropriate.

Restart the kubelet service.

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--event-qps is not set, or set to a value greater than 0
```

##### CIS Control 4.2.10

**Description:**

Ensure that the --tls-cert-file and --tls-private-key-file
arguments are set as appropriate


**Remediation:**

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each
worker node and
set the following arguments:

```
--tls-private-key-file=<path/to/private-key-file>
--tls-cert-file=<path/to/tls-certificate-file>
```

Restart the kubelet service.

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--tls-cert-file=/etc/kubernetes/pki/kubelet.crt and --tls-
private-key-file=/etc/kubernetes/pki/kubelet.key
```

##### CIS Control 4.2.11

**Description:**

Ensure that the --rotate-certificates argument is not set to
false


**Remediation:**

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each
worker node and
remove the --rotate-certificates=false argument.

Restart the kubelet service.

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--rotate-certificates is not set, or set to true
```

##### CIS Control 4.2.12

**Description:**

Verify that the RotateKubeletServerCertificate argument is
set to true


**Remediation:**

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each
worker node and
set the argument --feature-
gates=RotateKubeletServerCertificate=true
on each worker node.

Restart the kubelet service.

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
RotateKubeletServerCertificate feature gate is not set, or set
to true
```

##### CIS Control 4.2.13

**Description:**

Ensure that the Kubelet only makes use of Strong
Cryptographic Ciphers


**Remediation:**

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each
worker node and
set the --tls-cipher-suites parameter as follows, or to a
subset of these values.

```
--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_C
HACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_E
CDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256
_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES
_128_GCM_SHA256
```

Restart the kubelet service.

`snap restart k8s.kubelet`


**Audit (as root):**

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

**Expected output:**

```
--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_
ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_
POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WIT
H_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_
RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256
```

### Kubernetes policies

#### RBAC and service accounts

##### CIS Control 5.1.1

**Description:**

Ensure that the cluster-admin role is only used where
required


**Remediation:**

Identify all clusterrolebindings to the cluster-admin role.
Check if they are used and
if they need this role or if they could use a role with
fewer privileges.
Where possible, first bind users to a lower privileged role
and then remove the
clusterrolebinding to the cluster-admin role :
kubectl delete clusterrolebinding [name]


##### CIS Control 5.1.2

**Description:**

Minimize access to secrets


**Remediation:**

Where possible, remove get, list and watch access to Secret
objects in the cluster.


##### CIS Control 5.1.3

**Description:**

Minimize wildcard use in Roles and ClusterRoles


**Remediation:**

Where possible replace any use of wildcards in clusterroles
and roles with specific
objects or actions.


##### CIS Control 5.1.4

**Description:**

Minimize access to create pods


**Remediation:**

Where possible, remove create access to pod objects in the
cluster.


##### CIS Control 5.1.5

**Description:**

Ensure that default service accounts are not actively used.


**Remediation:**

Create explicit service accounts wherever a Kubernetes
workload requires specific access
to the Kubernetes API server.
Modify the configuration of each default service account to
include this value
automountServiceAccountToken: false


##### CIS Control 5.1.6

**Description:**

Ensure that Service Account Tokens are only mounted where
necessary


**Remediation:**

Modify the definition of pods and service accounts which do
not need to mount service
account tokens to disable it.


##### CIS Control 5.1.7

**Description:**

Avoid use of system:masters group


**Remediation:**

Remove the system:masters group from all users in the
cluster.


##### CIS Control 5.1.8

**Description:**

Limit use of the Bind, Impersonate and Escalate permissions
in the Kubernetes cluster


**Remediation:**

Where possible, remove the impersonate, bind and escalate
rights from subjects.


#### Pod security standards

##### CIS Control 5.2.1 / DISA STIG V-254800

**Description:**

Ensure that the cluster has at least one active policy
control mechanism in place


**Remediation:**

Ensure that either Pod Security Admission or an external
policy control system is in place
for every namespace which contains user workloads.


##### CIS Control 5.2.2 / DISA STIG V-254801

**Description:**

Minimize the admission of privileged containers


**Remediation:**

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of privileged containers.


##### CIS Control 5.2.3

**Description:**

Minimize the admission of containers wishing to share the
host process ID namespace


**Remediation:**

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of `hostPID` containers.


##### CIS Control 5.2.4

**Description:**

Minimize the admission of containers wishing to share the
host IPC namespace


**Remediation:**

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of `hostIPC` containers.


##### CIS Control 5.2.5

**Description:**

Minimize the admission of containers wishing to share the
host network namespace


**Remediation:**

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of `hostNetwork` containers.


##### CIS Control 5.2.6

**Description:**

Minimize the admission of containers with
allowPrivilegeEscalation


**Remediation:**

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers with
`.spec.allowPrivilegeEscalation` set to `true`.


##### CIS Control 5.2.7

**Description:**

Minimize the admission of root containers


**Remediation:**

Create a policy for each namespace in the cluster, ensuring
that either `MustRunAsNonRoot`
or `MustRunAs` with the range of UIDs not including 0, is
set.


##### CIS Control 5.2.8

**Description:**

Minimize the admission of containers with the NET_RAW
capability


**Remediation:**

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers with the `NET_RAW` capability.


##### CIS Control 5.2.9

**Description:**

Minimize the admission of containers with added capabilities


**Remediation:**

Ensure that `allowedCapabilities` is not present in policies
for the cluster unless
it is set to an empty array.


##### CIS Control 5.2.10

**Description:**

Minimize the admission of containers with capabilities
assigned


**Remediation:**

Review the use of capabilities in applications running on
your cluster. Where a namespace
contains applicaions which do not require any Linux
capabilities to operate consider adding
a PSP which forbids the admission of containers which do not
drop all capabilities.


##### CIS Control 5.2.11

**Description:**

Minimize the admission of Windows HostProcess containers


**Remediation:**

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers that have
`.securityContext.windowsOptions.hostProcess` set to `true`.


##### CIS Control 5.2.12

**Description:**

Minimize the admission of HostPath volumes


**Remediation:**

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers with `hostPath` volumes.


##### CIS Control 5.2.13

**Description:**

Minimize the admission of containers which use HostPorts


**Remediation:**

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers which use `hostPort` sections.


#### Network policies and CNI

##### CIS Control 5.3.1

**Description:**

Ensure that the CNI in use supports NetworkPolicies


**Remediation:**

If the CNI plugin in use does not support network policies,
consideration should be given to
making use of a different plugin, or finding an alternate
mechanism for restricting traffic
in the Kubernetes cluster.


##### CIS Control 5.3.2

**Description:**

Ensure that all Namespaces have NetworkPolicies defined


**Remediation:**

Follow the documentation and create NetworkPolicy objects as
you need them.


#### Secrets management

##### CIS Control 5.4.1

**Description:**

Prefer using Secrets as files over Secrets as environment
variables


**Remediation:**

If possible, rewrite application code to read Secrets from
mounted secret files, rather than
from environment variables.


##### CIS Control 5.4.2

**Description:**

Consider external secret storage


**Remediation:**

Refer to the Secrets management options offered by your
cloud provider or a third-party
secrets management solution.


#### Extensible admission control

##### CIS Control 5.5.1

**Description:**

Configure Image Provenance using ImagePolicyWebhook
admission controller


**Remediation:**

Follow the Kubernetes documentation and setup image
provenance.


#### General policies

##### CIS Control 5.7.1

**Description:**

Create administrative boundaries between resources using
namespaces


**Remediation:**

Follow the documentation and create namespaces for objects
in your deployment as you need
them.


##### CIS Control 5.7.2

**Description:**

Ensure that the seccomp profile is set to docker/default in
your Pod definitions


**Remediation:**

Use `securityContext` to enable the docker/default seccomp
profile in your pod definitions.
An example is as follows:

```
  securityContext:
    seccompProfile:
      type: RuntimeDefault
```

##### CIS Control 5.7.3

**Description:**

Apply SecurityContext to your Pods and Containers


**Remediation:**

Follow the Kubernetes documentation and apply
SecurityContexts to your Pods. For a
suggested list of SecurityContexts, you may refer to the CIS
Security Benchmark for Docker
Containers.


##### CIS Control 5.7.4

**Description:**

The default namespace should not be used


**Remediation:**

Ensure that namespaces are created to allow for appropriate
segregation of Kubernetes
resources and that all new resources are created in a
specific namespace.


<!-- Links -->
[Center for Internet Security (CIS)]:https://www.cisecurity.org/
[kube-bench]:https://aquasecurity.github.io/kube-bench/v0.6.15/
[CIS Kubernetes Benchmark]:https://www.cisecurity.org/benchmark/kubernetes
[getting started]: ../../tutorial/getting-started
[kube-bench release]: https://github.com/aquasecurity/kube-bench/releases
[post-deployment hardening]: hardening.md
