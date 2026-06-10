# PoC Tests

This directory contains Proof of Concept (PoC) validation tests for k8s-snap features.

## Available Tests

### CoreDNS ConfigMap PoC (`validate-coredns-configmap.sh`)

**Purpose:** Validates that ConfigMap-based helm value overrides work for CoreDNS feature.

**What it tests:**
1. ConfigMap override works (replicas: 2 → 3)
2. ConfigMap edit triggers reconcile (replicas: 3 → 4)
3. ConfigMap delete reverts to defaults (replicas: 4 → 2)
4. Existing CLI still works + merges correctly

**Usage:**
```bash
# Run all tests
sudo tests/poc/validate-coredns-configmap.sh

# Expected runtime: ~45 seconds
```

**Exit codes:**
- `0` = All tests passed
- `1` = One or more tests failed

**Requirements:**
- k8s-snap installed and running
- CoreDNS feature enabled
- kubectl available via `sudo k8s kubectl`

**For Agents:**

This test enables iterative development:
```bash
# 1. Make code changes
vim src/k8s/pkg/k8sd/features/dns/reconcile.go

# 2. Build and restart
make build
sudo systemctl restart snap.k8s.k8sd

# 3. Validate
sudo tests/poc/validate-coredns-configmap.sh

# 4. Fix failures and repeat
```

**See also:**
- [POC-PLAN-COREDNS-CONFIGMAP.md](../../docs/dev/POC-PLAN-COREDNS-CONFIGMAP.md) - Complete PoC plan
- [ADR-003](../../docs/dev/ADR-003-configmap-feature-configuration.md) - Architecture decision
- [SPEC-003](../../docs/dev/SPEC-003-configmap-feature-configuration.md) - Full specification
