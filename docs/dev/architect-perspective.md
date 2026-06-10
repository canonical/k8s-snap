# Architectural Analysis: k8sd Feature Configuration Approaches

**Date:** 2025-01-24  
**Status:** Recommendation  
**Decision Type:** Architecture

## Context

k8sd acts as an installer/operator for Canonical Kubernetes, deploying external helm charts (cilium, coredns, ingress, gateway) that it **does not own**. Users need advanced configuration capabilities (BGP, custom plugins) not exposed through the `k8s set` CLI. The current workaround uses annotations, which are string-typed and have poor discoverability.

**The fundamental architectural question:** How should k8sd expose configuration for components it deploys but doesn't own?

## Options

### Option 1: Custom k8sd CRDs (Wrapping Helm Values)

Create k8sd-owned CRDs (CiliumConfig, CoreDNSConfig) that mirror upstream helm chart schemas. A controller watches these CRDs and reconciles helm releases with values derived from the CR spec.

**Architecture:**
```
User → CiliumConfig CR → k8sd controller → helm chart → cilium
```

### Option 2: ConfigMaps for Direct Helm Values

Users create ConfigMaps containing standard helm `values.yaml` content. The controller reads ConfigMaps, merges values, and applies helm charts directly.

**Architecture:**
```
User → ConfigMap (helm values) → k8sd controller → helm chart → cilium
```

## Trade-off Analysis

| **Dimension** | **Custom CRDs** | **ConfigMaps** |
|---------------|-----------------|----------------|
| **Abstraction Level** | Abstraction leak: pretends to own cilium schema while actually proxying to helm. Creates a "k8sd dialect" of cilium configuration. | Honest: explicitly says "we deploy helm charts, here are the helm values." No pretense of ownership. |
| **Schema Maintenance** | HIGH BURDEN: Must version CRDs whenever upstream helm charts change. Breaking changes in cilium → breaking changes in k8sd CRDs. | LOW BURDEN: No schema to maintain. Upstream adds fields → users can use them immediately. |
| **Separation of Concerns** | Blurred: k8sd API surface includes detailed cilium configuration. Unclear where k8sd responsibility ends. | Clear: k8sd provides lifecycle (install/upgrade). Configuration is explicitly "bring your own helm values." |
| **Validation** | Structural validation via OpenAPI schema. Runtime validation complex (must validate against helm schema). | Structural validation limited. Validation happens at helm apply time (fail-fast at reconcile). |
| **Discoverability** | `kubectl explain CiliumConfig` works. But documentation is a shadow of upstream helm docs. | Must reference upstream helm documentation. Less Kubernetes-native discovery. |
| **Versioning Burden** | CRD version tracking: v1alpha1, v1beta1, v1. Deprecation policies. Conversion webhooks when upstream schema changes significantly. | ConfigMap schema is unversioned. Breaking changes are user's responsibility (same as helm users). |
| **Operational Complexity** | More CRDs in cluster. CRD lifecycle management (install, upgrade, cleanup). | Simpler: ConfigMaps are Kubernetes primitives. |
| **Upstream Ownership** | **Critical mismatch**: We don't control cilium helm schema evolution. We're creating an API surface we can't control. | Acknowledges we don't own the schema. Users understand they're working with external components. |
| **Ecosystem Precedent** | Rare pattern for deploying third-party helm charts. | **Strong precedent**: RKE2, K3s, Flux HelmRelease, ArgoCD ApplicationSet all use direct values passthrough. |
| **Technical Debt** | **High**: Each new chart = new CRDs. Schema drift detection. Conversion logic. Deprecation cycles. | **Low**: Adding new charts requires no new CRD types. |
| **User Mental Model** | "k8sd configures cilium" → implies k8sd owns/understands cilium configuration semantics. | "k8sd deploys cilium, I configure it via helm values" → clear responsibility boundary. |

## Recommendation

**Choose Option 2: ConfigMaps for Helm Values**

### Rationale

1. **Ownership Boundary Honesty**
   - You don't own cilium, coredns, or their helm charts. Creating CRDs suggests you do.
   - ConfigMaps make it explicit: "k8sd installs external charts; you configure them directly."
   - This is architecturally honest and sets correct expectations.

2. **Maintenance Burden is Unsustainable**
   - Custom CRDs create a **permanent schema synchronization tax**.
   - Every upstream helm chart change requires k8sd CRD versioning, testing, and potentially conversion webhooks.
   - Over 2-5 years, this compounds: 4 charts × yearly schema changes × version management = significant overhead.
   - **You're signing up to track and translate every upstream schema evolution.**

3. **Abstraction is a Liability, Not an Asset**
   - The CRD "abstraction" doesn't hide complexity—it duplicates it.
   - Users still need to understand cilium BGP semantics; the CRD doesn't simplify that.
   - What it does do: create a second documentation surface that must stay synchronized with upstream.
   - **Abstraction should reduce complexity. This increases it.**

4. **Ecosystem Validation**
   - RKE2, K3s, and GitOps tools (Flux, ArgoCD) all use direct helm values for third-party charts.
   - This isn't accidental—it's the architecturally sound pattern for deploying components you don't own.
   - These projects have years of operational experience validating this approach.

5. **Separation of Concerns**
   - Clear boundary: k8sd handles **lifecycle** (install, upgrade, version selection).
   - Users handle **configuration** via standard helm mechanisms.
   - This is a clean responsibility split that scales.

6. **Technical Debt Asymmetry**
   - Custom CRDs: debt accumulates over time (more schemas, more versions, more conversion logic).
   - ConfigMaps: debt is constant (controller logic is stable, no schema tracking).

### Key Concerns for ConfigMaps Approach

1. **Discoverability**
   - Mitigation: Provide example ConfigMaps in documentation. Link directly to upstream helm docs.
   - Tooling: `k8s config show cilium-helm-values` could display schema from embedded helm chart.

2. **Validation happens late**
   - Invalid helm values fail at reconcile, not at ConfigMap creation.
   - Mitigation: Controller should provide clear error messages pointing to helm validation errors.

3. **Less "Kubernetes-native" feel**
   - No `kubectl explain`. Users must read helm documentation.
   - Mitigation: This is the honest trade-off. Kubernetes-native feel isn't free—it costs schema maintenance.

### What Would Change This Recommendation?

**I would recommend Custom CRDs if:**

1. **You owned the helm charts** — If k8sd maintained cilium/coredns helm charts, CRDs make sense. You control schema evolution.

2. **You were providing a true abstraction** — If you were hiding cilium entirely behind a "NetworkConfig" that could swap implementations (cilium, calico, etc.), the abstraction earns its keep.

3. **Configuration was simple and stable** — If helm schemas were small (5-10 fields), rarely changed, and CRDs provided real value-add validation, the burden might be worth it.

4. **You had a large team dedicated to schema maintenance** — If maintaining CRD synchronization was a funded, staffed priority.

**None of these conditions are true here.**

## Risks

### Risks with ConfigMaps (Recommended)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Users struggle with helm values discoverability | Medium | Medium | Documentation with examples; `k8s config show` tooling |
| Invalid values cause reconciliation failures | Medium | Low | Clear error messages; validation in controller log output |
| Perceived as "less polished" than CRDs | Medium | Low | This is an educational issue, not a technical one |

### Risks with Custom CRDs (Not Recommended)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Schema drift from upstream | **High** | **High** | Automated schema sync tooling (complex) |
| Maintenance burden grows unsustainable | **High** | **High** | None—this is inherent to the approach |
| Breaking changes in upstream force CRD versioning | **High** | Medium | Conversion webhooks, deprecation cycles |
| Users confused about who owns configuration | Medium | Medium | Documentation clarifying k8sd is a proxy |

## Implementation Boundaries (ConfigMaps)

**Boundaries to establish:**

1. **CLI scope:** `k8s set` for common, lifecycle-affecting settings (enabled/disabled, version).
2. **ConfigMap scope:** Advanced configuration requiring helm-level control.
3. **Upstream CRDs:** Some components (cilium) have their own CRDs—users can still use those.

**Naming convention:**
```
k8sd-config-<component>  # e.g., k8sd-config-cilium
```

**Merge behavior:**
- k8sd provides default values
- User ConfigMap values override defaults
- Explicit merge semantics in documentation

**Discovery tooling:**
```bash
k8s config show cilium --available-values
k8s config show cilium --current-values  
k8s config validate cilium --from-file values.yaml
```

This provides discoverability without schema maintenance burden.

## Summary

**Choose ConfigMaps.** It's architecturally honest, maintainable long-term, and aligned with ecosystem patterns. Custom CRDs would saddle you with a permanent schema synchronization burden for components you don't own—that's technical debt you don't need.

The ConfigMap approach respects ownership boundaries: k8sd owns **lifecycle**, users own **configuration**. That's a clean, sustainable split.

---

**Architect's Note:** The appeal of CRDs is understandable—they feel more "Kubernetes-native." But architectural soundness isn't about following patterns blindly. It's about matching abstractions to ownership boundaries. When you don't own the schema, don't wrap it. Let it flow through.
