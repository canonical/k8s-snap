# ck-loadbalancer Chart Tests

This directory contains Helm chart rendering tests for the ck-loadbalancer chart.

## Running the tests

```bash
cd k8s/manifests/charts/ck-loadbalancer/tests
go test ./...
```

Or with verbose output:
```bash
cd k8s/manifests/charts/ck-loadbalancer/tests
go test -v ./...
```

**Note**: These tests require `helm` to be available in your PATH.
