## k8s refresh-certs

Refresh the certificates of the running node

```
k8s refresh-certs [flags]
```

### Options

```
      --expires-in string        the time until the certificates expire, e.g., 1h, 2d, 4mo, 5y. Aditionally, any valid time unit for ParseDuration is accepted.
      --extra-sans stringArray   extra SANs to add to the certificates.
  -h, --help                     help for refresh-certs
      --timeout duration         the max time to wait for the command to execute (default 1m30s)
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

