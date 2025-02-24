## k8s refresh-certs

Refresh the certificates of the running node

```
k8s refresh-certs [flags]
```

### Options

```
      --expires-in string              the time until the certificates expire, e.g., 1h, 2d, 4mo, 5y. Aditionally, any valid time unit for ParseDuration is accepted.
      --external-certificates string   path to a YAML file containing external certificate data in PEM format. If the cluster was bootstrapped with external certificates, the certificates will be updated.
      --extra-sans stringArray         extra SANs to add to the certificates.
  -h, --help                           help for refresh-certs
      --timeout duration               the max time to wait for the command to execute (default 1m30s)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

