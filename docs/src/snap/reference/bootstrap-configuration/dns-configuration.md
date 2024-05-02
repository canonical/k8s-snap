# DNS Configuration

## Format Specification

### dns.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled. If omitted defaults to `true`

### dns.cluster-domain
**Type:** `string`<br>
**Required:** `No` <br>

Sets the local domain of the cluster. If omitted defaults to `cluster.local`

### dns.service-ip
**Type:** `string`<br>
**Required:** `No` <br>

Sets the IP address of the dns service. If omitted defaults to the IP address of the Kubernetes service created by the feature.

Can be used to point to an external dns server when feature is disabled.


### dns.upstream-nameservers
**Type:** `list[string]`<br>
**Required:** `No` <br>

Sets the upstream nameservers used to forward queries for out-of-cluster endpoints. If omitted defaults to `/etc/resolv.conf` and uses the nameservers of the node.
