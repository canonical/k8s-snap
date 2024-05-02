# Load-balancer configuration

## Format Specification

### load-balancer.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled. If omitted defaults to `false`

### load-balancer.cidrs
**Type:** `list[string]`<br>
**Required:** `No` <br>

Sets the CIDRs used for assigning IP addresses to Kubernetes services with type `LoadBalancer`.

### load-balancer.l2-mode
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if L2 mode should be enabled. If omitted defaults to `false`

### load-balancer.l2-interfaces
**Type:** `list[string]`<br>
**Required:** `No` <br>

Sets the interfaces to be used for announcing IP addresses through ARP. If omitted all interfaces will be used.

### load-balancer.bgp-mode
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if BGP mode should be enabled. If omitted defaults to `false`

### load-balancer.bgp-local-asn
**Type:** `int`<br>
**Required:** `Yes if bgp-mode is true` <br>

Sets the ASN to be used for the local virtual BGP router.

### load-balancer.bgp-peer-address
**Type:** `string`<br>
**Required:** `Yes if bgp-mode is true` <br>

Sets the IP address of the BGP peer.

### load-balancer.bgp-peer-asn
**Type:** `int`<br>
**Required:** `Yes if bgp-mode is true` <br>

Sets the ASN of the BGP peer.

### load-balancer.bgp-peer-port
**Type:** `int`<br>
**Required:** `Yes if bgp-mode is true` <br>

Sets the port of the BGP peer.
