<!-- markdownlint-disable -->
## Control Plane Configuration
### Authentication and Authorization
Control **3.1.1**

Description: `Client certificate authentication should not be used for users (Manual)`

Remediation:
```
Alternative mechanisms provided by Kubernetes such as the use of OIDC should be
implemented in place of client certificates.
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
Create an audit policy file for your cluster.
```

Expected output:
```
test_items:
- flag: --audit-policy-file
  set: true
```

Control **3.2.2**

Description: `Ensure that the audit policy covers key security concerns (Manual)`

Remediation:
```
Review the audit policy provided for the cluster and ensure that it covers
at least the following areas,
- Access to Secrets managed by the cluster. Care should be taken to only
  log Metadata for requests to Secrets, ConfigMaps, and TokenReviews, in
  order to avoid risk of logging sensitive data.
- Modification of Pod and Deployment objects.
- Use of `pods/exec`, `pods/portforward`, `pods/proxy` and `services/proxy`.
For most requests, minimally logging at the Metadata level is recommended
(the most basic level of logging).
```

