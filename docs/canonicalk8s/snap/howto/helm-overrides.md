# How to override feature Helm values

{{product}} manages built-in features (DNS, network, ingress, etc.) by
deploying and reconciling Helm charts. You can pass extra Helm values to any
feature by creating a ConfigMap in the `kube-system` namespace. The cluster
controller picks up changes automatically — no restart is required.

## Prerequisites

- Root or sudo access to the machine.
- A bootstrapped {{product}} cluster (see the [Getting Started][getting-started-guide] guide).

## Naming convention

Each feature has a dedicated ConfigMap name:

| Feature | ConfigMap name |
|---------|----------------|
| DNS (CoreDNS) | `k8sd-coredns-values` |

Additional features will be listed here as support is added.

## ConfigMap format

The ConfigMap must contain a single key `values` whose value is a YAML
fragment. The values are deep-merged with the feature's defaults — only the
keys you specify are overridden.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: <configmap-name>
  namespace: kube-system
data:
  values: |
    key:
      nestedKey: value
```

## Example: scale CoreDNS replicas

By default CoreDNS uses an HPA with `minReplicas: 2`. To raise the minimum to
4 and cap the maximum at 20:

```
sudo k8s kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-coredns-values
  namespace: kube-system
data:
  values: |
    hpa:
      minReplicas: 4
      maxReplicas: 20
EOF
```

The controller reconciles within seconds. Verify the change:

```
sudo /snap/k8s/current/bin/helm get values ck-dns \
  --namespace kube-system --output yaml
```

## Update an override

Apply the same ConfigMap with updated values:

```
sudo k8s kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-coredns-values
  namespace: kube-system
data:
  values: |
    hpa:
      minReplicas: 6
      maxReplicas: 30
EOF
```

## Remove all overrides

Delete the ConfigMap to revert to the feature's defaults:

```
sudo k8s kubectl delete configmap k8sd-coredns-values -n kube-system
```

## Notes

- Overrides are merged on top of defaults. Keys you do not specify keep their
  default values.
- If the `values` key is missing from the ConfigMap, or if the YAML is
  invalid, the feature falls back to defaults and surfaces a warning in
  `sudo k8s status`.
- Overrides survive feature disable/enable cycles and cluster restarts.

<!-- LINKS -->
[getting-started-guide]: /snap/tutorial/getting-started
