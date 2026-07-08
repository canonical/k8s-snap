package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var helmPath = "/opt/homebrew/bin/helm"

func TestChartRender(t *testing.T) {
	// Check helm is available
	if _, err := os.Stat(helmPath); os.IsNotExist(err) {
		t.Skipf("helm not found at %s", helmPath)
	}

	chartPath, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get chart path: %v", err)
	}

	tests := []struct {
		name      string
		valuesFile string
		assertFunc func(t *testing.T, output string)
	}{
		{
			name:      "TC1: single peer regression",
			valuesFile: "values/tc1-single-peer.yaml",
			assertFunc: func(t *testing.T, output string) {
				// Check single BGPPeer with bare name (no suffix)
				if !strings.Contains(output, "kind: BGPPeer") {
					t.Error("Expected BGPPeer CR")
				}
				// Check that metadata name is bare (not suffixed with index)
				lines := strings.Split(output, "\n")
				for i, line := range lines {
					if strings.Contains(line, "kind: BGPPeer") {
						// Next few lines should contain metadata
						for j := i + 1; j < i+10 && j < len(lines); j++ {
							if strings.Contains(lines[j], "metadata:") {
								for k := j + 1; k < j+5 && k < len(lines); k++ {
									trimmed := strings.TrimSpace(lines[k])
									if strings.HasPrefix(trimmed, "name:") {
										// name should be exactly "name: ck-loadbalancer" (not name: ck-loadbalancer-0, etc)
										expectedName := "name: ck-loadbalancer"
										if trimmed != expectedName {
											t.Errorf("Expected exact name '%s', got '%s'", expectedName, trimmed)
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
				// Check BGPAdvertisement has ipAddressPools
				if !strings.Contains(output, "ipAddressPools:") {
					t.Error("Expected ipAddressPools in BGPAdvertisement")
				}
				// Check spec fields
				if !strings.Contains(output, "peerAddress: 10.0.0.1") {
					t.Error("Expected peerAddress: 10.0.0.1")
				}
				if !strings.Contains(output, "peerASN: 65001") {
					t.Error("Expected peerASN: 65001")
				}
				if !strings.Contains(output, "myASN: 64512") {
					t.Error("Expected myASN: 64512")
				}
				if !strings.Contains(output, "peerPort: 179") {
					t.Error("Expected peerPort: 179")
				}
			},
		},
		{
			name:      "TC2: three peers with nodeSelectors",
			valuesFile: "values/tc2-multi-peer.yaml",
			assertFunc: func(t *testing.T, output string) {
				// Check we have 3 BGPPeer CRs
				count := strings.Count(output, "kind: BGPPeer")
				if count != 3 {
					t.Errorf("Expected 3 BGPPeer CRs, got %d", count)
				}
				
				// Check naming: index 0 bare, index 1+ suffixed
				lines := strings.Split(output, "\n")
				foundBareName := false
				foundSuffix1 := false
				foundSuffix2 := false
				for _, line := range lines {
					if strings.TrimSpace(line) == "name: ck-loadbalancer" {
						foundBareName = true
					}
					if strings.Contains(line, "name: ck-loadbalancer-1") {
						foundSuffix1 = true
					}
					if strings.Contains(line, "name: ck-loadbalancer-2") {
						foundSuffix2 = true
					}
				}
				if !foundBareName {
					t.Error("Expected bare name for first peer")
				}
				if !foundSuffix1 {
					t.Error("Expected ck-loadbalancer-1 for second peer")
				}
				if !foundSuffix2 {
					t.Error("Expected ck-loadbalancer-2 for third peer")
				}

				// Check nodeSelectors
				if !strings.Contains(output, "topology.kubernetes.io/zone: i1") {
					t.Error("Expected zone i1 nodeSelector")
				}
				if !strings.Contains(output, "topology.kubernetes.io/zone: i2") {
					t.Error("Expected zone i2 nodeSelector")
				}
				if !strings.Contains(output, "topology.kubernetes.io/zone: i3") {
					t.Error("Expected zone i3 nodeSelector")
				}
				
				// Check ASNs
				if !strings.Contains(output, "peerASN: 65001") {
					t.Error("Expected peerASN: 65001")
				}
				if !strings.Contains(output, "peerASN: 65002") {
					t.Error("Expected peerASN: 65002")
				}
				if !strings.Contains(output, "peerASN: 65003") {
					t.Error("Expected peerASN: 65003")
				}
			},
		},
		{
			name:      "TC3: advertiseAllPools=true",
			valuesFile: "values/tc3-advertise-all.yaml",
			assertFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "kind: BGPAdvertisement") {
					t.Error("Expected BGPAdvertisement CR")
				}
				// Should NOT have ipAddressPools
				if strings.Contains(output, "ipAddressPools:") {
					t.Error("advertiseAllPools=true should omit ipAddressPools")
				}
				// Check spec is present (even if empty)
				if !strings.Contains(output, "spec:") {
					t.Error("Expected spec: in BGPAdvertisement")
				}
			},
		},
		{
			name:      "TC4: custom myASN",
			valuesFile: "values/tc4-custom-myasn.yaml",
			assertFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "myASN: 65099") {
					t.Error("Expected myASN: 65099 from neighbor config")
				}
				// Should not use localASN
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.Contains(line, "myASN:") && strings.Contains(line, "64512") {
						t.Error("Should use neighbor myASN, not localASN")
					}
				}
			},
		},
		{
			name:      "TC5: default peerPort",
			valuesFile: "values/tc5-no-peerport.yaml",
			assertFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "peerPort: 179") {
					t.Error("Expected peerPort: 179 (default)")
				}
			},
		},
		{
			name:      "TC6: bgp.enabled=false",
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
			name:      "TC7: empty neighbors list",
			valuesFile: "values/tc7-empty-neighbors.yaml",
			assertFunc: func(t *testing.T, output string) {
				if strings.Contains(output, "kind: BGPPeer") {
					t.Error("Should not render BGPPeer with empty neighbors")
				}
				// BGPAdvertisement should still be present
				if !strings.Contains(output, "kind: BGPAdvertisement") {
					t.Error("Expected BGPAdvertisement even with empty neighbors")
				}
			},
		},
		{
			name:      "TC8: cilium driver",
			valuesFile: "values/tc8-cilium-driver.yaml",
			assertFunc: func(t *testing.T, output string) {
				// MetalLB template should not render
				if strings.Contains(output, "kind: BGPPeer") {
					t.Error("MetalLB BGPPeer should not render for cilium driver")
				}
				if strings.Contains(output, "metallb.io/v1beta") {
					t.Error("MetalLB resources should not render for cilium driver")
				}
			},
		},
		{
			name:      "TC9: multi-label nodeSelector",
			valuesFile: "values/tc9-multi-label.yaml",
			assertFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "nodeSelectors:") {
					t.Error("Expected nodeSelectors")
				}
				if !strings.Contains(output, "topology.kubernetes.io/zone: i1") {
					t.Error("Expected zone label")
				}
				// Check for worker label - empty string value renders without quotes
				if !strings.Contains(output, "node-role.kubernetes.io/worker:") {
					t.Error("Expected worker role label")
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
	if _, err := os.Stat(helmPath); os.IsNotExist(err) {
		t.Skipf("helm not found at %s", helmPath)
	}

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
