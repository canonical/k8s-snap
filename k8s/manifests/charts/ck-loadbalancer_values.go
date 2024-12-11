package main

type CkLoadbalancerValues_L2 struct {
	Enabled any
	//  interfaces:
	//  - "^eth[0-9]+"
	Interfaces []any
}

type CkLoadbalancerValues_IpPool struct {
	//  cidrs:
	//  - cidr: "10.42.254.176/28"
	Cidrs []any
}

type CkLoadbalancerValues_Bgp struct {
	Enabled  any
	LocalAsn any
	//  neighbors:
	//  - peerAddress: '10.0.0.60/24'
	//    peerASN: 65100
	//    peerPort: 179
	Neighbors []any
}

type CkLoadbalancerValues struct {
	Driver any
	L2     CkLoadbalancerValues_L2
	IpPool CkLoadbalancerValues_IpPool
	Bgp    CkLoadbalancerValues_Bgp
}
