## Worker Node Security Configuration

### Worker Node Configuration Files

#### Control 4.1.1

Description: `Ensure that the kubelet service file permissions are set to 600
or more restrictive (Automated)`

Audit:

```
/bin/sh -c 'if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c permissions=%a /etc/systemd/system/snap.k8s.kubelet.service; fi'
```

Remediation:

```
Run the below command (based on the file location on your
system) on the each worker node.
For example, chmod 600 /etc/systemd/system/snap.k8s.kubelet.service
```

Expected output:

```
test_items:
- compare:
    op: bitmask
    value: '644'
  flag: permissions
```

#### Control 4.1.2

Description: `Ensure that the kubelet service file ownership is set to
root:root (Automated)`

Audit:

```
/bin/sh -c "if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c %U:%G /etc/systemd/system/snap.k8s.kubelet.service; else echo \"File not found\"; fi"
```

Remediation:

```
Run the below command (based on the file location on your
system) on the each worker node.
For example,
chown root:root /etc/systemd/system/snap.k8s.kubelet.service
```

Expected output:

```
bin_op: or
test_items:
- flag: root:root
- flag: File not found
```

#### Control 4.1.3

Description: `If proxy kubeconfig file exists ensure permissions are set to
600 or more restrictive (Manual)`

Audit:

```
/bin/sh -c 'if test -e /etc/kubernetes/proxy.conf; then stat -c permissions=%a /etc/kubernetes/proxy.conf; fi'
```

Remediation:

```
Run the below command (based on the file location on your
system) on the each worker node.
For example,
chmod 600 /etc/kubernetes/proxy.conf
```

Expected output:

```
bin_op: or
test_items:
- compare:
    op: bitmask
    value: '644'
  flag: permissions
  set: true
```

#### Control 4.1.4

Description: `If proxy kubeconfig file exists ensure ownership is set to
root:root (Manual)`

Audit:

```
/bin/sh -c 'if test -e /etc/kubernetes/proxy.conf; then stat -c %U:%G /etc/kubernetes/proxy.conf; fi'
```

Remediation:

```
Run the below command (based on the file location on your
system) on the each worker node.
For example, chown root:root /etc/kubernetes/proxy.conf
```

Expected output:

```
bin_op: or
test_items:
- flag: root:root
```

#### Control 4.1.5

Description: `Ensure that the --kubeconfig kubelet.conf file permissions are
set to 600 or more restrictive (Automated)`

Audit:

```
/bin/sh -c 'if test -e /etc/kubernetes/kubelet.conf; then stat -c permissions=%a /etc/kubernetes/kubelet.conf; fi'
```

Remediation:

```
Run the below command (based on the file location on your
system) on the each worker node.
For example,
chmod 600 /etc/kubernetes/kubelet.conf
```

Expected output:

```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 4.1.6

Description: `Ensure that the --kubeconfig kubelet.conf file ownership is set
to root:root (Automated)`

Audit:

```
/bin/sh -c 'if test -e /etc/kubernetes/kubelet.conf; then stat -c %U:%G /etc/kubernetes/kubelet.conf; fi'
```

Remediation:

```
Run the below command (based on the file location on your
system) on the each worker node.
For example,
chown root:root /etc/kubernetes/kubelet.conf
```

Expected output:

```
test_items:
- flag: root:root
```

#### Control 4.1.7

Description: `Ensure that the certificate authorities file permissions are set
to 600 or more restrictive (Manual)`

Audit:

```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c permissions=%a $CAFILE; fi
```

Remediation:

```
Run the following command to modify the file permissions of the
--client-ca-file chmod 600 <filename>
```

Expected output:

```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 4.1.8

Description: `Ensure that the client certificate authorities file ownership is
set to root:root (Manual)`

Audit:

```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c %U:%G $CAFILE; fi
```

Remediation:

```
Run the following command to modify the ownership of the
--client-ca-file.
chown root:root <filename>
```

Expected output:

```
test_items:
- compare:
    op: eq
    value: root:root
  flag: root:root
```

#### Control 4.1.9

Description: `If the kubelet config.yaml configuration file is being used
validate permissions set to 600 or more restrictive (Manual)`

Audit:

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kubelet; then stat -c permissions=%a /var/snap/k8s/common/args/kubelet; fi'
```

Remediation:

```
Run the following command (using the config file location
identified in the Audit step)
chmod 600 /var/snap/k8s/common/args/kubelet
```

Expected output:

```
test_items:
- compare:
    op: bitmask
    value: '600'
  flag: permissions
```

#### Control 4.1.10

Description: `If the kubelet config.yaml configuration file is being used
validate file ownership is set to root:root (Manual)`

Audit:

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kubelet; then stat -c %U:%G /var/snap/k8s/common/args/kubelet; fi'
```

Remediation:

```
Run the following command (using the config file location
identified in the Audit step)
chown root:root /var/snap/k8s/common/args/kubelet
```

Expected output:

```
test_items:
- flag: root:root
```

### Kubelet

#### Control 4.2.1

Description: `Ensure that the --anonymous-auth argument is set to false
(Automated)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`authentication: anonymous: enabled` to
`false`.
If using executable arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
set the below parameter in KUBELET_SYSTEM_PODS_ARGS variable.
`--anonymous-auth=false`
Based on your system, restart the kubelet service. For example,
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
test_items:
- compare:
    op: eq
    value: false
  flag: --anonymous-auth
  path: '{.authentication.anonymous.enabled}'
```

#### Control 4.2.2

Description: `Ensure that the --authorization-mode argument is not set to
AlwaysAllow (Automated)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`authorization.mode` to Webhook. If
using executable arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
set the below parameter in KUBELET_AUTHZ_ARGS variable.
--authorization-mode=Webhook
Based on your system, restart the kubelet service. For example,
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
test_items:
- compare:
    op: nothave
    value: AlwaysAllow
  flag: --authorization-mode
  path: '{.authorization.mode}'
```

#### Control 4.2.3

Description: `Ensure that the --client-ca-file argument is set as appropriate
(Automated)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`authentication.x509.clientCAFile` to
the location of the client CA file.
If using command line arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
set the below parameter in KUBELET_AUTHZ_ARGS variable.
--client-ca-file=<path/to/client-ca-file>
Based on your system, restart the kubelet service. For example,
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
test_items:
- flag: --client-ca-file
  path: '{.authentication.x509.clientCAFile}'
```

#### Control 4.2.4

Description: `Verify that the --read-only-port argument is set to 0 (Manual)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`readOnlyPort` to 0.
If using command line arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
set the below parameter in KUBELET_SYSTEM_PODS_ARGS variable.
--read-only-port=0
Based on your system, restart the kubelet service. For example,
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
bin_op: or
test_items:
- compare:
    op: eq
    value: 0
  flag: --read-only-port
  path: '{.readOnlyPort}'
- flag: --read-only-port
  path: '{.readOnlyPort}'
  set: false
```

#### Control 4.2.5

Description: `Ensure that the --streaming-connection-idle-timeout argument is
not set to 0 (Manual)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`streamingConnectionIdleTimeout` to a
value other than 0.
If using command line arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
set the below parameter in KUBELET_SYSTEM_PODS_ARGS variable.
--streaming-connection-idle-timeout=5m
Based on your system, restart the kubelet service. For example,
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
bin_op: or
test_items:
- compare:
    op: noteq
    value: 0
  flag: --streaming-connection-idle-timeout
  path: '{.streamingConnectionIdleTimeout}'
- flag: --streaming-connection-idle-timeout
  path: '{.streamingConnectionIdleTimeout}'
  set: false
```

#### Control 4.2.6

Description: `Ensure that the --protect-kernel-defaults argument is set to
true (Automated)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`protectKernelDefaults` to `true`.
If using command line arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
set the below parameter in KUBELET_SYSTEM_PODS_ARGS variable.
--protect-kernel-defaults=true
Based on your system, restart the kubelet service. For example:
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
test_items:
- compare:
    op: eq
    value: true
  flag: --protect-kernel-defaults
  path: '{.protectKernelDefaults}'
```

#### Control 4.2.7

Description: `Ensure that the --make-iptables-util-chains argument is set to
true (Automated)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`makeIPTablesUtilChains` to `true`.
If using command line arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
remove the --make-iptables-util-chains argument from the
KUBELET_SYSTEM_PODS_ARGS variable.
Based on your system, restart the kubelet service. For example:
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
bin_op: or
test_items:
- compare:
    op: eq
    value: true
  flag: --make-iptables-util-chains
  path: '{.makeIPTablesUtilChains}'
- flag: --make-iptables-util-chains
  path: '{.makeIPTablesUtilChains}'
  set: false
```

#### Control 4.2.8

Description: `Ensure that the --hostname-override argument is not set (Manual)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
Edit the kubelet service file /etc/systemd/system/snap.k8s.kubelet.service
on each worker node and remove the --hostname-override argument
from the
KUBELET_SYSTEM_PODS_ARGS variable.
Based on your system, restart the kubelet service. For example,
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
test_items:
- flag: --hostname-override
  set: false
```

#### Control 4.2.9

Description: `Ensure that the eventRecordQPS argument is set to a level which
ensures appropriate event capture (Manual)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`eventRecordQPS` to an appropriate level.
If using command line arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
set the below parameter in KUBELET_SYSTEM_PODS_ARGS variable.
Based on your system, restart the kubelet service. For example,
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
bin_op: or
test_items:
- compare:
    op: gte
    value: 0
  flag: --event-qps
  path: '{.eventRecordQPS}'
- flag: --event-qps
  path: '{.eventRecordQPS}'
  set: false
```

#### Control 4.2.10

Description: `Ensure that the --tls-cert-file and --tls-private-key-file
arguments are set as appropriate (Manual)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`tlsCertFile` to the location
of the certificate file to use to identify this Kubelet, and
`tlsPrivateKeyFile`
to the location of the corresponding private key file.
If using command line arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
set the below parameters in KUBELET_CERTIFICATE_ARGS variable.
--tls-cert-file=<path/to/tls-certificate-file>
--tls-private-key-file=<path/to/tls-key-file>
Based on your system, restart the kubelet service. For example,
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
test_items:
- flag: --tls-cert-file
  path: '{.tlsCertFile}'
- flag: --tls-private-key-file
  path: '{.tlsPrivateKeyFile}'
```

#### Control 4.2.11

Description: `Ensure that the --rotate-certificates argument is not set to
false (Automated)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to add the line
`rotateCertificates` to `true` or
remove it altogether to use the default value.
If using command line arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
remove --rotate-certificates=false argument from the
KUBELET_CERTIFICATE_ARGS
variable.
Based on your system, restart the kubelet service. For example,
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
bin_op: or
test_items:
- compare:
    op: eq
    value: true
  flag: --rotate-certificates
  path: '{.rotateCertificates}'
- flag: --rotate-certificates
  path: '{.rotateCertificates}'
  set: false
```

#### Control 4.2.12

Description: `Verify that the RotateKubeletServerCertificate argument is set
to true (Manual)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
Edit the kubelet service file /etc/systemd/system/snap.k8s.kubelet.service
on each worker node and set the below parameter in
KUBELET_CERTIFICATE_ARGS variable.
--feature-gates=RotateKubeletServerCertificate=true
Based on your system, restart the kubelet service. For example:
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
bin_op: or
test_items:
- compare:
    op: nothave
    value: false
  flag: RotateKubeletServerCertificate
  path: '{.featureGates.RotateKubeletServerCertificate}'
- flag: RotateKubeletServerCertificate
  path: '{.featureGates.RotateKubeletServerCertificate}'
  set: false
```

#### Control 4.2.13

Description: `Ensure that the Kubelet only makes use of Strong Cryptographic
Ciphers (Manual)`

Audit:

```
/bin/ps -fC kubelet
```

Remediation:

```
If using a Kubelet config file, edit the file to set
`TLSCipherSuites` to
TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_1
28_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_R
SA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM
_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256
or to a subset of these values.
If using executable arguments, edit the kubelet service file
/etc/systemd/system/snap.k8s.kubelet.service on each worker node and
set the --tls-cipher-suites parameter as follows, or to a subset
of these values.
--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_
ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_
POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WIT
H_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_
RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256
Based on your system, restart the kubelet service. For example:
systemctl daemon-reload
systemctl restart kubelet.service
```

Expected output:

```
test_items:
- compare:
    op: valid_elements
    value: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256
  flag: --tls-cipher-suites
  path: '{range .tlsCipherSuites[:]}{}{'',''}{end}'
```

