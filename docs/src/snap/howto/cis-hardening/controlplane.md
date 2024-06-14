
## Control Plane Configuration

### Authentication and Authorization

Control **3.1.1**
Description: `Client certificate authentication should not be used for users (Manual)`

Remediation:
```
Client certificate authentication should not be used for users (Manual)
```


### Logging

Control **3.2.1**
Description: `Ensure that a minimal audit policy is created (Manual)`

Audit:
```
/bin/ps -ef | grep kube-apiserver | grep -v grep
```

Remediation:
```
Ensure that a minimal audit policy is created (Manual)
```

Control **3.2.2**
Description: `Ensure that the audit policy covers key security concerns (Manual)`

Remediation:
```
Ensure that the audit policy covers key security concerns (Manual)
```

