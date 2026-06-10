# DevOps Perspective: k8sd Feature Configuration Approaches

**Author**: DevOps/SRE Practitioner
**Date**: 2024
**Audience**: Operations teams, SREs, platform engineers running Canonical Kubernetes in production

## Executive Summary

This document evaluates two k8sd feature configuration approaches from an **operational practitioner perspective** — the person receiving the 2am page when clusters fail. While architectural elegance matters, what keeps operators up at night is: **Can I debug this quickly? Can I roll back safely? Will this upgrade break my cluster?**

**Bottom Line**: The **ConfigMap approach** significantly reduces operational risk and cognitive load in production environments. It provides faster incident response, simpler troubleshooting, and more predictable upgrades.

---

## 1. Upgrade Scenarios: Real-World Breakage Patterns

### Scenario 1: k8s-snap Version Upgrade (1.30 → 1.31)

**Context**: k8s-snap versions are coupled to Kubernetes versions. Each new release may bump feature component versions (e.g., Cilium 1.15 → 1.16, CoreDNS 1.11 → 1.12).

#### ConfigMap Approach
```bash
# Pre-upgrade check
kubectl get cm -n kube-system k8sd-cilium-config -o yaml
# User sees THEIR helm values — direct transparency

# Upgrade k8s-snap
snap refresh k8s --channel=1.31/stable

# Post-upgrade validation
helm get values -n kube-system cilium
# If config invalid → helm will error IMMEDIATELY with upstream message
# Example: "Error: values don't meet the requirements of the schema"
```

**Failure Mode**: Helm validation fails → clear error → user fixes ConfigMap → reconcile
**Time to Recovery (TTR)**: 5-10 minutes (error message points directly to helm schema)
**Operational Complexity**: LOW — operator googles upstream helm chart docs, fixes values, applies

#### CRD Approach
```bash
# Pre-upgrade check
kubectl get ciliumconfig -n kube-system default -o yaml
# User sees k8sd's ABSTRACTION of cilium config

# Upgrade k8s-snap
snap refresh k8s --channel=1.31/stable

# Post-upgrade — SILENT BREAKAGE RISK
# If Cilium helm chart 1.16 deprecated a field that k8sd CRD still accepts...
# k8sd controller translates CRD → helm values → helm chart REJECTS
# Error surfaces in k8sd controller logs, NOT at kubectl apply time
```

**Failure Mode**: Version skew between CRD schema and helm chart → delayed error discovery
**Time to Recovery (TTR)**: 30-60 minutes (debug controller logs, understand translation layer, check if CRD or helm is wrong)
**Operational Complexity**: HIGH — operator must understand THREE layers: CRD schema, k8sd translation logic, helm chart schema

### Scenario 2: Feature Version Bump with Deprecated Values

**Real-World Example**: Cilium 1.15 → 1.16 deprecates `tunnel: vxlan` in favor of `tunnelProtocol: vxlan`

#### ConfigMap Approach
```yaml
# User's existing ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-cilium-config
  namespace: kube-system
data:
  values.yaml: |
    tunnel: vxlan  # DEPRECATED in Cilium 1.16
```

**What Happens**:
1. k8s-snap 1.31 ships Cilium helm chart 1.16
2. k8sd reconciles → applies helm chart with user's values
3. Helm chart validation layer catches deprecated field → **IMMEDIATE ERROR**
4. Helm error message: `"tunnel" is deprecated, use "tunnelProtocol"`
5. Operator updates ConfigMap, applies → fixed

**TTR**: 5-10 minutes
**Debugging Steps**: 1 (read helm error)

#### CRD Approach
```yaml
# User's existing CRD
apiVersion: k8sd.io/v1alpha1
kind: CiliumConfig
metadata:
  name: default
  namespace: kube-system
spec:
  tunnel: vxlan  # k8sd CRD schema STILL accepts this (schema not updated yet)
```

**What Happens**:
1. k8s-snap 1.31 ships Cilium helm chart 1.16
2. User applies CiliumConfig → k8sd CRD validation **PASSES** (CRD schema lags helm chart)
3. k8sd controller translates CRD → helm values → applies to helm
4. Helm chart rejects → reconciliation fails
5. Error only visible in **k8sd controller logs** (not in CRD status immediately)
6. Operator must: check CRD status → read controller logs → understand translation → check helm docs → realize CRD schema is out of sync

**TTR**: 30-90 minutes
**Debugging Steps**: 5+ (CRD status, controller logs, helm chart docs, schema comparison, CRD update wait)

---

## 2. Incident Response: 2AM Cluster Down Scenarios

### Scenario: Network Policy Misconfiguration Blocks CoreDNS

**2:00 AM**: PagerDuty alert → Cluster API server unreachable → All pods failing DNS lookups

#### ConfigMap Approach

**Investigation Path**:
```bash
# Step 1: Check what k8sd applied
kubectl get cm -n kube-system k8sd-coredns-config -o yaml
# Operator IMMEDIATELY sees the helm values that were applied
# Example: "enableIPv6: true" added, but cluster is IPv4-only

# Step 2: Validate helm release
helm get values -n kube-system coredns
# Confirms ConfigMap values are what helm received

# Step 3: Fix
kubectl edit cm -n kube-system k8sd-coredns-config
# Remove enableIPv6: true
# k8sd reconciles → helm updates → CoreDNS recovers

# TTR: 10-15 minutes
```

**Cognitive Load**: LOW
- One source of truth (helm values in ConfigMap)
- No translation layer to debug
- Helm error messages are authoritative
- Rollback = edit ConfigMap to previous values

#### CRD Approach

**Investigation Path**:
```bash
# Step 1: Check CRD
kubectl get corednsconfig -n kube-system default -o yaml
# Operator sees k8sd's representation (may not match helm)

# Step 2: Check if CRD was applied successfully
kubectl describe corednsconfig -n kube-system default
# Status field shows reconciliation state (if implemented)

# Step 3: Check k8sd controller logs
kubectl logs -n kube-system -l app=k8sd-controller --tail=500
# Search for helm errors in translation layer

# Step 4: Understand translation
# Q: Did CRD field "enableDualStack: true" translate to helm "enableIPv6: true"?
# Must read k8sd controller code or docs to understand mapping

# Step 5: Check actual helm release
helm get values -n kube-system coredns
# Confirms what was applied, but requires mapping back to CRD fields

# Step 6: Fix
kubectl edit corednsconfig -n kube-system default
# Modify CRD, wait for reconciliation

# TTR: 30-60 minutes
```

**Cognitive Load**: HIGH
- Two layers to correlate (CRD fields ≠ helm fields)
- Translation errors hidden in controller logs
- Must understand k8sd's field mapping
- Rollback = edit CRD + wait for reconciliation (no direct helm control)

---

## 3. Day-2 Operations: Configuration Change Workflows

### Use Case: Enable Cilium Hubble for Observability

#### ConfigMap Approach

```bash
# 1. User consults Cilium helm chart documentation
# https://github.com/cilium/cilium/tree/main/install/kubernetes/cilium

# 2. Create/update ConfigMap with helm values
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-cilium-config
  namespace: kube-system
data:
  values.yaml: |
    hubble:
      relay:
        enabled: true
      ui:
        enabled: true
EOF

# 3. k8sd reconciles → helm upgrade
# Helm validates → applies → done

# 4. Verify
helm get values -n kube-system cilium
kubectl get pods -n kube-system -l app.kubernetes.io/name=hubble-ui
```

**Workflow Complexity**: LOW
- Single source of truth (upstream helm docs)
- No CRD schema to check
- Helm validation is immediate
- GitOps-friendly (ConfigMap in git → apply)

#### CRD Approach

```bash
# 1. User consults TWO documentation sources
# - k8sd CiliumConfig CRD schema (does it support hubble fields?)
# - k8sd field mapping docs (how do CRD fields map to helm?)

# 2. Update CRD
kubectl edit ciliumconfig -n kube-system default
# Add hubble configuration (CRD field names may differ from helm)

# 3. Wait for reconciliation
# Controller watches CRD → translates → applies helm

# 4. Check reconciliation status
kubectl get ciliumconfig -n kube-system default -o yaml
# Look for status.conditions to see if applied successfully

# 5. Verify helm release
helm get values -n kube-system cilium
# Check if CRD fields translated correctly

# 6. If error → debug controller logs
kubectl logs -n kube-system -l app=k8sd-controller --tail=200 | grep cilium
```

**Workflow Complexity**: HIGH
- Two documentation sources (k8sd + helm)
- Reconciliation delay adds latency
- Validation happens asynchronously
- GitOps requires CRD in git (less common than ConfigMap patterns)

---

## 4. Rollback & Disaster Recovery

### Scenario: Bad Configuration Breaks Ingress

#### ConfigMap Approach

**Rollback Strategy**:
```bash
# Option 1: Edit ConfigMap to previous values
kubectl edit cm -n kube-system k8sd-ingress-config
# Revert changes → k8sd reconciles immediately

# Option 2: Use git history (GitOps)
git revert HEAD
git push
# Flux/ArgoCD syncs → ConfigMap restored → k8sd reconciles

# Option 3: Delete ConfigMap (fall back to defaults)
kubectl delete cm -n kube-system k8sd-ingress-config
# k8sd detects deletion → applies default helm values

# Option 4: Direct helm rollback (emergency escape hatch)
helm rollback -n kube-system ingress
# Bypasses k8sd temporarily for fast recovery
```

**Recovery Time Objective (RTO)**: 2-5 minutes
**Recovery Point Objective (RPO)**: Last known good ConfigMap (stored in git)

#### CRD Approach

**Rollback Strategy**:
```bash
# Option 1: Edit CRD to previous values
kubectl edit ingressconfig -n kube-system default
# Revert changes → wait for controller reconciliation

# Option 2: Use git history (GitOps)
git revert HEAD
git push
# Flux/ArgoCD syncs → CRD updated → controller reconciles → helm updates

# Option 3: Delete CRD (fall back to defaults)
kubectl delete ingressconfig -n kube-system default
# Controller detects deletion → reconciles with defaults

# Option 4: Direct helm rollback (BREAKS k8sd state)
helm rollback -n kube-system ingress
# ⚠️  Helm release now diverges from CRD state
# Controller will re-reconcile and overwrite rollback
# Operator must ALSO update CRD or disable controller
```

**Recovery Time Objective (RTO)**: 10-20 minutes (reconciliation loop adds latency)
**Recovery Point Objective (RPO)**: Last known good CRD (requires CRD-aware backup)

**Critical Risk**: Helm rollback without CRD update creates state drift → controller will re-apply bad config

---

## 5. Monitoring & Observability

### What Operators Need to Monitor

1. **Configuration drift detection**
2. **Reconciliation failures**
3. **Helm release health**
4. **Upgrade compatibility validation**

#### ConfigMap Approach

**Metrics**:
```prometheus
# k8sd controller metrics
k8sd_reconcile_duration_seconds{feature="cilium"}
k8sd_reconcile_errors_total{feature="cilium"}
k8sd_helm_release_version{feature="cilium",version="1.16.0"}

# Alerts
ALERT K8sdReconcileFailed
  IF rate(k8sd_reconcile_errors_total[5m]) > 0
  FOR 5m
  ANNOTATIONS {
    summary = "k8sd failing to reconcile {{ $labels.feature }}",
    description = "Check ConfigMap and helm errors"
  }
```

**Log Visibility**:
```bash
# Single log stream for all features
kubectl logs -n kube-system -l app=k8sd-controller -f

# Structured logs example
{"level":"error","feature":"cilium","msg":"helm upgrade failed","error":"values don't meet schema requirements"}
```

**Dashboard Queries**:
- ConfigMap version (git commit SHA annotation)
- Helm release version
- Last successful reconciliation timestamp
- Drift detection: `helm diff` between desired (ConfigMap) and actual

#### CRD Approach

**Metrics**:
```prometheus
# CRD controller metrics (per-CRD type)
k8sd_crd_reconcile_duration_seconds{crd_type="CiliumConfig"}
k8sd_crd_validation_errors_total{crd_type="CiliumConfig"}
k8sd_crd_translation_errors_total{crd_type="CiliumConfig"}
k8sd_helm_apply_errors_total{crd_type="CiliumConfig"}

# Alerts (more complex)
ALERT K8sdCRDReconcileFailed
  IF rate(k8sd_crd_reconcile_errors_total[5m]) > 0
  FOR 5m
  ANNOTATIONS {
    summary = "k8sd CRD {{ $labels.crd_type }} reconciliation failed",
    description = "Check CRD status, controller logs, and helm release state"
  }
```

**Log Visibility**:
```bash
# Multiple controllers (one per CRD type)
kubectl logs -n kube-system -l app=k8sd-cilium-controller -f
kubectl logs -n kube-system -l app=k8sd-coredns-controller -f
kubectl logs -n kube-system -l app=k8sd-ingress-controller -f

# Structured logs example (more verbose)
{"level":"error","crd":"CiliumConfig","namespace":"kube-system","name":"default","msg":"translation failed","field":"spec.tunnelMode","error":"unknown field in helm values"}
{"level":"error","crd":"CiliumConfig","namespace":"kube-system","name":"default","msg":"helm upgrade failed","error":"values don't meet schema"}
```

**Dashboard Queries**:
- CRD version (spec.version or metadata annotation)
- CRD status conditions (Ready, ValidationError, HelmApplyError)
- Helm release version
- Translation errors (CRD fields → helm values mismatches)
- Drift detection: CRD spec vs. helm values (requires custom logic)

**Operational Overhead**:
- ConfigMap: 1 controller, 3-5 metrics, simple logs
- CRD: N controllers (one per feature), 10-15 metrics, correlated logs across controllers

---

## 6. GitOps Integration Patterns

### Flux/ArgoCD Workflow Comparison

#### ConfigMap Approach

**Repository Structure**:
```
k8s-config/
├── clusters/
│   ├── prod/
│   │   ├── cilium-config.yaml      # ConfigMap with helm values
│   │   ├── coredns-config.yaml     # ConfigMap with helm values
│   │   └── ingress-config.yaml     # ConfigMap with helm values
│   └── staging/
│       └── ...
└── base/
    └── defaults/                    # Default helm values
```

**Flux HelmRelease Pattern** (familiar to GitOps operators):
```yaml
# This is how operators ALREADY manage helm in Flux
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: cilium
  namespace: kube-system
spec:
  chart:
    spec:
      chart: cilium
      version: 1.16.0
  valuesFrom:
    - kind: ConfigMap
      name: k8sd-cilium-config
      valuesKey: values.yaml
```

**k8sd ConfigMap Pattern** (mirrors Flux):
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-cilium-config
  namespace: kube-system
  annotations:
    config.gitops.toolkit/source: "git::https://github.com/org/k8s-config.git::clusters/prod/cilium-config.yaml"
data:
  values.yaml: |
    # Standard helm values — same format as Flux HelmRelease
    tunnel: disabled
    ipam:
      mode: kubernetes
```

**GitOps Workflow**:
1. Operator edits `clusters/prod/cilium-config.yaml` in git
2. Pull request → review → merge
3. Flux/ArgoCD syncs ConfigMap to cluster
4. k8sd detects ConfigMap change → reconciles helm release
5. Drift detection: `flux diff` or `argocd app diff` shows ConfigMap changes

**Benefits**:
- **Zero GitOps retraining**: Operators already know ConfigMap patterns
- **Standard Flux/ArgoCD diff tools** work out-of-box
- **Helm values are WYSIWYG**: What's in git is what helm receives
- **Audit trail**: Git commit = source of truth for compliance

#### CRD Approach

**Repository Structure**:
```
k8s-config/
├── clusters/
│   ├── prod/
│   │   ├── cilium-crd.yaml         # CiliumConfig CRD
│   │   ├── coredns-crd.yaml        # CoreDNSConfig CRD
│   │   └── ingress-crd.yaml        # IngressConfig CRD
│   └── staging/
│       └── ...
└── base/
    └── crds/                        # CRD definitions (schema)
```

**GitOps Workflow**:
1. Operator edits `clusters/prod/cilium-crd.yaml` in git
2. Pull request → review (but reviewer must understand CRD → helm translation)
3. Flux/ArgoCD syncs CRD to cluster
4. k8sd controller watches CRD → translates to helm values → reconciles
5. Drift detection: Must compare CRD spec + helm release (not just git diff)

**Challenges**:
- **Additional learning curve**: CRD fields ≠ helm values → operators must learn mapping
- **Diff tools less useful**: `flux diff` shows CRD changes, not resulting helm values
- **Audit trail gap**: Git shows CRD changes, but actual helm values are computed by controller
- **Schema version coupling**: CRD schema in `base/crds/` must stay in sync with k8sd version

**Multi-Cluster Complexity**:
- ConfigMap: Same helm values across clusters (easy to diff)
- CRD: CRD schema version must match k8sd version on each cluster → version skew risk

---

## 7. Enterprise Requirements: Compliance & Audit

### FedRAMP / CIS Kubernetes Benchmark Considerations

#### Audit Trail Requirements

**ConfigMap Approach**:
- ✅ **Git commit history** = complete audit trail of helm values
- ✅ **kubectl audit logs** show ConfigMap create/update/delete
- ✅ **Helm history** shows all releases: `helm history -n kube-system cilium`
- ✅ **Reproducibility**: ConfigMap + helm chart version = exact cluster state

**CRD Approach**:
- ✅ **Git commit history** = audit trail of CRD changes
- ✅ **kubectl audit logs** show CRD create/update/delete
- ⚠️  **Helm history** shows releases, but not CRD fields that generated them
- ⚠️  **Reproducibility requires**: CRD + controller version + helm chart version

**Compliance Gap**: CRD approach requires additional documentation of "CRD field X translates to helm value Y" for audit purposes.

#### RBAC Granularity

**ConfigMap Approach**:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: k8sd-config-manager
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["k8sd-cilium-config", "k8sd-coredns-config"]
  verbs: ["get", "update", "patch"]
```

**CRD Approach**:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8sd-crd-manager
rules:
- apiGroups: ["k8sd.io"]
  resources: ["ciliumconfigs", "corednsconfigs"]
  verbs: ["get", "update", "patch"]
- apiGroups: ["k8sd.io"]
  resources: ["ciliumconfigs/status", "corednsconfigs/status"]
  verbs: ["get"]  # Read-only access to reconciliation status
```

**Operational Note**: Both approaches support fine-grained RBAC. CRD has slight advantage of separate `/status` subresource for observability-only access.

---

## 8. What Keeps Operators Up at Night

### ConfigMap Approach: Operational Concerns

| Concern | Severity | Mitigation |
|---------|----------|------------|
| **Helm values validation** | Low | Helm validates immediately on apply |
| **Schema drift during upgrades** | Low | Helm chart errors are clear and actionable |
| **ConfigMap accidental deletion** | Medium | RBAC + Git source of truth + k8sd reconciles defaults |
| **Direct helm modifications** | Medium | k8sd reconciles ConfigMap state (eventual consistency) |
| **Upgrade breaking changes** | Low | Helm validation catches deprecated fields at apply time |

### CRD Approach: Operational Concerns

| Concern | Severity | Mitigation |
|---------|----------|------------|
| **CRD schema out of sync with helm** | **HIGH** | Requires k8sd releases synchronized with every helm chart bump |
| **Translation bugs** | **HIGH** | CRD field → helm value mapping errors surface only in controller logs |
| **Validation bypass** | **HIGH** | CRD validates at apply, helm validates at reconcile → delayed error discovery |
| **Controller version skew** | **MEDIUM** | Multi-cluster environments must pin k8sd + CRD versions together |
| **CRD migration on upgrades** | **MEDIUM** | Schema changes require conversion webhooks or manual migration |
| **Debugging translation layer** | **HIGH** | Operator must understand k8sd controller internals |
| **Direct helm modifications** | **HIGH** | CRD controller overwrites → requires disabling controller for emergency fixes |

---

## 9. Operational Recommendation

### Recommended Approach: **ConfigMap**

**Rationale**:
1. **Reduced Mean Time to Recovery (MTTR)**: 2-3x faster incident resolution due to direct helm visibility
2. **Lower cognitive load**: Single source of truth (helm values) vs. three-layer abstraction (CRD → translation → helm)
3. **Predictable upgrades**: Helm validation catches breaking changes immediately, not asynchronously
4. **GitOps compatibility**: Follows established patterns (Flux HelmRelease, RKE2 HelmChartConfig, K3s HelmChart)
5. **Operational simplicity**: One controller, simple logs, no translation debugging

### When CRD Might Be Acceptable

**Only if ALL of these are true**:
- k8sd owns and maintains the helm charts (Cilium, CoreDNS are external → this is FALSE)
- Schema stability is guaranteed (helm charts rarely change → this is FALSE for CNI/DNS)
- Operators are willing to learn custom CRD schemas instead of upstream helm docs
- Organization has mature CRD management practices (conversion webhooks, version migration playbooks)

**For k8s-snap's use case** (deploying external helm charts), the ConfigMap approach is **operationally superior**.

---

## 10. Operational Playbooks

### Playbook 1: Emergency Rollback

**ConfigMap**:
```bash
# 1. Identify last known good config
git log -- clusters/prod/cilium-config.yaml

# 2. Revert ConfigMap
kubectl apply -f clusters/prod/cilium-config.yaml.backup

# 3. Verify reconciliation
watch kubectl get pods -n kube-system -l app.kubernetes.io/name=cilium

# TTR: 3-5 minutes
```

**CRD**:
```bash
# 1. Identify last known good CRD
git log -- clusters/prod/cilium-crd.yaml

# 2. Revert CRD
kubectl apply -f clusters/prod/cilium-crd.yaml.backup

# 3. Wait for controller reconciliation
kubectl get ciliumconfig -n kube-system default -o yaml
# Check status.conditions for reconciliation state

# 4. Verify helm release
helm get values -n kube-system cilium

# 5. If controller is broken, emergency helm rollback
helm rollback -n kube-system cilium
# ⚠️  Must also disable controller to prevent re-apply

# TTR: 10-20 minutes
```

### Playbook 2: Configuration Validation Pre-Apply

**ConfigMap**:
```bash
# 1. Dry-run helm validation
helm upgrade --dry-run --debug \
  -n kube-system cilium cilium/cilium \
  -f proposed-cilium-config.yaml

# 2. If validation passes, apply ConfigMap
kubectl apply -f proposed-cilium-config.yaml

# TTR: 1-2 minutes for validation
```

**CRD**:
```bash
# 1. Validate CRD schema
kubectl apply --dry-run=client -f proposed-cilium-crd.yaml

# 2. Manual translation check (no automated tool)
# Compare CRD fields to helm chart documentation

# 3. Apply CRD
kubectl apply -f proposed-cilium-crd.yaml

# 4. Monitor controller logs for translation errors
kubectl logs -n kube-system -l app=k8sd-controller -f | grep cilium

# 5. Check helm release after reconciliation
helm get values -n kube-system cilium

# TTR: 5-10 minutes for validation + monitoring
```

### Playbook 3: Upgrade Testing (Staging → Prod)

**ConfigMap**:
```bash
# Staging cluster
kubectl apply -f staging/cilium-config.yaml
# Wait for helm reconcile
helm get values -n kube-system cilium
# Test cluster networking

# If success → promote to prod (same ConfigMap format)
kubectl apply -f prod/cilium-config.yaml --context=prod

# Confidence: HIGH (helm values are portable across environments)
```

**CRD**:
```bash
# Staging cluster
kubectl apply -f staging/cilium-crd.yaml
# Wait for controller reconcile
kubectl get ciliumconfig -n kube-system default -o yaml
# Check status + helm values
helm get values -n kube-system cilium

# Before promoting to prod → verify k8sd controller version matches
kubectl get deployment -n kube-system k8sd-controller -o yaml | grep image:

# If controller versions differ → CRD schema may be incompatible
# Must upgrade k8sd on prod first, THEN apply CRD

# Confidence: MEDIUM (version coupling risk)
```

---

## 11. Conclusion

From an operational perspective, the **ConfigMap approach** is the clear winner for k8sd's use case of deploying external helm charts. It provides:

- **Faster incident response** (2-3x faster MTTR)
- **Simpler troubleshooting** (single source of truth)
- **Predictable upgrades** (immediate helm validation)
- **Lower operational burden** (no translation layer to debug)
- **Better GitOps integration** (follows established patterns)

The CRD approach introduces significant operational complexity — a "permanent schema synchronization tax" — that does not provide commensurate value when k8sd does not own the helm charts it deploys.

**Recommendation**: Implement the ConfigMap approach for production deployments. Reserve CRDs for scenarios where k8sd owns the schema and can guarantee synchronization with the underlying implementation.

---

**Document Version**: 1.0
**Last Updated**: 2024
**Feedback**: Platform Engineering Team
