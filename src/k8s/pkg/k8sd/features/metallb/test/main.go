package main

import (
	"fmt"
	"log"

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

	/*
		values := map[string]any{
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
	*/

	lbValues := values.CkLoadbalancerValues{
		Driver: "metallb",
		L2: values.CkLoadbalancerValues_L2{
			Enabled:    loadbalancer.GetL2Mode(),
			Interfaces: values.ToAnyList(loadbalancer.GetL2Interfaces()),
		},
		IpPool: values.CkLoadbalancerValues_IpPool{
			Cidrs: values.ToAnyList(cidrs),
		},
		Bgp: values.CkLoadbalancerValues_Bgp{
			Enabled:  loadbalancer.GetBGPMode(),
			LocalAsn: loadbalancer.GetBGPLocalASN(),
			Neighbors: values.ToAnyList([]map[string]any{
				{
					"peerAddress": loadbalancer.GetBGPPeerAddress(),
					"peerASN":     loadbalancer.GetBGPPeerASN(),
					"peerPort":    loadbalancer.GetBGPPeerPort(),
				},
			}),
		},
	}

	lbValuesMap, err := lbValues.ToMap()
	if err != nil {
		log.Fatalf("failed to convert LoadBalancer values to map: %v", err)
	}

	fmt.Println(lbValuesMap)
	/* output:
	map[bgp:map[enabled:true localAsn:64512 neighbors:[map[peerASN:64513 peerAddress:10.0.0.1/32 peerPort:179]]] driver:metallb ipPool:map[cidrs:[map[cidr:192.0.2.0/24] map[start:20.0.20.100 stop:20.0.20.200]]] l2:map[enabled:true interfaces:[eth0 eth1]]]
	*/
}
