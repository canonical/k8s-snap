
## Worker Node Security Configuration

### Worker Node Configuration Files

Control **4.1.1**
Description: `Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c permissions=%a /etc/systemd/system/snap.k8s.kubelet.service; fi' 
```

Remediation:
```
Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated)
```

Control **4.1.2**
Description: `Ensure that the kubelet service file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c "if test -e /etc/systemd/system/snap.k8s.kubelet.service; then stat -c %U:%G /etc/systemd/system/snap.k8s.kubelet.service; else echo \"File not found\"; fi"
```

Remediation:
```
Ensure that the kubelet service file ownership is set to root:root (Automated)
```

Control **4.1.3**
Description: `If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/proxy.conf; then stat -c permissions=%a /etc/kubernetes/proxy.conf; fi' 
```

Remediation:
```
If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual)
```

Control **4.1.4**
Description: `If proxy kubeconfig file exists ensure ownership is set to root:root (Manual)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/proxy.conf; then stat -c %U:%G /etc/kubernetes/proxy.conf; fi' 
```

Remediation:
```
If proxy kubeconfig file exists ensure ownership is set to root:root (Manual)
```

Control **4.1.5**
Description: `Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/kubelet.conf; then stat -c permissions=%a /etc/kubernetes/kubelet.conf; fi' 
```

Remediation:
```
Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated)
```

Control **4.1.6**
Description: `Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated)`

Audit:
```
/bin/sh -c 'if test -e /etc/kubernetes/kubelet.conf; then stat -c %U:%G /etc/kubernetes/kubelet.conf; fi' 
```

Remediation:
```
Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated)
```

Control **4.1.7**
Description: `Ensure that the certificate authorities file permissions are set to 600 or more restrictive (Manual)`

Audit:
```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c permissions=%a $CAFILE; fi

```

Remediation:
```
Ensure that the certificate authorities file permissions are set to 600 or more restrictive (Manual)
```

Control **4.1.8**
Description: `Ensure that the client certificate authorities file ownership is set to root:root (Manual)`

Audit:
```
CAFILE=$(ps -ef | grep kubelet | grep -v apiserver | grep -- --client-ca-file= | awk -F '--client-ca-file=' '{print $2}' | awk '{print $1}' | uniq)
if test -z $CAFILE; then CAFILE=/etc/kubernetes/pki/client-ca.crt; fi
if test -e $CAFILE; then stat -c %U:%G $CAFILE; fi

```

Remediation:
```
Ensure that the client certificate authorities file ownership is set to root:root (Manual)
```

Control **4.1.9**
Description: `If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Manual)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kubelet; then stat -c permissions=%a /var/snap/k8s/common/args/kubelet; fi' 
```

Remediation:
```
If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Manual)
```

Control **4.1.10**
Description: `If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Manual)`

Audit:
```
/bin/sh -c 'if test -e /var/snap/k8s/common/args/kubelet; then stat -c %U:%G /var/snap/k8s/common/args/kubelet; fi' 
```

Remediation:
```
If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Manual)
```


### Kubelet

Control **4.2.1**
Description: `Ensure that the --anonymous-auth argument is set to false (Automated)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the --anonymous-auth argument is set to false (Automated)
```

Control **4.2.2**
Description: `Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)
```

Control **4.2.3**
Description: `Ensure that the --client-ca-file argument is set as appropriate (Automated)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the --client-ca-file argument is set as appropriate (Automated)
```

Control **4.2.4**
Description: `Verify that the --read-only-port argument is set to 0 (Manual)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Verify that the --read-only-port argument is set to 0 (Manual)
```

Control **4.2.5**
Description: `Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual)
```

Control **4.2.6**
Description: `Ensure that the --protect-kernel-defaults argument is set to true (Automated)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the --protect-kernel-defaults argument is set to true (Automated)
```

Control **4.2.7**
Description: `Ensure that the --make-iptables-util-chains argument is set to true (Automated)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the --make-iptables-util-chains argument is set to true (Automated)
```

Control **4.2.8**
Description: `Ensure that the --hostname-override argument is not set (Manual)`

Audit:
```
/bin/ps -fC kubelet 
```

Remediation:
```
Ensure that the --hostname-override argument is not set (Manual)
```

Control **4.2.9**
Description: `Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual)
```

Control **4.2.10**
Description: `Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual)
```

Control **4.2.11**
Description: `Ensure that the --rotate-certificates argument is not set to false (Automated)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the --rotate-certificates argument is not set to false (Automated)
```

Control **4.2.12**
Description: `Verify that the RotateKubeletServerCertificate argument is set to true (Manual)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Verify that the RotateKubeletServerCertificate argument is set to true (Manual)
```

Control **4.2.13**
Description: `Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual)`

Audit:
```
/bin/ps -fC kubelet
```

Remediation:
```
Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual)
```

