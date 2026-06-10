# CoreDNS ConfigMap PoC - Implementation Prompt

Use this prompt in a new agent session to implement the CoreDNS ConfigMap PoC.

---

## Prompt for Agent

```
Implement the CoreDNS ConfigMap PoC according to the plan in docs/dev/POC-PLAN-COREDNS-CONFIGMAP.md.

CONTEXT:
- We've decided to use ConfigMaps (not custom CRDs) for enterprise feature configuration
- This PoC validates the approach for the CoreDNS/DNS feature
- Architecture decision documented in docs/dev/ADR-003-configmap-feature-configuration.md
- Full specification in docs/dev/SPEC-003-configmap-feature-configuration.md

OBJECTIVE:
Implement minimal ConfigMap support for CoreDNS feature that:
1. Reads k8sd-coredns-values ConfigMap from kube-system namespace
2. Merges ConfigMap values with existing defaults (precedence: ConfigMap > DB > Base)
3. Watches ConfigMap changes and triggers reconciliation
4. Preserves existing CLI workflows (k8s set dns.*)

SCOPE (MINIMAL PoC):
- ✅ ConfigMap reading and merge logic
- ✅ Watcher for ConfigMap changes
- ✅ Integration into existing DNS reconcile function
- ❌ NO bootstrap file support (Day 2 only)
- ❌ NO validation webhook
- ❌ NO CLI commands (use kubectl directly)
- ❌ NO unit tests (validation test only)

SUCCESS CRITERIA:
Run the automated validation test and get all 4 tests passing:
```bash
sudo tests/poc/validate-coredns-configmap.sh
```

Expected output:
```
✓ PASS: Test 1: ConfigMap override works (replicas: 2 → 3)
✓ PASS: Test 2: Edit triggers reconcile (replicas: 3 → 4)
✓ PASS: Test 3: Delete reverts to defaults (replicas: 4 → 2)
✓ PASS: Test 4: CLI + ConfigMap merge works (cluster-domain + replicas)

✓ ALL TESTS PASSED
```

IMPLEMENTATION STEPS:
1. Read the complete plan: docs/dev/POC-PLAN-COREDNS-CONFIGMAP.md
2. Locate the DNS reconcile code (likely in src/k8s/pkg/k8sd/features/dns/ or similar)
3. Implement Phase 1: ConfigMap reader (getConfigMapOverrides function)
4. Implement Phase 2: Merge logic (mergeValues function)
5. Implement Phase 3: Integration into reconcile function
6. Implement Phase 4: ConfigMap watcher
7. Build and test iteratively using the validation script

ITERATIVE WORKFLOW:
```bash
# After each code change:
1. Build: make build (or equivalent build command)
2. Restart: sudo systemctl restart snap.k8s.k8sd
3. Test: sudo tests/poc/validate-coredns-configmap.sh
4. If failures, check logs and fix: sudo journalctl -u snap.k8s.k8sd -f
5. Repeat until all tests pass
```

KEY TECHNICAL DETAILS:
- ConfigMap name: k8sd-coredns-values
- ConfigMap namespace: kube-system
- ConfigMap data key: values (YAML content)
- Merge precedence: base defaults ← microcluster DB ← ConfigMap (right wins)
- Use deep merge for nested maps
- Handle NotFound gracefully (no ConfigMap = no overrides)

FILES TO MODIFY (approximate paths):
- src/k8s/pkg/k8sd/features/dns/reconcile.go (~80 lines added)
- src/k8s/pkg/k8sd/setup/features.go (~25 lines for watcher)

CONSTRAINTS:
- Must not break existing CLI workflows (k8s set dns.*)
- Must not require changes to microcluster DB schema
- Must handle ConfigMap not existing (graceful degradation)
- Must handle invalid YAML in ConfigMap (log error, continue)

DEBUGGING TIPS:
- Add debug logs liberally: log.Info("Reading ConfigMap overrides...")
- Check k8sd logs: sudo journalctl -u snap.k8s.k8sd -f
- Verify ConfigMap created: sudo k8s kubectl get configmap k8sd-coredns-values -n kube-system -o yaml
- Check merge output: Add log.Info("Merged values: %+v", helmValues)
- Verify reconcile triggered: Log entry point of reconcile function

DOCUMENTATION:
Read these files in order for full context:
1. docs/dev/POC-PLAN-COREDNS-CONFIGMAP.md (implementation plan with code examples)
2. docs/dev/ADR-003-configmap-feature-configuration.md (architecture decision rationale)
3. docs/dev/SPEC-003-configmap-feature-configuration.md (complete specification)

Start by exploring the codebase to locate the DNS reconcile code, then implement the phases from the PoC plan. Use the validation test to verify your implementation works correctly.
```

---

## Alternative: Minimal Version

If you want a more concise prompt:

```
Implement ConfigMap-based helm value overrides for CoreDNS. Read docs/dev/POC-PLAN-COREDNS-CONFIGMAP.md for the complete plan. Success = all tests pass when running:

sudo tests/poc/validate-coredns-configmap.sh

Key requirements:
- Read k8sd-coredns-values ConfigMap from kube-system namespace
- Merge with existing defaults (ConfigMap > DB > Base precedence)
- Watch ConfigMap changes and trigger reconcile
- Don't break existing k8s set dns.* commands

Files to modify: src/k8s/pkg/k8sd/features/dns/reconcile.go + watcher setup

Build → Restart → Test iteratively until all 4 tests pass.
```

---

## Tips for Agent Success

**Before starting:**
- Read POC-PLAN-COREDNS-CONFIGMAP.md completely
- Understand current architecture (mermaid diagrams show data flow)
- Locate DNS reconcile code in codebase

**During implementation:**
- Follow the 4 phases in order (reader → merge → integrate → watcher)
- Test after each phase using the validation script
- Add debug logs to understand execution flow
- Don't skip the watcher - it's critical for Test 2

**When tests fail:**
- Read the failure reason carefully
- Check k8sd logs for errors
- Verify ConfigMap was created correctly
- Check merge precedence order
- Ensure watcher is triggering reconcile

**Success indicators:**
- All 4 tests pass
- Existing CLI still works (Test 4)
- ConfigMap edits trigger updates (Test 2)
- Delete reverts to defaults (Test 3)

---

## Expected Timeline

**Optimistic:** 2-3 hours (experienced Go developer, familiar with k8s-snap)
**Realistic:** 4-6 hours (includes exploration, debugging, iterations)
**Pessimistic:** 1-2 days (unfamiliar codebase, need to learn k8sd architecture)

The automated test enables tight iteration loops, so expect 3-5 test runs before success.
