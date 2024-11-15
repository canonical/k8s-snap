# CIS compliance 

CIS Hardening refers to the process of implementing security configurations that
align with the benchmarks set by the [Center for Internet Security (CIS)][].
Out of the box {{product}} complies with the majority of the recommended
CIS security configurations. Since implementing all security recommendations
would comes at the expense of compatibility and/or performance we expect
cluster administrators to follow post deployment hardening steps based on their
needs. This guide covers:

  * post deployment harden steps you could consider for your {{product}}
  * use [kube-bench][] to automatically check whether your Kubernetes clusters
   are configured according to the [CIS Kubernetes Benchmark][]
  * manually configure and audit each configuration CIS hardening recommendation


## What you'll need

This guide assumes the following:

- You have a bootstrapped {{product}} cluster (see the [Getting Started]
[getting-started-guide] guide)
- You have root or sudo access to the machine


## Post-deployment extra hardening steps

The following hardening configurations are not part of the default {{product}} setup
as they might:

  * impact performance
  * affect the compatibility with workloads
  * require input from the administrator

Please, consider the effects of each configuration suggested before applying it.

### Control plane nodes

#### Configure log auditing

```{note}
Configuring log auditing requires the cluster administrator's input and
may incurr performance penalties in the form of disk I/O.
```

Create an audit-policy.yaml file under `/var/snap/k8s/common/etc/` and specify 
the level of auditing you desire based on the [upstream instructions][]. Here is 
a minimal example of such a policy file.

```
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
--audit-log-path=/var/log/apiserver/audit.log
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


Create a configuration file with the [rate limits][] and place it under 
`/var/snap/k8s/common/etc/`.
For example:

```
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
This configuration may affect compatibility with workloads and metrics collection.
```

Edit the Kubernetes scheduler arguments file `/var/snap/k8s/common/args/kube-scheduler`
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

Run the following commands on nodes that host workloads. In the default deployment
the control plane nodes functions as workers and they may need to be hardened.

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
Fully complying with the spirit of this hardening recommendation calls for systemd configuration
that is out of the scope of this documentation page. 
```

Ensure that only the owner of `/etc/systemd/system/snap.k8s.kubelet.service`
has full read and write access to it. Setting the kubelet service file permission
needs to be performed every time the k8s snap refreshes.

```
chmod 600 /etc/systemd/system/snap.k8s.kubelet.service
```

Restart `kubelet`.

```
sudo systemctl restart snap.k8s.kubelet
``` 



## Assess CIS hardening with kube-bench 

Download the latest [kube-bench release][] on your Kubernetes nodes. Make sure 
to select the appropriate binary version.

For example, to download the Linux binary, use the following command. Replace 
`KB` by the version listed in the releases page.

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

Verify kube-bench installation.

```
kube-bench version
``` 

The output should list the version installed.

Install `kubectl` and configure it to interact with the cluster.

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

In what follows we iterate over all CIS hardening recommendations
and, when possible, provide information on how to comply with each
one manually. This can be used for manually auditing the CIS
hardening state of a cluster.

### Control Plane Security Configuration

#### Control Plane Node Configuration Files

##### Control 1.1.1

###### Description:

Ensure that the API server configuration file permissions are
set to 600 (Automated)


###### Remediation:

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/kube-apiserver`


###### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c permissions=%a /var/snap/k8s/common/args/kube-apiserver; fi'
```

###### Expected output:

```
permissions=600
```

##### Control 1.1.2

###### Description:

Ensure that the API server configuration file ownership is set
to root:root (Automated)


###### Remediation:

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/kube-apiserver`


###### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-apiserver; then stat -c %U:%G /var/snap/k8s/common/args/kube-apiserver; fi'
```

###### Expected output:

```
root:root
```

##### Control 1.1.3

###### Description:

Ensure that the controller manager configuration file
permissions are set to 600 (Automated)


###### Remediation:

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/kube-controller-manager`


###### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c permissions=%a /var/snap/k8s/common/args/kube-controller-manager; fi'
```

###### Expected output:

```
permissions=600
```

##### Control 1.1.4

###### Description:

Ensure that the controller manager configuration file ownership
is set to root:root (Automated)


###### Remediation:

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/kube-controller-manager`


###### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-controller-manager; then stat -c %U:%G /var/snap/k8s/common/args/kube-controller-manager; fi'
```

###### Expected output:

```
root:root
```

##### Control 1.1.5

###### Description:

Ensure that the scheduler configuration file permissions are set
to 600 (Automated)


###### Remediation:

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/kube-scheduler`


###### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c permissions=%a /var/snap/k8s/common/args/kube-scheduler; fi'
```

###### Expected output:

```
permissions=600
```

##### Control 1.1.6

###### Description:

Ensure that the scheduler configuration file ownership is set to
root:root (Automated)


###### Remediation:

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/kube-scheduler`


###### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kube-scheduler; then stat -c %U:%G /var/snap/k8s/common/args/kube-scheduler; fi'
```

###### Expected output:

```
root:root
```

##### Control 1.1.7

###### Description:

Ensure that the dqlite configuration file permissions are set to
644 or more restrictive (Automated)


###### Remediation:

Run the following command on the control plane node.

`chmod 600 /var/snap/k8s/common/args/k8s-dqlite`


###### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/k8s-dqlite; then stat -c permissions=%a /var/snap/k8s/common/args/k8s-dqlite; fi'
```

###### Expected output:

```
permissions=600
```

##### Control 1.1.8

###### Description:

Ensure that the dqlite configuration file ownership is set to
root:root (Automated)


###### Remediation:

Run the following command on the control plane node.

`chown root:root /var/snap/k8s/common/args/k8s-dqlite`


###### Audit (as root):

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/k8s-dqlite; then stat -c %U:%G /var/snap/k8s/common/args/k8s-dqlite; fi'
```

###### Expected output:

```
root:root
```

##### Control 1.1.9

###### Description:

Ensure that the Container Network Interface file permissions are
set to 600 (Manual)


###### Remediation:

Run the following command on the control plane node.

`chmod 600 /etc/cni/net.d/05-cilium.conflist`


###### Audit (as root):

```
ps -ef | grep bin/kubelet | grep -- --cni-conf-dir | sed 's%.*cni-conf-dir[= ]\([^ ]*\).*%\1%' | xargs -I{} find {} -mindepth 1 | xargs --no-run-if-empty stat -c permissions=%a
find /etc/cni/net.d/05-cilium.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c permissions=%a
```

###### Expected output:

```
permissions=600
```

##### Control 1.1.10

###### Description:

Ensure that the Container Network Interface file ownership is
set to root:root (Manual)


###### Remediation:

Run the following command on the control plane node.

`chown root:root /etc/cni/net.d/05-cilium.conflist`


###### Audit (as root):

```
find /etc/cni/net.d/05-cilium.conflist -type f 2> /dev/null | xargs --no-run-if-empty stat -c %U:%G
```

###### Expected output:

```
root:root
```

##### Control 1.1.11

###### Description:

Ensure that the dqlite data directory permissions are set to 700
or more restrictive (Automated)


###### Remediation:

Dqlite data are kept by default under
/var/snap/k8s/common/var/lib/k8s-dqlite.
To comply with the spirit of this CIS recommendation:

`chmod 700 /var/snap/k8s/common/var/lib/k8s-dqlite`


###### Audit (as root):

```
DATA_DIR='/var/snap/k8s/common/var/lib/k8s-dqlite'
stat -c permissions=%a "$DATA_DIR"
```

###### Expected output:

```
permissions=700
```

##### Control 1.1.12

###### Description:

Ensure that the dqlite data directory ownership is set to
root:root (Automated)


###### Remediation:

Dqlite data are kept by default under
/var/snap/k8s/common/var/lib/k8s-dqlite.
To comply with the spirit of this CIS recommendation:

`chown root:root /var/snap/k8s/common/var/lib/k8s-dqlite`


###### Audit (as root):

```
DATA_DIR='/var/snap/k8s/common/var/lib/k8s-dqlite'
stat -c %U:%G "$DATA_DIR"
```

###### Expected output:

```
root:root
```

##### Control 1.1.13

###### Description:

Ensure that the admin.conf file permissions are set to 600
(Automated)


###### Remediation:

Run the following command on the control plane node.

`chmod 600 /etc/kubernetes/admin.conf`


###### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c permissions=%a /etc/kubernetes/admin.conf; fi'
```

###### Expected output:

```
permissions=600
```

##### Control 1.1.14

###### Description:

Ensure that the admin.conf file ownership is set to root:root
(Automated)


###### Remediation:

Run the following command on the control plane node.

`chown root:root /etc/kubernetes/admin.conf`


###### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/admin.conf; then stat -c %U:%G /etc/kubernetes/admin.conf; fi'
```

###### Expected output:

```
root:root
```

##### Control 1.1.15

###### Description:

Ensure that the scheduler.conf file permissions are set to 600
(Automated)


###### Remediation:

Run the following command on the control plane node.

`chmod 600 /etc/kubernetes/scheduler.conf`


###### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c permissions=%a /etc/kubernetes/scheduler.conf; fi'
```

###### Expected output:

```
permissions=600
```

##### Control 1.1.16

###### Description:

Ensure that the scheduler.conf file ownership is set to
root:root (Automated)


###### Remediation:

Run the following command on the control plane node.

`chown root:root /etc/kubernetes/scheduler.conf`


###### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/scheduler.conf; then stat -c %U:%G /etc/kubernetes/scheduler.conf; fi'
```

###### Expected output:

```
root:root
```

##### Control 1.1.17

###### Description:

Ensure that the controller-manager.conf file permissions are set
to 600 (Automated)


###### Remediation:

Run the following command on the control plane node.

`chmod 600 /etc/kubernetes/controller.conf`


###### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c permissions=%a /etc/kubernetes/controller.conf; fi'
```

###### Expected output:

```
permissions=600
```

##### Control 1.1.18

###### Description:

Ensure that the controller-manager.conf file ownership is set to
root:root (Automated)


###### Remediation:

Run the following command on the control plane node.

`chown root:root /etc/kubernetes/controller.conf`


###### Audit (as root):

```
/bin/sh -c 'if test -e /etc/kubernetes/controller.conf; then stat -c %U:%G /etc/kubernetes/controller.conf; fi'
```

###### Expected output:

```
root:root
```

##### Control 1.1.19

###### Description:

Ensure that the Kubernetes PKI directory and file ownership is
set to root:root (Automated)


###### Remediation:

Run the following command on the control plane node.

`chown -R root:root /etc/kubernetes/pki/`


###### Audit (as root):

```
find /etc/kubernetes/pki/ | xargs stat -c %U:%G
```

###### Expected output:

```
root:root
```

##### Control 1.1.20

###### Description:

Ensure that the Kubernetes PKI certificate file permissions are
set to 600 (Manual)


###### Remediation:

Run the following command on the control plane node.

`chmod -R 600 /etc/kubernetes/pki/*.crt`


###### Audit (as root):

```
find /etc/kubernetes/pki/ -name '*.crt' | xargs stat -c permissions=%a
```

###### Expected output:

```
permissions=600
```

##### Control 1.1.21

###### Description:

Ensure that the Kubernetes PKI key file permissions are set to
600 (Manual)


###### Remediation:

Run the following command on the control plane node.

`chmod -R 600 /etc/kubernetes/pki/*.key`


###### Audit (as root):

```
find /etc/kubernetes/pki/ -name '*.key' | xargs stat -c permissions=%a
```

###### Expected output:

```
permissions=600
```

#### API Server

##### Control 1.2.1

###### Description:

Ensure that the --anonymous-auth argument is set to false
(Manual)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the following argument.

`--anonymous-auth=false`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--anonymous-auth=false
```

##### Control 1.2.2

###### Description:

Ensure that the --token-auth-file parameter is not set
(Automated)


###### Remediation:

Follow the documentation and configure alternate mechanisms for
authentication. Then,
edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and remove the --token-auth-file
argument.


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--token-auth-file is not set
```

##### Control 1.2.3

###### Description:

Ensure that the --DenyServiceExternalIPs is not set (Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and remove the
`DenyServiceExternalIPs`
from enabled admission plugins.


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--enable-admission-plugins does not contain
DenyServiceExternalIPs
```

##### Control 1.2.4

###### Description:

Ensure that the --kubelet-client-certificate and --kubelet-
client-key arguments are set as appropriate (Automated)


###### Remediation:

Follow the Kubernetes documentation and set up the TLS
connection between the
apiserver and kubelets. Then, edit API server configuration file
/var/snap/k8s/common/args/kube-apiserver on the control plane node and set the
kubelet client certificate and key parameters as follows.

```
--kubelet-client-certificate=<path/to/client-certificate-file>
--kubelet-client-key=<path/to/client-key-file>
```


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-
kubelet-client.crt and --kubelet-client-
key=/etc/kubernetes/pki/apiserver-kubelet-client.key
```

##### Control 1.2.5

###### Description:

Ensure that the --kubelet-certificate-authority argument is set
as appropriate (Automated)


###### Remediation:

Follow the Kubernetes documentation and setup the TLS connection
between
the apiserver and kubelets. Then, edit the API server
configuration file
/var/snap/k8s/common/args/kube-apiserver on the control plane node and set the
--kubelet-certificate-authority parameter to the path to the
cert file for the certificate authority.

`--kubelet-certificate-authority=<ca-string>`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--kubelet-certificate-authority=/etc/kubernetes/pki/ca.crt
```

##### Control 1.2.6

###### Description:

Ensure that the --authorization-mode argument is not set to
AlwaysAllow (Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode
parameter to values other than AlwaysAllow.
One such example could be as follows.

`--authorization-mode=Node,RBAC`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--authorization-mode=Node,RBAC
```

##### Control 1.2.7

###### Description:

Ensure that the --authorization-mode argument includes Node
(Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode
parameter to a value that includes Node.

`--authorization-mode=Node,RBAC`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--authorization-mode=Node,RBAC
```

##### Control 1.2.8

###### Description:

Ensure that the --authorization-mode argument includes RBAC
(Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --authorization-mode
parameter to a value that includes RBAC,

`--authorization-mode=Node,RBAC`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--authorization-mode=Node,RBAC
```

##### Control 1.2.9

###### Description:

Ensure that the admission control plugin EventRateLimit is set
(Manual)


###### Remediation:

Follow the Kubernetes documentation and set the desired limits
in a configuration file.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
and set the following arguments.

```
--enable-admission-plugins=...,EventRateLimit,...
--admission-control-config-file=<path/to/configuration/file>
```


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### Control 1.2.10

###### Description:

Ensure that the admission control plugin AlwaysAdmit is not set
(Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and either remove the --enable-
admission-plugins parameter, or set it to a
value that does not include AlwaysAdmit.


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### Control 1.2.11

###### Description:

Ensure that the admission control plugin AlwaysPullImages is set
(Manual)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-plugins
parameter to include
AlwaysPullImages.

`--enable-admission-plugins=...,AlwaysPullImages,...`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### Control 1.2.12

###### Description:

Ensure that the admission control plugin SecurityContextDeny is
set if PodSecurityPolicy is not used (Manual)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-plugins
parameter to include
SecurityContextDeny, unless PodSecurityPolicy is already in
place.

`--enable-admission-plugins=...,SecurityContextDeny,...`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### Control 1.2.13

###### Description:

Ensure that the admission control plugin ServiceAccount is set
(Automated)


###### Remediation:

Follow the documentation and create ServiceAccount objects as
per your environment.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and ensure that the --disable-
admission-plugins parameter is set to a
value that does not include ServiceAccount.


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--disable-admission-plugins is not set
```

##### Control 1.2.14

###### Description:

Ensure that the admission control plugin NamespaceLifecycle is
set (Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --disable-admission-
plugins parameter to
ensure it does not include NamespaceLifecycle.


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--disable-admission-plugins is not set
```

##### Control 1.2.15

###### Description:

Ensure that the admission control plugin NodeRestriction is set
(Automated)


###### Remediation:

Follow the Kubernetes documentation and configure
NodeRestriction plug-in on kubelets.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --enable-admission-plugins
parameter to a
value that includes NodeRestriction.

`--enable-admission-plugins=...,NodeRestriction,...`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--enable-admission-
plugins=NodeRestriction,EventRateLimit,AlwaysPullImages
```

##### Control 1.2.16

###### Description:

Ensure that the --secure-port argument is not set to 0
(Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and either remove the --secure-port
parameter or
set it to a different (non-zero) desired port.


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--secure-port=6443
```

##### Control 1.2.17

###### Description:

Ensure that the --profiling argument is set to false (Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the following argument.

`--profiling=false`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--profiling=false
```

##### Control 1.2.18

###### Description:

Ensure that the --audit-log-path argument is set (Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-path parameter
to a suitable path and
file where you would like audit logs to be written.

`--audit-log-path=/var/log/apiserver/audit.log`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--audit-log-path=/var/log/apiserver/audit.log
```

##### Control 1.2.19

###### Description:

Ensure that the --audit-log-maxage argument is set to 30 or as
appropriate (Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxage
parameter to 30
or as an appropriate number of days.

`--audit-log-maxage=30`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--audit-log-maxage=30
```

##### Control 1.2.20

###### Description:

Ensure that the --audit-log-maxbackup argument is set to 10 or
as appropriate (Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxbackup
parameter to 10 or to an appropriate
value.

`--audit-log-maxbackup=10`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--audit-log-maxbackup=10
```

##### Control 1.2.21

###### Description:

Ensure that the --audit-log-maxsize argument is set to 100 or as
appropriate (Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --audit-log-maxsize
parameter to an appropriate size in MB.

`--audit-log-maxsize=100`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--audit-log-maxsize=100
```

##### Control 1.2.22

###### Description:

Ensure that the --request-timeout argument is set as appropriate
(Manual)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
and set the following argument as appropriate and if needed.

`--request-timeout=300s`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--request-timeout=300s
```

##### Control 1.2.23

###### Description:

Ensure that the --service-account-lookup argument is set to true
(Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the following argument.

`--service-account-lookup=true`

Alternatively, you can delete the --service-account-lookup
argument from this file so
that the default takes effect.


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--service-account-lookup is not set
```

##### Control 1.2.24

###### Description:

Ensure that the --service-account-key-file argument is set as
appropriate (Automated)


###### Remediation:

Edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --service-account-key-file
parameter
to the public key file for service accounts.

`--service-account-key-file=<filename>`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--service-account-key-
file=/etc/kubernetes/pki/serviceaccount.key
```

##### Control 1.2.25

###### Description:

Ensure that the --etcd-certfile and --etcd-keyfile arguments are
set as appropriate (Automated)


###### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### Control 1.2.26

###### Description:

Ensure that the --tls-cert-file and --tls-private-key-file
arguments are set as appropriate (Automated)


###### Remediation:

Follow the Kubernetes documentation and set up the TLS
connection on the apiserver.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the TLS certificate and
private key file parameters.

```
--tls-cert-file=<path/to/tls-certificate-file>
--tls-private-key-file=<path/to/tls-key-file>
```


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--tls-cert-file=/etc/kubernetes/pki/apiserver.crt and --tls-
private-key-file=/etc/kubernetes/pki/apiserver.key
```

##### Control 1.2.27

###### Description:

Ensure that the --client-ca-file argument is set as appropriate
(Automated)


###### Remediation:

Follow the Kubernetes documentation and set up the TLS
connection on the apiserver.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the client certificate
authority file.

`--client-ca-file=<path/to/client-ca-file>`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--client-ca-file=/etc/kubernetes/pki/client-ca.crt
```

##### Control 1.2.28

###### Description:

Ensure that the --etcd-cafile argument is set as appropriate
(Automated)


###### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


###### Expected output:

```
--etcd-cafile=/etc/kubernetes/pki/etcd/ca.crt
```

##### Control 1.2.29

###### Description:

Ensure that the --encryption-provider-config argument is set as
appropriate (Manual)


###### Remediation:

Follow the Kubernetes documentation and configure a
EncryptionConfig file.
Then, edit the API server configuration file /var/snap/k8s/common/args/kube-apiserver
on the control plane node and set the --encryption-provider-
config parameter to the path of that file.

`--encryption-provider-config=</path/to/EncryptionConfig/File>`


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--encryption-provider-config is set
```

##### Control 1.2.30

###### Description:

Ensure that encryption providers are appropriately configured
(Manual)


###### Remediation:

Follow the Kubernetes documentation and configure a
EncryptionConfig file.
In this file, choose aescbc, kms or secretbox as the encryption
provider.


###### Audit (as root):

```
ENCRYPTION_PROVIDER_CONFIG=$(ps -ef | grep kube-apiserver | grep -- --encryption-provider-config | sed 's%.*encryption-provider-config[= ]\([^ ]*\).*%\1%')
if test -e $ENCRYPTION_PROVIDER_CONFIG; then grep -A1 'providers:' $ENCRYPTION_PROVIDER_CONFIG | tail -n1 | grep -o "[A-Za-z]*" | sed 's/^/provider=/'; fi
```

###### Expected output:

```
--encryption-provider-config is one of or all of
aescbc,kms,secretbox
```

##### Control 1.2.31

###### Description:

Ensure that the API Server only makes use of Strong
Cryptographic Ciphers (Manual)


###### Remediation:

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


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

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

#### Controller Manager

##### Control 1.3.1

###### Description:

Ensure that the --terminated-pod-gc-threshold argument is set as
appropriate (Manual)


###### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --terminated-pod-gc-
threshold to an appropriate threshold.

`--terminated-pod-gc-threshold=10`


###### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

###### Expected output:

```
--terminated-pod-gc-threshold=12500
```

##### Control 1.3.2

###### Description:

Ensure that the --profiling argument is set to false (Automated)


###### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the following argument.

`--profiling=false`


###### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

###### Expected output:

```
--profiling=false
```

##### Control 1.3.3

###### Description:

Ensure that the --use-service-account-credentials argument is
set to true (Automated)


###### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node to set the following argument.

`--use-service-account-credentials=true`


###### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

###### Expected output:

```
--user-service-account-credentials=true
```

##### Control 1.3.4

###### Description:

Ensure that the --service-account-private-key-file argument is
set as appropriate (Automated)


###### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --service-account-private-
key-file parameter
to the private key file for service accounts.

`--service-account-private-key-file=<filename>`


###### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

###### Expected output:

```
--service-account-private-key-
file=/etc/kubernetes/pki/serviceaccount.key
```

##### Control 1.3.5

###### Description:

Ensure that the --root-ca-file argument is set as appropriate
(Automated)


###### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --root-ca-file parameter
to the certificate bundle file.

`--root-ca-file=<path/to/file>`


###### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

###### Expected output:

```
--root-ca-file=/etc/kubernetes/pki/ca.crt
```

##### Control 1.3.6

###### Description:

Ensure that the RotateKubeletServerCertificate argument is set
to true (Automated)


###### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and set the --feature-gates parameter
to include RotateKubeletServerCertificate=true.

`--feature-gates=RotateKubeletServerCertificate=true`


###### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

###### Expected output:

```
--feature-gates is not set
```

##### Control 1.3.7

###### Description:

Ensure that the --bind-address argument is set to 127.0.0.1
(Automated)


###### Remediation:

Edit the Controller Manager configuration file
/var/snap/k8s/common/args/kube-controller-manager
on the control plane node and ensure the correct value for the
--bind-address parameter
and restart the controller manager service


###### Audit (as root):

```
/bin/ps -ef | grep kube-controller-manager | grep -v grep
```

###### Expected output:

```
---bind-address=127.0.0.1
```

#### Scheduler

##### Control 1.4.1

###### Description:

Ensure that the --profiling argument is set to false (Automated)


###### Remediation:

Edit the Scheduler configuration file /var/snap/k8s/common/args/kube-scheduler file
on the control plane node and set the following argument.

`--profiling=false`


###### Audit (as root):

```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

###### Expected output:

```
--profiling=false
```

##### Control 1.4.2

###### Description:

Ensure that the --bind-address argument is set to 127.0.0.1
(Automated)


###### Remediation:

Edit the Scheduler configuration file /var/snap/k8s/common/args/kube-scheduler
on the control plane node and ensure the correct value for the
--bind-address parameter
and restart the kube-scheduler service


###### Audit (as root):

```
/bin/ps -ef | grep kube-scheduler | grep -v grep
```

###### Expected output:

```
---bind-address=127.0.0.1
```

### Datastore Node Configuration

#### Datastore Node Configuration

##### Control 2.1

###### Description:

Ensure that the --cert-file and --key-file arguments are set as
appropriate (Automated)


###### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### Control 2.2

###### Description:

Ensure that the --client-cert-auth argument is set to true
(Automated)


###### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### Control 2.3

###### Description:

Ensure that the --auto-tls argument is not set to true
(Automated)


###### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


##### Control 2.4

###### Description:

Ensure that the --peer-cert-file and --peer-key-file arguments
are set as appropriate (Automated)


###### Remediation:

The certificate pair for dqlite and tls peer communication is
/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt and
/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key.


###### Audit (as root):

```
if test -e /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt && test -e /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key; then echo 'certs-found'; fi
```

###### Expected output:

```
certs-found
```

##### Control 2.5

###### Description:

Ensure that the --peer-client-cert-auth argument is set to true
(Automated)


###### Remediation:

Dqlite peer communication uses TLS unless the --enable-tls is
set to false in
/var/snap/k8s/common/args/k8s-dqlite.


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/k8s-dqlite | /bin/grep enable-tls || true; echo $?
```

###### Expected output:

```
0
```

##### Control 2.6

###### Description:

Ensure that the --peer-auto-tls argument is not set to true
(Automated)


###### Remediation:

Not applicable. Canonical K8s uses dqlite and tls peer
communication uses the certificates
created upon the snap creation.


##### Control 2.7

###### Description:

Ensure that a unique Certificate Authority is used for the
datastore (Manual)


###### Remediation:

Not applicable. Canonical K8s uses dqlite and tls peer
communication uses certificates
created upon cluster setup.


### Control Plane Configuration

#### Authentication and Authorization

##### Control 3.1.1

###### Description:

Client certificate authentication should not be used for users
(Manual)


###### Remediation:

Alternative mechanisms provided by Kubernetes such as the use of
OIDC should be
implemented in place of client certificates.


#### Logging

##### Control 3.2.1

###### Description:

Ensure that a minimal audit policy is created (Manual)


###### Remediation:

Create an audit policy file for your cluster.


###### Audit (as root):

```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

###### Expected output:

```
--audit-policy-file=/var/snap/k8s/common/etc/audit-policy.yaml
```

##### Control 3.2.2

###### Description:

Ensure that the audit policy covers key security concerns
(Manual)


###### Remediation:

Review the audit policy provided for the cluster and ensure that
it covers
at least the following areas,
- Access to Secrets managed by the cluster. Care should be taken
to only
  log Metadata for requests to Secrets, ConfigMaps, and
TokenReviews, in
  order to avoid risk of logging sensitive data.
- Modification of Pod and Deployment objects.
- Use of `pods/exec`, `pods/portforward`, `pods/proxy` and
`services/proxy`.
For most requests, minimally logging at the Metadata level is
recommended
(the most basic level of logging).


### Worker Node Security Configuration

#### Worker Node Configuration Files

##### Control 4.1.1

###### Description:

Ensure that the kubelet service file permissions are set to 600
(Automated)


###### Remediation:

Run the following command on each worker node.


`chmod 600 /etc/systemd/system/snap.k8s.kubelet.service`


###### Audit (as root):

```
/bin/sh -c "if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c permissions=%a /etc/systemd/system/snap.k8s.kubelet.service; fi"
```

###### Expected output:

```
permissions=600
```

##### Control 4.1.2

###### Description:

Ensure that the kubelet service file ownership is set to
root:root (Automated)


###### Remediation:

Run the following command on each worker node.


`chown root:root /etc/systemd/system/snap.k8s.kubelet.service`


###### Audit (as root):

```
/bin/sh -c "if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c %U:%G /etc/systemd/system/snap.k8s.kubelet.service; else echo \"File not found\"; fi"
```

###### Expected output:

```
root:root
```

##### Control 4.1.3

###### Description:

If proxy kubeconfig file exists ensure permissions are set to
600 (Manual)


###### Remediation:

Run the following command on each worker node.


`chmod 600 /etc/kubernetes/proxy.conf`


###### Audit (as root):

```
/bin/sh -c "if test -e /etc/kubernetes/proxy.conf; then stat -c permissions=%a /etc/kubernetes/proxy.conf; fi"
```

###### Expected output:

```
permissions=600
```

##### Control 4.1.4

###### Description:

If proxy kubeconfig file exists ensure ownership is set to
root:root (Manual)


###### Remediation:

Run the following command on each worker node.


`chown root:root /etc/kubernetes/proxy.conf`


###### Audit (as root):

```
/bin/sh -c "if test -e /etc/kubernetes/proxy.conf; then stat -c %U:%G /etc/kubernetes/proxy.conf; fi"
```

###### Expected output:

```
root:root
```

##### Control 4.1.5

###### Description:

Ensure that the --kubeconfig kubelet.conf file permissions are
set to 600 (Automated)


###### Remediation:

Run the following command on each worker node.


`chmod 600 /etc/kubernetes/kubelet.conf`


###### Audit (as root):

```
/bin/sh -c "if test -e /etc/kubernetes/kubelet.conf; then stat -c permissions=%a /etc/kubernetes/kubelet.conf; fi"
```

###### Expected output:

```
permissions=600
```

##### Control 4.1.6

###### Description:

Ensure that the --kubeconfig kubelet.conf file ownership is set
to root:root (Automated)


###### Remediation:

Run the following command on each worker node.


`chown root:root /etc/kubernetes/kubelet.conf`


###### Audit (as root):

```
/bin/sh -c "if test -e /etc/kubernetes/kubelet.conf; then stat -c %U:%G /etc/kubernetes/kubelet.conf; fi"
```

###### Expected output:

```
root:root
```

##### Control 4.1.7

###### Description:

Ensure that the certificate authorities file permissions are set
to 600 (Manual)


###### Remediation:

Run the following command to modify the file permissions of the
--client-ca-file.

`chmod 600 /etc/kubernetes/pki/client-ca.crt`


###### Audit (as root):

```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c permissions=%a $CAFILE; fi
```

###### Expected output:

```
permissions=600
```

##### Control 4.1.8

###### Description:

Ensure that the client certificate authorities file ownership is
set to root:root (Manual)


###### Remediation:

Run the following command to modify the ownership of the
--client-ca-file.

`chown root:root /etc/kubernetes/pki/client-ca.crt`


###### Audit (as root):

```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c %U:%G $CAFILE; fi
```

###### Expected output:

```
root:root
```

##### Control 4.1.9

###### Description:

If the kubelet config.yaml configuration file is being used
validate permissions set to 600 (Manual)


###### Remediation:

Run the following command (using the config file location
identified in the Audit step)


`chmod 600 /var/snap/k8s/common/args/kubelet`


###### Audit (as root):

```
/bin/sh -c "if test -e /var/snap/k8s/common/args/kubelet; then stat -c permissions=%a /var/snap/k8s/common/args/kubelet; fi"
```

###### Expected output:

```
permissions=600
```

##### Control 4.1.10

###### Description:

If the kubelet config.yaml configuration file is being used
validate file ownership is set to root:root (Manual)


###### Remediation:

Run the following command (using the config file location
identified in the Audit step)


`chown root:root /var/snap/k8s/common/args/kubelet`


###### Audit (as root):

```
/bin/sh -c "if test -e /var/snap/k8s/common/args/kubelet; then stat -c %U:%G /var/snap/k8s/common/args/kubelet; fi"
```

###### Expected output:

```
root:root
```

#### Kubelet

##### Control 4.2.1

###### Description:

Ensure that the --anonymous-auth argument is set to false
(Automated)


###### Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--anonymous-auth=false`

Restart the kubelet service.

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--anonymous-auth=false
```

##### Control 4.2.2

###### Description:

Ensure that the --authorization-mode argument is not set to
AlwaysAllow (Automated)


###### Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--authorization-mode=Webhook`

Restart the kubelet service:

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--authorization-mode=Webhook
```

##### Control 4.2.3

###### Description:

Ensure that the --client-ca-file argument is set as appropriate
(Automated)


###### Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--client-ca-file=/etc/kubernetes/pki/client-ca.crt`

Restart the kubelet service:

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--client-ca-file=/etc/kubernetes/pki/client-ca.crt
```

##### Control 4.2.4

###### Description:

Verify that the --read-only-port argument is set to 0 (Manual)


###### Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--read-only-port=0`

Restart the kubelet service:

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--read-only-port=0
```

##### Control 4.2.5

###### Description:

Ensure that the --streaming-connection-idle-timeout argument is
not set to 0 (Manual)


###### Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--streaming-connection-idle-timeout=5m`

Restart the kubelet service:

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--streaming-connection-idle-timeout is not set
```

##### Control 4.2.6

###### Description:

Ensure that the --protect-kernel-defaults argument is set to
true (Automated)


###### Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument:

`--protect-kernel-defaults=true`

Restart the kubelet service:

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--protect-kernel-defaults=true
```

##### Control 4.2.7

###### Description:

Ensure that the --make-iptables-util-chains argument is set to
true (Automated)


###### Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and
set the following argument:

`--make-iptables-util-chains=true`

Restart the kubelet service.

For example: `snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--make-iptables-util-chains is not set, or set to a value
greater or equal than 5m
```

##### Control 4.2.8

###### Description:

Ensure that the --hostname-override argument is not set (Manual)


###### Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet
on each worker node and remove the --hostname-override argument.

Restart the kubelet service.

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--hostname-override is set to false
```

##### Control 4.2.9

###### Description:

Ensure that the --event-qps argument is set to a level which
ensures appropriate event capture (Manual)


###### Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each worker
node and
set the --event-qps parameter as appropriate.

Restart the kubelet service.

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--event-qps is not set
```

##### Control 4.2.10

###### Description:

Ensure that the --tls-cert-file and --tls-private-key-file
arguments are set as appropriate (Manual)


###### Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each worker
node and
set the following arguments:

```
--tls-private-key-file=<path/to/private-key-file>
--tls-cert-file=<path/to/tls-certificate-file>
```

Restart the kubelet service.

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--tls-cert-file=/etc/kubernetes/pki/kubelet.crt and --tls-
private-key-file=/etc/kubernetes/pki/kubelet.key
```

##### Control 4.2.11

###### Description:

Ensure that the --rotate-certificates argument is not set to
false (Automated)


###### Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each worker
node and
remove the --rotate-certificates=false argument.

Restart the kubelet service.

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--rotate-certificates is not set
```

##### Control 4.2.12

###### Description:

Verify that the RotateKubeletServerCertificate argument is set
to true (Manual)


###### Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each worker
node and
set the argument --feature-
gates=RotateKubeletServerCertificate=true
on each worker node.

Restart the kubelet service.

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
the RotateKubeletServerCertificate feature gate is not set
```

##### Control 4.2.13

###### Description:

Ensure that the Kubelet only makes use of Strong Cryptographic
Ciphers (Manual)


###### Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each worker
node and
set the --tls-cipher-suites parameter as follows, or to a subset
of these values.

```
--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_
ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_
POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WIT
H_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_
RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256
```

Restart the kubelet service.

`snap restart k8s.kubelet`


###### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/kubelet
```

###### Expected output:

```
--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_
ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_
POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WIT
H_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_
RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256
```

### Kubernetes Policies

#### RBAC and Service Accounts

##### Control 5.1.1

###### Description:

Ensure that the cluster-admin role is only used where required
(Manual)


###### Remediation:

Identify all clusterrolebindings to the cluster-admin role.
Check if they are used and
if they need this role or if they could use a role with fewer
privileges.
Where possible, first bind users to a lower privileged role and
then remove the
clusterrolebinding to the cluster-admin role :
kubectl delete clusterrolebinding [name]


##### Control 5.1.2

###### Description:

Minimize access to secrets (Manual)


###### Remediation:

Where possible, remove get, list and watch access to Secret
objects in the cluster.


##### Control 5.1.3

###### Description:

Minimize wildcard use in Roles and ClusterRoles (Manual)


###### Remediation:

Where possible replace any use of wildcards in clusterroles and
roles with specific
objects or actions.


##### Control 5.1.4

###### Description:

Minimize access to create pods (Manual)


###### Remediation:

Where possible, remove create access to pod objects in the
cluster.


##### Control 5.1.5

###### Description:

Ensure that default service accounts are not actively used.
(Manual)


###### Remediation:

Create explicit service accounts wherever a Kubernetes workload
requires specific access
to the Kubernetes API server.
Modify the configuration of each default service account to
include this value
automountServiceAccountToken: false


##### Control 5.1.6

###### Description:

Ensure that Service Account Tokens are only mounted where
necessary (Manual)


###### Remediation:

Modify the definition of pods and service accounts which do not
need to mount service
account tokens to disable it.


##### Control 5.1.7

###### Description:

Avoid use of system:masters group (Manual)


###### Remediation:

Remove the system:masters group from all users in the cluster.


##### Control 5.1.8

###### Description:

Limit use of the Bind, Impersonate and Escalate permissions in
the Kubernetes cluster (Manual)


###### Remediation:

Where possible, remove the impersonate, bind and escalate rights
from subjects.


#### Pod Security Standards

##### Control 5.2.1

###### Description:

Ensure that the cluster has at least one active policy control
mechanism in place (Manual)


###### Remediation:

Ensure that either Pod Security Admission or an external policy
control system is in place
for every namespace which contains user workloads.


##### Control 5.2.2

###### Description:

Minimize the admission of privileged containers (Manual)


###### Remediation:

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of privileged containers.


##### Control 5.2.3

###### Description:

Minimize the admission of containers wishing to share the host
process ID namespace (Automated)


###### Remediation:

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of `hostPID` containers.


##### Control 5.2.4

###### Description:

Minimize the admission of containers wishing to share the host
IPC namespace (Automated)


###### Remediation:

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of `hostIPC` containers.


##### Control 5.2.5

###### Description:

Minimize the admission of containers wishing to share the host
network namespace (Automated)


###### Remediation:

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of `hostNetwork` containers.


##### Control 5.2.6

###### Description:

Minimize the admission of containers with
allowPrivilegeEscalation (Automated)


###### Remediation:

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers with `.spec.allowPrivilegeEscalation`
set to `true`.


##### Control 5.2.7

###### Description:

Minimize the admission of root containers (Automated)


###### Remediation:

Create a policy for each namespace in the cluster, ensuring that
either `MustRunAsNonRoot`
or `MustRunAs` with the range of UIDs not including 0, is set.


##### Control 5.2.8

###### Description:

Minimize the admission of containers with the NET_RAW capability
(Automated)


###### Remediation:

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers with the `NET_RAW` capability.


##### Control 5.2.9

###### Description:

Minimize the admission of containers with added capabilities
(Automated)


###### Remediation:

Ensure that `allowedCapabilities` is not present in policies for
the cluster unless
it is set to an empty array.


##### Control 5.2.10

###### Description:

Minimize the admission of containers with capabilities assigned
(Manual)


###### Remediation:

Review the use of capabilites in applications running on your
cluster. Where a namespace
contains applicaions which do not require any Linux capabities
to operate consider adding
a PSP which forbids the admission of containers which do not
drop all capabilities.


##### Control 5.2.11

###### Description:

Minimize the admission of Windows HostProcess containers
(Manual)


###### Remediation:

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers that have
`.securityContext.windowsOptions.hostProcess` set to `true`.


##### Control 5.2.12

###### Description:

Minimize the admission of HostPath volumes (Manual)


###### Remediation:

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers with `hostPath` volumes.


##### Control 5.2.13

###### Description:

Minimize the admission of containers which use HostPorts
(Manual)


###### Remediation:

Add policies to each namespace in the cluster which has user
workloads to restrict the
admission of containers which use `hostPort` sections.


#### Network Policies and CNI

##### Control 5.3.1

###### Description:

Ensure that the CNI in use supports NetworkPolicies (Manual)


###### Remediation:

If the CNI plugin in use does not support network policies,
consideration should be given to
making use of a different plugin, or finding an alternate
mechanism for restricting traffic
in the Kubernetes cluster.


##### Control 5.3.2

###### Description:

Ensure that all Namespaces have NetworkPolicies defined (Manual)


###### Remediation:

Follow the documentation and create NetworkPolicy objects as you
need them.


#### Secrets Management

##### Control 5.4.1

###### Description:

Prefer using Secrets as files over Secrets as environment
variables (Manual)


###### Remediation:

If possible, rewrite application code to read Secrets from
mounted secret files, rather than
from environment variables.


##### Control 5.4.2

###### Description:

Consider external secret storage (Manual)


###### Remediation:

Refer to the Secrets management options offered by your cloud
provider or a third-party
secrets management solution.


#### Extensible Admission Control

##### Control 5.5.1

###### Description:

Configure Image Provenance using ImagePolicyWebhook admission
controller (Manual)


###### Remediation:

Follow the Kubernetes documentation and setup image provenance.


#### General Policies

##### Control 5.7.1

###### Description:

Create administrative boundaries between resources using
namespaces (Manual)


###### Remediation:

Follow the documentation and create namespaces for objects in
your deployment as you need
them.


##### Control 5.7.2

###### Description:

Ensure that the seccomp profile is set to docker/default in your
Pod definitions (Manual)


###### Remediation:

Use `securityContext` to enable the docker/default seccomp
profile in your pod definitions.
An example is as follows:
  securityContext:
    seccompProfile:
      type: RuntimeDefault


##### Control 5.7.3

###### Description:

Apply SecurityContext to your Pods and Containers (Manual)


###### Remediation:

Follow the Kubernetes documentation and apply SecurityContexts
to your Pods. For a
suggested list of SecurityContexts, you may refer to the CIS
Security Benchmark for Docker
Containers.


##### Control 5.7.4

###### Description:

The default namespace should not be used (Manual)


###### Remediation:

Ensure that namespaces are created to allow for appropriate
segregation of Kubernetes
resources and that all new resources are created in a specific
namespace.




<!-- Links -->
[Center for Internet Security (CIS)]:https://www.cisecurity.org/
[kube-bench]:https://aquasecurity.github.io/kube-bench/v0.6.15/
[CIS Kubernetes Benchmark]:https://www.cisecurity.org/benchmark/kubernetes
[getting-started-guide]: ../tutorial/getting-started
[kube-bench release]: https://github.com/aquasecurity/kube-bench/releases
[upstream instructions]:https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[rate limits]:https://kubernetes.io/docs/reference/config-api/apiserver-eventratelimit.v1alpha1
