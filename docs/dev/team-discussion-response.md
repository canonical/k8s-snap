# Addressing Team Discussion Points

**Context:** The spec includes a "Team Discussion" section with concerns about the CRD approach. This document directly addresses each point.

---

## Team Concerns from Spec

### 1. "Helm is an implementation detail we do not want to expose"

**Concern:** Creating CRDs that mirror helm values exposes helm as an implementation detail.

**Analysis:**
This concern is valid, but **both approaches expose helm**:

| Approach | How it exposes helm |
|----------|---------------------|
| **CRD** | CRD schema IS the helm values schema. Field names match helm values exactly. Users must understand "this k8sd field maps to this helm value maps to this cilium flag." |
| **ConfigMap** | Users directly provide helm values. Explicitly documented: "k8sd uses helm, here's how to configure it." |
| **Annotations (current)** | Already expose helm values as strings. Example: `--annotations cilium.bpf.preallocateMaps=true` |

**Key insight:** We're already exposing helm via annotations. The question is: **should we pretend not to or be honest about it?**

**CRD approach:** Pretends to abstract helm, but it's a leaky abstraction. CRD fields are 1:1 helm values.

**ConfigMap approach:** Honest about helm usage. Clear contract: "We use helm to deploy cilium. Provide values.yaml to customize."

**Verdict:** ConfigMap is more honest. Users who need BGP configuration are sophisticated enough to understand helm values.

**Alternative perspective:** What if we embraced helm as a **feature** not a secret?
- Helm is the industry standard for k8s app packaging
- Helm values are well-documented upstream
- Users can leverage existing helm knowledge

---

### 2. "Can feature CRDs not be directly edited instead of through k8sd feature CRs?"

**Concern (from Berkay):** Cilium CRDs can't be modified manually because deployments are managed through helm.

**Analysis:**
This is a critical distinction between:
- **Upstream CRDs** (CiliumBGPPeeringPolicy, CiliumNetworkPolicy): Cilium's native resources
- **k8sd CRDs** (CiliumConfig in the spec): Proposed wrapper for helm values

**Problem:**
Users need to configure:
1. **Helm values** (tunnelPort, bpf.preallocateMaps): controls cilium deployment
2. **Upstream CRDs** (BGPPeeringPolicy): runtime cilium behavior

**CRD approach creates three layers:**
```
k8s set CLI → k8sd CRD (helm values) → upstream CRD (cilium behavior)
```

**ConfigMap approach creates two clear paths:**
```
Simple: k8s set CLI (stored in microcluster)
Advanced: ConfigMap (helm values) + upstream CRDs (cilium behavior)
```

**Why can't users edit upstream CRDs?**
They can! Upstream CRDs are separate from helm deployment config:
- Helm values configure **how cilium is deployed**
- Upstream CRDs configure **what cilium does at runtime**

**Example:**
```yaml
# ConfigMap: controls deployment
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-cilium-values
data:
  values.yaml: |
    bgp:
      enabled: true  # enables BGP feature in cilium

# Upstream CRD: controls BGP behavior
apiVersion: cilium.io/v2alpha1
kind: CiliumBGPPeeringPolicy
metadata:
  name: my-bgp-peer
spec:
  nodeSelector:
    matchLabels:
      rack: rack1
  virtualRouters:
  - localASN: 64512
    neighbors:
    - peerAddress: 10.0.0.1/32
      peerASN: 64512
```

**Verdict:** ConfigMap approach maintains clear separation: helm values (deployment) vs upstream CRDs (behavior).

---

### 3. "We wrap another values layer around the existing one"

**Concern:** CRDs introduce indirection that creates sub-optimal UX. Another layer of values.

**Analysis:**
This is the **core problem** with the CRD approach.

**Current state (for simple config):**
```
User intent → k8s set CLI → microcluster → k8sd reconciler → helm values → deployed
```

**CRD approach (for advanced config):**
```
User intent → k8sd CRD → k8sd reconciler → helm values → deployed
                ↓
           (must match helm schema exactly)
```

**ConfigMap approach (for advanced config):**
```
User intent → ConfigMap (helm values.yaml) → k8sd reconciler → helm values → deployed
```

**Key difference:**
- **CRD:** k8sd CRD schema is a **translation layer** (must be maintained to match helm)
- **ConfigMap:** User provides helm values **directly** (no translation needed)

**Maintenance burden:**

| When upstream helm changes | CRD Approach | ConfigMap Approach |
|---------------------------|--------------|---------------------|
| New value added | Must update k8sd CRD schema | User can use it immediately |
| Value deprecated | Must deprecate k8sd CRD field, handle migration | User sees helm warning, updates configmap |
| Value type changed | Must update k8sd CRD schema, test | User's invalid type fails at helm apply |
| Value removed | Must version k8sd CRD, write conversion logic | User's configmap has ignored field, logged as warning |

**Verdict:** The "extra layer" concern is valid. ConfigMap removes the translation layer.

---

### 4. "The CRDs introduce an indirection that creates a sub-optimal UX"

**Concern:** User needs to modify configuration in three different places without understanding why.

**Analysis:**
This is about **cognitive load and learning curve**.

**User journey: "I want to configure BGP on cilium"**

**CRD approach:**
1. Read canonical k8s docs
2. Learn about k8sd CRDs
3. Find `CiliumConfig` CRD reference
4. Discover BGP options in k8sd CRD schema
5. Understand which values go in k8sd CRD vs upstream cilium CRDs
6. Create/edit CiliumConfig CR
7. Wait for reconcile
8. Create upstream BGPPeeringPolicy CR

**Questions user has:**
- "Why do I set `bgp.enabled` in CiliumConfig but BGP peers in CiliumBGPPeeringPolicy?"
- "How do I know which config goes in k8sd CRD vs upstream CRD?"
- "What if k8sd CRD doesn't expose the helm value I need?"

**ConfigMap approach:**
1. Read canonical k8s docs
2. Learn that k8sd uses helm for cilium
3. Read upstream [cilium helm docs](https://docs.cilium.io/helm-reference)
4. Create configmap with values.yaml (standard helm pattern)
5. Apply configmap
6. Create upstream BGPPeeringPolicy CR (separate concern)

**Questions user has:**
- "What helm values can I set?" → Refer to upstream helm docs (canonical source)
- "What's the precedence order?" → CLI < annotations < configmap (documented once)

**Cognitive load comparison:**

| Concept | CRD | ConfigMap |
|---------|-----|-----------|
| k8sd-specific API | Must learn ❌ | Not needed ✅ |
| Helm values | Indirect (via CRD) ❌ | Direct ✅ |
| Upstream CRDs | Must learn ❌ | Must learn ❌ |
| Which config goes where? | k8sd internals knowledge required ❌ | Functional split: simple=CLI, advanced=configmap ✅ |

**Verdict:** ConfigMap has lower cognitive load. Users leverage existing helm knowledge.

---

### 5. "It is not obvious for the user why and which certain values need to be set in k8sd CRD vs Cilium CRD"

**Concern:** Without understanding k8sd internals, users don't know where to configure what.

**Analysis:**
This is about **intuitive mental models**.

**CRD approach mental model:**
- k8sd CRD: "k8sd's opinion about cilium configuration"
- Upstream CRD: "cilium's native resources"
- But: k8sd CRD fields are actually helm values, which control cilium deployment
- Confusing: "Is `bgp.enabled` a k8sd opinion or a helm value?"

**Answer:** It's a helm value, wrapped in k8sd CRD format.

**ConfigMap approach mental model:**
- CLI: "quick settings for common cases"
- ConfigMap: "advanced helm values for power users"
- Upstream CRD: "cilium's runtime behavior resources"

**Clarity:** Each layer has a clear purpose:
1. **CLI** (microcluster config): Simple settings, k8s snap managed
2. **ConfigMap** (helm values): Advanced deployment config, user managed
3. **Upstream CRD** (cilium resources): Runtime behavior, user managed

**Example decision tree:**

| User need | Where to configure | Why |
|-----------|-------------------|-----|
| Enable cilium | `k8s enable network` | CLI for common action |
| Change tunnel port | `k8s set network.tunnel-port=9999` | CLI for simple setting |
| Enable BGP | ConfigMap: `bgp.enabled: true` | Helm value, not common setting |
| Define BGP peers | Upstream CRD: `CiliumBGPPeeringPolicy` | Cilium runtime resource |
| Custom IPAM mode | ConfigMap: `ipam.mode: cluster-pool` | Helm value, advanced setting |

**With CRDs:**

| User need | Where to configure | Why |
|-----------|-------------------|-----|
| Enable cilium | `k8s enable network` OR `CiliumConfig.enabled: true` | Two ways, unclear which |
| Change tunnel port | `k8s set` OR `CiliumConfig.spec.tunnelPort` | Two ways, unclear precedence |
| Enable BGP | `CiliumConfig.spec.bgp.enabled: true` | If exposed in schema... |
| Define BGP peers | Upstream CRD: `CiliumBGPPeeringPolicy` | Still needed for peers |
| Custom IPAM mode | `CiliumConfig.spec.ipam.mode` | If exposed in schema... |

**Problem:** k8sd CRD duplicates CLI capability (enabled, tunnelPort) while also exposing helm values (bgp, ipam). No clear boundary.

**Verdict:** ConfigMap has clearer boundaries: CLI for simple, ConfigMap for advanced, upstream CRD for runtime.

---

### 6. "Instead, we should only support two user stories: k8s set for opinionated small configurations and Upstream CRD for power users"

**Concern:** Why introduce k8sd CRDs at all? Just use CLI + upstream CRDs.

**Analysis:**
This is the **minimalist alternative**: No ConfigMaps, no k8sd CRDs.

**Two-path model:**
1. **Simple users:** `k8s set` CLI only
2. **Power users:** Manually configure upstream CRDs

**Why this doesn't solve Banca d'Italia's problem:**

BGP configuration requires **both** helm values AND upstream CRDs:

```yaml
# Helm value: enables BGP feature in cilium deployment
bgp.enabled: true

# Upstream CRD: defines BGP peering behavior
apiVersion: cilium.io/v2alpha1
kind: CiliumBGPPeeringPolicy
spec: ...
```

**Without ConfigMap or k8sd CRD for helm values:**
- User creates `CiliumBGPPeeringPolicy` CR
- Cilium doesn't respond (BGP feature not enabled in deployment)
- User confused: "Why doesn't my BGP policy work?"

**Answer:** Because `bgp.enabled: true` is a **helm value**, not a CLI option or upstream CRD.

**Options to enable bgp.enabled:**
1. **Annotation:** `k8s set --annotations cilium.bgp.enabled=true`
   - Already works today
   - Unstructured string
   - Hard to discover
   - No GitOps tooling

2. **Add to CLI:** `k8s set network.bgp-enabled=true`
   - Requires k8sd code change for every new advanced option
   - CLI becomes bloated with niche options
   - Still doesn't help users who need dozens of BGP-related values

3. **ConfigMap:** `bgp.enabled: true` in values.yaml
   - Standard helm pattern
   - Supports all helm values (not just common ones)
   - GitOps-friendly
   - **Proposed solution**

4. **k8sd CRD:** `CiliumConfig.spec.bgp.enabled: true`
   - Structured alternative to configmap
   - Maintenance burden
   - **Not recommended**

**Verdict:** The two-path model (CLI + upstream CRD) is insufficient. Users need a way to configure helm values beyond what CLI exposes. ConfigMap is the right mechanism.

---

## Summary: Why ConfigMap Addresses Team Concerns Better

| Team Concern | CRD Approach | ConfigMap Approach |
|--------------|--------------|---------------------|
| **Exposing helm** | Leaky abstraction: CRD fields ARE helm values | Honest: "We use helm, here's values.yaml" |
| **Extra values layer** | k8sd CRD is translation layer (must maintain schema) | Direct helm values (no translation) |
| **Indirection UX** | Three unclear layers (CLI/k8sd CRD/upstream CRD) | Two clear paths (CLI simple / ConfigMap advanced) |
| **Config placement** | Unclear: which goes in k8sd CRD vs upstream? | Clear: deployment (helm) vs behavior (upstream) |
| **Power user needs** | Must wait for k8sd to expose values in CRD schema | Can use any helm value immediately |

**Conclusion:** ConfigMap approach directly addresses the team's concerns:
1. ✅ Honest about helm (no pretense of hiding it)
2. ✅ No translation layer (direct helm values)
3. ✅ Clear boundaries (simple=CLI, advanced=configmap)
4. ✅ Intuitive split (deployment config vs runtime behavior)
5. ✅ Unblocks power users (all helm values accessible)

---

## Recommendation Reinforced

**Adopt ConfigMap approach:**
- Addresses all team concerns
- Faster to deliver (2-3 weeks vs 4-6 weeks)
- Lower maintenance burden
- Aligns with ecosystem patterns (RKE2, K3s, Flux)
- Clear user mental model

**Do not adopt CRD approach:**
- Creates the problems the team identified
- High maintenance burden
- Doesn't solve core UX issue (still three layers)
- Goes against ecosystem patterns

**Next steps:**
1. Prototype configmap implementation (1 week)
2. Validate with Banca d'Italia use case
3. Document clearly: when to use CLI vs ConfigMap vs upstream CRDs
4. Deliver MVP in 2-3 weeks

---

**The team's concerns validate the ConfigMap recommendation.**
