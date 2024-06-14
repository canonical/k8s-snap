
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
Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated)
```

Control **2.2**
Description: `Ensure that the --client-cert-auth argument is set to true (Automated)`

Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Ensure that the --client-cert-auth argument is set to true (Automated)
```

Control **2.3**
Description: `Ensure that the --auto-tls argument is not set to true (Automated)`

Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Ensure that the --auto-tls argument is not set to true (Automated)
```

Control **2.4**
Description: `Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated)`

Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated)
```

Control **2.5**
Description: `Ensure that the --peer-client-cert-auth argument is set to true (Automated)`

Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Ensure that the --peer-client-cert-auth argument is set to true (Automated)
```

Control **2.6**
Description: `Ensure that the --peer-auto-tls argument is not set to true (Automated)`

Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Ensure that the --peer-auto-tls argument is not set to true (Automated)
```

Control **2.7**
Description: `Ensure that a unique Certificate Authority is used for etcd (Manual)`

Audit:
```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Remediation:
```
Ensure that a unique Certificate Authority is used for etcd (Manual)
```

