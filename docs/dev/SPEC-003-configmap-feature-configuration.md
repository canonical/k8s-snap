# Proposal information

- **Index**: 003

- **Status**: **DRAFTING**

- **Name**: ConfigMap-Based Feature Configuration for Enterprise Users

- **Owner**: Louise Schmidt-Gen / [@louiseschmidtgen](https://github.com/louiseschmidtgen)

# Proposal Details

## Summary

This proposal introduces ConfigMap-based advanced feature configuration for Canonical Kubernetes, enabling enterprise users to configure features (cilium, coredns, ingress, load-balancer, gateway) with full helm values support while maintaining audit trails, RBAC, and GitOps workflows required by regulated environments.

Users can provide standard helm `values.yaml` via ConfigMaps. k8sd reconcilers merge these values with defaults and apply the helm chart. This approach provides enterprise-ready configuration without creating custom k8sd CRDs that would require permanent schema maintenance.

## Rationale

### Current Limitation

Canonical Kubernetes supports feature configuration via `k8s set` CLI commands, which works well for simple use cases:

```bash
k8s enable network
k8s set network.tunnel-port=9999
```

However, **enterprise users need advanced configuration not exposed via CLI:**

**Real-world example: Banca d'Italia**
- Requires BGP configuration for cilium
- Needs network policies for regulatory compliance
- Current workaround: annotations (string-typed, hard to discover, no GitOps support)

**Enterprise requirements:**
- **Compliance:** FedRAMP, CIS, STIG require audit trails and change tracking
- **RBAC:** Network team configures cilium, security team configures ingress (delegation)
- **GitOps:** Configuration tracked in git with Flux/ArgoCD
- **Validation:** Catch errors before production apply

**Current approach (annotations) is insufficient:**
```bash
k8s set --annotations cilium.bgp.enabled=true
```
- String-typed (no validation)
- Hard to discover (no schema)
- No GitOps tooling support
- No RBAC granularity
- Manual parsing error-prone

### Why ConfigMaps?

**Ecosystem validation:** Projects deploying external helm charts universally use ConfigMaps or direct values:
- **Rancher RKE2:** HelmChartConfig CRD → configmap values
- **K3s:** HelmChartConfig → valuesContent
- **Flux:** HelmRelease.valuesFrom configmap
- **ArgoCD:** values.yaml in git
- **Cluster API:** HelmChartProxy.valuesFrom

**This pattern has years of operational validation.**

**Why not custom k8sd CRDs?**
- We deploy **external** helm charts (cilium, coredns, ingress) we don't own
- Creating CRDs that mirror upstream schemas creates permanent schema synchronization burden
- Upstream helm chart changes require k8sd CRD updates → users blocked
- Analysis shows 78% higher maintenance cost (2175 hours over 5 years)
- Goes against ecosystem patterns

**ConfigMaps are architecturally honest:** "We deploy via helm. Configure via standard helm values."

### User Scenarios

#### Scenario 1: Enterprise Power User (Banca d'Italia)

**Need:** Configure cilium BGP for load balancer IP announcement

**Current (annotations):**
```bash
k8s set --annotations cilium.bgp.enabled=true
k8s set --annotations cilium.bgp.announce.loadbalancerIP=true
```
Problems: String-typed, manual parsing, no validation, no GitOps

**Proposed (ConfigMap):**
```yaml
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
```

Benefits:
- Standard helm values format (upstream docs apply directly)
- GitOps compatible (track in git, Flux/ArgoCD sync)
- Audit trail (kubectl managed fields + git history)
- Validation available (`k8s validate cilium-values values.yaml`)

#### Scenario 2: Regulated Environment (Financial Services)

**Need:** Enable Cilium Hubble for network observability with compliance audit trail

**Proposed:**
```yaml
# In git repo: k8s-config/network-observability.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-cilium-values
  namespace: kube-system
  labels:
    compliance: "STIG-required"
data:
  values.yaml: |
    hubble:
      enabled: true
      ui:
        enabled: true
      relay:
        enabled: true
      metrics:
        enabled:
        - dns
        - drop
        - tcp
        - flow
        - port-distribution
```

Benefits:
- Full audit trail (git commits + kubectl logs)
- RBAC (network team owns this configmap)
- Compliance-ready (FedRAMP/CIS requirements met)
- Self-service validation before apply

#### Scenario 3: Multi-Team Organization

**Need:** Network team configures cilium, security team configures ingress, separate RBAC

**Proposed RBAC:**
```yaml
# Network team role
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: network-config-manager
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["k8sd-cilium-values"]
  verbs: ["get", "update", "patch"]

# Security team role
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: security-config-manager
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["k8sd-ingress-values"]
  verbs: ["get", "update", "patch"]
```

Benefits:
- Clean separation of concerns
- Least-privilege access
- Standard Kubernetes RBAC (no new concepts)

## User facing changes

### Bootstrap Configuration

**Before:** Only simple cluster-config options available
```yaml
cluster-config:
  network:
    pod-cidr: 10.0.0.0/16
```

**After:** Optional helm values file paths
```yaml
cluster-config:
  network:
    pod-cidr: 10.0.0.0/16
  cilium-values-file: /path/to/cilium-values.yaml
  coredns-values-file: /path/to/coredns-values.yaml
  ingress-values-file: /path/to/ingress-values.yaml
```

Values in these files are standard helm `values.yaml` format, merged at bootstrap time.

### Day 2 Configuration

**New capability:** Create ConfigMaps for advanced configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-cilium-values
  namespace: kube-system
data:
  values.yaml: |
    # Standard helm values for cilium
    tunnelPort: 9999
    bgp:
      enabled: true
```

Apply via: `kubectl apply -f cilium-values.yaml`

k8sd reconciler automatically watches ConfigMap, merges values, and updates the helm chart.

### New CLI Commands

**Phase 2 additions:**

```bash
# Validate configmap before apply
k8s validate cilium-values my-values.yaml

# Show current merged values (what's actually applied)
k8s show cilium-values

# Show available values (helm chart schema with docs)
k8s config show cilium --available-values

# Force reconcile (optional, automatic on configmap change)
k8s refresh network
```

### Precedence Order

Configuration precedence (higher overrides lower):

```
base defaults → cluster-config (microcluster) → annotations → configmap values
```

Example:
- Base: `tunnelPort: 8472`
- Cluster-config: `tunnelPort: 9000`
- ConfigMap: `tunnelPort: 9999`
- **Applied:** `9999`

Check merged result: `k8s show cilium-values`

### Unchanged

**Simple user experience unchanged:**
```bash
k8s enable network
k8s set network.tunnel-port=9999
k8s disable network
```

CLI commands continue to work exactly as before. ConfigMaps are additive for power users.

## Alternative solutions

### Alternative 1: Custom k8sd CRDs (Rejected)

**Approach:** Create typed CRDs (CiliumConfig, CoreDNSConfig, etc.) that mirror helm values

**Why considered:**
- Typed validation (schema enforcement)
- Native RBAC (Role for resource type)
- Kubernetes-native feel
- Discoverability (`kubectl explain`)

**Why rejected:**
- **78% higher maintenance cost** (2175 hours over 5 years)
- **Schema synchronization burden:** Must track every upstream helm chart change
- **Version coupling:** CRD versioning required, users blocked until k8sd updates
- **Architectural dishonesty:** Suggests control over schemas we don't own
- **Against ecosystem patterns:** RKE2, K3s, Flux all use ConfigMaps for external charts
- **Unanimous expert rejection:** Architect, developer, devops, security all recommend ConfigMap

Detailed analysis in `docs/dev/FINAL-RECOMMENDATION.md` (87% confidence, 5 perspectives)

### Alternative 2: Extend Annotations (Rejected)

**Approach:** Continue using annotations, add structure via naming conventions

**Why rejected:**
- String-typed (no validation)
- Poor discoverability (no schema)
- No GitOps tooling support
- Doesn't meet enterprise compliance requirements
- Already the problematic workaround we're replacing

### Alternative 3: Direct Helm Values in Microcluster (Rejected)

**Approach:** Store helm values directly in microcluster database (not ConfigMaps)

**Why rejected:**
- No GitOps support (microcluster not git-trackable)
- No Kubernetes-native RBAC
- No audit trail via kubectl
- Doesn't meet enterprise requirements for compliance
- Violates Kubernetes-native principles

## Out of scope

### Feature CRD Versioning Strategy

**Future work:** When k8s-snap upgrades features (cilium 1.15→1.16), how do we communicate breaking changes to users?

**Current approach:** User configmap values overlay new helm chart. If incompatible, helm apply fails with error. User references upstream helm docs for 1.16 values.

**Potential enhancement:** `k8s config diff cilium 1.15 1.16` to show breaking changes.

**Out of scope for initial release** - address if upgrade pain becomes significant.

### Upstream CRD Configuration

**Scope clarification:** This proposal covers **helm values** (deployment configuration).

**Separate concern:** Upstream CRDs like `CiliumBGPPeeringPolicy`, `CiliumNetworkPolicy` are configured directly by users (not via k8sd ConfigMaps).

**Example:**
- **ConfigMap** (this proposal): Enable BGP feature (`bgp.enabled: true`)
- **Upstream CRD** (separate): Define BGP peers (`CiliumBGPPeeringPolicy`)

Clear boundary: helm values (deployment) vs runtime resources (upstream CRDs).

### Secrets Management

**Out of scope:** How users reference secrets in ConfigMaps

**Current approach:** Users can reference secrets via helm values:
```yaml
# ConfigMap references secret
secretKeyRef:
  name: my-secret
  key: api-key
```

**Future enhancement:** Sealed secrets, external secrets operator integration

### Multi-Cluster Configuration

**Out of scope:** Syncing ConfigMaps across multiple clusters

**Future enhancement:** Document Flux/ArgoCD patterns for multi-cluster sync

### Configuration Drift Remediation UI

**Out of scope:** Dashboard showing config drift

**Future enhancement:** `k8s status` could show "config drift detected" warnings

# Implementation Details

## API Changes

**No k8sd API changes required.**

Reconcilers read ConfigMaps via standard Kubernetes API.

## CLI Changes

### Phase 1 (MVP)

**No CLI changes** - ConfigMaps applied via `kubectl`

### Phase 2 (Validation Tooling)

**New commands:**

```bash
k8s validate <feature>-values <file>
  Validates helm values file using helm dry-run
  Returns: success or error with line numbers
  Example: k8s validate cilium-values cilium-values.yaml

k8s show <feature>-values
  Shows current merged helm values actually applied
  Returns: YAML with precedence annotations
  Example: k8s show cilium-values

k8s config show <feature> --available-values
  Shows helm chart schema with documentation
  Returns: Available values with types and descriptions
  Example: k8s config show cilium --available-values

k8s refresh <feature>
  Forces reconciliation (optional, automatic on configmap change)
  Example: k8s refresh network
```

## Database Changes

**No database schema changes required.**

Microcluster continues to store CLI-configured options. ConfigMaps are read-only from k8sd's perspective (stored in Kubernetes etcd).

## Configuration Changes

### Bootstrap File Changes

**New optional fields in bootstrap config:**

```yaml
cluster-config:
  # Existing fields unchanged
  network:
    enabled: true
  
  # New optional fields (Phase 1)
  cilium-values-file: /path/to/cilium-values.yaml
  coredns-values-file: /path/to/coredns-values.yaml
  ingress-values-file: /path/to/ingress-values.yaml
  loadbalancer-values-file: /path/to/loadbalancer-values.yaml
  gateway-values-file: /path/to/gateway-values.yaml
```

Each file contains standard helm `values.yaml` format.

### ConfigMap Naming Convention

**Fixed names per feature:**
- `k8sd-cilium-values` (network feature)
- `k8sd-coredns-values` (dns feature)
- `k8sd-ingress-values` (ingress feature)
- `k8sd-loadbalancer-values` (load-balancer feature)
- `k8sd-gateway-values` (gateway feature)
- `k8sd-localstorage-values` (local-storage feature)

**Namespace:** `kube-system` (where features are deployed)

**Data key:** `values.yaml` (ConfigMap data contains single key)

## Documentation Changes

### New Pages Required

#### How-To: Configure Advanced Feature Options
- Target: Enterprise users, power users
- Content: ConfigMap creation, precedence order, validation workflow
- Examples: BGP, Hubble, custom plugins per feature

#### How-To: GitOps Integration
- Target: DevOps teams
- Content: Flux/ArgoCD setup, git workflow, drift detection
- Examples: Multi-environment setup (dev/staging/prod)

#### How-To: RBAC for Feature Configuration
- Target: Security teams, compliance
- Content: Team-based access control, least-privilege patterns
- Examples: Network team vs security team separation

#### Reference: Helm Values Precedence
- Target: All users
- Content: Detailed precedence explanation with examples
- Examples: What happens when CLI, annotations, and configmap conflict

#### Explanation: Why ConfigMaps Not CRDs
- Target: Architects, decision makers
- Content: Architectural rationale, ecosystem patterns, maintenance burden
- Link: Full analysis in docs/dev/FINAL-RECOMMENDATION.md

### Pages to Update

#### Bootstrap Configuration
- Add: cilium-values-file, coredns-values-file sections
- Example: Bootstrap with BGP pre-configured

#### Feature Configuration Overview
- Add: Section on simple (CLI) vs advanced (ConfigMap) configuration
- Clarify: When to use each approach

#### k8s CLI Reference
- Add: New commands (validate, show, config show, refresh)

#### Troubleshooting Guide
- Add: ConfigMap-specific debugging (precedence, validation, merge conflicts)

## Testing

### Unit Tests

**ConfigMap reader:**
- Parse valid YAML
- Handle invalid YAML gracefully
- Merge precedence order correct
- Missing configmap handled (no error)

**Validation logic:**
- Blacklist enforcement (image, ports, hostNetwork)
- Helm dry-run integration
- Error messages clear and actionable

### Integration Tests

**Bootstrap with values file:**
```bash
# Test: Bootstrap with cilium-values-file
k8s bootstrap --file bootstrap.yaml
# Verify: cilium deployed with custom values
kubectl get configmap cilium-config -n kube-system -o yaml
# Assert: tunnelPort=9999 (from values file)
```

**Day 2 configmap apply:**
```bash
# Test: Apply configmap, verify reconcile
kubectl apply -f cilium-values.yaml
# Wait: for reconcile (max 30s)
# Verify: helm release updated
helm get values cilium -n kube-system
# Assert: bgp.enabled=true
```

**Precedence testing:**
```bash
# Test: CLI vs ConfigMap precedence
k8s set network.tunnel-port=8000  # CLI
kubectl apply -f cilium-values.yaml  # ConfigMap with tunnelPort: 9999
k8s show cilium-values
# Assert: tunnelPort=9999 (ConfigMap wins)
```

**Upgrade testing:**
```bash
# Test: k8s-snap upgrade with user configmap
# Setup: Apply configmap with cilium 1.15 values
# Upgrade: k8s refresh (upgrade to cilium 1.16)
# Verify: Helm warnings for deprecated values logged
# Verify: Cluster remains functional
```

### E2E Tests

**Full workflow:**
1. Bootstrap cluster with default config
2. Enable network feature
3. Apply configmap with BGP config
4. Verify BGP peers established
5. Update configmap (change BGP settings)
6. Verify changes applied within 30s
7. Delete configmap
8. Verify reverts to cluster-config values

**GitOps workflow:**
1. Setup Flux watching git repo
2. Commit configmap to git
3. Verify Flux applies configmap
4. Verify k8sd reconciles
5. Verify feature configuration updated

### Security Testing

**Blacklist enforcement:**
```bash
# Test: Attempt to change image
kubectl apply -f evil-cilium-values.yaml  # contains image: evil/cilium
# Assert: Admission webhook rejects OR runtime validation reverts
```

**RBAC testing:**
```bash
# Test: User with cilium-config-manager role
# Can: update k8sd-cilium-values
# Cannot: update k8sd-ingress-values
# Assert: kubectl apply fails for ingress configmap
```

## Considerations for backwards compatibility

### Existing CLI Configuration (Preserved)

**No breaking changes to CLI behavior:**
- `k8s set network.tunnel-port=9999` continues to work
- Stored in microcluster as before
- Applied with same precedence rules

### Existing Annotations (Deprecated, Not Removed)

**Annotations continue to work** (for backwards compatibility):
- Precedence: below ConfigMaps, above cluster-config
- Documented as deprecated
- Recommend migration to ConfigMaps
- Potential removal in future major version (with deprecation period)

**Migration path documented:** `k8s migrate-annotations cilium` (future tool)

### Microcluster Database

**No schema changes:**
- Existing tables unchanged
- CLI-configured options stored as before
- ConfigMaps stored in Kubernetes etcd (not microcluster)

### Feature Enablement

**No changes to enable/disable:**
```bash
k8s enable network   # unchanged
k8s disable network  # unchanged
```

When feature enabled: reconciler checks for configmap, applies if present.
When feature disabled: configmap preserved (not deleted), not applied.

### Rollback Strategy

**If upgrade to ConfigMap-supporting version fails:**
1. Rollback k8s-snap package
2. ConfigMaps remain in cluster (no-op for older version)
3. CLI configuration continues to work
4. No data loss

## Implementation notes and guidelines

### Phase 1: MVP with Cilium (1 week, ~310 LOC)

**File:** `src/k8s/pkg/k8sd/features/cilium/reconcile.go`

**Add ConfigMap reading:**
```go
func (r *Reconciler) getConfigMapValues(ctx context.Context) (map[string]interface{}, error) {
    cm, err := r.clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "k8sd-cilium-values", metav1.GetOptions{})
    if err != nil {
        if apierrors.IsNotFound(err) {
            return nil, nil  // No configmap, not an error
        }
        return nil, err
    }
    
    valuesYAML := cm.Data["values.yaml"]
    var values map[string]interface{}
    if err := yaml.Unmarshal([]byte(valuesYAML), &values); err != nil {
        return nil, fmt.Errorf("invalid values.yaml in configmap: %w", err)
    }
    
    return values, nil
}
```

**Merge logic (extend existing):**
```go
func (r *Reconciler) buildHelmValues(ctx context.Context) (map[string]interface{}, error) {
    // Existing code for base defaults + cluster-config + annotations
    baseValues := r.getDefaultValues()
    clusterConfigValues := r.getClusterConfigValues()
    annotationValues := r.getAnnotationValues()
    
    // New: ConfigMap values (highest precedence)
    configMapValues, err := r.getConfigMapValues(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get configmap values: %w", err)
    }
    
    // Merge with precedence: base < cluster-config < annotations < configmap
    finalValues := merge(baseValues, clusterConfigValues, annotationValues, configMapValues)
    
    return finalValues, nil
}
```

**Watch ConfigMap for changes:**
```go
func (r *Reconciler) watchConfigMap(ctx context.Context) {
    watcher, err := r.clientset.CoreV1().ConfigMaps("kube-system").Watch(ctx, metav1.ListOptions{
        FieldSelector: "metadata.name=k8sd-cilium-values",
    })
    if err != nil {
        log.Errorf("failed to watch configmap: %v", err)
        return
    }
    
    for event := range watcher.ResultChan() {
        if event.Type == watch.Modified || event.Type == watch.Added {
            log.Info("configmap changed, triggering reconcile")
            r.reconcile(ctx)
        }
    }
}
```

**Bootstrap file support:**

**File:** `src/k8s/cmd/k8s/bootstrap.go`

```go
func loadBootstrapConfig(path string) (*BootstrapConfig, error) {
    // Existing code loads bootstrap.yaml
    
    // New: Load values files if specified
    if config.CiliumValuesFile != "" {
        valuesYAML, err := ioutil.ReadFile(config.CiliumValuesFile)
        if err != nil {
            return nil, fmt.Errorf("failed to read cilium-values-file: %w", err)
        }
        
        // Create configmap with values
        cm := &corev1.ConfigMap{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "k8sd-cilium-values",
                Namespace: "kube-system",
            },
            Data: map[string]string{
                "values.yaml": string(valuesYAML),
            },
        }
        
        // Apply configmap after cluster bootstrap
        config.InitialConfigMaps = append(config.InitialConfigMaps, cm)
    }
    
    return config, nil
}
```

### Phase 2: Validation Tooling (1-2 weeks)

**CLI command: `k8s validate`**

**File:** `src/k8s/cmd/k8s/validate.go`

```go
func validateValues(feature, file string) error {
    valuesYAML, err := ioutil.ReadFile(file)
    if err != nil {
        return err
    }
    
    // Use helm dry-run to validate
    cmd := exec.Command("helm", "template", feature, ".", 
        "--values", file,
        "--dry-run",
    )
    cmd.Dir = getFeatureChartPath(feature)
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("validation failed:\n%s", output)
    }
    
    fmt.Println("✓ Validation successful")
    return nil
}
```

**CLI command: `k8s show values`**

```go
func showValues(feature string) error {
    // Get current merged values from reconciler
    values, err := getAppliedHelmValues(feature)
    if err != nil {
        return err
    }
    
    // Pretty-print with precedence annotations
    fmt.Println("# Merged values for", feature)
    fmt.Println("# Precedence: base < cluster-config < annotations < configmap")
    fmt.Println()
    
    valuesYAML, _ := yaml.Marshal(values)
    fmt.Println(string(valuesYAML))
    
    return nil
}
```

**Runtime validation (blacklist):**

```go
func validateAgainstBlacklist(values map[string]interface{}) error {
    blacklist := []string{
        "image",
        "imageRepository", 
        "imageTag",
        "hostNetwork",
        "privileged",
        "hostPID",
        "hostIPC",
    }
    
    for _, key := range blacklist {
        if hasKey(values, key) {
            return fmt.Errorf("forbidden field: %s (security policy)", key)
        }
    }
    
    return nil
}
```

### Phase 3: Multi-Feature Rollout (2-3 weeks)

**Apply same pattern to each feature:**
- coredns: `k8sd-coredns-values`
- ingress: `k8sd-ingress-values`
- load-balancer: `k8sd-loadbalancer-values`
- gateway: `k8sd-gateway-values`
- local-storage: `k8sd-localstorage-values`

**Code structure:**
- Extract ConfigMap reading to shared utility
- Each feature reconciler calls utility
- Consistent naming convention enforced

### Testing Strategy

**Unit tests:** `src/k8s/pkg/k8sd/features/cilium/reconcile_test.go`
```go
func TestConfigMapMerge(t *testing.T) {
    base := map[string]interface{}{"tunnelPort": 8472}
    configMap := map[string]interface{}{"tunnelPort": 9999}
    
    result := merge(base, configMap)
    
    assert.Equal(t, 9999, result["tunnelPort"])
}
```

**Integration tests:** `tests/integration/feature_config_test.go`
```go
func TestConfigMapReconcile(t *testing.T) {
    // Create configmap
    kubectl("apply", "-f", "testdata/cilium-values.yaml")
    
    // Wait for reconcile
    time.Sleep(30 * time.Second)
    
    // Verify helm values applied
    output := helm("get", "values", "cilium", "-n", "kube-system")
    assert.Contains(t, output, "bgp.enabled: true")
}
```

### Rollout Plan

**Week 1:** Phase 1 MVP with cilium
- Implement ConfigMap reading + merge
- Bootstrap file support
- Basic testing
- Validate with Banca d'Italia use case

**Week 2-3:** Phase 2 validation tooling
- CLI commands (validate, show, config show)
- Runtime validation + blacklist
- Admission webhook (optional)
- Security testing

**Week 4-6:** Phase 3 multi-feature
- Extend to coredns, ingress, load-balancer, gateway
- Extract common utilities
- Full test coverage

**Week 7:** Phase 4 documentation
- How-to guides
- GitOps examples
- Compliance documentation

**Total:** 6-9 weeks to production-ready

### Code Review Focus Areas

**Phase 1 review:**
- [ ] Merge logic correct (precedence order)
- [ ] ConfigMap not found handled gracefully
- [ ] Invalid YAML doesn't crash reconciler
- [ ] Watch configmap for changes
- [ ] Bootstrap file integration clean

**Phase 2 review:**
- [ ] Helm dry-run integration secure
- [ ] Blacklist comprehensive (all security fields)
- [ ] Error messages actionable
- [ ] CLI commands UX-tested

**Phase 3 review:**
- [ ] Code duplication minimized (shared utilities)
- [ ] Naming conventions consistent
- [ ] All features tested equivalently

---

**Implementation owner:** To be assigned  
**Target release:** k8s-snap 1.32  
**Dependencies:** None (additive feature)
