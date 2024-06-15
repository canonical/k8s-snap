
## Etcd Node Configuration
### Etcd Node Configuration
Control **2.1**

Description: `Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated)`
Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Follow the etcd service documentation and configure TLS encryption.
Then, edit the etcd pod specification file /etc/kubernetes/manifests/etcd.yaml
on the master node and set the below parameters.
--cert-file=</path/to/ca-file>
--key-file=</path/to/key-file>
```

Expected output:
```
bin_op: and
test_items:
- env: ETCD_CERT_FILE
  flag: --cert-file
- env: ETCD_KEY_FILE
  flag: --key-file
```

Control **2.2**

Description: `Ensure that the --client-cert-auth argument is set to true (Automated)`
Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Edit the etcd pod specification file /etc/default/etcd on the master
node and set the below parameter.
--client-cert-auth="true"
```

Expected output:
```
test_items:
- compare:
    op: eq
    value: true
  env: ETCD_CLIENT_CERT_AUTH
  flag: --client-cert-auth
```

Control **2.3**

Description: `Ensure that the --auto-tls argument is not set to true (Automated)`
Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Edit the etcd pod specification file /etc/default/etcd on the master
node and either remove the --auto-tls parameter or set it to false.
  --auto-tls=false
```

Expected output:
```
bin_op: or
test_items:
- env: ETCD_AUTO_TLS
  flag: --auto-tls
  set: false
- compare:
    op: eq
    value: false
  env: ETCD_AUTO_TLS
  flag: --auto-tls
```

Control **2.4**

Description: `Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated)`
Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Follow the etcd service documentation and configure peer TLS encryption as appropriate
for your etcd cluster.
Then, edit the etcd pod specification file /etc/default/etcd on the
master node and set the below parameters.
--peer-client-file=</path/to/peer-cert-file>
--peer-key-file=</path/to/peer-key-file>
```

Expected output:
```
bin_op: and
test_items:
- env: ETCD_PEER_CERT_FILE
  flag: --peer-cert-file
- env: ETCD_PEER_KEY_FILE
  flag: --peer-key-file
```

Control **2.5**

Description: `Ensure that the --peer-client-cert-auth argument is set to true (Automated)`
Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Edit the etcd pod specification file /etc/default/etcd on the master
node and set the below parameter.
--peer-client-cert-auth=true
```

Expected output:
```
test_items:
- compare:
    op: eq
    value: true
  env: ETCD_PEER_CLIENT_CERT_AUTH
  flag: --peer-client-cert-auth
```

Control **2.6**

Description: `Ensure that the --peer-auto-tls argument is not set to true (Automated)`
Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Edit the etcd pod specification file /etc/default/etcd on the master
node and either remove the --peer-auto-tls parameter or set it to false.
--peer-auto-tls=false
```

Expected output:
```
bin_op: or
test_items:
- env: ETCD_PEER_AUTO_TLS
  flag: --peer-auto-tls
  set: false
- compare:
    op: eq
    value: false
  env: ETCD_PEER_AUTO_TLS
  flag: --peer-auto-tls
```

Control **2.7**

Description: `Ensure that a unique Certificate Authority is used for etcd (Manual)`
Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
[Manual test]
Follow the etcd documentation and create a dedicated certificate authority setup for the
etcd service.
Then, edit the etcd pod specification file /etc/default/etcd on the
master node and set the below parameter.
--trusted-ca-file=</path/to/ca-file>
```

Expected output:
```
test_items:
- env: ETCD_TRUSTED_CA_FILE
  flag: --trusted-ca-file
```

