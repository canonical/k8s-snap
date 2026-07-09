// Run these tests with:
//
//	cd k8s/manifests/charts/ck-loadbalancer/tests && go test ./...
//
// or from repo root:
//
//	cd k8s/manifests/charts/ck-loadbalancer/tests && go test -v ./...
package tests

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func getHelmPath(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("helm")
	if err != nil {
		t.Skip("helm not found in PATH, skipping chart render tests")
	}
	return p
}

func TestChartRender(t *testing.T) {
	helmPath := getHelmPath(t)

	chartPath, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get chart path: %v", err)
	}

	tests := []struct {
		name       string
		valuesFile string
		assertFunc func(t *testing.T, output string)
	}{
		{
			name:       "TC1: single peer regression",
			valuesFile: "values/tc1-single-peer.yaml",
			assertFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "kind: BGPPeer") {
					t.Error("Expected BGPPeer CR")
				}
				// Index-0 peer must use the bare name (no numeric suffix) so upgrades
				// do not flap existing BGP sessions.
				lines := strings.Split(output, "\n")
				for i, line := range lines {
					if strings.Contains(line, "kind: BGPPeer") {
						for j := i + 1; j < i+10 && j < len(lines); j++ {
							if strings.Contains(lines[j], "metadata:") {
								for k := j + 1; k < j+5 && k < len(lines); k++ {
									trimmed := strings.TrimSpace(lines[k])
									if strings.HasPrefix(trimmed, "name:") {
										if trimmed != "name: ck-loadbalancer" {
											t.Errorf("Expected bare name 'ck-loadbalancer', got %q", trimmed)
										}
										break
									}
								}
								break
							}
						}
						break
					}
				}
				if !strings.Contains(output, "ipAddressPools:") {
					t.Error("Default BGPAdvertisement must restrict to named pool")
				}
			},
		},
		{
			name:       "TC2: three peers with per-zone nodeSelectors",
			valuesFile: "values/tc2-multi-peer.yaml",
			assertFunc: func(t *testing.T, output string) {
				if count := strings.Count(output, "kind: BGPPeer"); count != 3 {
					t.Errorf("Expected 3 BGPPeer CRs, got %d", count)
				}
				// Naming: bare, -1, -2
				for _, name := range []string{"name: ck-loadbalancer", "name: ck-loadbalancer-1", "name: ck-loadbalancer-2"} {
					if !strings.Contains(output, name) {
						t.Errorf("Expected peer name %q", name)
					}
				}
				// Per-zone nodeSelectors
				for _, zone := range []string{"i1", "i2", "i3"} {
					if !strings.Contains(output, "topology.kubernetes.io/zone: "+zone) {
						t.Errorf("Expected nodeSelector zone %q", zone)
					}
				}
			},
		},
		{
			name:       "TC3: advertiseAllPools=true omits ipAddressPools",
			valuesFile: "values/tc3-advertise-all.yaml",
			assertFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "kind: BGPAdvertisement") {
					t.Error("Expected BGPAdvertisement CR")
				}
				if strings.Contains(output, "ipAddressPools:") {
					t.Error("advertiseAllPools=true must not emit ipAddressPools")
				}
			},
		},
		{
			name:       "TC6: bgp.enabled=false renders no MetalLB BGP CRs",
			valuesFile: "values/tc6-bgp-disabled.yaml",
			assertFunc: func(t *testing.T, output string) {
				if strings.Contains(output, "kind: BGPPeer") {
					t.Error("Should not render BGPPeer when bgp.enabled=false")
				}
				if strings.Contains(output, "kind: BGPAdvertisement") {
					t.Error("Should not render BGPAdvertisement when bgp.enabled=false")
				}
			},
		},
		{
			name:       "TC8: cilium driver renders no MetalLB resources",
			valuesFile: "values/tc8-cilium-driver.yaml",
			assertFunc: func(t *testing.T, output string) {
				if strings.Contains(output, "metallb.io/v1beta") {
					t.Error("MetalLB resources must not render for cilium driver")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valuesPath := filepath.Join(chartPath, "tests", tt.valuesFile)
			cmd := exec.Command(helmPath, "template", "ck-loadbalancer", chartPath, "-f", valuesPath)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("helm template failed: %v\nOutput: %s", err, output)
			}
			tt.assertFunc(t, string(output))
		})
	}
}

func TestHelmLint(t *testing.T) {
	helmPath := getHelmPath(t)

	chartPath, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get chart path: %v", err)
	}

	cmd := exec.Command(helmPath, "lint", chartPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm lint failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(string(output), "1 chart(s) linted, 0 chart(s) failed") {
		t.Errorf("Expected successful lint, got: %s", output)
	}
}
