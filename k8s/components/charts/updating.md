# Updating component charts versions

Helm can be utilized to pull in new versions of charts as `tgz` archives.


```
helm pull --repo <helm-repository-url> <chart-name>
```

This command would pull in the latest version for the given `chart-name` from the given repository. An additional flag can also be defined to pull in a specific version, or a range with `--version`. Example: `--version ^2.0.0`


## metrics-server

```
helm pull --repo https://kubernetes-sigs.github.io/metrics-server/ metrics-server
```
