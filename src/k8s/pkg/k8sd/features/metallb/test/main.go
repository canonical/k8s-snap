package main

import (
	"fmt"
	"log"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/values"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"k8s.io/utils/ptr"
)

func main() {
	loadbalancer := &types.LoadBalancer{
		Enabled:        ptr.To(true),
		L2Mode:         ptr.To(true),
		L2Interfaces:   ptr.To([]string{"eth0", "eth1"}),
		BGPMode:        ptr.To(true),
		BGPLocalASN:    ptr.To(64512),
		BGPPeerAddress: ptr.To("10.0.0.1/32"),
		BGPPeerASN:     ptr.To(64513),
		BGPPeerPort:    ptr.To(179),
		CIDRs:          ptr.To([]string{"192.0.2.0/24"}),
		IPRanges: ptr.To([]types.LoadBalancer_IPRange{
			{Start: "20.0.20.100", Stop: "20.0.20.200"},
		}),
	}

	cidrs := []map[string]any{}
	for _, cidr := range loadbalancer.GetCIDRs() {
		cidrs = append(cidrs, map[string]any{"cidr": cidr})
	}
	for _, ipRange := range loadbalancer.GetIPRanges() {
		cidrs = append(cidrs, map[string]any{"start": ipRange.Start, "stop": ipRange.Stop})
	}

	oldValues := map[string]any{
		"driver": "metallb",
		"l2": map[string]any{
			"enabled":    loadbalancer.GetL2Mode(),
			"interfaces": loadbalancer.GetL2Interfaces(),
		},
		"ipPool": map[string]any{
			"cidrs": cidrs,
		},
		"bgp": map[string]any{
			"enabled":  loadbalancer.GetBGPMode(),
			"localASN": loadbalancer.GetBGPLocalASN(),
			"neighbors": []map[string]any{
				{
					"peerAddress": loadbalancer.GetBGPPeerAddress(),
					"peerASN":     loadbalancer.GetBGPPeerASN(),
					"peerPort":    loadbalancer.GetBGPPeerPort(),
				},
			},
		},
	}

	lbValues := values.CkLoadbalancerValues{
		Driver: ptr.To("metallb"),
		L2: &values.CkLoadbalancerValues_L2{
			// Enabled:    ptr.To(loadbalancer.GetL2Mode()),
			Interfaces: ptr.To(features.ToAnyList(loadbalancer.GetL2Interfaces())),
			UNSAFE_MISC_FIELDS: map[string]any{
				"enabled": true,
			},
		},
		IpPool: &values.CkLoadbalancerValues_IpPool{
			Cidrs: ptr.To(features.ToAnyList(cidrs)),
		},
		Bgp: &values.CkLoadbalancerValues_Bgp{
			Enabled:  ptr.To(loadbalancer.GetBGPMode()),
			LocalAsn: ptr.To(int64(loadbalancer.GetBGPLocalASN())),
			Neighbors: ptr.To(features.ToAnyList([]map[string]any{
				{
					"peerAddress": loadbalancer.GetBGPPeerAddress(),
					"peerASN":     loadbalancer.GetBGPPeerASN(),
					"peerPort":    loadbalancer.GetBGPPeerPort(),
				},
			})),
		},
		UNSAFE_MISC_FIELDS: map[string]any{
			"l2": map[string]any{
				"enabled": false,
			},
		},
	}

	newValues, err := lbValues.ToMap()
	if err != nil {
		log.Fatalf("failed to convert LoadBalancer values to map: %v", err)
	}

	fmt.Println(newValues)
	fmt.Println()
	fmt.Println(oldValues)
}
