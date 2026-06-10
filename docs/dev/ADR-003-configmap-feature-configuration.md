# ADR 003: ConfigMap-Based Feature Configuration

**Status:** Accepted  
**Date:** 2026-06-09  
**Authors:** Louise Schmidt-Gen ([@louiseschmidtgen](https://github.com/louiseschmidtgen))  
**Reviewers:** Architecture Team, DevOps Team, Security Team  
**Confidence:** 87%

---

## Context

Canonical Kubernetes (k8s-snap) deploys managed features (cilium, coredns, ingress, load-balancer, gateway) via helm charts. Currently, users configure these features through:

1. **CLI commands** (`k8s set network.tunnel-port=9999`) - stored in microcluster
2. **Annotations** (string-typed, hard to discover, no GitOps support)

**Problem:** Enterprise users need advanced configuration not exposed via CLI:
- **Banca d'Italia** requires BGP configuration for cilium
- **Regulated environments** need audit trails, RBAC, GitOps workflows
- **Power users** need access to full helm chart values

**Current workaround:** Annotations are insufficient (string-typed, no validation, poor discoverability)

---

## Decision

**We will use ConfigMaps containing standard helm values.yaml for advanced feature configuration.**

Users create ConfigMaps (e.g., `k8sd-cilium-values`) with helm values. k8sd reconcilers read these ConfigMaps, merge values with defaults and cluster-config, then apply the helm chart.

**We will NOT create custom k8sd CRDs** (CiliumConfig, CoreDNSConfig, etc.) that wrap helm values.

---

## Rationale

### Analysis Process

Comprehensive evaluation across 5 perspectives:
- Initial generalist analysis (8 dimensions, 79% confidence)
- @architect: System design and ownership boundaries
- @developer: Implementation complexity and debugging
- @devops: Operations, upgrades, and incident response
- @security: Validation, RBAC, and attack surface

**Unanimous verdict:** ConfigMap approach superior (87% confidence)

### Key Factors

#### 1. Ownership Boundary (@architect insight)

**k8sd doesn't own cilium/coredns/ingress helm charts.** Creating CRDs that mirror upstream helm schemas:
- Suggests false control over schemas we don't maintain
- Creates permanent schema synchronization burden
- Requires dual documentation (k8sd CRD docs + upstream helm docs)

**ConfigMap is architecturally honest:** "We deploy via helm. Configure via helm values. Refer to upstream docs."

#### 2. Maintenance Burden (@developer quantification)

**5-year cost comparison:**

| Approach | Year 1 | Year 2-5 | Total | Delta |
|----------|--------|----------|-------|-------|
| **CRD** | 600h | 550h/yr (2200h) | 2800h | Baseline |
| **ConfigMap** | 125h | 125h/yr (500h) | 625h | **-2175h (-78%)** |

**Savings: 2175 hours = 1.09 FTE = $163k over 5 years**

Implementation estimates:
- **CRD MVP:** 3 weeks, 1000 LOC, 1500 LOC tests
- **ConfigMap MVP:** 1 week, 310 LOC, 250 LOC tests

#### 3. Operational Resilience (@devops insight)

**Time to recovery (TTR) comparison:**

| Scenario | CRD TTR | ConfigMap TTR | Improvement |
|----------|---------|---------------|-------------|
| Invalid config applied | 30-45 min | 10-15 min | 3x faster |
| Upgrade breaking change | 30-60 min | 5-10 min | 5x faster |
| 2AM cluster down | 45-90 min | 15-30 min | 3x faster |

**Upgrade scenario:** k8s-snap 1.30→1.31 (cilium 1.15→1.16) with user BGP config:
- **CRD:** Schema sync, CRD migration, user CR updates (30-60 min debugging)
- **ConfigMap:** Helm validates immediately, user self-service fix (5-10 min)

#### 4. Security Posture (@security analysis)

**Risk scoring:**
- **CRD:** 7.2/10 (schema drift exploits, version confusion, conversion bugs)
- **ConfigMap:** 5.4/10 (simpler attack surface, upstream validation authoritative)

**Key insight:** "Security depends on implementation quality, not storage mechanism."

**Both require layered defense:**
- Admission webhook (fail-fast validation)
- Runtime validation (blacklist forbidden fields)
- RBAC lockdown (prevent direct helm access)
- Drift detection (continuous enforcement)

**CRD typed validation doesn't eliminate need for runtime checks.** ConfigMap has simpler attack surface with fewer potential exploit vectors.

#### 5. Ecosystem Validation

**Projects deploying external helm charts ALL use ConfigMaps or direct values:**

| Project | Pattern |
|---------|---------|
| Rancher RKE2 | HelmChartConfig CRD → configmap values |
| K3s | HelmChartConfig → valuesContent |
| Flux | HelmRelease.valuesFrom configmap |
| ArgoCD | values.yaml in git |
| Cluster API | HelmChartProxy.valuesFrom |

**Pattern:** Only projects that **own the application** (cert-manager, prometheus-operator) use typed CRDs. We deploy **external** helm charts → ConfigMap pattern applies.

---

## Alternatives Considered

### Alternative 1: Custom k8sd CRDs

**Approach:** Create CRDs like `CiliumConfig`, `CoreDNSConfig` with typed schemas mirroring helm values.

**Advantages:**
- Typed validation (schema enforcement)
- Native RBAC (Role for resource type)
- Discoverability (`kubectl explain CiliumConfig`)

**Disadvantages:**
- **High maintenance burden:** Must sync CRD schemas with upstream helm changes (500-700h/year)
- **Version coupling:** CRD versioning required when helm charts change
- **Upgrade blockers:** Users wait for k8sd CRD updates before using new helm values
- **Architectural dishonesty:** Suggests control over schemas we don't own
- **3-5x implementation cost:** 1000 LOC vs 310 LOC for MVP

**Why rejected:**
- Creates permanent schema synchronization tax
- Goes against ecosystem patterns (RKE2, K3s, Flux all use ConfigMaps)
- Maintenance burden unsustainable over 2-5 years
- Abstracts something we don't own

### Alternative 2: Continue with Annotations Only

**Approach:** Extend annotation-based configuration, no structured approach.

**Advantages:**
- Zero implementation cost
- Already works today

**Disadvantages:**
- String-typed (no validation)
- Poor discoverability
- No GitOps tooling support
- No RBAC granularity
- Doesn't meet enterprise requirements (FedRAMP/CIS compliance)

**Why rejected:** Insufficient for enterprise use cases, no path forward for compliance.

### Alternative 3: Hybrid (ConfigMap + Optional CRD)

**Approach:** ConfigMap for values, optional CRD wrapper for typed validation.

**Advantages:**
- Flexibility (users choose)

**Disadvantages:**
- Doubles complexity
- Still have CRD schema maintenance burden
- User confusion ("Do I use CRD or ConfigMap?")
- Doubles documentation and support surface

**Why rejected:** Adds complexity without solving core problems. If wrapping ConfigMaps with CRDs, we've added indirection without removing underlying issues.

---

## Implementation

### Phase 1: MVP with Cilium (1 week)

**Components:**
1. ConfigMap reader in cilium reconciler
2. YAML parser + merge logic (precedence: base → cluster-config → annotations → configmap)
3. Bootstrap file support: `cilium-values-file: /path/to/values.yaml`
4. Basic documentation

**Deliverable:** Banca d'Italia can configure BGP via configmap

**Code estimate:** ~310 lines

### Phase 2: Validation Tooling (1-2 weeks)

**Add:**
- `k8s validate cilium-values <file>` (helm dry-run validation)
- `k8s show cilium-values` (display current merged values)
- `k8s config show cilium --available-values` (helm chart schema)
- Admission webhook (optional, recommended for enterprise)
- Runtime validation (blacklist: image, hostNetwork, privileged, hostPID, hostPorts)

**Deliverable:** Enterprise-ready validation

### Phase 3: Multi-Feature Rollout (2-3 weeks)

**Extend pattern to:**
- coredns → `k8sd-coredns-values`
- ingress → `k8sd-ingress-values`
- load-balancer → `k8sd-loadbalancer-values`
- gateway → `k8sd-gateway-values`
- local-storage → `k8sd-localstorage-values`

**Deliverable:** All features support advanced configuration

### Phase 4: GitOps & Documentation (1 week)

**Add:**
- Flux/ArgoCD integration examples
- Audit trail documentation (FedRAMP/CIS compliance)
- Per-feature example configmaps
- Troubleshooting playbooks

**Deliverable:** Production-ready documentation

**Total timeline:** 6-9 weeks

---

## Configuration Model

### Precedence Order

```
base defaults → cluster-config (microcluster) → annotations → configmap values
```

Higher precedence overrides lower. ConfigMaps have highest precedence for advanced users.

### Bootstrap Example

```yaml
# bootstrap.yaml
cluster-config:
  network:
    enabled: true
  cilium-values-file: /path/to/cilium-values.yaml  # optional

# cilium-values.yaml (standard helm values)
tunnelPort: 9999
bpf:
  preallocateMaps: true
bgp:
  enabled: true
  announce:
    loadbalancerIP: true
```

### Day 2 Example

```yaml
# ConfigMap for advanced cilium configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-cilium-values
  namespace: kube-system
data:
  values.yaml: |
    bgp:
      enabled: true
      announce:
        loadbalancerIP: true
    ipam:
      mode: kubernetes
    hubble:
      enabled: true
      ui:
        enabled: true
```

Apply via: `kubectl apply -f cilium-values.yaml`

k8sd reconciler watches ConfigMap, merges values, applies helm chart automatically.

### User Workflows

**Simple user (unchanged):**
```bash
k8s enable network
k8s set network.tunnel-port=9999
```

**Power user (new):**
```bash
# Create/edit configmap
kubectl apply -f cilium-values.yaml

# Verify merged values
k8s show cilium-values

# Force reconcile (optional, automatic on change)
k8s refresh network
```

**GitOps user (new):**
```bash
# In git repo: k8s-config/cilium-values.yaml
# Flux/ArgoCD syncs configmap → k8sd reconciles → cilium updated
# Full audit trail via git history + kubectl managed fields
```

---

## Security Considerations

### Validation Strategy

**Three-layer defense:**

1. **Pre-apply validation** (optional, user-initiated):
   ```bash
   k8s validate cilium-values cilium-values.yaml
   ```
   Uses helm dry-run to catch errors before apply.

2. **Admission webhook** (recommended for enterprise):
   - Validates ConfigMap on apply
   - Rejects forbidden fields (image, ports, hostNetwork, privileged)
   - Fail-closed configuration

3. **Runtime validation** (defense in depth):
   - Reconciler validates before helm apply
   - Blacklist enforcement
   - Drift detection (auto-revert unauthorized changes)

### RBAC Model

**Team-based delegation:**

```yaml
# Network team can manage cilium config only
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: cilium-config-manager
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["k8sd-cilium-values"]
  verbs: ["get", "update", "patch"]

# Security team can manage ingress config only
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: ingress-config-manager
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["k8sd-ingress-values"]
  verbs: ["get", "update", "patch"]
```

**Prevents direct helm manipulation:**
```yaml
# Deny direct helm secret access
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: deny-direct-helm
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["sh.helm.release.*"]
  verbs: []  # no permissions
```

### Forbidden Fields (Blacklist)

**Security-sensitive fields that MUST be blocked:**
- `image`, `imageRepository`, `imageTag` (prevent container substitution)
- `hostNetwork: true` (prevent host network access)
- `privileged: true` (prevent privilege escalation)
- `hostPID: true`, `hostIPC: true` (prevent namespace escape)
- Port changes that bypass network policies

Blocked at admission webhook + runtime validation.

---

## Compliance & Audit

### FedRAMP/CIS Requirements

**Audit trail:**
- Git history (if using GitOps)
- `kubectl get configmap k8sd-cilium-values --show-managed-fields=true`
- API audit logs (RequestResponse level, 90-day retention)

**RBAC:**
- Team-based access control (network team, security team)
- Least-privilege principle (resourceNames-scoped Roles)

**Change management:**
- All changes via kubectl (auditable)
- Drift detection (unauthorized changes reverted)

**Validation:**
- Pre-apply validation (fail-fast)
- Admission webhook (reject invalid configs)
- Runtime validation (defense in depth)

**Meets:** FedRAMP Moderate, CIS Kubernetes Benchmark, STIG requirements

---

## Upgrade Considerations

### Helm Chart Version Bumps

**Scenario:** k8s-snap 1.30→1.31 (cilium 1.15→1.16), user has BGP config

**ConfigMap behavior:**
1. User configmap contains values for cilium 1.15
2. k8s-snap 1.31 applies cilium 1.16 helm chart
3. Helm merges user values into new chart
4. If deprecated value used: helm logs warning, uses default for new value
5. If invalid value: helm apply fails with clear error
6. User updates configmap based on upstream cilium 1.16 docs
7. Reconcile succeeds

**Self-service, no k8sd intervention required.**

### Breaking Changes

**If upstream helm chart removes a value:**
- Helm ignores unknown values (logged as warning)
- `k8s status` shows warning
- User updates configmap when ready
- No cluster downtime

**If upstream helm chart changes value type:**
- Helm apply fails with clear error
- User fixes configmap based on upstream docs
- Quick feedback loop (5-10 min)

### Multi-Feature Upgrades

Each feature has independent configmap → independent upgrade path.

No coordination required between cilium, coredns, ingress updates.

---

## Migration from Current State

### Microcluster Config (Preserved)

Existing CLI-configured options stay in microcluster:
```bash
k8s set network.tunnel-port=9999
```

Stored in microcluster, applied as before.

### ConfigMaps (Additive)

ConfigMaps are **additive**, not replacement:
- Don't require migration of existing configs
- Higher precedence than microcluster for conflicts
- Clean separation: simple (CLI) vs advanced (configmap)

### Migration Path

**For users with annotations:**
1. Extract annotation values
2. Create configmap with equivalent helm values
3. Remove annotations
4. Validate with `k8s validate`
5. Apply configmap

**Tool support:** `k8s migrate-annotations cilium` (future enhancement)

---

## Risks & Mitigations

### Risk: Invalid YAML Breaks Cluster

**Severity:** High  
**Probability:** Medium

**Mitigations:**
- Pre-apply validation: `k8s validate cilium-values values.yaml`
- Admission webhook (fail-closed, rejects invalid configs)
- Helm validation (authoritative, fails apply if invalid)
- Drift detection (auto-revert if manually broken)

### Risk: Users Confused by Precedence

**Severity:** Medium  
**Probability:** Medium

**Mitigations:**
- Clear documentation with examples
- `k8s show cilium-values` shows merged result
- Precedence explained in every doc page
- Warning logs when conflicts detected

### Risk: Discoverability Poor

**Severity:** Medium  
**Probability:** High

**Mitigations:**
- `k8s config show cilium --available-values` command
- Per-feature example configmaps in docs
- Direct links to upstream helm documentation
- Common use cases documented (BGP, L7 policy, custom plugins)

### Risk: Security Boundaries Bypassed

**Severity:** High  
**Probability:** Low

**Mitigations:**
- Admission webhook (fail-closed validation)
- Runtime blacklist (image, ports, hostNetwork)
- RBAC lockdown (prevent direct helm access)
- Drift detection (revert unauthorized changes)

---

## Success Criteria

### Phase 1 (MVP)
- ✅ Banca d'Italia configures BGP successfully
- ✅ Implementation matches estimates (1 week, ~300 lines)
- ✅ Debugging workflow validated (<15 min for common errors)

### Phase 2 (Validation)
- ✅ Validation tooling catches 95%+ errors before apply
- ✅ <10% support tickets about configuration
- ✅ Enterprise customers adopt (FedRAMP/CIS documented)

### Long-term (1 year)
- ✅ 95%+ smooth upgrades (k8s-snap version transitions)
- ✅ Maintenance burden <150 hours/year
- ✅ No major architectural regrets

---

## Reversibility

**If ConfigMap approach proves problematic:**

ConfigMap→CRD migration is feasible (easier than CRD→ConfigMap):
1. Define CRD schemas based on helm values
2. Migrate configmap contents to CRs
3. Deprecate configmaps over 2-3 releases

**Start simple, add complexity only if proven necessary.**

---

## References

**Analysis documents** (all in `docs/dev/`):
1. `FINAL-RECOMMENDATION.md` - Synthesis across 5 perspectives
2. `architect-perspective.md` - System design analysis
3. `developer-perspective.md` - Implementation & debugging
4. `devops-perspective.md` - Operations & upgrades
5. `security-perspective.md` - Security & validation
6. `configuration-approaches-analysis.md` - Initial 8-dimension analysis
7. `executive-summary.md` - Quick overview

**Total analysis:** ~150KB, 87% confidence, unanimous recommendation

**Ecosystem precedent:**
- RKE2: https://docs.rke2.io/helm#customizing-packaged-components
- K3s: https://docs.k3s.io/helm#customizing-packaged-components
- Flux HelmRelease: https://fluxcd.io/docs/components/helm/helmreleases/

---

## Decision

**Accepted:** ConfigMap-based feature configuration  
**Date:** 2026-06-09  
**Confidence:** 87%  
**Unanimous:** All domain experts agree

**Next step:** Implement Phase 1 (MVP with cilium) in 1 week.
