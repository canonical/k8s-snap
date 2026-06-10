# Executive Summary: Configuration Approach Recommendation

**Date:** 2026-06-09  
**Decision:** Adopt ConfigMap Approach  
**Confidence:** High (76% weighted score vs 50%)

---

## TL;DR

**Use ConfigMaps to provide helm values overrides, NOT custom k8sd CRDs.**

**Why:** Simpler, faster, more maintainable, aligns with Kubernetes ecosystem patterns (RKE2, K3s, Flux all do this).

---

## The Question

How should enterprise users configure advanced features (like BGP in Cilium) in Canonical Kubernetes?

**Two options:**
1. **CRD Approach:** Create k8sd-specific CRDs (`CiliumConfig`, `CoreDNSConfig`) that wrap helm values
2. **ConfigMap Approach:** Let users provide helm values directly via ConfigMaps

---

## The Recommendation

### ✅ ConfigMap Approach

**Implementation:**
```yaml
# Bootstrap
cluster-config:
  cilium-values-file: /path/to/values.yaml

# Day 2
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-cilium-values
data:
  values.yaml: |
    bgp:
      enabled: true
      announce:
        loadbalancerIP: true
```

**Reconcile logic:**
```
base values → cluster-config → annotations → configmap → helm apply
```

---

## Score Comparison

| Dimension | Weight | CRD | ConfigMap | Winner |
|-----------|--------|-----|-----------|--------|
| **User Experience** | 10 | 4/10 | 7/10 | ConfigMap |
| **Implementation Cost** | 8 | 3/10 | 9/10 | ConfigMap |
| **Enterprise Features** | 10 | 9/10 | 7/10 | CRD |
| **Upstream Alignment** | 9 | 5/10 | 8/10 | ConfigMap |
| **Upgrade Resilience** | 9 | 4/10 | 8/10 | ConfigMap |
| **Abstraction Quality** | 7 | 4/10 | 8/10 | ConfigMap |
| **Migration Ease** | 6 | 5/10 | 7/10 | ConfigMap |
| **Debugging** | 8 | 5/10 | 7/10 | ConfigMap |

**Weighted Total:**
- **ConfigMap:** 510/670 (76%)
- **CRD:** 333/670 (50%)

**Winner:** ConfigMap by 53% margin

---

## Why ConfigMap Wins

### 1. Simpler Implementation (9/10 vs 3/10)
**ConfigMap:**
- Read configmap
- Merge into helm values
- Apply helm chart
- **~50-100 lines of code**

**CRD:**
- Define CRD schemas for each feature
- Implement controller + watchers
- Version CRDs on upstream changes
- Migration logic
- **~5000+ lines of code + ongoing maintenance**

### 2. Better Upgrades (8/10 vs 4/10)
**Scenario:** Cilium helm chart removes a deprecated value

**ConfigMap:**
- User's configmap has old value
- Helm logs warning, ignores it
- User updates configmap when ready
- **Self-service, no blocker**

**CRD:**
- k8sd CRD schema has old value
- Must update CRD schema
- Must migrate user CRs
- Users blocked until k8sd releases new version
- **k8sd gates the upgrade**

### 3. Upstream Alignment (8/10 vs 5/10)
**ConfigMap:** Aligns with ecosystem patterns
- RKE2: `HelmChartConfig` CRD → values configmap
- K3s: `HelmChartConfig` → valuesContent
- Flux: `HelmRelease.valuesFrom` configmaps
- ArgoCD: values.yaml in git

**Pattern:** Projects deploying **external helm charts** use **ConfigMaps or direct values**.

**CRD:** Creates k8sd-specific abstraction layer
- User must learn k8sd schema
- k8sd schema must mirror helm values
- Adds indirection: `k8sd CRD → helm values → upstream CRD`
- Maintenance burden: track upstream changes

### 4. Honest Abstraction (8/10 vs 4/10)
**ConfigMap:** "We use helm. Provide values.yaml."
- Clear contract
- Users reference upstream helm docs
- No pretense of hiding implementation

**CRD:** "We hide helm... but CRD fields are helm values"
- Leaky abstraction
- Duplicates helm schema in k8sd format
- Must document "this k8sd field maps to this helm value"

**Analogy:** CRD approach is like creating a custom API for SQL queries instead of letting users write SQL. Doesn't hide complexity; duplicates it.

---

## Where CRD Was Better

### Typed Validation ✅
**CRD:** Schema enforcement at API level  
**ConfigMap:** Unstructured YAML

**Mitigation:** Add `k8s validate cilium-values values.yaml` command (helm dry-run)

### RBAC Elegance ✅
**CRD:** `Role` for `ciliumconfigs` resource  
**ConfigMap:** `Role` for `configmap/k8sd-cilium-values`

**Verdict:** ConfigMap RBAC works, just less elegant

### Discoverability ✅
**CRD:** `kubectl explain CiliumConfig.spec`  
**ConfigMap:** Read upstream helm docs

**Mitigation:** 
- `k8s show cilium-schema` command
- Documentation with examples per use case

**Enterprise Features:** CRD scored 9/10 vs ConfigMap 7/10. But configmaps still provide:
- ✅ Audit trail (kubectl managed fields)
- ✅ RBAC (configmap-based)
- ✅ GitOps (standard pattern)

**Gaps are mitigatable.**

---

## Implementation Timeline

### ConfigMap Approach
**Phase 1 (MVP):** 2-3 weeks
- Configmap reading + merge logic
- Bootstrap file support
- **Unlocks Banca d'Italia BGP use case**

**Phase 2 (Validation):** 1-2 weeks
- `k8s validate` command
- `k8s show values` command
- Runtime security validation

**Phase 3 (Rollout):** 2-3 weeks
- Apply to all features (coredns, ingress, lb, etc.)
- Multi-feature documentation

**Total:** 6-9 weeks

### CRD Approach
**Phase 1 (Foundation):** 4-6 weeks
- CRD schema design
- Controller implementation
- Migration logic from microcluster

**Phase 2 (Features):** 3-4 weeks
- CRDs for each feature
- Testing + validation

**Phase 3 (Versioning):** 2-3 weeks
- CRD version strategy
- Conversion logic

**Total:** 12-16 weeks

**ConfigMap delivers in half the time.**

---

## Addressing Team Concerns

### "Helm is an implementation detail we shouldn't expose"
**Counter:** We already expose it via annotations. ConfigMap is honest about helm usage rather than pretending to hide it with a k8sd wrapper.

Users who need advanced config are sophisticated enough to understand helm values.

### "Three config layers (CLI, ConfigMap, upstream CRD) is confusing"
**Response:** CRDs don't eliminate this:
- CLI for basic options
- k8sd CRD for helm values
- Upstream CRD for cilium-native features

**ConfigMap makes the distinction functional:**
- CLI: quick settings
- ConfigMap: advanced helm values
- Upstream CRD: cilium-native resources

### "CRDs provide better enterprise UX"
**Data:** 
- Enterprise precedent uses ConfigMaps (RKE2, K3s, Flux)
- Typed validation nice-to-have, not blocker (helm validates anyway)
- GitOps works with ConfigMaps (standard pattern)

---

## What About Both? (Hybrid)

**Could we:** ConfigMap for values + optional CRD for typed validation?

**Why not:**
- Doubles complexity
- Still have schema maintenance burden
- Users confused: "Do I use CRD or ConfigMap?"
- Adds indirection without solving core problem

**Verdict:** Pick one approach, execute well.

---

## Risk Assessment

### ConfigMap Risks

| Risk | Mitigation |
|------|------------|
| Invalid YAML breaks cluster | `k8s validate` command, helm dry-run |
| Users confused by precedence | Clear docs, `k8s show values` command |
| Upgrade compatibility | Log warnings, documentation |

**All risks are mitigatable.**

### CRD Risks

| Risk | Mitigation |
|------|------------|
| Schema maintenance burden | Must track upstream changes (ongoing) |
| CRD versioning complexity | Conversion webhooks, migration tooling (high cost) |
| User confusion | Extensive documentation (doesn't solve root cause) |
| Upgrade blockers | Users wait for k8sd updates (unmitigatable) |

**CRD risks are higher likelihood and harder to mitigate.**

---

## Decision Rationale

**Core insight:** k8sd doesn't **own** cilium/coredns/ingress. We deploy them via helm.

**Pattern from ecosystem:**
- Projects that **own the app** (cert-manager, prometheus-operator): use typed CRDs ✅
- Projects that **deploy external helm charts** (RKE2, K3s, Flux, ArgoCD): use ConfigMaps or direct values ✅

**k8sd is in the second category.**

Wrapping helm values in k8sd CRDs creates:
- Schema maintenance burden without ownership benefits
- Version coupling without control over upstream
- Abstraction leak (CRD fields ARE helm values, just renamed)

**ConfigMap approach:**
- Honest about helm usage
- Aligns with ecosystem patterns
- Faster to deliver, easier to maintain
- 90% of enterprise benefits at 20% of cost

---

## Recommendation

### ✅ Adopt ConfigMap Approach

**Next Steps:**
1. **Week 1:** Prototype configmap implementation for cilium
2. **Weeks 2-3:** Deliver Phase 1 MVP (unlocks Banca d'Italia)
3. **Weeks 4-5:** Add validation tooling (`k8s validate`, `k8s show values`)
4. **Weeks 6-9:** Roll out to all features

**Success Criteria:**
- Banca d'Italia configures BGP successfully
- <10% support tickets related to configuration
- 95%+ smooth upgrades without config changes

### 🚫 Do Not Implement CRD Approach

**Reason:** Complexity doesn't justify marginal benefits. We'd be fighting the ecosystem pattern, increasing maintenance burden, and creating user confusion.

---

## Conclusion

**The ConfigMap approach wins decisively:**
- Faster delivery (half the time)
- Lower maintenance (no schema versioning)
- Better upgrades (self-service)
- Ecosystem alignment (proven pattern)
- Honest abstraction (embraces helm)

**Trade-offs are acceptable:**
- Unstructured YAML mitigated with validation tooling
- ConfigMap RBAC works (less elegant but functional)
- Discoverability through docs (upstream helm docs are canonical anyway)

**Ship the ConfigMap approach.**

---

**Questions or Concerns?**

See full analysis: `configuration-approaches-analysis.md` (22KB, comprehensive edge case coverage)
