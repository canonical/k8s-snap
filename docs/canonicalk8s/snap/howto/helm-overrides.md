# How to override feature values with Helm

{{product}} manages built-in features (DNS, network, ingress, etc.) by
deploying and reconciling Helm charts. You can pass extra Helm values to any
feature by creating a ConfigMap in the `kube-system` namespace. The cluster
controller picks up changes automatically — no restart is required.

## Prerequisites

- Root or sudo access to the machine.
- A bootstrapped {{product}} cluster (see the [Getting Started][getting-started-guide] guide).
- [Helm](https://helm.sh/docs/intro/install/) installed on your machine to inspect release values.

## Naming convention

Each feature has a dedicated ConfigMap name:

| Feature | ConfigMap name |
|---------|----------------|
| DNS (CoreDNS) | `k8sd-coredns-values` |
| Network, Ingress, Gateway (Cilium) | `k8sd-cilium-values` |
| Load Balancer (MetalLB) | `k8sd-metallb-values` |
| Local Storage (LocalPV) | `k8sd-localpv-values` |
| Metrics Server | `k8sd-metrics-server-values` |

> **Note:** Network, Ingress, and Gateway all share the same Cilium Helm chart and therefore use the same ConfigMap (`k8sd-cilium-values`).

All ConfigMaps live in the `kube-system` namespace.

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

By default CoreDNS uses a Horizontal Pod Autoscaler (HPA) with `minReplicas: 2`. To raise the minimum to
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
helm get values ck-dns --namespace kube-system --output yaml
```

> **Note:** `ck-dns` is the internal Helm release name for CoreDNS. Use `k8s helm list -n kube-system` to list all managed releases.

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

## Remove overrides

Once a value has been overridden, removing the key from the ConfigMap's
`values` field **does not revert it** to the chart default. The last-applied
value persists in the Helm release. To revert a key, explicitly set it back to
the chart's default value in the ConfigMap. Deleting the ConfigMap entirely
also does **not** revert the release — the previously deployed values remain.

## Notes

- Overrides are merged on top of defaults. Keys you do not specify keep their
  default values.
- If the `values` key is missing from the ConfigMap, or if the YAML is
  invalid, the override is ignored and a warning is surfaced in
  `sudo k8s status`. Errors in the values themselves (e.g. an unknown chart
  key) are only surfaced by Helm at reconcile time, not at `kubectl apply`.
- Overrides survive feature disable/enable cycles and cluster restarts.

<!-- LINKS -->
[getting-started-guide]: /snap/tutorial/getting-started
