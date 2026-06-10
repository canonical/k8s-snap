#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0
TOTAL=4

echo "=========================================="
echo "CoreDNS ConfigMap PoC Validation Test"
echo "=========================================="
echo ""

# Helper functions
pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((PASSED++))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    echo -e "  ${YELLOW}Reason:${NC} $2"
    ((FAILED++))
}

info() {
    echo -e "${YELLOW}→${NC} $1"
}

cleanup() {
    info "Cleaning up test resources..."
    sudo k8s kubectl delete configmap k8sd-coredns-values -n kube-system --ignore-not-found=true 2>/dev/null || true
}

# Ensure cleanup on exit
trap cleanup EXIT

# Wait for pods to stabilize
wait_for_reconcile() {
    info "Waiting for reconcile (10s)..."
    sleep 10
}

# Get current replica count
get_replicas() {
    sudo k8s kubectl get deployment -n kube-system coredns -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0"
}

# Check if deployment has specific annotation/label/config
check_deployment_config() {
    local key=$1
    local expected=$2
    sudo k8s kubectl get deployment -n kube-system coredns -o yaml | grep -q "$key.*$expected"
}

echo ""
echo "Test 1: ConfigMap Override Works"
echo "-----------------------------------"

# Clean state
cleanup
wait_for_reconcile

# Get baseline
BASELINE_REPLICAS=$(get_replicas)
info "Baseline replicas: $BASELINE_REPLICAS"

if [ "$BASELINE_REPLICAS" != "2" ]; then
    fail "Test 1: Baseline replicas" "Expected 2, got $BASELINE_REPLICAS"
else
    # Create ConfigMap with overrides
    cat <<EOF | sudo k8s kubectl apply -f - >/dev/null 2>&1
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-coredns-values
  namespace: kube-system
data:
  values: |
    replicas: 3
EOF

    wait_for_reconcile
    
    # Check replicas updated
    NEW_REPLICAS=$(get_replicas)
    
    if [ "$NEW_REPLICAS" = "3" ]; then
        pass "Test 1: ConfigMap override works (replicas: 2 → 3)"
    else
        fail "Test 1: ConfigMap override works" "Expected replicas=3, got $NEW_REPLICAS"
    fi
fi

echo ""
echo "Test 2: ConfigMap Edit Triggers Reconcile"
echo "------------------------------------------"

# Edit ConfigMap (change replicas to 4)
sudo k8s kubectl patch configmap k8sd-coredns-values -n kube-system \
    --type merge -p '{"data":{"values":"replicas: 4\n"}}' >/dev/null 2>&1

wait_for_reconcile

UPDATED_REPLICAS=$(get_replicas)

if [ "$UPDATED_REPLICAS" = "4" ]; then
    pass "Test 2: Edit triggers reconcile (replicas: 3 → 4)"
else
    fail "Test 2: Edit triggers reconcile" "Expected replicas=4, got $UPDATED_REPLICAS"
fi

echo ""
echo "Test 3: ConfigMap Delete Reverts to Defaults"
echo "---------------------------------------------"

# Delete ConfigMap
sudo k8s kubectl delete configmap k8sd-coredns-values -n kube-system >/dev/null 2>&1

wait_for_reconcile

REVERTED_REPLICAS=$(get_replicas)

if [ "$REVERTED_REPLICAS" = "2" ]; then
    pass "Test 3: Delete reverts to defaults (replicas: 4 → 2)"
else
    fail "Test 3: Delete reverts to defaults" "Expected replicas=2, got $REVERTED_REPLICAS"
fi

echo ""
echo "Test 4: Existing CLI Still Works"
echo "---------------------------------"

# Use existing CLI to set cluster-domain
sudo k8s set dns.cluster-domain=poc-test.local >/dev/null 2>&1

wait_for_reconcile

# Check cluster domain was applied
if sudo k8s kubectl get configmap coredns -n kube-system -o yaml | grep -q "poc-test.local"; then
    # Now add ConfigMap override
    cat <<EOF | sudo k8s kubectl apply -f - >/dev/null 2>&1
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8sd-coredns-values
  namespace: kube-system
data:
  values: |
    replicas: 5
EOF

    wait_for_reconcile
    
    # Check both CLI config and ConfigMap override are applied
    MERGE_REPLICAS=$(get_replicas)
    CLI_PRESERVED=$(sudo k8s kubectl get configmap coredns -n kube-system -o yaml | grep -c "poc-test.local" || echo "0")
    
    if [ "$MERGE_REPLICAS" = "5" ] && [ "$CLI_PRESERVED" != "0" ]; then
        pass "Test 4: CLI + ConfigMap merge works (cluster-domain + replicas)"
    else
        fail "Test 4: CLI + ConfigMap merge" "replicas=$MERGE_REPLICAS (expected 5), cluster-domain found=$CLI_PRESERVED (expected >0)"
    fi
else
    fail "Test 4: CLI config" "Cluster domain 'poc-test.local' not found in CoreDNS config"
fi

# Reset cluster-domain to default
sudo k8s set dns.cluster-domain=cluster.local >/dev/null 2>&1

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Total:  $TOTAL tests"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED${NC}"
    echo "The ConfigMap PoC implementation is working correctly."
    exit 0
else
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    echo "Review the implementation and fix the failing tests."
    exit 1
fi
