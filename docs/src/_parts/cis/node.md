## Worker Node Security Configuration

### Worker Node Configuration Files

#### Control 4.1.1

Description: Ensure that the kubelet service file permissions are set to 600
or more restrictive (Automated)

Audit:

```
/bin/sh -c 'if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c permissions=%a /etc/systemd/system/snap.k8s.kubelet.service; fi'
```

Expected output:

```
permissions=644
```

Remediation:

Run the following command on each worker node.


`chmod 600 /etc/systemd/system/snap.k8s.kubelet.service`

#### Control 4.1.2

Description: Ensure that the kubelet service file ownership is set to
root:root (Automated)

Audit:

```
/bin/sh -c "if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c %U:%G /etc/systemd/system/snap.k8s.kubelet.service; else echo \"File not found\"; fi"
```

Expected output:

```
root:root
```

Remediation:

Run the following command on each worker node.


`chown root:root /etc/systemd/system/snap.k8s.kubelet.service`

#### Control 4.1.3

Description: If proxy kubeconfig file exists ensure permissions are set to
600 or more restrictive (Manual)

Audit:

```
/bin/sh -c 'if test -e /etc/kubernetes/proxy.conf; then stat -c permissions=%a /etc/kubernetes/proxy.conf; fi'
```

Expected output:

```
permissions=644
```

Remediation:

Run the following command on the each worker node.


`chmod 600 /etc/kubernetes/proxy.conf`

#### Control 4.1.4

Description: If proxy kubeconfig file exists ensure ownership is set to
root:root (Manual)

Audit:

```
/bin/sh -c 'if test -e /etc/kubernetes/proxy.conf; then stat -c %U:%G /etc/kubernetes/proxy.conf; fi'
```

Expected output:

```
root:root
```

Remediation:

Run the following command on the each worker node.


`chown root:root /etc/kubernetes/proxy.conf`

#### Control 4.1.5

Description: Ensure that the --kubeconfig kubelet.conf file permissions are
set to 600 or more restrictive (Automated)

Audit:

```
/bin/sh -c 'if test -e /etc/kubernetes/kubelet.conf; then stat -c permissions=%a /etc/kubernetes/kubelet.conf; fi'
```

Expected output:

```
permissions=600
```

Remediation:

Run the following command on the each worker node.


`chmod 600 /etc/kubernetes/kubelet.conf`

#### Control 4.1.6

Description: Ensure that the --kubeconfig kubelet.conf file ownership is set
to root:root (Automated)

Audit:

```
/bin/sh -c 'if test -e /etc/kubernetes/kubelet.conf; then stat -c %U:%G /etc/kubernetes/kubelet.conf; fi'
```

Expected output:

```
root:root
```

Remediation:

Run the following command on the each worker node.


`chown root:root /etc/kubernetes/kubelet.conf`

#### Control 4.1.7

Description: Ensure that the certificate authorities file permissions are set
to 600 or more restrictive (Manual)

Audit:

```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c permissions=%a $CAFILE; fi
```

Expected output:

```
permissions=600
```

Remediation:

Run the following command to modify the file permissions of the
--client-ca-file.


`chmod 600 <filename>`

#### Control 4.1.8

Description: Ensure that the client certificate authorities file ownership is
set to root:root (Manual)

Audit:

```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c %U:%G $CAFILE; fi
```

Expected output:

```
root:root
```

Remediation:

Run the following command to modify the ownership of the
--client-ca-file.


`chown root:root <filename>`

#### Control 4.1.9

Description: If the kubelet config.yaml configuration file is being used
validate permissions set to 600 or more restrictive (Manual)

Audit:

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kubelet; then stat -c permissions=%a /var/snap/k8s/common/args/kubelet; fi'
```

Expected output:

```
permissions=600
```

Remediation:

Run the following command (using the config file location
identified in the Audit step)


`chmod 600 /var/snap/k8s/common/args/kubelet`

#### Control 4.1.10

Description: If the kubelet config.yaml configuration file is being used
validate file ownership is set to root:root (Manual)

Audit:

```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kubelet; then stat -c %U:%G /var/snap/k8s/common/args/kubelet; fi'
```

Expected output:

```
root:root
```

Remediation:

Run the following command (using the config file location
identified in the Audit step)


`chown root:root /var/snap/k8s/common/args/kubelet`

### Kubelet

#### Control 4.2.1

Description: Ensure that the --anonymous-auth argument is set to false
(Automated)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--anonymous-auth=false
```

Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.


`--anonymous-auth=false`


Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.2

Description: Ensure that the --authorization-mode argument is not set to
AlwaysAllow (Automated)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--authorization-mode=Webhook
```

Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--authorization-mode=Webhook`

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.3

Description: Ensure that the --client-ca-file argument is set as appropriate
(Automated)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--client-ca-file=/etc/kubernetes/pki/client-ca.crt
```

Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--client-ca-file=<path/to/client-ca-file>`

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.4

Description: Verify that the --read-only-port argument is set to 0 (Manual)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--read-only-port=0
```

Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--read-only-port=0`

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.5

Description: Ensure that the --streaming-connection-idle-timeout argument is
not set to 0 (Manual)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--streaming-connection-idle-timeout is not set
```

Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--streaming-connection-idle-timeout=5m`

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.6

Description: Ensure that the --protect-kernel-defaults argument is set to
true (Automated)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--protect-kernel-defaults=true
```

Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and set the following argument.

`--protect-kernel-defaults=true`

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.7

Description: Ensure that the --make-iptables-util-chains argument is set to
true (Automated)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--make-iptables-util-chains is not set
```

Remediation:

Edit the kubelet configuration file
/var/snap/k8s/common/args/kubelet on each worker node and
remove the --make-iptables-util-chains argument.

Restart the kubelet service.

For example: `snap restart k8s.kubelet`

#### Control 4.2.8

Description: Ensure that the --hostname-override argument is not set (Manual)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--hostname-override is set to false
```

Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet
on each worker node and remove the --hostname-override argument.

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.9

Description: Ensure that the eventRecordQPS argument is set to a level which
ensures appropriate event capture (Manual)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--event-qps is not set
```

Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each worker
node and
set the --event-qps parameter as appropriate.

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.10

Description: Ensure that the --tls-cert-file and --tls-private-key-file
arguments are set as appropriate (Manual)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--tls-cert-file=/etc/kubernetes/pki/kubelet.crt and --tls-
private-key-file=/etc/kubernetes/pki/kubelet.key
```

Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each worker
node and
set the following arguments:

```
--tls-cert-file=<path/to/tls-certificate-file>
--tls-cert-file=<path/to/tls-certificate-file>
```

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.11

Description: Ensure that the --rotate-certificates argument is not set to
false (Automated)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--rotate-certificates is not set
```

Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each worker
node and
remove the --rotate-certificates=false argument.

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.12

Description: Verify that the RotateKubeletServerCertificate argument is set
to true (Manual)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
the RotateKubeletServerCertificate feature gate is not set
```

Remediation:

Edit the kubelet configuration file /var/snap/k8s/common/args/kubelet on each worker
node and
set the argument --feature-
gates=RotateKubeletServerCertificate=true
on each worker node.

Restart the kubelet service.

For example, `snap restart k8s.kubelet`

#### Control 4.2.13

Description: Ensure that the Kubelet only makes use of Strong Cryptographic
Ciphers (Manual)

Audit:

```
/bin/ps -fC kubelet
```

Expected output:

```
--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_
ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_
POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WIT
H_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_
RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256
```

Remediation:

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

For example, `snap restart k8s.kubelet`

