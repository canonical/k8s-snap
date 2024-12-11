package main

type Cilium1163Values_Debug struct {
	//  -- Enable debug logging
	Enabled any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Configure verbosity levels for debug logging
	//  This option is used to enable debug messages for operations related to such
	//  sub-system such as (e.g. kvstore, envoy, datapath or policy), and flow is
	//  for enabling debug messages emitted per request, message and connection.
	//  Multiple values can be set via a space-separated string (e.g. "datapath envoy").
	//
	//  Applicable values:
	//  - flow
	//  - kvstore
	//  - envoy
	//  - datapath
	//  - policy
	Verbose any
}

type Cilium1163Values_Rbac struct {
	//  -- Enable creation of Resource-Based Access Control configuration.
	Create any
}

type Cilium1163Values_K8SClientRateLimit struct {
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- (int) The sustained request rate in requests per second.
	//  @default -- 5 for k8s up to 1.26. 10 for k8s version 1.27+
	Qps any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- (int) The burst request rate in requests per second.
	//  The rate limiter will allow short bursts with a higher rate.
	//  @default -- 10 for k8s up to 1.26. 20 for k8s version 1.27+
	Burst any
}

type Cilium1163Values_Cluster struct {
	//  -- Name of the cluster. Only required for Cluster Mesh and mutual authentication with SPIRE.
	//  It must respect the following constraints:
	//  * It must contain at most 32 characters;
	//  * It must begin and end with a lower case alphanumeric character;
	//  * It may contain lower case alphanumeric characters and dashes between.
	//  The "default" name cannot be used if the Cluster ID is different from 0.
	Name any
	//  -- (int) Unique ID of the cluster. Must be unique across all connected
	//  clusters and in the range of 1 to 255. Only required for Cluster Mesh,
	//  may be 0 if Cluster Mesh is not used.
	Id any
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium struct {
	Create      any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium_Annotations
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Nodeinit_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Nodeinit struct {
	Create any
	//  -- Enabled is temporary until https://github.com/cilium/cilium-cli/issues/1396 is implemented.
	//  Cilium CLI doesn't create the SAs for node-init, thus the workaround. Helm is not affected by
	//  this issue. Name and automount can be configured, if enabled is set to true.
	//  Otherwise, they are ignored. Enabled can be removed once the issue is fixed.
	//  Cilium-nodeinit DS must also be fixed.
	Enabled     any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Nodeinit_Annotations
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Envoy_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Envoy struct {
	Create      any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Envoy_Annotations
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Operator_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Operator struct {
	Create      any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Operator_Annotations
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Preflight_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Preflight struct {
	Create      any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Preflight_Annotations
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Relay_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Relay struct {
	Create      any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Relay_Annotations
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Ui_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Ui struct {
	Create      any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Ui_Annotations
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_ClustermeshApiserver_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_ClustermeshApiserver struct {
	Create      any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_ClustermeshApiserver_Annotations
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Clustermeshcertgen_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Clustermeshcertgen struct {
	Create      any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Clustermeshcertgen_Annotations
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Hubblecertgen_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_ServiceAccounts_Hubblecertgen struct {
	Create      any
	Name        any
	Automount   any
	Annotations Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium1163Values_Cilium1163Values_ServiceAccounts_Hubblecertgen_Annotations
}

type Cilium1163Values_ServiceAccounts struct {
	Cilium               Cilium1163Values_Cilium1163Values_ServiceAccounts_Cilium
	Nodeinit             Cilium1163Values_Cilium1163Values_ServiceAccounts_Nodeinit
	Envoy                Cilium1163Values_Cilium1163Values_ServiceAccounts_Envoy
	Operator             Cilium1163Values_Cilium1163Values_ServiceAccounts_Operator
	Preflight            Cilium1163Values_Cilium1163Values_ServiceAccounts_Preflight
	Relay                Cilium1163Values_Cilium1163Values_ServiceAccounts_Relay
	Ui                   Cilium1163Values_Cilium1163Values_ServiceAccounts_Ui
	ClustermeshApiserver Cilium1163Values_Cilium1163Values_ServiceAccounts_ClustermeshApiserver
	//  -- Clustermeshcertgen is used if clustermesh.apiserver.tls.auto.method=cronJob
	Clustermeshcertgen Cilium1163Values_Cilium1163Values_ServiceAccounts_Clustermeshcertgen
	//  -- Hubblecertgen is used if hubble.tls.auto.method=cronJob
	Hubblecertgen Cilium1163Values_Cilium1163Values_ServiceAccounts_Hubblecertgen
}

type Cilium1163Values_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	PullPolicy any
	//  cilium-digest
	Digest    any
	UseDigest any
}

type Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels struct {
	K8SApp any
}

type Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector struct {
	MatchLabels Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels
}

type Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem struct {
	TopologyKey   any
	LabelSelector Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector
}

type Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution []Cilium1163Values_Cilium1163Values_Affinity_Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem
}

type Cilium1163Values_Affinity struct {
	PodAntiAffinity Cilium1163Values_Cilium1163Values_Affinity_PodAntiAffinity
}

type Cilium1163Values_NodeSelector struct {
	KubernetesIoos any
}

type Cilium1163Values_TolerationsItem struct {
	//  - key: "key"
	//    operator: "Equal|Exists"
	//    value: "value"
	//    effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"
	Operator any
}

type Cilium1163Values_ExtraConfig struct {
}

type Cilium1163Values_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_PodSecurityContext_AppArmorProfile struct {
	Type any
}

type Cilium1163Values_PodSecurityContext struct {
	//  -- AppArmorProfile options for the `cilium-agent` and init containers
	AppArmorProfile Cilium1163Values_Cilium1163Values_PodSecurityContext_AppArmorProfile
}

type Cilium1163Values_PodAnnotations struct {
}

type Cilium1163Values_PodLabels struct {
}

type Cilium1163Values_Resources struct {
}

type Cilium1163Values_InitResources struct {
}

type Cilium1163Values_Cilium1163Values_SecurityContext_SeLinuxOptions struct {
	Level any
	//  Running with spc_t since we have removed the privileged mode.
	//  Users can change it to a different type as long as they have the
	//  type available on the system.
	Type any
}

type Cilium1163Values_Cilium1163Values_SecurityContext_Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities_CiliumAgentItem struct {
}

type Cilium1163Values_Cilium1163Values_SecurityContext_Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities_MountCgroupItem struct {
}

type Cilium1163Values_Cilium1163Values_SecurityContext_Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities_ApplySysctlOverwritesItem struct {
}

type Cilium1163Values_Cilium1163Values_SecurityContext_Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities_CleanCiliumStateItem struct {
}

type Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities struct {
	//  -- Capabilities for the `cilium-agent` container
	CiliumAgent []Cilium1163Values_Cilium1163Values_SecurityContext_Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities_CiliumAgentItem
	//  -- Capabilities for the `mount-cgroup` init container
	MountCgroup []Cilium1163Values_Cilium1163Values_SecurityContext_Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities_MountCgroupItem
	//  -- capabilities for the `apply-sysctl-overwrites` init container
	ApplySysctlOverwrites []Cilium1163Values_Cilium1163Values_SecurityContext_Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities_ApplySysctlOverwritesItem
	//  -- Capabilities for the `clean-cilium-state` init container
	CleanCiliumState []Cilium1163Values_Cilium1163Values_SecurityContext_Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities_CleanCiliumStateItem
}

type Cilium1163Values_SecurityContext struct {
	//  -- User to run the pod with
	//  runAsUser: 0
	//  -- Run the pod with elevated privileges
	Privileged any
	//  -- SELinux options for the `cilium-agent` and init containers
	SeLinuxOptions Cilium1163Values_Cilium1163Values_SecurityContext_SeLinuxOptions
	Capabilities   Cilium1163Values_Cilium1163Values_SecurityContext_Capabilities
}

type Cilium1163Values_Cilium1163Values_UpdateStrategy_RollingUpdate struct {
	//  @schema
	//  type: [integer, string]
	//  @schema
	MaxUnavailable any
}

type Cilium1163Values_UpdateStrategy struct {
	Type          any
	RollingUpdate Cilium1163Values_Cilium1163Values_UpdateStrategy_RollingUpdate
}

type Cilium1163Values_Aksbyocni struct {
	//  -- Enable AKS BYOCNI integration.
	//  Note that this is incompatible with AKS clusters not created in BYOCNI mode:
	//  use Azure integration (`azure.enabled`) instead.
	Enabled any
}

type Cilium1163Values_Azure struct {
	//  -- Enable Azure integration.
	//  Note that this is incompatible with AKS clusters created in BYOCNI mode: use
	//  AKS BYOCNI integration (`aksbyocni.enabled`) instead.
	//  usePrimaryAddress: false
	//  resourceGroup: group1
	//  subscriptionID: 00000000-0000-0000-0000-000000000000
	//  tenantID: 00000000-0000-0000-0000-000000000000
	//  clientID: 00000000-0000-0000-0000-000000000000
	//  clientSecret: 00000000-0000-0000-0000-000000000000
	//  userAssignedIdentityID: 00000000-0000-0000-0000-000000000000
	Enabled any
}

type Cilium1163Values_Alibabacloud struct {
	//  -- Enable AlibabaCloud ENI integration
	Enabled any
}

type Cilium1163Values_BandwidthManager struct {
	//  -- Enable bandwidth manager infrastructure (also prerequirement for BBR)
	Enabled any
	//  -- Activate BBR TCP congestion control for Pods
	Bbr any
}

type Cilium1163Values_Nat46X64Gateway struct {
	//  -- Enable RFC8215-prefixed translation
	Enabled any
}

type Cilium1163Values_HighScaleIpcache struct {
	//  -- Enable the high scale mode for the ipcache.
	Enabled any
}

type Cilium1163Values_L2Announcements struct {
	//  -- Enable L2 announcements
	//  -- If a lease is not renewed for X duration, the current leader is considered dead, a new leader is picked
	//  leaseDuration: 15s
	//  -- The interval at which the leader will renew the lease
	//  leaseRenewDeadline: 5s
	//  -- The timeout between retries if renewal fails
	//  leaseRetryPeriod: 2s
	Enabled any
}

type Cilium1163Values_L2PodAnnouncements struct {
	//  -- Enable L2 pod announcements
	Enabled any
	//  -- Interface used for sending Gratuitous ARP pod announcements
	Interface any
}

type Cilium1163Values_Cilium1163Values_Bgp_Announce struct {
	//  -- Enable allocation and announcement of service LoadBalancer IPs
	LoadbalancerIp any
	//  -- Enable announcement of node pod CIDR
	PodCidr any
}

type Cilium1163Values_Bgp struct {
	//  -- Enable BGP support inside Cilium; embeds a new ConfigMap for BGP inside
	//  cilium-agent and cilium-operator
	Enabled  any
	Announce Cilium1163Values_Cilium1163Values_Bgp_Announce
}

type Cilium1163Values_Cilium1163Values_BgpControlPlane_SecretsNamespace struct {
	//  -- Create secrets namespace for BGP secrets.
	Create any
	//  -- The name of the secret namespace to which Cilium agents are given read access
	Name any
}

type Cilium1163Values_BgpControlPlane struct {
	//  -- Enables the BGP control plane.
	Enabled any
	//  -- SecretsNamespace is the namespace which BGP support will retrieve secrets from.
	SecretsNamespace Cilium1163Values_Cilium1163Values_BgpControlPlane_SecretsNamespace
}

type Cilium1163Values_PmtuDiscovery struct {
	//  -- Enable path MTU discovery to send ICMP fragmentation-needed replies to
	//  the client.
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Bpf_AutoMount struct {
	//  -- Enable automatic mount of BPF filesystem
	//  When `autoMount` is enabled, the BPF filesystem is mounted at
	//  `bpf.root` path on the underlying host and inside the cilium agent pod.
	//  If users disable `autoMount`, it's expected that users have mounted
	//  bpffs filesystem at the specified `bpf.root` volume, and then the
	//  volume will be mounted inside the cilium agent pod at the same path.
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Bpf_Cilium1163Values_Cilium1163Values_Bpf_Events_Drop struct {
	//  -- Enable drop events.
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Bpf_Cilium1163Values_Cilium1163Values_Bpf_Events_PolicyVerdict struct {
	//  -- Enable policy verdict events.
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Bpf_Cilium1163Values_Cilium1163Values_Bpf_Events_Trace struct {
	//  -- Enable trace events.
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Bpf_Events struct {
	Drop          Cilium1163Values_Cilium1163Values_Bpf_Cilium1163Values_Cilium1163Values_Bpf_Events_Drop
	PolicyVerdict Cilium1163Values_Cilium1163Values_Bpf_Cilium1163Values_Cilium1163Values_Bpf_Events_PolicyVerdict
	Trace         Cilium1163Values_Cilium1163Values_Bpf_Cilium1163Values_Cilium1163Values_Bpf_Events_Trace
}

type Cilium1163Values_Bpf struct {
	AutoMount Cilium1163Values_Cilium1163Values_Bpf_AutoMount
	//  -- Configure the mount point for the BPF filesystem
	Root any
	//  -- Enables pre-allocation of eBPF map values. This increases
	//  memory usage but can reduce latency.
	PreallocateMaps any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- (int) Configure the maximum number of entries in auth map.
	//  @default -- `524288`
	AuthMapMax any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- (int) Configure the maximum number of entries in the TCP connection tracking
	//  table.
	//  @default -- `524288`
	CtTcpMax any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- (int) Configure the maximum number of entries for the non-TCP connection
	//  tracking table.
	//  @default -- `262144`
	CtAnyMax any
	//  -- Control events generated by the Cilium datapath exposed to Cilium monitor and Hubble.
	Events Cilium1163Values_Cilium1163Values_Bpf_Events
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- Configure the maximum number of service entries in the
	//  load balancer maps.
	LbMapMax any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- (int) Configure the maximum number of entries for the NAT table.
	//  @default -- `524288`
	NatMax any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- (int) Configure the maximum number of entries for the neighbor table.
	//  @default -- `524288`
	NeighMax any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  @default -- `16384`
	//  -- (int) Configures the maximum number of entries for the node table.
	NodeMapMax any
	//  -- Configure the maximum number of entries in endpoint policy map (per endpoint).
	//  @schema
	//  type: [null, integer]
	//  @schema
	PolicyMapMax any
	//  @schema
	//  type: [null, number]
	//  @schema
	//  -- (float64) Configure auto-sizing for all BPF maps based on available memory.
	//  ref: https://docs.cilium.io/en/stable/network/ebpf/maps/
	//  @default -- `0.0025`
	MapDynamicSizeRatio any
	//  -- Configure the level of aggregation for monitor notifications.
	//  Valid options are none, low, medium, maximum.
	MonitorAggregation any
	//  -- Configure the typical time between monitor notifications for
	//  active connections.
	MonitorInterval any
	//  -- Configure which TCP flags trigger notifications when seen for the
	//  first time in a connection.
	MonitorFlags any
	//  -- Allow cluster external access to ClusterIP services.
	LbExternalClusterIp any
	//  @schema
	//  type: [null, boolean]
	//  @schema
	//  -- (bool) Enable native IP masquerade support in eBPF
	//  @default -- `false`
	Masquerade any
	//  @schema
	//  type: [null, boolean]
	//  @schema
	//  -- (bool) Configure whether direct routing mode should route traffic via
	//  host stack (true) or directly and more efficiently out of BPF (false) if
	//  the kernel supports it. The latter has the implication that it will also
	//  bypass netfilter in the host namespace.
	//  @default -- `false`
	HostLegacyRouting any
	//  @schema
	//  type: [null, boolean]
	//  @schema
	//  -- (bool) Configure the eBPF-based TPROXY to reduce reliance on iptables rules
	//  for implementing Layer 7 policy.
	//  @default -- `false`
	Tproxy any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- (list) Configure explicitly allowed VLAN id's for bpf logic bypass.
	//  [0] will allow all VLAN id's without any filtering.
	//  @default -- `[]`
	VlanBypass any
	//  -- (bool) Disable ExternalIP mitigation (CVE-2020-8554)
	//  @default -- `false`
	DisableExternalIpmitigation any
	//  -- (bool) Attach endpoint programs using tcx instead of legacy tc hooks on
	//  supported kernels.
	//  @default -- `true`
	EnableTcx any
	//  -- (string) Mode for Pod devices for the core datapath (veth, netkit, netkit-l2, lb-only)
	//  @default -- `veth`
	DatapathMode any
}

type Cilium1163Values_Cilium1163Values_Cni_Cilium1163Values_Cilium1163Values_Cni_Resources_Requests struct {
	Cpu    any
	Memory any
}

type Cilium1163Values_Cilium1163Values_Cni_Resources struct {
	Requests Cilium1163Values_Cilium1163Values_Cni_Cilium1163Values_Cilium1163Values_Cni_Resources_Requests
}

type Cilium1163Values_Cni struct {
	//  -- Install the CNI configuration and binary files into the filesystem.
	Install any
	//  -- Remove the CNI configuration and binary files on agent shutdown. Enable this
	//  if you're removing Cilium from the cluster. Disable this to prevent the CNI
	//  configuration file from being removed during agent upgrade, which can cause
	//  nodes to go unmanageable.
	Uninstall any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Configure chaining on top of other CNI plugins. Possible values:
	//   - none
	//   - aws-cni
	//   - flannel
	//   - generic-veth
	//   - portmap
	ChainingMode any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- A CNI network name in to which the Cilium plugin should be added as a chained plugin.
	//  This will cause the agent to watch for a CNI network with this network name. When it is
	//  found, this will be used as the basis for Cilium's CNI configuration file. If this is
	//  set, it assumes a chaining mode of generic-veth. As a special case, a chaining mode
	//  of aws-cni implies a chainingTarget of aws-cni.
	ChainingTarget any
	//  -- Make Cilium take ownership over the `/etc/cni/net.d` directory on the
	//  node, renaming all non-Cilium CNI configurations to `*.cilium_bak`.
	//  This ensures no Pods can be scheduled using other CNI plugins during Cilium
	//  agent downtime.
	Exclusive any
	//  -- Configure the log file for CNI logging with retention policy of 7 days.
	//  Disable CNI file logging by setting this field to empty explicitly.
	LogFile any
	//  -- Skip writing of the CNI configuration. This can be used if
	//  writing of the CNI configuration is performed by external automation.
	CustomConf any
	//  -- Configure the path to the CNI configuration directory on the host.
	ConfPath any
	//  -- Configure the path to the CNI binary directory on the host.
	//  -- Specify the path to a CNI config to read from on agent start.
	//  This can be useful if you want to manage your CNI
	//  configuration outside of a Kubernetes environment. This parameter is
	//  mutually exclusive with the 'cni.configMap' parameter. The agent will
	//  write this to 05-cilium.conflist on startup.
	//  readCniConf: /host/etc/cni/net.d/05-sample.conflist.input
	BinPath any
	//  -- When defined, configMap will mount the provided value as ConfigMap and
	//  interpret the cniConf variable as CNI configuration file and write it
	//  when the agent starts up
	//  configMap: cni-configuration
	//
	//  -- Configure the key in the CNI ConfigMap to read the contents of
	//  the CNI configuration from.
	ConfigMapKey any
	//  -- Configure the path to where to mount the ConfigMap inside the agent pod.
	ConfFileMountPath any
	//  -- Configure the path to where the CNI configuration directory is mounted
	//  inside the agent pod.
	HostConfDirMountPath any
	//  -- Specifies the resources for the cni initContainer
	Resources Cilium1163Values_Cilium1163Values_Cni_Resources
	//  -- Enable route MTU for pod netns when CNI chaining is used
	EnableRouteMtuforCnichaining any
}

type Cilium1163Values_CustomCalls struct {
	//  -- Enable tail call hooks for custom eBPF programs.
	Enabled any
}

type Cilium1163Values_Daemon struct {
	//  -- Configure where Cilium runtime state should be stored.
	RunPath any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Configure a custom list of possible configuration override sources
	//  The default is "config-map:cilium-config,cilium-node-config". For supported
	//  values, see the help text for the build-config subcommand.
	//  Note that this value should be a comma-separated string.
	ConfigSources any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- allowedConfigOverrides is a list of config-map keys that can be overridden.
	//  That is to say, if this value is set, config sources (excepting the first one) can
	//  only override keys in this list.
	//
	//  This takes precedence over blockedConfigOverrides.
	//
	//  By default, all keys may be overridden. To disable overrides, set this to "none" or
	//  change the configSources variable.
	AllowedConfigOverrides any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- blockedConfigOverrides is a list of config-map keys that may not be overridden.
	//  In other words, if any of these keys appear in a configuration source excepting the
	//  first one, they will be ignored
	//
	//  This is ignored if allowedConfigOverrides is set.
	//
	//  By default, all keys may be overridden.
	BlockedConfigOverrides any
}

type Cilium1163Values_Cilium1163Values_CiliumEndpointSlice_RateLimitsItem struct {
	Nodes any
	Limit any
	Burst any
}

type Cilium1163Values_CiliumEndpointSlice struct {
	//  -- Enable Cilium EndpointSlice feature.
	Enabled any
	//  -- List of rate limit options to be used for the CiliumEndpointSlice controller.
	//  Each object in the list must have the following fields:
	//  nodes: Count of nodes at which to apply the rate limit.
	//  limit: The sustained request rate in requests per second. The maximum rate that can be configured is 50.
	//  burst: The burst request rate in requests per second. The maximum burst that can be configured is 100.
	RateLimits []Cilium1163Values_Cilium1163Values_CiliumEndpointSlice_RateLimitsItem
}

type Cilium1163Values_Cilium1163Values_EnvoyConfig_SecretsNamespace struct {
	//  -- Create secrets namespace for CiliumEnvoyConfig CRDs.
	Create any
	//  -- The name of the secret namespace to which Cilium agents are given read access.
	Name any
}

type Cilium1163Values_EnvoyConfig struct {
	//  -- Enable CiliumEnvoyConfig CRD
	//  CiliumEnvoyConfig CRD can also be implicitly enabled by other options.
	Enabled any
	//  -- SecretsNamespace is the namespace in which envoy SDS will retrieve secrets from.
	SecretsNamespace Cilium1163Values_Cilium1163Values_EnvoyConfig_SecretsNamespace
	//  -- Interval in which an attempt is made to reconcile failed EnvoyConfigs. If the duration is zero, the retry is deactivated.
	RetryInterval any
}

type Cilium1163Values_Cilium1163Values_IngressController_IngressLbannotationPrefixesItem struct {
}

type Cilium1163Values_Cilium1163Values_IngressController_SecretsNamespace struct {
	//  -- Create secrets namespace for Ingress.
	Create any
	//  -- Name of Ingress secret namespace.
	Name any
	//  -- Enable secret sync, which will make sure all TLS secrets used by Ingress are synced to secretsNamespace.name.
	//  If disabled, TLS secrets must be maintained externally.
	Sync any
}

type Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_Service_Labels struct {
}

type Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_Service_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_IngressController_Service struct {
	//  -- Service name
	Name any
	//  -- Labels to be added for the shared LB service
	Labels Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_Service_Labels
	//  -- Annotations to be added for the shared LB service
	Annotations Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_Service_Annotations
	//  -- Service type for the shared LB service
	Type any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- Configure a specific nodePort for insecure HTTP traffic on the shared LB service
	InsecureNodePort any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- Configure a specific nodePort for secure HTTPS traffic on the shared LB service
	SecureNodePort any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Configure a specific loadBalancerClass on the shared LB service (requires Kubernetes 1.24+)
	LoadBalancerClass any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Configure a specific loadBalancerIP on the shared LB service
	LoadBalancerIp any
	//  @schema
	//  type: [null, boolean]
	//  @schema
	//  -- Configure if node port allocation is required for LB service
	//  ref: https://kubernetes.io/docs/concepts/services-networking/service/#load-balancer-nodeport-allocation
	AllocateLoadBalancerNodePorts any
	//  -- Control how traffic from external sources is routed to the LoadBalancer Kubernetes Service for Cilium Ingress in shared mode.
	//  Valid values are "Cluster" and "Local".
	//  ref: https://kubernetes.io/docs/reference/networking/virtual-ips/#external-traffic-policy
	ExternalTrafficPolicy any
}

type Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_HostNetwork_Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_HostNetwork_Nodes_MatchLabels struct {
}

type Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_HostNetwork_Nodes struct {
	//  -- Specify the labels of the nodes where the Ingress listeners should be exposed
	//
	//  matchLabels:
	//    kubernetes.io/os: linux
	//    kubernetes.io/hostname: kind-worker
	MatchLabels Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_HostNetwork_Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_HostNetwork_Nodes_MatchLabels
}

type Cilium1163Values_Cilium1163Values_IngressController_HostNetwork struct {
	//  -- Configure whether the Envoy listeners should be exposed on the host network.
	Enabled any
	//  -- Configure a specific port on the host network that gets used for the shared listener.
	SharedListenerPort any
	//  Specify the nodes where the Ingress listeners should be exposed
	Nodes Cilium1163Values_Cilium1163Values_IngressController_Cilium1163Values_Cilium1163Values_IngressController_HostNetwork_Nodes
}

type Cilium1163Values_IngressController struct {
	//  -- Enable cilium ingress controller
	//  This will automatically set enable-envoy-config as well.
	Enabled any
	//  -- Set cilium ingress controller to be the default ingress controller
	//  This will let cilium ingress controller route entries without ingress class set
	Default any
	//  -- Default ingress load balancer mode
	//  Supported values: shared, dedicated
	//  For granular control, use the following annotations on the ingress resource:
	//  "ingress.cilium.io/loadbalancer-mode: dedicated" (or "shared").
	LoadbalancerMode any
	//  -- Enforce https for host having matching TLS host in Ingress.
	//  Incoming traffic to http listener will return 308 http error code with respective location in header.
	EnforceHttps any
	//  -- Enable proxy protocol for all Ingress listeners. Note that _only_ Proxy protocol traffic will be accepted once this is enabled.
	EnableProxyProtocol any
	//  -- IngressLBAnnotations are the annotation and label prefixes, which are used to filter annotations and/or labels to propagate from Ingress to the Load Balancer service
	IngressLbannotationPrefixes []Cilium1163Values_Cilium1163Values_IngressController_IngressLbannotationPrefixesItem
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Default secret namespace for ingresses without .spec.tls[].secretName set.
	DefaultSecretNamespace any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Default secret name for ingresses without .spec.tls[].secretName set.
	DefaultSecretName any
	//  -- SecretsNamespace is the namespace in which envoy SDS will retrieve TLS secrets from.
	SecretsNamespace Cilium1163Values_Cilium1163Values_IngressController_SecretsNamespace
	//  -- Load-balancer service in shared mode.
	//  This is a single load-balancer service for all Ingress resources.
	Service Cilium1163Values_Cilium1163Values_IngressController_Service
	//  Host Network related configuration
	HostNetwork Cilium1163Values_Cilium1163Values_IngressController_HostNetwork
}

type Cilium1163Values_Cilium1163Values_GatewayApi_GatewayClass struct {
	//  -- Enable creation of GatewayClass resource
	//  The default value is 'auto' which decides according to presence of gateway.networking.k8s.io/v1/GatewayClass in the cluster.
	//  Other possible values are 'true' and 'false', which will either always or never create the GatewayClass, respectively.
	Create any
}

type Cilium1163Values_Cilium1163Values_GatewayApi_SecretsNamespace struct {
	//  -- Create secrets namespace for Gateway API.
	Create any
	//  -- Name of Gateway API secret namespace.
	Name any
	//  -- Enable secret sync, which will make sure all TLS secrets used by Ingress are synced to secretsNamespace.name.
	//  If disabled, TLS secrets must be maintained externally.
	Sync any
}

type Cilium1163Values_Cilium1163Values_GatewayApi_Cilium1163Values_Cilium1163Values_GatewayApi_HostNetwork_Cilium1163Values_Cilium1163Values_GatewayApi_Cilium1163Values_Cilium1163Values_GatewayApi_HostNetwork_Nodes_MatchLabels struct {
}

type Cilium1163Values_Cilium1163Values_GatewayApi_Cilium1163Values_Cilium1163Values_GatewayApi_HostNetwork_Nodes struct {
	//  -- Specify the labels of the nodes where the Ingress listeners should be exposed
	//
	//  matchLabels:
	//    kubernetes.io/os: linux
	//    kubernetes.io/hostname: kind-worker
	MatchLabels Cilium1163Values_Cilium1163Values_GatewayApi_Cilium1163Values_Cilium1163Values_GatewayApi_HostNetwork_Cilium1163Values_Cilium1163Values_GatewayApi_Cilium1163Values_Cilium1163Values_GatewayApi_HostNetwork_Nodes_MatchLabels
}

type Cilium1163Values_Cilium1163Values_GatewayApi_HostNetwork struct {
	//  -- Configure whether the Envoy listeners should be exposed on the host network.
	Enabled any
	//  Specify the nodes where the Ingress listeners should be exposed
	Nodes Cilium1163Values_Cilium1163Values_GatewayApi_Cilium1163Values_Cilium1163Values_GatewayApi_HostNetwork_Nodes
}

type Cilium1163Values_GatewayApi struct {
	//  -- Enable support for Gateway API in cilium
	//  This will automatically set enable-envoy-config as well.
	Enabled any
	//  -- Enable proxy protocol for all GatewayAPI listeners. Note that _only_ Proxy protocol traffic will be accepted once this is enabled.
	EnableProxyProtocol any
	//  -- Enable Backend Protocol selection support (GEP-1911) for Gateway API via appProtocol.
	EnableAppProtocol any
	//  -- Enable ALPN for all listeners configured with Gateway API. ALPN will attempt HTTP/2, then HTTP 1.1.
	//  Note that this will also enable `appProtocol` support, and services that wish to use HTTP/2 will need to indicate that via their `appProtocol`.
	EnableAlpn any
	//  -- The number of additional GatewayAPI proxy hops from the right side of the HTTP header to trust when determining the origin client's IP address.
	XffNumTrustedHops any
	//  -- Control how traffic from external sources is routed to the LoadBalancer Kubernetes Service for all Cilium GatewayAPI Gateway instances. Valid values are "Cluster" and "Local".
	//  Note that this value will be ignored when `hostNetwork.enabled == true`.
	//  ref: https://kubernetes.io/docs/reference/networking/virtual-ips/#external-traffic-policy
	ExternalTrafficPolicy any
	GatewayClass          Cilium1163Values_Cilium1163Values_GatewayApi_GatewayClass
	//  -- SecretsNamespace is the namespace in which envoy SDS will retrieve TLS secrets from.
	SecretsNamespace Cilium1163Values_Cilium1163Values_GatewayApi_SecretsNamespace
	//  Host Network related configuration
	HostNetwork Cilium1163Values_Cilium1163Values_GatewayApi_HostNetwork
}

type Cilium1163Values_Cilium1163Values_Encryption_StrictMode struct {
	//  -- Enable WireGuard Pod2Pod strict mode.
	Enabled any
	//  -- CIDR for the WireGuard Pod2Pod strict mode.
	Cidr any
	//  -- Allow dynamic lookup of remote node identities.
	//  This is required when tunneling is used or direct routing is used and the node CIDR and pod CIDR overlap.
	AllowRemoteNodeIdentities any
}

type Cilium1163Values_Cilium1163Values_Encryption_Ipsec struct {
	//  -- Name of the key file inside the Kubernetes secret configured via secretName.
	KeyFile any
	//  -- Path to mount the secret inside the Cilium pod.
	MountPath any
	//  -- Name of the Kubernetes secret containing the encryption keys.
	SecretName any
	//  -- The interface to use for encrypted traffic.
	Interface any
	//  -- Enable the key watcher. If disabled, a restart of the agent will be
	//  necessary on key rotations.
	KeyWatcher any
	//  -- Maximum duration of the IPsec key rotation. The previous key will be
	//  removed after that delay.
	KeyRotationDuration any
	//  -- Enable IPsec encrypted overlay
	EncryptedOverlay any
}

type Cilium1163Values_Cilium1163Values_Encryption_Wireguard struct {
	//  -- Enables the fallback to the user-space implementation (deprecated).
	UserspaceFallback any
	//  -- Controls WireGuard PersistentKeepalive option. Set 0s to disable.
	PersistentKeepalive any
}

type Cilium1163Values_Encryption struct {
	//  -- Enable transparent network encryption.
	Enabled any
	//  -- Encryption method. Can be either ipsec or wireguard.
	Type any
	//  -- Enable encryption for pure node to node traffic.
	//  This option is only effective when encryption.type is set to "wireguard".
	NodeEncryption any
	//  -- Configure the WireGuard Pod2Pod strict mode.
	StrictMode Cilium1163Values_Cilium1163Values_Encryption_StrictMode
	Ipsec      Cilium1163Values_Cilium1163Values_Encryption_Ipsec
	Wireguard  Cilium1163Values_Cilium1163Values_Encryption_Wireguard
}

type Cilium1163Values_EndpointHealthChecking struct {
	//  -- Enable connectivity health checking between virtual endpoints.
	Enabled any
}

type Cilium1163Values_EndpointRoutes struct {
	//  @schema
	//  type: [boolean, string]
	//  @schema
	//  -- Enable use of per endpoint routes instead of routing via
	//  the cilium_host interface.
	Enabled any
}

type Cilium1163Values_K8SNetworkPolicy struct {
	//  -- Enable support for K8s NetworkPolicy
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Eni_EniTags struct {
}

type Cilium1163Values_Cilium1163Values_Eni_GcTags struct {
}

type Cilium1163Values_Eni struct {
	//  -- Enable Elastic Network Interface (ENI) integration.
	Enabled any
	//  -- Update ENI Adapter limits from the EC2 API
	UpdateEc2AdapterLimitViaApi any
	//  -- Release IPs not used from the ENI
	AwsReleaseExcessIps any
	//  -- Enable ENI prefix delegation
	AwsEnablePrefixDelegation any
	//  -- EC2 API endpoint to use
	Ec2Apiendpoint any
	//  -- Tags to apply to the newly created ENIs
	EniTags Cilium1163Values_Cilium1163Values_Eni_EniTags
	//  -- Interval for garbage collection of unattached ENIs. Set to "0s" to disable.
	//  @default -- `"5m"`
	GcInterval any
	//  -- Additional tags attached to ENIs created by Cilium.
	//  Dangling ENIs with this tag will be garbage collected
	//  @default -- `{"io.cilium/cilium-managed":"true,"io.cilium/cluster-name":"<auto-detected>"}`
	GcTags Cilium1163Values_Cilium1163Values_Eni_GcTags
	//  -- If using IAM role for Service Accounts will not try to
	//  inject identity values from cilium-aws kubernetes secret.
	//  Adds annotation to service account if managed by Helm.
	//  See https://github.com/aws/amazon-eks-pod-identity-webhook
	IamRole any
	//  -- Filter via subnet IDs which will dictate which subnets are going to be used to create new ENIs
	//  Important note: This requires that each instance has an ENI with a matching subnet attached
	//  when Cilium is deployed. If you only want to control subnets for ENIs attached by Cilium,
	//  use the CNI configuration file settings (cni.customConf) instead.
	SubnetIdsFilter []any
	//  -- Filter via tags (k=v) which will dictate which subnets are going to be used to create new ENIs
	//  Important note: This requires that each instance has an ENI with a matching subnet attached
	//  when Cilium is deployed. If you only want to control subnets for ENIs attached by Cilium,
	//  use the CNI configuration file settings (cni.customConf) instead.
	SubnetTagsFilter []any
	//  -- Filter via AWS EC2 Instance tags (k=v) which will dictate which AWS EC2 Instances
	//  are going to be used to create new ENIs
	InstanceTagsFilter []any
}

type Cilium1163Values_ExternalIps struct {
	//  -- Enable ExternalIPs service support.
	Enabled any
}

type Cilium1163Values_Gke struct {
	//  -- Enable Google Kubernetes Engine integration
	Enabled any
}

type Cilium1163Values_HostFirewall struct {
	//  -- Enables the enforcement of host policies in the eBPF datapath.
	Enabled any
}

type Cilium1163Values_HostPort struct {
	//  -- Enable hostPort service support.
	Enabled any
}

type Cilium1163Values_SocketLb struct {
	//  -- Enable socket LB
	//  -- Disable socket lb for non-root ns. This is used to enable Istio routing rules.
	//  hostNamespaceOnly: false
	//  -- Enable terminating pod connections to deleted service backends.
	//  terminatePodConnections: true
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Certgen_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Certgen_PodLabels struct {
}

type Cilium1163Values_Cilium1163Values_Certgen_Cilium1163Values_Cilium1163Values_Certgen_Annotations_Job struct {
}

type Cilium1163Values_Cilium1163Values_Certgen_Cilium1163Values_Cilium1163Values_Certgen_Annotations_CronJob struct {
}

type Cilium1163Values_Cilium1163Values_Certgen_Annotations struct {
	Job     Cilium1163Values_Cilium1163Values_Certgen_Cilium1163Values_Cilium1163Values_Certgen_Annotations_Job
	CronJob Cilium1163Values_Cilium1163Values_Certgen_Cilium1163Values_Cilium1163Values_Certgen_Annotations_CronJob
}

type Cilium1163Values_Cilium1163Values_Certgen_Affinity struct {
}

type Cilium1163Values_Certgen struct {
	Image Cilium1163Values_Cilium1163Values_Certgen_Image
	//  -- Seconds after which the completed job pod will be deleted
	TtlSecondsAfterFinished any
	//  -- Labels to be added to hubble-certgen pods
	PodLabels Cilium1163Values_Cilium1163Values_Certgen_PodLabels
	//  -- Annotations to be added to the hubble-certgen initial Job and CronJob
	Annotations Cilium1163Values_Cilium1163Values_Certgen_Annotations
	//  -- Node tolerations for pod assignment on nodes with taints
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []any
	//  -- Additional certgen volumes.
	ExtraVolumes []any
	//  -- Additional certgen volumeMounts.
	ExtraVolumeMounts []any
	//  -- Affinity for certgen
	Affinity Cilium1163Values_Cilium1163Values_Certgen_Affinity
}

type Cilium1163Values_Cilium1163Values_Hubble_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Tls_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Tls_Server_Mtls struct {
	//  When set to true enforces mutual TLS between Hubble Metrics server and its clients.
	//  False allow non-mutual TLS connections.
	//  This option has no effect when TLS is disabled.
	Enabled   any
	UseSecret any
	//  -- Name of the ConfigMap containing the CA to validate client certificates against.
	//  If mTLS is enabled and this is unspecified, it will default to the
	//  same CA used for Hubble metrics server certificates.
	Name any
	//  -- Entry of the ConfigMap containing the CA.
	Key any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Tls_Server struct {
	//  -- Name of the Secret containing the certificate and key for the Hubble metrics server.
	//  If specified, cert and key are ignored.
	ExistingSecret any
	//  -- base64 encoded PEM values for the Hubble metrics server certificate (deprecated).
	//  Use existingSecret instead.
	Cert any
	//  -- base64 encoded PEM values for the Hubble metrics server key (deprecated).
	//  Use existingSecret instead.
	Key any
	//  -- Extra DNS names added to certificate when it's auto generated
	ExtraDnsNames []any
	//  -- Extra IP addresses added to certificate when it's auto generated
	ExtraIpAddresses []any
	//  -- Configure mTLS for the Hubble metrics server.
	Mtls Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Tls_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Tls_Server_Mtls
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Tls struct {
	//  Enable hubble metrics server TLS.
	Enabled any
	//  Configure hubble metrics server TLS.
	Server Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Tls_Server
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceAnnotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_RelabelingsItem_SourceLabelsItem struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_RelabelingsItem struct {
	SourceLabels []Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_RelabelingsItem_SourceLabelsItem
	TargetLabel  any
	Replacement  any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_TlsConfig struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor struct {
	//  -- Create ServiceMonitor resources for Prometheus Operator.
	//  This requires the prometheus CRDs to be available.
	//  ref: https://github.com/prometheus-operator/prometheus-operator/blob/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml)
	Enabled any
	//  -- Labels to add to ServiceMonitor hubble
	Labels Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_Labels
	//  -- Annotations to add to ServiceMonitor hubble
	Annotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_Annotations
	//  -- jobLabel to add for ServiceMonitor hubble
	JobLabel any
	//  -- Interval for scrape metrics.
	Interval any
	//  -- Relabeling configs for the ServiceMonitor hubble
	Relabelings []Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_RelabelingsItem
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Metrics relabeling configs for the ServiceMonitor hubble
	MetricRelabelings any
	//  Configure TLS for the ServiceMonitor.
	//  Note, when using TLS you will either need to specify
	//  tlsConfig.insecureSkipVerify or specify a CA to use.
	TlsConfig Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor_TlsConfig
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Dashboards_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Dashboards struct {
	Enabled any
	Label   any
	//  @schema
	//  type: [null, string]
	//  @schema
	Namespace   any
	LabelValue  any
	Annotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Dashboards_Annotations
}

type Cilium1163Values_Cilium1163Values_Hubble_Metrics struct {
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Configures the list of metrics to collect. If empty or null, metrics
	//  are disabled.
	//  Example:
	//
	//    enabled:
	//    - dns:query;ignoreAAAA
	//    - drop
	//    - tcp
	//    - flow
	//    - icmp
	//    - http
	//
	//  You can specify the list of metrics from the helm CLI:
	//
	//    --set hubble.metrics.enabled="{dns:query;ignoreAAAA,drop,tcp,flow,icmp,http}"
	//
	Enabled any
	//  -- Enables exporting hubble metrics in OpenMetrics format.
	EnableOpenMetrics any
	//  -- Configure the port the hubble metric server listens on.
	Port any
	Tls  Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Tls
	//  -- Annotations to be added to hubble-metrics service.
	ServiceAnnotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceAnnotations
	ServiceMonitor     Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_ServiceMonitor
	//  -- Grafana dashboards for hubble
	//  grafana can import dashboards based on the label and value
	//  ref: https://github.com/grafana/helm-charts/tree/main/charts/grafana#sidecar-for-dashboards
	Dashboards Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Metrics_Dashboards
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Redact_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Redact_Http_Headers struct {
	//  -- List of HTTP headers to allow: headers not matching will be redacted. Note: `allow` and `deny` lists cannot be used both at the same time, only one can be present.
	//  Example:
	//    redact:
	//      enabled: true
	//      http:
	//        headers:
	//          allow:
	//            - traceparent
	//            - tracestate
	//            - Cache-Control
	//
	//  You can specify the options from the helm CLI:
	//    --set hubble.redact.enabled="true"
	//    --set hubble.redact.http.headers.allow="traceparent,tracestate,Cache-Control"
	Allow []any
	//  -- List of HTTP headers to deny: matching headers will be redacted. Note: `allow` and `deny` lists cannot be used both at the same time, only one can be present.
	//  Example:
	//    redact:
	//      enabled: true
	//      http:
	//        headers:
	//          deny:
	//            - Authorization
	//            - Proxy-Authorization
	//
	//  You can specify the options from the helm CLI:
	//    --set hubble.redact.enabled="true"
	//    --set hubble.redact.http.headers.deny="Authorization,Proxy-Authorization"
	Deny []any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Redact_Http struct {
	//  -- Enables redacting URL query (GET) parameters.
	//  Example:
	//
	//    redact:
	//      enabled: true
	//      http:
	//        urlQuery: true
	//
	//  You can specify the options from the helm CLI:
	//
	//    --set hubble.redact.enabled="true"
	//    --set hubble.redact.http.urlQuery="true"
	UrlQuery any
	//  -- Enables redacting user info, e.g., password when basic auth is used.
	//  Example:
	//
	//    redact:
	//      enabled: true
	//      http:
	//        userInfo: true
	//
	//  You can specify the options from the helm CLI:
	//
	//    --set hubble.redact.enabled="true"
	//    --set hubble.redact.http.userInfo="true"
	UserInfo any
	Headers  Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Redact_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Redact_Http_Headers
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Redact_Kafka struct {
	//  -- Enables redacting Kafka's API key.
	//  Example:
	//
	//    redact:
	//      enabled: true
	//      kafka:
	//        apiKey: true
	//
	//  You can specify the options from the helm CLI:
	//
	//    --set hubble.redact.enabled="true"
	//    --set hubble.redact.kafka.apiKey="true"
	ApiKey any
}

type Cilium1163Values_Cilium1163Values_Hubble_Redact struct {
	Enabled any
	Http    Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Redact_Http
	Kafka   Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Redact_Kafka
}

type Cilium1163Values_Cilium1163Values_Hubble_PeerService struct {
	//  -- Service Port for the Peer service.
	//  If not set, it is dynamically assigned to port 443 if TLS is enabled and to
	//  port 80 if not.
	//  servicePort: 80
	//  -- Target Port for the Peer service, must match the hubble.listenAddress'
	//  port.
	TargetPort any
	//  -- The cluster domain to use to query the Hubble Peer service. It should
	//  be the local cluster.
	ClusterDomain any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Tls_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Tls_Auto_CertManagerIssuerRef struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Tls_Auto struct {
	//  -- Auto-generate certificates.
	//  When set to true, automatically generate a CA and certificates to
	//  enable mTLS between Hubble server and Hubble Relay instances. If set to
	//  false, the certs for Hubble server need to be provided by setting
	//  appropriate values below.
	Enabled any
	//  -- Set the method to auto-generate certificates. Supported values:
	//  - helm:         This method uses Helm to generate all certificates.
	//  - cronJob:      This method uses a Kubernetes CronJob the generate any
	//                  certificates not provided by the user at installation
	//                  time.
	//  - certmanager:  This method use cert-manager to generate & rotate certificates.
	Method any
	//  -- Generated certificates validity duration in days.
	CertValidityDuration any
	//  -- Schedule for certificates regeneration (regardless of their expiration date).
	//  Only used if method is "cronJob". If nil, then no recurring job will be created.
	//  Instead, only the one-shot job is deployed to generate the certificates at
	//  installation time.
	//
	//  Defaults to midnight of the first day of every fourth month. For syntax, see
	//  https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#schedule-syntax
	Schedule any
	//  [Example]
	//  certManagerIssuerRef:
	//    group: cert-manager.io
	//    kind: ClusterIssuer
	//    name: ca-issuer
	//  -- certmanager issuer used when hubble.tls.auto.method=certmanager.
	CertManagerIssuerRef Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Tls_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Tls_Auto_CertManagerIssuerRef
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Tls_Server struct {
	//  -- Name of the Secret containing the certificate and key for the Hubble server.
	//  If specified, cert and key are ignored.
	ExistingSecret any
	//  -- base64 encoded PEM values for the Hubble server certificate (deprecated).
	//  Use existingSecret instead.
	Cert any
	//  -- base64 encoded PEM values for the Hubble server key (deprecated).
	//  Use existingSecret instead.
	Key any
	//  -- Extra DNS names added to certificate when it's auto generated
	ExtraDnsNames []any
	//  -- Extra IP addresses added to certificate when it's auto generated
	ExtraIpAddresses []any
}

type Cilium1163Values_Cilium1163Values_Hubble_Tls struct {
	//  -- Enable mutual TLS for listenAddress. Setting this value to false is
	//  highly discouraged as the Hubble API provides access to potentially
	//  sensitive network flow metadata and is exposed on the host network.
	Enabled any
	//  -- Configure automatic TLS certificates generation.
	Auto Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Tls_Auto
	//  -- The Hubble server certificate and private key
	Server Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Tls_Server
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	//  hubble-relay-digest
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels struct {
	K8SApp any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector struct {
	MatchLabels Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem struct {
	TopologyKey   any
	LabelSelector Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution []Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity struct {
	PodAffinity Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity_PodAffinity
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_NodeSelector struct {
	KubernetesIoos any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_PodAnnotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_PodLabels struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_PodDisruptionBudget struct {
	//  -- enable PodDisruptionBudget
	//  ref: https://kubernetes.io/docs/concepts/workloads/pods/disruptions/
	Enabled any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Minimum number/percentage of pods that should remain scheduled.
	//  When it's set, maxUnavailable must be disabled by `maxUnavailable: null`
	MinAvailable any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Maximum number/percentage of pods that may be made unavailable
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_UpdateStrategy_RollingUpdate struct {
	//  @schema
	//  type: [integer, string]
	//  @schema
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_UpdateStrategy struct {
	Type          any
	RollingUpdate Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_UpdateStrategy_RollingUpdate
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_PodSecurityContext struct {
	FsGroup any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_SecurityContext_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_SecurityContext_Capabilities_DropItem struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_SecurityContext_Capabilities struct {
	Drop []Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_SecurityContext_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_SecurityContext_Capabilities_DropItem
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_SecurityContext struct {
	//  readOnlyRootFilesystem: true
	RunAsNonRoot any
	RunAsUser    any
	RunAsGroup   any
	Capabilities Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_SecurityContext_Capabilities
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Service struct {
	//  --- The type of service used for Hubble Relay access, either ClusterIP or NodePort.
	Type any
	//  --- The port to use when the service type is set to NodePort.
	NodePort any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Tls_Client struct {
	//  -- Name of the Secret containing the certificate and key for the Hubble metrics server.
	//  If specified, cert and key are ignored.
	ExistingSecret any
	//  -- base64 encoded PEM values for the Hubble relay client certificate (deprecated).
	//  Use existingSecret instead.
	Cert any
	//  -- base64 encoded PEM values for the Hubble relay client key (deprecated).
	//  Use existingSecret instead.
	Key any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Tls_Server struct {
	//  When set to true, enable TLS on for Hubble Relay server
	//  (ie: for clients connecting to the Hubble Relay API).
	Enabled any
	//  When set to true enforces mutual TLS between Hubble Relay server and its clients.
	//  False allow non-mutual TLS connections.
	//  This option has no effect when TLS is disabled.
	Mtls any
	//  -- Name of the Secret containing the certificate and key for the Hubble relay server.
	//  If specified, cert and key are ignored.
	ExistingSecret any
	//  -- base64 encoded PEM values for the Hubble relay server certificate (deprecated).
	//  Use existingSecret instead.
	Cert any
	//  -- base64 encoded PEM values for the Hubble relay server key (deprecated).
	//  Use existingSecret instead.
	Key any
	//  -- extra DNS names added to certificate when its auto gen
	ExtraDnsNames []any
	//  -- extra IP addresses added to certificate when its auto gen
	ExtraIpAddresses []any
	//  DNS name used by the backend to connect to the relay
	//  This is a simple workaround as the relay certificates are currently hardcoded to
	//  *.hubble-relay.cilium.io
	//  See https://github.com/cilium/cilium/pull/28709#discussion_r1371792546
	//  For GKE Dataplane V2 this should be set to relay.kube-system.svc.cluster.local
	RelayName any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Tls struct {
	//  -- The hubble-relay client certificate and private key.
	//  This keypair is presented to Hubble server instances for mTLS
	//  authentication and is required when hubble.tls.enabled is true.
	//  These values need to be set manually if hubble.tls.auto.enabled is false.
	Client Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Tls_Client
	//  -- The hubble-relay server certificate and private key
	Server Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Tls_Server
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_ServiceMonitor_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_ServiceMonitor_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_ServiceMonitor struct {
	//  -- Enable service monitors.
	//  This requires the prometheus CRDs to be available (see https://github.com/prometheus-operator/prometheus-operator/blob/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml)
	Enabled any
	//  -- Labels to add to ServiceMonitor hubble-relay
	Labels Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_ServiceMonitor_Labels
	//  -- Annotations to add to ServiceMonitor hubble-relay
	Annotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_ServiceMonitor_Annotations
	//  -- Interval for scrape metrics.
	Interval any
	//  -- Specify the Kubernetes namespace where Prometheus expects to find
	//  service monitors configured.
	//  namespace: ""
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Relabeling configs for the ServiceMonitor hubble-relay
	Relabelings any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Metrics relabeling configs for the ServiceMonitor hubble-relay
	MetricRelabelings any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus struct {
	Enabled        any
	Port           any
	ServiceMonitor Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus_ServiceMonitor
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Gops struct {
	//  -- Enable gops for hubble-relay
	Enabled any
	//  -- Configure gops listen port for hubble-relay
	Port any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Pprof struct {
	//  -- Enable pprof for hubble-relay
	Enabled any
	//  -- Configure pprof listen address for hubble-relay
	Address any
	//  -- Configure pprof listen port for hubble-relay
	Port any
}

type Cilium1163Values_Cilium1163Values_Hubble_Relay struct {
	//  -- Enable Hubble Relay (requires hubble.enabled=true)
	Enabled any
	//  -- Roll out Hubble Relay pods automatically when configmap is updated.
	RollOutPods any
	//  -- Hubble-relay container image.
	Image Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Image
	//  -- Specifies the resources for the hubble-relay pods
	Resources Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Resources
	//  -- Number of replicas run for the hubble-relay deployment.
	Replicas any
	//  -- Affinity for hubble-replay
	Affinity Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Affinity
	//  -- Pod topology spread constraints for hubble-relay
	//    - maxSkew: 1
	//      topologyKey: topology.kubernetes.io/zone
	//      whenUnsatisfiable: DoNotSchedule
	TopologySpreadConstraints []any
	//  -- Node labels for pod assignment
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
	NodeSelector Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_NodeSelector
	//  -- Node tolerations for pod assignment on nodes with taints
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []any
	//  -- Additional hubble-relay environment variables.
	ExtraEnv []any
	//  -- Annotations to be added to all top-level hubble-relay objects (resources under templates/hubble-relay)
	Annotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Annotations
	//  -- Annotations to be added to hubble-relay pods
	PodAnnotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_PodAnnotations
	//  -- Labels to be added to hubble-relay pods
	PodLabels Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_PodLabels
	//  PodDisruptionBudget settings
	PodDisruptionBudget Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_PodDisruptionBudget
	//  -- The priority class to use for hubble-relay
	PriorityClassName any
	//  -- Configure termination grace period for hubble relay Deployment.
	TerminationGracePeriodSeconds any
	//  -- hubble-relay update strategy
	UpdateStrategy Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_UpdateStrategy
	//  -- Additional hubble-relay volumes.
	ExtraVolumes []any
	//  -- Additional hubble-relay volumeMounts.
	ExtraVolumeMounts []any
	//  -- hubble-relay pod security context
	PodSecurityContext Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_PodSecurityContext
	//  -- hubble-relay container security context
	SecurityContext Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_SecurityContext
	//  -- hubble-relay service configuration.
	Service Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Service
	//  -- Host to listen to. Specify an empty string to bind to all the interfaces.
	ListenHost any
	//  -- Port to listen to.
	ListenPort any
	//  -- TLS configuration for Hubble Relay
	Tls Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Tls
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Dial timeout to connect to the local hubble instance to receive peer information (e.g. "30s").
	DialTimeout any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Backoff duration to retry connecting to the local hubble instance in case of failure (e.g. "30s").
	RetryTimeout any
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- (int) Max number of flows that can be buffered for sorting before being sent to the
	//  client (per request) (e.g. 100).
	SortBufferLenMax any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- When the per-request flows sort buffer is not full, a flow is drained every
	//  time this timeout is reached (only affects requests in follow-mode) (e.g. "1s").
	//  -- Port to use for the k8s service backed by hubble-relay pods.
	//  If not set, it is dynamically assigned to port 443 if TLS is enabled and to
	//  port 80 if not.
	//  servicePort: 80
	SortBufferDrainTimeout any
	//  -- Enable prometheus metrics for hubble-relay on the configured port at
	//  /metrics
	Prometheus Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Prometheus
	Gops       Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Gops
	Pprof      Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Relay_Pprof
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Standalone_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Standalone_Tls_CertsVolume struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Standalone_Tls struct {
	//  -- When deploying Hubble UI in standalone, with tls enabled for Hubble relay, it is required
	//  to provide a volume for mounting the client certificates.
	//    projected:
	//      defaultMode: 0400
	//      sources:
	//      - secret:
	//          name: hubble-ui-client-certs
	//          items:
	//          - key: tls.crt
	//            path: client.crt
	//          - key: tls.key
	//            path: client.key
	//          - key: ca.crt
	//            path: hubble-relay-ca.crt
	CertsVolume Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Standalone_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Standalone_Tls_CertsVolume
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Standalone struct {
	//  -- When true, it will allow installing the Hubble UI only, without checking dependencies.
	//  It is useful if a cluster already has cilium and Hubble relay installed and you just
	//  want Hubble UI to be deployed.
	//  When installed via helm, installing UI should be done via `helm upgrade` and when installed via the cilium cli, then `cilium hubble enable --ui`
	Enabled any
	Tls     Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Standalone_Tls
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Tls_Client struct {
	//  -- Name of the Secret containing the client certificate and key for Hubble UI
	//  If specified, cert and key are ignored.
	ExistingSecret any
	//  -- base64 encoded PEM values for the Hubble UI client certificate (deprecated).
	//  Use existingSecret instead.
	Cert any
	//  -- base64 encoded PEM values for the Hubble UI client key (deprecated).
	//  Use existingSecret instead.
	Key any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Tls struct {
	Client Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Tls_Client
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_SecurityContext struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_LivenessProbe struct {
	//  -- Enable liveness probe for Hubble-ui backend (requires Hubble-ui 0.12+)
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_ReadinessProbe struct {
	//  -- Enable readiness probe for Hubble-ui backend (requires Hubble-ui 0.12+)
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend struct {
	//  -- Hubble-ui backend image.
	Image Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_Image
	//  -- Hubble-ui backend security context.
	SecurityContext Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_SecurityContext
	//  -- Additional hubble-ui backend environment variables.
	ExtraEnv []any
	//  -- Additional hubble-ui backend volumes.
	ExtraVolumes []any
	//  -- Additional hubble-ui backend volumeMounts.
	ExtraVolumeMounts []any
	LivenessProbe     Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_LivenessProbe
	ReadinessProbe    Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_ReadinessProbe
	//  -- Resource requests and limits for the 'backend' container of the 'hubble-ui' deployment.
	//    limits:
	//      cpu: 1000m
	//      memory: 1024M
	//    requests:
	//      cpu: 100m
	//      memory: 64Mi
	Resources Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend_Resources
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_SecurityContext struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Server_Ipv6 struct {
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Server struct {
	//  -- Controls server listener for ipv6
	Ipv6 Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Server_Ipv6
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend struct {
	//  -- Hubble-ui frontend image.
	Image Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Image
	//  -- Hubble-ui frontend security context.
	SecurityContext Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_SecurityContext
	//  -- Additional hubble-ui frontend environment variables.
	ExtraEnv []any
	//  -- Additional hubble-ui frontend volumes.
	ExtraVolumes []any
	//  -- Additional hubble-ui frontend volumeMounts.
	ExtraVolumeMounts []any
	//  -- Resource requests and limits for the 'frontend' container of the 'hubble-ui' deployment.
	Resources Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Resources
	//    limits:
	//      cpu: 1000m
	//      memory: 1024M
	//    requests:
	//      cpu: 100m
	//      memory: 64Mi
	Server Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend_Server
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_PodAnnotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_PodLabels struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_PodDisruptionBudget struct {
	//  -- enable PodDisruptionBudget
	//  ref: https://kubernetes.io/docs/concepts/workloads/pods/disruptions/
	Enabled any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Minimum number/percentage of pods that should remain scheduled.
	//  When it's set, maxUnavailable must be disabled by `maxUnavailable: null`
	MinAvailable any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Maximum number/percentage of pods that may be made unavailable
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Affinity struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_NodeSelector struct {
	KubernetesIoos any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_UpdateStrategy_RollingUpdate struct {
	//  @schema
	//  type: [integer, string]
	//  @schema
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_UpdateStrategy struct {
	Type          any
	RollingUpdate Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_UpdateStrategy_RollingUpdate
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_SecurityContext struct {
	RunAsUser  any
	RunAsGroup any
	FsGroup    any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Service_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Service struct {
	//  -- Annotations to be added for the Hubble UI service
	Annotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Service_Annotations
	//  --- The type of service used for Hubble UI access, either ClusterIP or NodePort.
	Type any
	//  --- The port to use when the service type is set to NodePort.
	NodePort any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Ingress_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Ingress_HostsItem struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Ingress_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Ingress struct {
	Enabled     any
	Annotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Ingress_Annotations
	//  kubernetes.io/ingress.class: nginx
	//  kubernetes.io/tls-acme: "true"
	ClassName any
	Hosts     []Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Ingress_HostsItem
	Labels    Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Ingress_Labels
	//   - secretName: chart-example-tls
	//     hosts:
	//       - chart-example.local
	Tls []any
}

type Cilium1163Values_Cilium1163Values_Hubble_Ui struct {
	//  -- Whether to enable the Hubble UI.
	Enabled    any
	Standalone Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Standalone
	//  -- Roll out Hubble-ui pods automatically when configmap is updated.
	RollOutPods any
	Tls         Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Tls
	Backend     Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Backend
	Frontend    Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Frontend
	//  -- The number of replicas of Hubble UI to deploy.
	Replicas any
	//  -- Annotations to be added to all top-level hubble-ui objects (resources under templates/hubble-ui)
	Annotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Annotations
	//  -- Annotations to be added to hubble-ui pods
	PodAnnotations Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_PodAnnotations
	//  -- Labels to be added to hubble-ui pods
	PodLabels Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_PodLabels
	//  PodDisruptionBudget settings
	PodDisruptionBudget Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_PodDisruptionBudget
	//  -- Affinity for hubble-ui
	Affinity Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Affinity
	//  -- Pod topology spread constraints for hubble-ui
	//    - maxSkew: 1
	//      topologyKey: topology.kubernetes.io/zone
	//      whenUnsatisfiable: DoNotSchedule
	TopologySpreadConstraints []any
	//  -- Node labels for pod assignment
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
	NodeSelector Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_NodeSelector
	//  -- Node tolerations for pod assignment on nodes with taints
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []any
	//  -- The priority class to use for hubble-ui
	PriorityClassName any
	//  -- hubble-ui update strategy.
	UpdateStrategy Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_UpdateStrategy
	//  -- Security context to be added to Hubble UI pods
	SecurityContext Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_SecurityContext
	//  -- hubble-ui service configuration.
	Service Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Service
	//  -- Defines base url prefix for all hubble-ui http requests.
	//  It needs to be changed in case if ingress for hubble-ui is configured under some sub-path.
	//  Trailing `/` is required for custom path, ex. `/service-map/`
	BaseUrl any
	//  -- hubble-ui ingress configuration.
	Ingress Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Ui_Ingress
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Static struct {
	Enabled   any
	FilePath  any
	FieldMask []any
	//  - time
	//  - source
	//  - destination
	//  - verdict
	AllowList []any
	//  - '{"verdict":["DROPPED","ERROR"]}'
	//  - '{"source_pod":["kube-system/"]}'
	//  - '{"destination_pod":["kube-system/"]}'
	DenyList []any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Dynamic_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Dynamic_Config_ContentItem struct {
	Name           any
	FieldMask      []any
	IncludeFilters []any
	ExcludeFilters []any
	//    - name: "test002"
	//      filePath: "/var/log/network/flow-log/pa/test002.log"
	//      fieldMask: ["source.namespace", "source.pod_name", "destination.namespace", "destination.pod_name", "verdict"]
	//      includeFilters:
	//      - source_pod: ["default/"]
	//        event_type:
	//        - type: 1
	//      - destination_pod: ["frontend/nginx-975996d4c-7hhgt"]
	//      excludeFilters: []
	//      end: "2023-10-09T23:59:59-07:00"
	FilePath any
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Dynamic_Config struct {
	//  ---- Name of configmap with configuration that may be altered to reconfigure exporters within a running agents.
	ConfigMapName any
	//  ---- True if helm installer should create config map.
	//  Switch to false if you want to self maintain the file content.
	CreateConfigMap any
	//  ---- Exporters configuration in YAML format.
	Content []Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Dynamic_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Dynamic_Config_ContentItem
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Dynamic struct {
	Enabled any
	Config  Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Dynamic_Config
}

type Cilium1163Values_Cilium1163Values_Hubble_Export struct {
	//  --- Defines max file size of output file before it gets rotated.
	FileMaxSizeMb any
	//  --- Defines max number of backup/rotated files.
	FileMaxBackups any
	//  --- Static exporter configuration.
	//  Static exporter is bound to agent lifecycle.
	Static Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Static
	//  --- Dynamic exporters configuration.
	//  Dynamic exporters may be reconfigured without a need of agent restarts.
	Dynamic Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_Export_Dynamic
}

type Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_DropEventEmitter_ReasonsItem struct {
}

type Cilium1163Values_Cilium1163Values_Hubble_DropEventEmitter struct {
	Enabled any
	//  --- Minimum time between emitting same events.
	Interval any
	//  --- Drop reasons to emit events for.
	//  ref: https://docs.cilium.io/en/stable/_api/v1/flow/README/#dropreason
	Reasons []Cilium1163Values_Cilium1163Values_Hubble_Cilium1163Values_Cilium1163Values_Hubble_DropEventEmitter_ReasonsItem
}

type Cilium1163Values_Hubble struct {
	//  -- Enable Hubble (true by default).
	Enabled any
	//  -- Annotations to be added to all top-level hubble objects (resources under templates/hubble)
	//  -- Buffer size of the channel Hubble uses to receive monitor events. If this
	//  value is not set, the queue size is set to the default monitor queue size.
	//  eventQueueSize: ""
	Annotations Cilium1163Values_Cilium1163Values_Hubble_Annotations
	//  -- Number of recent flows for Hubble to cache. Defaults to 4095.
	//  Possible values are:
	//    1, 3, 7, 15, 31, 63, 127, 255, 511, 1023,
	//    2047, 4095, 8191, 16383, 32767, 65535
	//  eventBufferCapacity: "4095"
	//
	//  -- Hubble metrics configuration.
	//  See https://docs.cilium.io/en/stable/observability/metrics/#hubble-metrics
	//  for more comprehensive documentation about Hubble metrics.
	Metrics Cilium1163Values_Cilium1163Values_Hubble_Metrics
	//  -- Unix domain socket path to listen to when Hubble is enabled.
	SocketPath any
	//  -- Enables redacting sensitive information present in Layer 7 flows.
	Redact Cilium1163Values_Cilium1163Values_Hubble_Redact
	//  -- An additional address for Hubble to listen to.
	//  Set this field ":4244" if you are enabling Hubble Relay, as it assumes that
	//  Hubble is listening on port 4244.
	ListenAddress any
	//  -- Whether Hubble should prefer to announce IPv6 or IPv4 addresses if both are available.
	PreferIpv6 any
	//  @schema
	//  type: [null, boolean]
	//  @schema
	//  -- (bool) Skip Hubble events with unknown cgroup ids
	//  @default -- `true`
	SkipUnknownCgroupIds any
	PeerService          Cilium1163Values_Cilium1163Values_Hubble_PeerService
	//  -- TLS configuration for Hubble
	Tls   Cilium1163Values_Cilium1163Values_Hubble_Tls
	Relay Cilium1163Values_Cilium1163Values_Hubble_Relay
	Ui    Cilium1163Values_Cilium1163Values_Hubble_Ui
	//  -- Hubble flows export.
	Export Cilium1163Values_Cilium1163Values_Hubble_Export
	//  -- Emit v1.Events related to pods on detection of packet drops.
	//     This feature is alpha, please provide feedback at https://github.com/cilium/cilium/issues/33975.
	DropEventEmitter Cilium1163Values_Cilium1163Values_Hubble_DropEventEmitter
}

type Cilium1163Values_Cilium1163Values_Ipam_Cilium1163Values_Cilium1163Values_Ipam_Operator_ClusterPoolIpv4PodCidrlistItem struct {
}

type Cilium1163Values_Cilium1163Values_Ipam_Cilium1163Values_Cilium1163Values_Ipam_Operator_ClusterPoolIpv6PodCidrlistItem struct {
}

type Cilium1163Values_Cilium1163Values_Ipam_Cilium1163Values_Cilium1163Values_Ipam_Operator_AutoCreateCiliumPodIppools struct {
}

type Cilium1163Values_Cilium1163Values_Ipam_Operator struct {
	//  @schema
	//  type: [array, string]
	//  @schema
	//  -- IPv4 CIDR list range to delegate to individual nodes for IPAM.
	ClusterPoolIpv4PodCidrlist []Cilium1163Values_Cilium1163Values_Ipam_Cilium1163Values_Cilium1163Values_Ipam_Operator_ClusterPoolIpv4PodCidrlistItem
	//  -- IPv4 CIDR mask size to delegate to individual nodes for IPAM.
	ClusterPoolIpv4MaskSize any
	//  @schema
	//  type: [array, string]
	//  @schema
	//  -- IPv6 CIDR list range to delegate to individual nodes for IPAM.
	ClusterPoolIpv6PodCidrlist []Cilium1163Values_Cilium1163Values_Ipam_Cilium1163Values_Cilium1163Values_Ipam_Operator_ClusterPoolIpv6PodCidrlistItem
	//  -- IPv6 CIDR mask size to delegate to individual nodes for IPAM.
	ClusterPoolIpv6MaskSize any
	//  -- IP pools to auto-create in multi-pool IPAM mode.
	AutoCreateCiliumPodIppools Cilium1163Values_Cilium1163Values_Ipam_Cilium1163Values_Cilium1163Values_Ipam_Operator_AutoCreateCiliumPodIppools
	//    default:
	//      ipv4:
	//        cidrs:
	//          - 10.10.0.0/8
	//        maskSize: 24
	//    other:
	//      ipv6:
	//        cidrs:
	//          - fd00:100::/80
	//        maskSize: 96
	//  @schema
	//  type: [null, integer]
	//  @schema
	//  -- (int) The maximum burst size when rate limiting access to external APIs.
	//  Also known as the token bucket capacity.
	//  @default -- `20`
	ExternalApilimitBurstSize any
	//  @schema
	//  type: [null, number]
	//  @schema
	//  -- (float) The maximum queries per second when rate limiting access to
	//  external APIs. Also known as the bucket refill rate, which is used to
	//  refill the bucket up to the burst size capacity.
	//  @default -- `4.0`
	ExternalApilimitQps any
}

type Cilium1163Values_Ipam struct {
	//  -- Configure IP Address Management mode.
	//  ref: https://docs.cilium.io/en/stable/network/concepts/ipam/
	Mode any
	//  -- Maximum rate at which the CiliumNode custom resource is updated.
	CiliumNodeUpdateRate any
	Operator             Cilium1163Values_Cilium1163Values_Ipam_Operator
}

type Cilium1163Values_NodeIpam struct {
	//  -- Configure Node IPAM
	//  ref: https://docs.cilium.io/en/stable/network/node-ipam/
	Enabled any
}

type Cilium1163Values_IpMasqAgent struct {
	Enabled any
}

type Cilium1163Values_Ipv4 struct {
	//  -- Enable IPv4 support.
	Enabled any
}

type Cilium1163Values_Ipv6 struct {
	//  -- Enable IPv6 support.
	Enabled any
}

type Cilium1163Values_K8S struct {
	//  -- requireIPv4PodCIDR enables waiting for Kubernetes to provide the PodCIDR
	//  range via the Kubernetes node resource
	RequireIpv4PodCidr any
	//  -- requireIPv6PodCIDR enables waiting for Kubernetes to provide the PodCIDR
	//  range via the Kubernetes node resource
	RequireIpv6PodCidr any
}

type Cilium1163Values_StartupProbe struct {
	//  -- failure threshold of startup probe.
	//  105 x 2s translates to the old behaviour of the readiness probe (120s delay + 30 x 3s)
	FailureThreshold any
	//  -- interval between checks of the startup probe
	PeriodSeconds any
}

type Cilium1163Values_LivenessProbe struct {
	//  -- failure threshold of liveness probe
	FailureThreshold any
	//  -- interval between checks of the liveness probe
	PeriodSeconds any
}

type Cilium1163Values_ReadinessProbe struct {
	//  -- failure threshold of readiness probe
	FailureThreshold any
	//  -- interval between checks of the readiness probe
	PeriodSeconds any
}

type Cilium1163Values_L2NeighDiscovery struct {
	//  -- Enable L2 neighbor discovery in the agent
	Enabled any
	//  -- Override the agent's default neighbor resolution refresh period.
	RefreshPeriod any
}

type Cilium1163Values_Maglev struct {
}

type Cilium1163Values_Nat struct {
	//  -- Number of the top-k SNAT map connections to track in Cilium statedb.
	MapStatsEntries any
	//  -- Interval between how often SNAT map is counted for stats.
	MapStatsInterval any
}

type Cilium1163Values_EgressGateway struct {
	//  -- Enables egress gateway to redirect and SNAT the traffic that leaves the
	//  cluster.
	Enabled any
	//  -- Time between triggers of egress gateway state reconciliations
	//  -- Maximum number of entries in egress gateway policy map
	//  maxPolicyEntries: 16384
	ReconciliationTriggerInterval any
}

type Cilium1163Values_Vtep struct {
	//  -- Enables VXLAN Tunnel Endpoint (VTEP) Integration (beta) to allow
	//  Cilium-managed pods to talk to third party VTEP devices over Cilium tunnel.
	Enabled any
	//  -- A space separated list of VTEP device endpoint IPs, for example "1.1.1.1  1.1.2.1"
	Endpoint any
	//  -- A space separated list of VTEP device CIDRs, for example "1.1.1.0/24 1.1.2.0/24"
	Cidr any
	//  -- VTEP CIDRs Mask that applies to all VTEP CIDRs, for example "255.255.255.0"
	Mask any
	//  -- A space separated list of VTEP device MAC addresses (VTEP MAC), for example "x:x:x:x:x:x  y:y:y:y:y:y:y"
	Mac any
}

type Cilium1163Values_Monitor struct {
	//  -- Enable the cilium-monitor sidecar.
	Enabled any
}

type Cilium1163Values_Cilium1163Values_LoadBalancer_L7 struct {
	//  -- Enable L7 service load balancing via envoy proxy.
	//  The request to a k8s service, which has specific annotation e.g. service.cilium.io/lb-l7,
	//  will be forwarded to the local backend proxy to be load balanced to the service endpoints.
	//  Please refer to docs for supported annotations for more configuration.
	//
	//  Applicable values:
	//    - envoy: Enable L7 load balancing via envoy proxy. This will automatically set enable-envoy-config as well.
	//    - disabled: Disable L7 load balancing by way of service annotation.
	Backend any
	//  -- List of ports from service to be automatically redirected to above backend.
	//  Any service exposing one of these ports will be automatically redirected.
	//  Fine-grained control can be achieved by using the service annotation.
	Ports []any
	//  -- Default LB algorithm
	//  The default LB algorithm to be used for services, which can be overridden by the
	//  service annotation (e.g. service.cilium.io/lb-l7-algorithm)
	//  Applicable values: round_robin, least_request, random
	Algorithm any
}

type Cilium1163Values_LoadBalancer struct {
	//  -- standalone enables the standalone L4LB which does not connect to
	//  kube-apiserver.
	//  standalone: false
	//
	//  -- algorithm is the name of the load balancing algorithm for backend
	//  selection e.g. random or maglev
	//  algorithm: random
	//
	//  -- mode is the operation mode of load balancing for remote backends
	//  e.g. snat, dsr, hybrid
	//  mode: snat
	//
	//  -- acceleration is the option to accelerate service handling via XDP
	//  Applicable values can be: disabled (do not use XDP), native (XDP BPF
	//  program is run directly out of the networking driver's early receive
	//  path), or best-effort (use native mode XDP acceleration on devices
	//  that support it).
	//  -- dsrDispatch configures whether IP option or IPIP encapsulation is
	//  used to pass a service IP and port to remote backend
	//  dsrDispatch: opt
	Acceleration any
	//  -- serviceTopology enables K8s Topology Aware Hints -based service
	//  endpoints filtering
	//  serviceTopology: false
	//
	//  -- L7 LoadBalancer
	L7 Cilium1163Values_Cilium1163Values_LoadBalancer_L7
}

type Cilium1163Values_NodePort struct {
	//  -- Enable the Cilium NodePort service implementation.
	//  -- Port range to use for NodePort services.
	//  range: "30000,32767"
	Enabled any
	//  @schema
	//  type: [null, string, array]
	//  @schema
	//  -- List of CIDRs for choosing which IP addresses assigned to native devices are used for NodePort load-balancing.
	//  By default this is empty and the first suitable, preferably private, IPv4 and IPv6 address assigned to each device is used.
	//
	//  Example:
	//
	//    addresses: ["192.168.1.0/24", "2001::/64"]
	//
	Addresses any
	//  -- Set to true to prevent applications binding to service ports.
	BindProtection any
	//  -- Append NodePort range to ip_local_reserved_ports if clash with ephemeral
	//  ports is detected.
	AutoProtectPortRange any
	//  -- Enable healthcheck nodePort server for NodePort services
	EnableHealthCheck any
	//  -- Enable access of the healthcheck nodePort on the LoadBalancerIP. Needs
	//  EnableHealthCheck to be enabled
	EnableHealthCheckLoadBalancerIp any
}

type Cilium1163Values_Pprof struct {
	//  -- Enable pprof for cilium-agent
	Enabled any
	//  -- Configure pprof listen address for cilium-agent
	Address any
	//  -- Configure pprof listen port for cilium-agent
	Port any
}

type Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_RelabelingsItem_SourceLabelsItem struct {
}

type Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_RelabelingsItem struct {
	SourceLabels []Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_RelabelingsItem_SourceLabelsItem
	TargetLabel  any
	Replacement  any
}

type Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor struct {
	//  -- Enable service monitors.
	//  This requires the prometheus CRDs to be available (see https://github.com/prometheus-operator/prometheus-operator/blob/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml)
	Enabled any
	//  -- Labels to add to ServiceMonitor cilium-agent
	Labels Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_Labels
	//  -- Annotations to add to ServiceMonitor cilium-agent
	Annotations Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_Annotations
	//  -- jobLabel to add for ServiceMonitor cilium-agent
	JobLabel any
	//  -- Interval for scrape metrics.
	Interval any
	//  -- Specify the Kubernetes namespace where Prometheus expects to find
	//  service monitors configured.
	//  namespace: ""
	//  -- Relabeling configs for the ServiceMonitor cilium-agent
	Relabelings []Cilium1163Values_Cilium1163Values_Prometheus_Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor_RelabelingsItem
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Metrics relabeling configs for the ServiceMonitor cilium-agent
	MetricRelabelings any
	//  -- Set to `true` and helm will not check for monitoring.coreos.com/v1 CRDs before deploying
	TrustCrdsExist any
}

type Cilium1163Values_Cilium1163Values_Prometheus_ControllerGroupMetricsItem struct {
}

type Cilium1163Values_Prometheus struct {
	Enabled        any
	Port           any
	ServiceMonitor Cilium1163Values_Cilium1163Values_Prometheus_ServiceMonitor
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Metrics that should be enabled or disabled from the default metric list.
	//  The list is expected to be separated by a space. (+metric_foo to enable
	//  metric_foo , -metric_bar to disable metric_bar).
	//  ref: https://docs.cilium.io/en/stable/observability/metrics/
	Metrics any
	//  --- Enable controller group metrics for monitoring specific Cilium
	//  subsystems. The list is a list of controller group names. The special
	//  values of "all" and "none" are supported. The set of controller
	//  group names is not guaranteed to be stable between Cilium versions.
	ControllerGroupMetrics []Cilium1163Values_Cilium1163Values_Prometheus_ControllerGroupMetricsItem
}

type Cilium1163Values_Cilium1163Values_Dashboards_Annotations struct {
}

type Cilium1163Values_Dashboards struct {
	Enabled any
	Label   any
	//  @schema
	//  type: [null, string]
	//  @schema
	Namespace   any
	LabelValue  any
	Annotations Cilium1163Values_Cilium1163Values_Dashboards_Annotations
}

type Cilium1163Values_Cilium1163Values_Envoy_Log struct {
	//  -- The format string to use for laying out the log message metadata of Envoy.
	Format any
	//  -- Path to a separate Envoy log file, if any. Defaults to /dev/stdout.
	Path any
}

type Cilium1163Values_Cilium1163Values_Envoy_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	PullPolicy any
	Digest     any
	UseDigest  any
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_UpdateStrategy_RollingUpdate struct {
	//  @schema
	//  type: [integer, string]
	//  @schema
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Envoy_UpdateStrategy struct {
	Type          any
	RollingUpdate Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_UpdateStrategy_RollingUpdate
}

type Cilium1163Values_Cilium1163Values_Envoy_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_PodSecurityContext_AppArmorProfile struct {
	Type any
}

type Cilium1163Values_Cilium1163Values_Envoy_PodSecurityContext struct {
	//  -- AppArmorProfile options for the `cilium-agent` and init containers
	AppArmorProfile Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_PodSecurityContext_AppArmorProfile
}

type Cilium1163Values_Cilium1163Values_Envoy_PodAnnotations struct {
}

type Cilium1163Values_Cilium1163Values_Envoy_PodLabels struct {
}

type Cilium1163Values_Cilium1163Values_Envoy_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Envoy_StartupProbe struct {
	//  -- failure threshold of startup probe.
	//  105 x 2s translates to the old behaviour of the readiness probe (120s delay + 30 x 3s)
	FailureThreshold any
	//  -- interval between checks of the startup probe
	PeriodSeconds any
}

type Cilium1163Values_Cilium1163Values_Envoy_LivenessProbe struct {
	//  -- failure threshold of liveness probe
	FailureThreshold any
	//  -- interval between checks of the liveness probe
	PeriodSeconds any
}

type Cilium1163Values_Cilium1163Values_Envoy_ReadinessProbe struct {
	//  -- failure threshold of readiness probe
	FailureThreshold any
	//  -- interval between checks of the readiness probe
	PeriodSeconds any
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_SecurityContext_SeLinuxOptions struct {
	Level any
	//  Running with spc_t since we have removed the privileged mode.
	//  Users can change it to a different type as long as they have the
	//  type available on the system.
	Type any
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_SecurityContext_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_SecurityContext_Capabilities_EnvoyItem struct {
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_SecurityContext_Capabilities struct {
	//  -- Capabilities for the `cilium-envoy` container.
	//  Even though granted to the container, the cilium-envoy-starter wrapper drops
	//  all capabilities after forking the actual Envoy process.
	//  `NET_BIND_SERVICE` is the only capability that can be passed to the Envoy process by
	//  setting `envoy.securityContext.capabilities.keepNetBindService=true` (in addition to granting the
	//  capability to the container).
	//  Note: In case of embedded envoy, the capability must  be granted to the cilium-agent container.
	Envoy []Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_SecurityContext_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_SecurityContext_Capabilities_EnvoyItem
	//  -- Keep capability `NET_BIND_SERVICE` for Envoy process.
	KeepCapNetBindService any
}

type Cilium1163Values_Cilium1163Values_Envoy_SecurityContext struct {
	//  -- User to run the pod with
	//  runAsUser: 0
	//  -- Run the pod with elevated privileges
	Privileged any
	//  -- SELinux options for the `cilium-envoy` container
	SeLinuxOptions Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_SecurityContext_SeLinuxOptions
	Capabilities   Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_SecurityContext_Capabilities
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels struct {
	K8SApp any
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector struct {
	MatchLabels Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem struct {
	TopologyKey   any
	LabelSelector Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution []Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels struct {
	K8SApp any
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector struct {
	MatchLabels Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem struct {
	TopologyKey   any
	LabelSelector Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution []Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_NodeSelectorTermsItem_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_NodeSelectorTermsItem_MatchExpressionsItem_ValuesItem struct {
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_NodeSelectorTermsItem_MatchExpressionsItem struct {
	Key      any
	Operator any
	Values   []Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_NodeSelectorTermsItem_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_NodeSelectorTermsItem_MatchExpressionsItem_ValuesItem
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_NodeSelectorTermsItem struct {
	MatchExpressions []Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_NodeSelectorTermsItem_MatchExpressionsItem
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution struct {
	NodeSelectorTerms []Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution_NodeSelectorTermsItem
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity_RequiredDuringSchedulingIgnoredDuringExecution
}

type Cilium1163Values_Cilium1163Values_Envoy_Affinity struct {
	PodAntiAffinity Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAntiAffinity
	PodAffinity     Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_PodAffinity
	NodeAffinity    Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Affinity_NodeAffinity
}

type Cilium1163Values_Cilium1163Values_Envoy_NodeSelector struct {
	KubernetesIoos any
}

type Cilium1163Values_Cilium1163Values_Envoy_TolerationsItem struct {
	//  - key: "key"
	//    operator: "Equal|Exists"
	//    value: "value"
	//    effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"
	Operator any
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Debug_Admin struct {
	//  -- Enable admin interface for cilium-envoy.
	//  This is useful for debugging and should not be enabled in production.
	Enabled any
	//  -- Port number (bound to loopback interface).
	//  kubectl port-forward can be used to access the admin interface.
	Port any
}

type Cilium1163Values_Cilium1163Values_Envoy_Debug struct {
	Admin Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Debug_Admin
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_RelabelingsItem_SourceLabelsItem struct {
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_RelabelingsItem struct {
	SourceLabels []Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_RelabelingsItem_SourceLabelsItem
	TargetLabel  any
	Replacement  any
}

type Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor struct {
	//  -- Enable service monitors.
	//  This requires the prometheus CRDs to be available (see https://github.com/prometheus-operator/prometheus-operator/blob/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml)
	//  Note that this setting applies to both cilium-envoy _and_ cilium-agent
	//  with Envoy enabled.
	Enabled any
	//  -- Labels to add to ServiceMonitor cilium-envoy
	Labels Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_Labels
	//  -- Annotations to add to ServiceMonitor cilium-envoy
	Annotations Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_Annotations
	//  -- Interval for scrape metrics.
	Interval any
	//  -- Specify the Kubernetes namespace where Prometheus expects to find
	//  service monitors configured.
	//  namespace: ""
	//  -- Relabeling configs for the ServiceMonitor cilium-envoy
	//  or for cilium-agent with Envoy configured.
	Relabelings []Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor_RelabelingsItem
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Metrics relabeling configs for the ServiceMonitor cilium-envoy
	//  or for cilium-agent with Envoy configured.
	MetricRelabelings any
}

type Cilium1163Values_Cilium1163Values_Envoy_Prometheus struct {
	//  -- Enable prometheus metrics for cilium-envoy
	Enabled        any
	ServiceMonitor Cilium1163Values_Cilium1163Values_Envoy_Cilium1163Values_Cilium1163Values_Envoy_Prometheus_ServiceMonitor
	//  -- Serve prometheus metrics for cilium-envoy on the configured port
	Port any
}

type Cilium1163Values_Envoy struct {
	//  @schema
	//  type: [null, boolean]
	//  @schema
	//  -- Enable Envoy Proxy in standalone DaemonSet.
	//  This field is enabled by default for new installation.
	//  @default -- `true` for new installation
	Enabled any
	//  -- (int)
	//  Set Envoy'--base-id' to use when allocating shared memory regions.
	//  Only needs to be changed if multiple Envoy instances will run on the same node and may have conflicts. Supported values: 0 - 4294967295. Defaults to '0'
	BaseId any
	Log    Cilium1163Values_Cilium1163Values_Envoy_Log
	//  -- Time in seconds after which a TCP connection attempt times out
	ConnectTimeoutSeconds any
	//  -- ProxyMaxRequestsPerConnection specifies the max_requests_per_connection setting for Envoy
	MaxRequestsPerConnection any
	//  -- Set Envoy HTTP option max_connection_duration seconds. Default 0 (disable)
	MaxConnectionDurationSeconds any
	//  -- Set Envoy upstream HTTP idle connection timeout seconds.
	//  Does not apply to connections with pending requests. Default 60s
	IdleTimeoutDurationSeconds any
	//  -- Number of trusted hops regarding the x-forwarded-for and related HTTP headers for the ingress L7 policy enforcement Envoy listeners.
	XffNumTrustedHopsL7PolicyIngress any
	//  -- Number of trusted hops regarding the x-forwarded-for and related HTTP headers for the egress L7 policy enforcement Envoy listeners.
	XffNumTrustedHopsL7PolicyEgress any
	//  -- Envoy container image.
	Image Cilium1163Values_Cilium1163Values_Envoy_Image
	//  -- Additional containers added to the cilium Envoy DaemonSet.
	ExtraContainers []any
	//  -- Additional envoy container arguments.
	ExtraArgs []any
	//  -- Additional envoy container environment variables.
	ExtraEnv []any
	//  -- Additional envoy hostPath mounts.
	//  - name: host-mnt-data
	//    mountPath: /host/mnt/data
	//    hostPath: /mnt/data
	//    hostPathType: Directory
	//    readOnly: true
	//    mountPropagation: HostToContainer
	ExtraHostPathMounts []any
	//  -- Additional envoy volumes.
	ExtraVolumes []any
	//  -- Additional envoy volumeMounts.
	ExtraVolumeMounts []any
	//  -- Configure termination grace period for cilium-envoy DaemonSet.
	TerminationGracePeriodSeconds any
	//  -- TCP port for the health API.
	HealthPort any
	//  -- cilium-envoy update strategy
	//  ref: https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/#updating-a-daemonset
	UpdateStrategy Cilium1163Values_Cilium1163Values_Envoy_UpdateStrategy
	//  -- Roll out cilium envoy pods automatically when configmap is updated.
	RollOutPods any
	//  -- Annotations to be added to all top-level cilium-envoy objects (resources under templates/cilium-envoy)
	Annotations Cilium1163Values_Cilium1163Values_Envoy_Annotations
	//  -- Security Context for cilium-envoy pods.
	PodSecurityContext Cilium1163Values_Cilium1163Values_Envoy_PodSecurityContext
	//  -- Annotations to be added to envoy pods
	PodAnnotations Cilium1163Values_Cilium1163Values_Envoy_PodAnnotations
	//  -- Labels to be added to envoy pods
	PodLabels Cilium1163Values_Cilium1163Values_Envoy_PodLabels
	//  -- Envoy resource limits & requests
	//  ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	//    limits:
	//      cpu: 4000m
	//      memory: 4Gi
	//    requests:
	//      cpu: 100m
	//      memory: 512Mi
	Resources       Cilium1163Values_Cilium1163Values_Envoy_Resources
	StartupProbe    Cilium1163Values_Cilium1163Values_Envoy_StartupProbe
	LivenessProbe   Cilium1163Values_Cilium1163Values_Envoy_LivenessProbe
	ReadinessProbe  Cilium1163Values_Cilium1163Values_Envoy_ReadinessProbe
	SecurityContext Cilium1163Values_Cilium1163Values_Envoy_SecurityContext
	//  -- Affinity for cilium-envoy.
	Affinity Cilium1163Values_Cilium1163Values_Envoy_Affinity
	//  -- Node selector for cilium-envoy.
	NodeSelector Cilium1163Values_Cilium1163Values_Envoy_NodeSelector
	//  -- Node tolerations for envoy scheduling to nodes with taints
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []Cilium1163Values_Cilium1163Values_Envoy_TolerationsItem
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- The priority class to use for cilium-envoy.
	PriorityClassName any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- DNS policy for Cilium envoy pods.
	//  Ref: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
	DnsPolicy any
	Debug     Cilium1163Values_Cilium1163Values_Envoy_Debug
	//  -- Configure Cilium Envoy Prometheus options.
	//  Note that some of these apply to either cilium-agent or cilium-envoy.
	Prometheus Cilium1163Values_Cilium1163Values_Envoy_Prometheus
}

type Cilium1163Values_Cilium1163Values_ResourceQuotas_Cilium1163Values_Cilium1163Values_ResourceQuotas_Cilium_Hard struct {
	//  5k nodes * 2 DaemonSets (Cilium and cilium node init)
	Pods any
}

type Cilium1163Values_Cilium1163Values_ResourceQuotas_Cilium struct {
	Hard Cilium1163Values_Cilium1163Values_ResourceQuotas_Cilium1163Values_Cilium1163Values_ResourceQuotas_Cilium_Hard
}

type Cilium1163Values_Cilium1163Values_ResourceQuotas_Cilium1163Values_Cilium1163Values_ResourceQuotas_Operator_Hard struct {
	//  15 "clusterwide" Cilium Operator pods for HA
	Pods any
}

type Cilium1163Values_Cilium1163Values_ResourceQuotas_Operator struct {
	Hard Cilium1163Values_Cilium1163Values_ResourceQuotas_Cilium1163Values_Cilium1163Values_ResourceQuotas_Operator_Hard
}

type Cilium1163Values_ResourceQuotas struct {
	Enabled  any
	Cilium   Cilium1163Values_Cilium1163Values_ResourceQuotas_Cilium
	Operator Cilium1163Values_Cilium1163Values_ResourceQuotas_Operator
}

type Cilium1163Values_Cilium1163Values_Tls_Ca struct {
	//  -- Optional CA cert. If it is provided, it will be used by cilium to
	//  generate all other certificates. Otherwise, an ephemeral CA is generated.
	Cert any
	//  -- Optional CA private key. If it is provided, it will be used by cilium to
	//  generate all other certificates. Otherwise, an ephemeral CA is generated.
	Key any
	//  -- Generated certificates validity duration in days. This will be used for auto generated CA.
	CertValidityDuration any
}

type Cilium1163Values_Cilium1163Values_Tls_CaBundle struct {
	//  -- Enable the use of the CA trust bundle.
	Enabled any
	//  -- Name of the ConfigMap containing the CA trust bundle.
	Name any
	//  -- Entry of the ConfigMap containing the CA trust bundle.
	Key any
	//  -- Use a Secret instead of a ConfigMap.
	//  If uncommented, creates the ConfigMap and fills it with the specified content.
	//  Otherwise, the ConfigMap is assumed to be already present in .Release.Namespace.
	//
	//  content: |
	//    -----BEGIN CERTIFICATE-----
	//    ...
	//    -----END CERTIFICATE-----
	//    -----BEGIN CERTIFICATE-----
	//    ...
	//    -----END CERTIFICATE-----
	UseSecret any
}

type Cilium1163Values_Tls struct {
	//  -- This configures how the Cilium agent loads the secrets used TLS-aware CiliumNetworkPolicies
	//  (namely the secrets referenced by terminatingTLS and originatingTLS).
	//  Possible values:
	//    - local
	//    - k8s
	SecretsBackend any
	//  -- Base64 encoded PEM values for the CA certificate and private key.
	//  This can be used as common CA to generate certificates used by hubble and clustermesh components.
	//  It is neither required nor used when cert-manager is used to generate the certificates.
	Ca Cilium1163Values_Cilium1163Values_Tls_Ca
	//  -- Configure the CA trust bundle used for the validation of the certificates
	//  leveraged by hubble and clustermesh. When enabled, it overrides the content of the
	//  'ca.crt' field of the respective certificates, allowing for CA rotation with no down-time.
	CaBundle Cilium1163Values_Cilium1163Values_Tls_CaBundle
}

type Cilium1163Values_WellKnownIdentities struct {
	//  -- Enable the use of well-known identities.
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Etcd_EndpointsItem struct {
}

type Cilium1163Values_Etcd struct {
	//  -- Enable etcd mode for the agent.
	Enabled any
	//  -- List of etcd endpoints
	Endpoints []Cilium1163Values_Cilium1163Values_Etcd_EndpointsItem
	//  -- Enable use of TLS/SSL for connectivity to etcd.
	Ssl any
}

type Cilium1163Values_Cilium1163Values_Operator_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	//  operator-generic-digest
	GenericDigest any
	//  operator-azure-digest
	AzureDigest any
	//  operator-aws-digest
	AwsDigest any
	//  operator-alibabacloud-digest
	AlibabacloudDigest any
	UseDigest          any
	PullPolicy         any
	Suffix             any
}

type Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_UpdateStrategy_RollingUpdate struct {
	//  @schema
	//  type: [integer, string]
	//  @schema
	MaxSurge any
	//  @schema
	//  type: [integer, string]
	//  @schema
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Operator_UpdateStrategy struct {
	Type          any
	RollingUpdate Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_UpdateStrategy_RollingUpdate
}

type Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels struct {
	IoCiliumapp any
}

type Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector struct {
	MatchLabels Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels
}

type Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem struct {
	TopologyKey   any
	LabelSelector Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector
}

type Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution []Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem
}

type Cilium1163Values_Cilium1163Values_Operator_Affinity struct {
	PodAntiAffinity Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Affinity_PodAntiAffinity
}

type Cilium1163Values_Cilium1163Values_Operator_NodeSelector struct {
	KubernetesIoos any
}

type Cilium1163Values_Cilium1163Values_Operator_TolerationsItem struct {
	//  - key: "key"
	//    operator: "Equal|Exists"
	//    value: "value"
	//    effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"
	Operator any
}

type Cilium1163Values_Cilium1163Values_Operator_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Operator_PodSecurityContext struct {
}

type Cilium1163Values_Cilium1163Values_Operator_PodAnnotations struct {
}

type Cilium1163Values_Cilium1163Values_Operator_PodLabels struct {
}

type Cilium1163Values_Cilium1163Values_Operator_PodDisruptionBudget struct {
	//  -- enable PodDisruptionBudget
	//  ref: https://kubernetes.io/docs/concepts/workloads/pods/disruptions/
	Enabled any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Minimum number/percentage of pods that should remain scheduled.
	//  When it's set, maxUnavailable must be disabled by `maxUnavailable: null`
	MinAvailable any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Maximum number/percentage of pods that may be made unavailable
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Operator_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Operator_SecurityContext struct {
}

type Cilium1163Values_Cilium1163Values_Operator_Pprof struct {
	//  -- Enable pprof for cilium-operator
	Enabled any
	//  -- Configure pprof listen address for cilium-operator
	Address any
	//  -- Configure pprof listen port for cilium-operator
	Port any
}

type Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_ServiceMonitor_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_ServiceMonitor_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_ServiceMonitor struct {
	//  -- Enable service monitors.
	//  This requires the prometheus CRDs to be available (see https://github.com/prometheus-operator/prometheus-operator/blob/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml)
	Enabled any
	//  -- Labels to add to ServiceMonitor cilium-operator
	Labels Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_ServiceMonitor_Labels
	//  -- Annotations to add to ServiceMonitor cilium-operator
	Annotations Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_ServiceMonitor_Annotations
	//  -- jobLabel to add for ServiceMonitor cilium-operator
	JobLabel any
	//  -- Interval for scrape metrics.
	Interval any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Relabeling configs for the ServiceMonitor cilium-operator
	Relabelings any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Metrics relabeling configs for the ServiceMonitor cilium-operator
	MetricRelabelings any
}

type Cilium1163Values_Cilium1163Values_Operator_Prometheus struct {
	Enabled        any
	Port           any
	ServiceMonitor Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Prometheus_ServiceMonitor
}

type Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Dashboards_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Operator_Dashboards struct {
	Enabled any
	Label   any
	//  @schema
	//  type: [null, string]
	//  @schema
	Namespace   any
	LabelValue  any
	Annotations Cilium1163Values_Cilium1163Values_Operator_Cilium1163Values_Cilium1163Values_Operator_Dashboards_Annotations
}

type Cilium1163Values_Cilium1163Values_Operator_UnmanagedPodWatcher struct {
	//  -- Restart any pod that are not managed by Cilium.
	Restart any
	//  -- Interval, in seconds, to check if there are any pods that are not
	//  managed by Cilium.
	IntervalSeconds any
}

type Cilium1163Values_Operator struct {
	//  -- Enable the cilium-operator component (required).
	Enabled any
	//  -- Roll out cilium-operator pods automatically when configmap is updated.
	RollOutPods any
	//  -- cilium-operator image.
	Image Cilium1163Values_Cilium1163Values_Operator_Image
	//  -- Number of replicas to run for the cilium-operator deployment
	Replicas any
	//  -- The priority class to use for cilium-operator
	PriorityClassName any
	//  -- DNS policy for Cilium operator pods.
	//  Ref: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
	DnsPolicy any
	//  -- cilium-operator update strategy
	UpdateStrategy Cilium1163Values_Cilium1163Values_Operator_UpdateStrategy
	//  -- Affinity for cilium-operator
	Affinity Cilium1163Values_Cilium1163Values_Operator_Affinity
	//  -- Pod topology spread constraints for cilium-operator
	//    - maxSkew: 1
	//      topologyKey: topology.kubernetes.io/zone
	//      whenUnsatisfiable: DoNotSchedule
	TopologySpreadConstraints []any
	//  -- Node labels for cilium-operator pod assignment
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
	NodeSelector Cilium1163Values_Cilium1163Values_Operator_NodeSelector
	//  -- Node tolerations for cilium-operator scheduling to nodes with taints
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []Cilium1163Values_Cilium1163Values_Operator_TolerationsItem
	//  -- Additional cilium-operator container arguments.
	ExtraArgs []any
	//  -- Additional cilium-operator environment variables.
	ExtraEnv []any
	//  -- Additional cilium-operator hostPath mounts.
	//  - name: host-mnt-data
	//    mountPath: /host/mnt/data
	//    hostPath: /mnt/data
	//    hostPathType: Directory
	//    readOnly: true
	//    mountPropagation: HostToContainer
	ExtraHostPathMounts []any
	//  -- Additional cilium-operator volumes.
	ExtraVolumes []any
	//  -- Additional cilium-operator volumeMounts.
	ExtraVolumeMounts []any
	//  -- Annotations to be added to all top-level cilium-operator objects (resources under templates/cilium-operator)
	Annotations Cilium1163Values_Cilium1163Values_Operator_Annotations
	//  -- HostNetwork setting
	HostNetwork any
	//  -- Security context to be added to cilium-operator pods
	PodSecurityContext Cilium1163Values_Cilium1163Values_Operator_PodSecurityContext
	//  -- Annotations to be added to cilium-operator pods
	PodAnnotations Cilium1163Values_Cilium1163Values_Operator_PodAnnotations
	//  -- Labels to be added to cilium-operator pods
	PodLabels Cilium1163Values_Cilium1163Values_Operator_PodLabels
	//  PodDisruptionBudget settings
	PodDisruptionBudget Cilium1163Values_Cilium1163Values_Operator_PodDisruptionBudget
	//  -- cilium-operator resource limits & requests
	//  ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	//    limits:
	//      cpu: 1000m
	//      memory: 1Gi
	//    requests:
	//      cpu: 100m
	//      memory: 128Mi
	Resources Cilium1163Values_Cilium1163Values_Operator_Resources
	//  -- Security context to be added to cilium-operator pods
	//  runAsUser: 0
	SecurityContext Cilium1163Values_Cilium1163Values_Operator_SecurityContext
	//  -- Interval for endpoint garbage collection.
	EndpointGcinterval any
	//  -- Interval for cilium node garbage collection.
	NodeGcinterval any
	//  -- Interval for identity garbage collection.
	IdentityGcinterval any
	//  -- Timeout for identity heartbeats.
	IdentityHeartbeatTimeout any
	Pprof                    Cilium1163Values_Cilium1163Values_Operator_Pprof
	//  -- Enable prometheus metrics for cilium-operator on the configured port at
	//  /metrics
	Prometheus Cilium1163Values_Cilium1163Values_Operator_Prometheus
	//  -- Grafana dashboards for cilium-operator
	//  grafana can import dashboards based on the label and value
	//  ref: https://github.com/grafana/helm-charts/tree/main/charts/grafana#sidecar-for-dashboards
	Dashboards Cilium1163Values_Cilium1163Values_Operator_Dashboards
	//  -- Skip CRDs creation for cilium-operator
	SkipCrdcreation any
	//  -- Remove Cilium node taint from Kubernetes nodes that have a healthy Cilium
	//  pod running.
	RemoveNodeTaints any
	//  @schema
	//  type: [null, boolean]
	//  @schema
	//  -- Taint nodes where Cilium is scheduled but not running. This prevents pods
	//  from being scheduled to nodes where Cilium is not the default CNI provider.
	//  @default -- same as removeNodeTaints
	SetNodeTaints any
	//  -- Set Node condition NetworkUnavailable to 'false' with the reason
	//  'CiliumIsUp' for nodes that have a healthy Cilium pod.
	SetNodeNetworkStatus any
	UnmanagedPodWatcher  Cilium1163Values_Cilium1163Values_Operator_UnmanagedPodWatcher
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Nodeinit_UpdateStrategy struct {
	Type any
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Affinity struct {
}

type Cilium1163Values_Cilium1163Values_Nodeinit_NodeSelector struct {
	KubernetesIoos any
}

type Cilium1163Values_Cilium1163Values_Nodeinit_TolerationsItem struct {
	//  - key: "key"
	//    operator: "Equal|Exists"
	//    value: "value"
	//    effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"
	Operator any
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Nodeinit_PodAnnotations struct {
}

type Cilium1163Values_Cilium1163Values_Nodeinit_PodLabels struct {
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_PodSecurityContext_AppArmorProfile struct {
	Type any
}

type Cilium1163Values_Cilium1163Values_Nodeinit_PodSecurityContext struct {
	//  -- AppArmorProfile options for the `cilium-node-init` and init containers
	AppArmorProfile Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_PodSecurityContext_AppArmorProfile
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_Resources_Requests struct {
	Cpu    any
	Memory any
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Resources struct {
	Requests Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_Resources_Requests
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext_SeLinuxOptions struct {
	Level any
	//  Running with spc_t since we have removed the privileged mode.
	//  Users can change it to a different type as long as they have the
	//  type available on the system.
	Type any
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext_Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext_Capabilities_AddItem struct {
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext_Capabilities struct {
	Add []Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext_Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext_Capabilities_AddItem
}

type Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext struct {
	Privileged     any
	SeLinuxOptions Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext_SeLinuxOptions
	Capabilities   Cilium1163Values_Cilium1163Values_Nodeinit_Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext_Capabilities
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Startup struct {
	PreScript  any
	PostScript any
}

type Cilium1163Values_Cilium1163Values_Nodeinit_Prestop struct {
	PreScript  any
	PostScript any
}

type Cilium1163Values_Nodeinit struct {
	//  -- Enable the node initialization DaemonSet
	Enabled any
	//  -- node-init image.
	Image Cilium1163Values_Cilium1163Values_Nodeinit_Image
	//  -- The priority class to use for the nodeinit pod.
	PriorityClassName any
	//  -- node-init update strategy
	UpdateStrategy Cilium1163Values_Cilium1163Values_Nodeinit_UpdateStrategy
	//  -- Additional nodeinit environment variables.
	ExtraEnv []any
	//  -- Additional nodeinit volumes.
	ExtraVolumes []any
	//  -- Additional nodeinit volumeMounts.
	ExtraVolumeMounts []any
	//  -- Affinity for cilium-nodeinit
	Affinity Cilium1163Values_Cilium1163Values_Nodeinit_Affinity
	//  -- Node labels for nodeinit pod assignment
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
	NodeSelector Cilium1163Values_Cilium1163Values_Nodeinit_NodeSelector
	//  -- Node tolerations for nodeinit scheduling to nodes with taints
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []Cilium1163Values_Cilium1163Values_Nodeinit_TolerationsItem
	//  -- Annotations to be added to all top-level nodeinit objects (resources under templates/cilium-nodeinit)
	Annotations Cilium1163Values_Cilium1163Values_Nodeinit_Annotations
	//  -- Annotations to be added to node-init pods.
	PodAnnotations Cilium1163Values_Cilium1163Values_Nodeinit_PodAnnotations
	//  -- Labels to be added to node-init pods.
	PodLabels Cilium1163Values_Cilium1163Values_Nodeinit_PodLabels
	//  -- Security Context for cilium-node-init pods.
	PodSecurityContext Cilium1163Values_Cilium1163Values_Nodeinit_PodSecurityContext
	//  -- nodeinit resource limits & requests
	//  ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	Resources Cilium1163Values_Cilium1163Values_Nodeinit_Resources
	//  -- Security context to be added to nodeinit pods.
	SecurityContext Cilium1163Values_Cilium1163Values_Nodeinit_SecurityContext
	//  -- bootstrapFile is the location of the file where the bootstrap timestamp is
	//  written by the node-init DaemonSet
	BootstrapFile any
	//  -- startup offers way to customize startup nodeinit script (pre and post position)
	Startup Cilium1163Values_Cilium1163Values_Nodeinit_Startup
	//  -- prestop offers way to customize prestop nodeinit script (pre and post position)
	Prestop Cilium1163Values_Cilium1163Values_Nodeinit_Prestop
}

type Cilium1163Values_Cilium1163Values_Preflight_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	//  cilium-digest
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Preflight_UpdateStrategy struct {
	Type any
}

type Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels struct {
	K8SApp any
}

type Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector struct {
	MatchLabels Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector_MatchLabels
}

type Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem struct {
	TopologyKey   any
	LabelSelector Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem_LabelSelector
}

type Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution []Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity_RequiredDuringSchedulingIgnoredDuringExecutionItem
}

type Cilium1163Values_Cilium1163Values_Preflight_Affinity struct {
	PodAffinity Cilium1163Values_Cilium1163Values_Preflight_Cilium1163Values_Cilium1163Values_Preflight_Affinity_PodAffinity
}

type Cilium1163Values_Cilium1163Values_Preflight_NodeSelector struct {
	KubernetesIoos any
}

type Cilium1163Values_Cilium1163Values_Preflight_TolerationsItem struct {
	//  - key: "key"
	//    operator: "Equal|Exists"
	//    value: "value"
	//    effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"
	Operator any
}

type Cilium1163Values_Cilium1163Values_Preflight_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Preflight_PodSecurityContext struct {
}

type Cilium1163Values_Cilium1163Values_Preflight_PodAnnotations struct {
}

type Cilium1163Values_Cilium1163Values_Preflight_PodLabels struct {
}

type Cilium1163Values_Cilium1163Values_Preflight_PodDisruptionBudget struct {
	//  -- enable PodDisruptionBudget
	//  ref: https://kubernetes.io/docs/concepts/workloads/pods/disruptions/
	Enabled any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Minimum number/percentage of pods that should remain scheduled.
	//  When it's set, maxUnavailable must be disabled by `maxUnavailable: null`
	MinAvailable any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Maximum number/percentage of pods that may be made unavailable
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Preflight_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Preflight_ReadinessProbe struct {
	//  -- For how long kubelet should wait before performing the first probe
	InitialDelaySeconds any
	//  -- interval between checks of the readiness probe
	PeriodSeconds any
}

type Cilium1163Values_Cilium1163Values_Preflight_SecurityContext struct {
}

type Cilium1163Values_Preflight struct {
	//  -- Enable Cilium pre-flight resources (required for upgrade)
	Enabled any
	//  -- Cilium pre-flight image.
	Image Cilium1163Values_Cilium1163Values_Preflight_Image
	//  -- The priority class to use for the preflight pod.
	PriorityClassName any
	//  -- preflight update strategy
	UpdateStrategy Cilium1163Values_Cilium1163Values_Preflight_UpdateStrategy
	//  -- Additional preflight environment variables.
	ExtraEnv []any
	//  -- Additional preflight volumes.
	ExtraVolumes []any
	//  -- Additional preflight volumeMounts.
	ExtraVolumeMounts []any
	//  -- Affinity for cilium-preflight
	Affinity Cilium1163Values_Cilium1163Values_Preflight_Affinity
	//  -- Node labels for preflight pod assignment
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
	NodeSelector Cilium1163Values_Cilium1163Values_Preflight_NodeSelector
	//  -- Node tolerations for preflight scheduling to nodes with taints
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []Cilium1163Values_Cilium1163Values_Preflight_TolerationsItem
	//  -- Annotations to be added to all top-level preflight objects (resources under templates/cilium-preflight)
	Annotations Cilium1163Values_Cilium1163Values_Preflight_Annotations
	//  -- Security context to be added to preflight pods.
	PodSecurityContext Cilium1163Values_Cilium1163Values_Preflight_PodSecurityContext
	//  -- Annotations to be added to preflight pods
	PodAnnotations Cilium1163Values_Cilium1163Values_Preflight_PodAnnotations
	//  -- Labels to be added to the preflight pod.
	PodLabels Cilium1163Values_Cilium1163Values_Preflight_PodLabels
	//  PodDisruptionBudget settings
	PodDisruptionBudget Cilium1163Values_Cilium1163Values_Preflight_PodDisruptionBudget
	//  -- preflight resource limits & requests
	//  ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	//    limits:
	//      cpu: 4000m
	//      memory: 4Gi
	//    requests:
	//      cpu: 100m
	//      memory: 512Mi
	Resources      Cilium1163Values_Cilium1163Values_Preflight_Resources
	ReadinessProbe Cilium1163Values_Cilium1163Values_Preflight_ReadinessProbe
	//  -- Security context to be added to preflight pods
	//    runAsUser: 0
	SecurityContext Cilium1163Values_Cilium1163Values_Preflight_SecurityContext
	//  -- Path to write the `--tofqdns-pre-cache` file to.
	TofqdnsPreCache any
	//  -- Configure termination grace period for preflight Deployment and DaemonSet.
	TerminationGracePeriodSeconds any
	//  -- By default we should always validate the installed CNPs before upgrading
	//  Cilium. This will make sure the user will have the policies deployed in the
	//  cluster with the right schema.
	ValidateCnps any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Config struct {
	//  -- Enable the Clustermesh explicit configuration.
	Enabled any
	//  -- Default dns domain for the Clustermesh API servers
	//  This is used in the case cluster addresses are not provided
	//  and IPs are used.
	Domain any
	//  -- List of clusters to be peered in the mesh.
	//  clusters:
	//  # -- Name of the cluster
	//  - name: cluster1
	//  # -- Address of the cluster, use this if you created DNS records for
	//  # the cluster Clustermesh API server.
	//    address: cluster1.mesh.cilium.io
	//  # -- Port of the cluster Clustermesh API server.
	//    port: 2379
	//  # -- IPs of the cluster Clustermesh API server, use multiple ones when
	//  # you have multiple IPs to access the Clustermesh API server.
	//    ips:
	//    - 172.18.255.201
	//  # -- base64 encoded PEM values for the cluster client certificate, private key and certificate authority.
	//  # These fields can (and should) be omitted in case the CA is shared across clusters. In that case, the
	//  # "remote" private key and certificate available in the local cluster are automatically used instead.
	//    tls:
	//      cert: ""
	//      key: ""
	//      caCert: ""
	Clusters []any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	//  clustermesh-apiserver-digest
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_ReadinessProbe struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_SecurityContext_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_SecurityContext_Capabilities_DropItem struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_SecurityContext_Capabilities struct {
	Drop []Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_SecurityContext_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_SecurityContext_Capabilities_DropItem
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_SecurityContext struct {
	AllowPrivilegeEscalation any
	Capabilities             Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_SecurityContext_Capabilities
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Lifecycle struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Init_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Init struct {
	//  -- Specifies the resources for etcd init container in the apiserver
	//    requests:
	//      cpu: 100m
	//      memory: 100Mi
	//    limits:
	//      cpu: 100m
	//      memory: 100Mi
	Resources Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Init_Resources
	//  -- Additional arguments to `clustermesh-apiserver etcdinit`.
	ExtraArgs []any
	//  -- Additional environment variables to `clustermesh-apiserver etcdinit`.
	ExtraEnv []any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd struct {
	//  The etcd binary is included in the clustermesh API server image, so the same image from above is reused.
	//  Independent override isn't supported, because clustermesh-apiserver is tested against the etcd version it is
	//  built with.
	//
	//  -- Specifies the resources for etcd container in the apiserver
	//    requests:
	//      cpu: 200m
	//      memory: 256Mi
	//    limits:
	//      cpu: 1000m
	//      memory: 256Mi
	Resources Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Resources
	//  -- Security context to be added to clustermesh-apiserver etcd containers
	SecurityContext Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_SecurityContext
	//  -- lifecycle setting for the etcd container
	Lifecycle Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Lifecycle
	Init      Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd_Init
	//  @schema
	//  enum: [Disk, Memory]
	//  @schema
	//  -- Specifies whether etcd data is stored in a temporary volume backed by
	//  the node's default medium, such as disk, SSD or network storage (Disk), or
	//  RAM (Memory). The Memory option enables improved etcd read and write
	//  performance at the cost of additional memory usage, which counts against
	//  the memory limits of the container.
	StorageMedium any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_ReadinessProbe struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_SecurityContext_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_SecurityContext_Capabilities_DropItem struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_SecurityContext_Capabilities struct {
	Drop []Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_SecurityContext_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_SecurityContext_Capabilities_DropItem
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_SecurityContext struct {
	AllowPrivilegeEscalation any
	Capabilities             Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_SecurityContext_Capabilities
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Lifecycle struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh struct {
	//  -- Enable KVStoreMesh. KVStoreMesh caches the information retrieved
	//  from the remote clusters in the local etcd instance.
	Enabled any
	//  -- TCP port for the KVStoreMesh health API.
	HealthPort any
	//  -- Configuration for the KVStoreMesh readiness probe.
	ReadinessProbe Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_ReadinessProbe
	//  -- Additional KVStoreMesh arguments.
	ExtraArgs []any
	//  -- Additional KVStoreMesh environment variables.
	ExtraEnv []any
	//  -- Resource requests and limits for the KVStoreMesh container
	//    requests:
	//      cpu: 100m
	//      memory: 64Mi
	//    limits:
	//      cpu: 1000m
	//      memory: 1024M
	Resources Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Resources
	//  -- Additional KVStoreMesh volumeMounts.
	ExtraVolumeMounts []any
	//  -- KVStoreMesh Security context
	SecurityContext Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_SecurityContext
	//  -- lifecycle setting for the KVStoreMesh container
	Lifecycle Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh_Lifecycle
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Service_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Service struct {
	//  -- The type of service used for apiserver access.
	Type any
	//  -- Optional port to use as the node port for apiserver access.
	//
	//  WARNING: make sure to configure a different NodePort in each cluster if
	//  kube-proxy replacement is enabled, as Cilium is currently affected by a known
	//  bug (#24692) when NodePorts are handled by the KPR implementation. If a service
	//  with the same NodePort exists both in the local and the remote cluster, all
	//  traffic originating from inside the cluster and targeting the corresponding
	//  NodePort will be redirected to a local backend, regardless of whether the
	//  destination node belongs to the local or the remote cluster.
	NodePort any
	//  -- Annotations for the clustermesh-apiserver
	//  For GKE LoadBalancer, use annotation cloud.google.com/load-balancer-type: "Internal"
	//  For EKS LoadBalancer, use annotation service.beta.kubernetes.io/aws-load-balancer-internal: "true"
	Annotations Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Service_Annotations
	//  @schema
	//  enum: [Local, Cluster]
	//  @schema
	//  -- The externalTrafficPolicy of service used for apiserver access.
	ExternalTrafficPolicy any
	//  @schema
	//  enum: [Local, Cluster]
	//  @schema
	//  -- The internalTrafficPolicy of service used for apiserver access.
	InternalTrafficPolicy any
	//  @schema
	//  enum: [HAOnly, Always, Never]
	//  @schema
	//  -- Defines when to enable session affinity.
	//  Each replica in a clustermesh-apiserver deployment runs its own discrete
	//  etcd cluster. Remote clients connect to one of the replicas through a
	//  shared Kubernetes Service. A client reconnecting to a different backend
	//  will require a full resync to ensure data integrity. Session affinity
	//  can reduce the likelihood of this happening, but may not be supported
	//  by all cloud providers.
	//  Possible values:
	//   - "HAOnly" (default) Only enable session affinity for deployments with more than 1 replica.
	//   - "Always" Always enable session affinity.
	//   - "Never" Never enable session affinity. Useful in environments where
	//             session affinity is not supported, but may lead to slightly
	//             degraded performance due to more frequent reconnections.
	EnableSessionAffinity any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Configure a loadBalancerClass.
	//  Allows to configure the loadBalancerClass on the clustermesh-apiserver
	//  LB service in case the Service type is set to LoadBalancer
	//  (requires Kubernetes 1.24+).
	LoadBalancerClass any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- Configure a specific loadBalancerIP.
	//  Allows to configure a specific loadBalancerIP on the clustermesh-apiserver
	//  LB service in case the Service type is set to LoadBalancer.
	LoadBalancerIp any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Lifecycle struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_SecurityContext_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_SecurityContext_Capabilities_DropItem struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_SecurityContext_Capabilities struct {
	Drop []Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_SecurityContext_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_SecurityContext_Capabilities_DropItem
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_SecurityContext struct {
	AllowPrivilegeEscalation any
	Capabilities             Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_SecurityContext_Capabilities
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_PodSecurityContext struct {
	RunAsNonRoot any
	RunAsUser    any
	RunAsGroup   any
	FsGroup      any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_PodAnnotations struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_PodLabels struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_PodDisruptionBudget struct {
	//  -- enable PodDisruptionBudget
	//  ref: https://kubernetes.io/docs/concepts/workloads/pods/disruptions/
	Enabled any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Minimum number/percentage of pods that should remain scheduled.
	//  When it's set, maxUnavailable must be disabled by `maxUnavailable: null`
	MinAvailable any
	//  @schema
	//  type: [null, integer, string]
	//  @schema
	//  -- Maximum number/percentage of pods that may be made unavailable
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_PodAffinityTerm_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_PodAffinityTerm_LabelSelector_MatchLabels struct {
	K8SApp any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_PodAffinityTerm_LabelSelector struct {
	MatchLabels Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_PodAffinityTerm_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_PodAffinityTerm_LabelSelector_MatchLabels
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_PodAffinityTerm struct {
	LabelSelector Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_PodAffinityTerm_LabelSelector
	TopologyKey   any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem struct {
	Weight          any
	PodAffinityTerm Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem_PodAffinityTerm
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity struct {
	PreferredDuringSchedulingIgnoredDuringExecution []Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity_PreferredDuringSchedulingIgnoredDuringExecutionItem
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity struct {
	PodAntiAffinity Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity_PodAntiAffinity
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_NodeSelector struct {
	KubernetesIoos any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_UpdateStrategy_RollingUpdate struct {
	//  @schema
	//  type: [integer, string]
	//  @schema
	MaxSurge any
	//  @schema
	//  type: [integer, string]
	//  @schema
	MaxUnavailable any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_UpdateStrategy struct {
	Type          any
	RollingUpdate Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_UpdateStrategy_RollingUpdate
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Auto_CertManagerIssuerRef struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Auto struct {
	//  -- When set to true, automatically generate a CA and certificates to
	//  enable mTLS between clustermesh-apiserver and external workload instances.
	//  If set to false, the certs to be provided by setting appropriate values below.
	Enabled any
	//  Sets the method to auto-generate certificates. Supported values:
	//  - helm:         This method uses Helm to generate all certificates.
	//  - cronJob:      This method uses a Kubernetes CronJob the generate any
	//                  certificates not provided by the user at installation
	//                  time.
	//  - certmanager:  This method use cert-manager to generate & rotate certificates.
	Method any
	//  -- Generated certificates validity duration in days.
	//  -- Schedule for certificates regeneration (regardless of their expiration date).
	//  Only used if method is "cronJob". If nil, then no recurring job will be created.
	//  Instead, only the one-shot job is deployed to generate the certificates at
	//  installation time.
	//
	//  Due to the out-of-band distribution of client certs to external workloads the
	//  CA is (re)regenerated only if it is not provided as a helm value and the k8s
	//  secret is manually deleted.
	//
	//  Defaults to none. Commented syntax gives midnight of the first day of every
	//  fourth month. For syntax, see
	//  https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#schedule-syntax
	//  schedule: "0 0 1 */4 *"
	CertValidityDuration any
	//  [Example]
	//  certManagerIssuerRef:
	//    group: cert-manager.io
	//    kind: ClusterIssuer
	//    name: ca-issuer
	//  -- certmanager issuer used when clustermesh.apiserver.tls.auto.method=certmanager.
	CertManagerIssuerRef Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Auto_CertManagerIssuerRef
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Server struct {
	Cert any
	Key  any
	//  -- Extra DNS names added to certificate when it's auto generated
	ExtraDnsNames []any
	//  -- Extra IP addresses added to certificate when it's auto generated
	ExtraIpAddresses []any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Admin struct {
	Cert any
	Key  any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Client struct {
	Cert any
	Key  any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Remote struct {
	Cert any
	Key  any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls struct {
	//  -- Configure the clustermesh authentication mode.
	//  Supported values:
	//  - legacy:     All clusters access remote clustermesh instances with the same
	//                username (i.e., remote). The "remote" certificate must be
	//                generated with CN=remote if provided manually.
	//  - migration:  Intermediate mode required to upgrade from legacy to cluster
	//                (and vice versa) with no disruption. Specifically, it enables
	//                the creation of the per-cluster usernames, while still using
	//                the common one for authentication. The "remote" certificate must
	//                be generated with CN=remote if provided manually (same as legacy).
	//  - cluster:    Each cluster accesses remote etcd instances with a username
	//                depending on the local cluster name (i.e., remote-<cluster-name>).
	//                The "remote" certificate must be generated with CN=remote-<cluster-name>
	//                if provided manually. Cluster mode is meaningful only when the same
	//                CA is shared across all clusters part of the mesh.
	AuthMode any
	//  -- Allow users to provide their own certificates
	//  Users may need to provide their certificates using
	//  a mechanism that requires they provide their own secrets.
	//  This setting does not apply to any of the auto-generated
	//  mechanisms below, it only restricts the creation of secrets
	//  via the `tls-provided` templates.
	EnableSecrets any
	//  -- Configure automatic TLS certificates generation.
	//  A Kubernetes CronJob is used the generate any
	//  certificates not provided by the user at installation
	//  time.
	Auto Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Auto
	//  -- base64 encoded PEM values for the clustermesh-apiserver server certificate and private key.
	//  Used if 'auto' is not enabled.
	Server Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Server
	//  -- base64 encoded PEM values for the clustermesh-apiserver admin certificate and private key.
	//  Used if 'auto' is not enabled.
	Admin Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Admin
	//  -- base64 encoded PEM values for the clustermesh-apiserver client certificate and private key.
	//  Used if 'auto' is not enabled.
	Client Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Client
	//  -- base64 encoded PEM values for the clustermesh-apiserver remote cluster certificate and private key.
	//  Used if 'auto' is not enabled.
	Remote Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls_Remote
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Kvstoremesh struct {
	//  -- Enables exporting KVStoreMesh metrics in OpenMetrics format.
	Enabled any
	//  -- Configure the port the KVStoreMesh metric server listens on.
	Port any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Etcd struct {
	//  -- Enables exporting etcd metrics in OpenMetrics format.
	Enabled any
	//  -- Set level of detail for etcd metrics; specify 'extensive' to include server side gRPC histogram metrics.
	Mode any
	//  -- Configure the port the etcd metric server listens on.
	Port any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor_Kvstoremesh struct {
	//  -- Interval for scrape metrics (KVStoreMesh metrics)
	Interval any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Relabeling configs for the ServiceMonitor clustermesh-apiserver (KVStoreMesh metrics)
	Relabelings any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Metrics relabeling configs for the ServiceMonitor clustermesh-apiserver (KVStoreMesh metrics)
	MetricRelabelings any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor_Etcd struct {
	//  -- Interval for scrape metrics (etcd metrics)
	Interval any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Relabeling configs for the ServiceMonitor clustermesh-apiserver (etcd metrics)
	Relabelings any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Metrics relabeling configs for the ServiceMonitor clustermesh-apiserver (etcd metrics)
	MetricRelabelings any
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor struct {
	//  -- Enable service monitor.
	//  This requires the prometheus CRDs to be available (see https://github.com/prometheus-operator/prometheus-operator/blob/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml)
	Enabled any
	//  -- Labels to add to ServiceMonitor clustermesh-apiserver
	Labels Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor_Labels
	//  -- Annotations to add to ServiceMonitor clustermesh-apiserver
	//  -- Specify the Kubernetes namespace where Prometheus expects to find
	//  service monitors configured.
	//  namespace: ""
	Annotations Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor_Annotations
	//  -- Interval for scrape metrics (apiserver metrics)
	Interval any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Relabeling configs for the ServiceMonitor clustermesh-apiserver (apiserver metrics)
	Relabelings any
	//  @schema
	//  type: [null, array]
	//  @schema
	//  -- Metrics relabeling configs for the ServiceMonitor clustermesh-apiserver (apiserver metrics)
	MetricRelabelings any
	Kvstoremesh       Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor_Kvstoremesh
	Etcd              Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor_Etcd
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics struct {
	//  -- Enables exporting apiserver metrics in OpenMetrics format.
	Enabled any
	//  -- Configure the port the apiserver metric server listens on.
	Port           any
	Kvstoremesh    Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Kvstoremesh
	Etcd           Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_Etcd
	ServiceMonitor Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics_ServiceMonitor
}

type Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver struct {
	//  -- Clustermesh API server image.
	Image Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Image
	//  -- TCP port for the clustermesh-apiserver health API.
	HealthPort any
	//  -- Configuration for the clustermesh-apiserver readiness probe.
	ReadinessProbe Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_ReadinessProbe
	Etcd           Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Etcd
	Kvstoremesh    Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Kvstoremesh
	Service        Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Service
	//  -- Number of replicas run for the clustermesh-apiserver deployment.
	Replicas any
	//  -- lifecycle setting for the apiserver container
	Lifecycle Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Lifecycle
	//  -- terminationGracePeriodSeconds for the clustermesh-apiserver deployment
	TerminationGracePeriodSeconds any
	//  -- Additional clustermesh-apiserver arguments.
	ExtraArgs []any
	//  -- Additional clustermesh-apiserver environment variables.
	ExtraEnv []any
	//  -- Additional clustermesh-apiserver volumes.
	ExtraVolumes []any
	//  -- Additional clustermesh-apiserver volumeMounts.
	ExtraVolumeMounts []any
	//  -- Security context to be added to clustermesh-apiserver containers
	SecurityContext Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_SecurityContext
	//  -- Security context to be added to clustermesh-apiserver pods
	PodSecurityContext Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_PodSecurityContext
	//  -- Annotations to be added to clustermesh-apiserver pods
	PodAnnotations Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_PodAnnotations
	//  -- Labels to be added to clustermesh-apiserver pods
	PodLabels Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_PodLabels
	//  PodDisruptionBudget settings
	PodDisruptionBudget Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_PodDisruptionBudget
	//  -- Resource requests and limits for the clustermesh-apiserver
	//    requests:
	//      cpu: 100m
	//      memory: 64Mi
	//    limits:
	//      cpu: 1000m
	//      memory: 1024M
	Resources Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Resources
	//  -- Affinity for clustermesh.apiserver
	Affinity Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Affinity
	//  -- Pod topology spread constraints for clustermesh-apiserver
	//    - maxSkew: 1
	//      topologyKey: topology.kubernetes.io/zone
	//      whenUnsatisfiable: DoNotSchedule
	TopologySpreadConstraints []any
	//  -- Node labels for pod assignment
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
	NodeSelector Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_NodeSelector
	//  -- Node tolerations for pod assignment on nodes with taints
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []any
	//  -- clustermesh-apiserver update strategy
	UpdateStrategy Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_UpdateStrategy
	//  -- The priority class to use for clustermesh-apiserver
	PriorityClassName any
	Tls               Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Tls
	//  clustermesh-apiserver Prometheus metrics configuration
	Metrics Cilium1163Values_Cilium1163Values_Clustermesh_Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver_Metrics
}

type Cilium1163Values_Clustermesh struct {
	//  -- Deploy clustermesh-apiserver for clustermesh
	UseApiserver any
	//  -- The maximum number of clusters to support in a ClusterMesh. This value
	//  cannot be changed on running clusters, and all clusters in a ClusterMesh
	//  must be configured with the same value. Values > 255 will decrease the
	//  maximum allocatable cluster-local identities.
	//  Supported values are 255 and 511.
	MaxConnectedClusters any
	//  -- Enable the synchronization of Kubernetes EndpointSlices corresponding to
	//  the remote endpoints of appropriately-annotated global services through ClusterMesh
	EnableEndpointSliceSynchronization any
	//  -- Enable Multi-Cluster Services API support
	EnableMcsapisupport any
	//  -- Annotations to be added to all top-level clustermesh objects (resources under templates/clustermesh-apiserver and templates/clustermesh-config)
	Annotations Cilium1163Values_Cilium1163Values_Clustermesh_Annotations
	//  -- Clustermesh explicit configuration.
	Config    Cilium1163Values_Cilium1163Values_Clustermesh_Config
	Apiserver Cilium1163Values_Cilium1163Values_Clustermesh_Apiserver
}

type Cilium1163Values_ExternalWorkloads struct {
	//  -- Enable support for external workloads, such as VMs (false by default).
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Cgroup_Cilium1163Values_Cilium1163Values_Cgroup_AutoMount_Resources struct {
}

type Cilium1163Values_Cilium1163Values_Cgroup_AutoMount struct {
	//  -- Enable auto mount of cgroup2 filesystem.
	//  When `autoMount` is enabled, cgroup2 filesystem is mounted at
	//  `cgroup.hostRoot` path on the underlying host and inside the cilium agent pod.
	//  If users disable `autoMount`, it's expected that users have mounted
	//  cgroup2 filesystem at the specified `cgroup.hostRoot` volume, and then the
	//  volume will be mounted inside the cilium agent pod at the same path.
	Enabled any
	//  -- Init Container Cgroup Automount resource limits & requests
	//    limits:
	//      cpu: 100m
	//      memory: 128Mi
	//    requests:
	//      cpu: 100m
	//      memory: 128Mi
	Resources Cilium1163Values_Cilium1163Values_Cgroup_Cilium1163Values_Cilium1163Values_Cgroup_AutoMount_Resources
}

type Cilium1163Values_Cgroup struct {
	AutoMount Cilium1163Values_Cilium1163Values_Cgroup_AutoMount
	//  -- Configure cgroup root where cgroup2 filesystem is mounted on the host (see also: `cgroup.autoMount`)
	HostRoot any
}

type Cilium1163Values_Sysctlfix struct {
	//  -- Enable the sysctl override. When enabled, the init container will mount the /proc of the host so that the `sysctlfix` utility can execute.
	Enabled any
}

type Cilium1163Values_DnsProxy struct {
	//  -- Timeout (in seconds) when closing the connection between the DNS proxy and the upstream server. If set to 0, the connection is closed immediately (with TCP RST). If set to -1, the connection is closed asynchronously in the background.
	SocketLingerTimeout any
	//  -- DNS response code for rejecting DNS requests, available options are '[nameError refused]'.
	DnsRejectResponseCode any
	//  -- Allow the DNS proxy to compress responses to endpoints that are larger than 512 Bytes or the EDNS0 option, if present.
	EnableDnsCompression any
	//  -- Maximum number of IPs to maintain per FQDN name for each endpoint.
	EndpointMaxIpPerHostname any
	//  -- Time during which idle but previously active connections with expired DNS lookups are still considered alive.
	IdleConnectionGracePeriod any
	//  -- Maximum number of IPs to retain for expired DNS lookups with still-active connections.
	MaxDeferredConnectionDeletes any
	//  -- The minimum time, in seconds, to use DNS data for toFQDNs policies. If
	//  the upstream DNS server returns a DNS record with a shorter TTL, Cilium
	//  overwrites the TTL with this value. Setting this value to zero means that
	//  Cilium will honor the TTLs returned by the upstream DNS server.
	MinTtl any
	//  -- DNS cache data at this path is preloaded on agent startup.
	PreCache any
	//  -- Global port on which the in-agent DNS proxy should listen. Default 0 is a OS-assigned port.
	ProxyPort any
	//  -- The maximum time the DNS proxy holds an allowed DNS response before sending it along. Responses are sent as soon as the datapath is updated with the new IP information.
	//  -- DNS proxy operation mode (true/false, or unset to use version dependent defaults)
	//  enableTransparentMode: true
	ProxyResponseMaxDelay any
}

type Cilium1163Values_Sctp struct {
	//  -- Enable SCTP support. NOTE: Currently, SCTP support does not support rewriting ports or multihoming.
	Enabled any
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_InitImage struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_ServiceAccount struct {
	Create any
	Name   any
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_TolerationsItem struct {
	Key    any
	Effect any
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_Affinity struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_NodeSelector struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_PodSecurityContext struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_SecurityContext struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent struct {
	//  -- SPIRE agent image
	Image Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_Image
	//  -- SPIRE agent service account
	ServiceAccount Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_ServiceAccount
	//  -- SPIRE agent annotations
	Annotations Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_Annotations
	//  -- SPIRE agent labels
	Labels Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_Labels
	//  -- SPIRE Workload Attestor kubelet verification.
	SkipKubeletVerification any
	//  -- SPIRE agent tolerations configuration
	//  By default it follows the same tolerations as the agent itself
	//  to allow the Cilium agent on this node to connect to SPIRE.
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_TolerationsItem
	//  -- SPIRE agent affinity configuration
	Affinity Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_Affinity
	//  -- SPIRE agent nodeSelector configuration
	//  ref: ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
	NodeSelector Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_NodeSelector
	//  -- Security context to be added to spire agent pods.
	//  SecurityContext holds pod-level security attributes and common container settings.
	//  ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-pod
	PodSecurityContext Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_PodSecurityContext
	//  -- Security context to be added to spire agent containers.
	//  SecurityContext holds pod-level security attributes and common container settings.
	//  ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-container
	SecurityContext Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent_SecurityContext
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Image struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	Override   any
	Repository any
	Tag        any
	Digest     any
	UseDigest  any
	PullPolicy any
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_ServiceAccount struct {
	Create any
	Name   any
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Service_Annotations struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Service_Labels struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Service struct {
	//  -- Service type for the SPIRE server service
	Type any
	//  -- Annotations to be added to the SPIRE server service
	Annotations Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Service_Annotations
	//  -- Labels to be added to the SPIRE server service
	Labels Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Service_Labels
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Affinity struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_NodeSelector struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_DataStorage struct {
	//  -- Enable SPIRE server data storage
	Enabled any
	//  -- Size of the SPIRE server data storage
	Size any
	//  -- Access mode of the SPIRE server data storage
	AccessMode any
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- StorageClass of the SPIRE server data storage
	StorageClass any
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_PodSecurityContext struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_SecurityContext struct {
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Ca_Subject struct {
	Country      any
	Organization any
	CommonName   any
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Ca struct {
	//  -- SPIRE CA key type
	//  AWS requires the use of RSA. EC cryptography is not supported
	KeyType any
	//  -- SPIRE CA Subject
	Subject Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Ca_Subject
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server struct {
	//  -- SPIRE server image
	Image Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Image
	//  -- SPIRE server service account
	ServiceAccount Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_ServiceAccount
	//  -- SPIRE server init containers
	InitContainers []any
	//  -- SPIRE server annotations
	Annotations Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Annotations
	//  -- SPIRE server labels
	Labels Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Labels
	//  SPIRE server service configuration
	Service Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Service
	//  -- SPIRE server affinity configuration
	Affinity Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Affinity
	//  -- SPIRE server nodeSelector configuration
	//  ref: ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
	NodeSelector Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_NodeSelector
	//  -- SPIRE server tolerations configuration
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []any
	//  SPIRE server datastorage configuration
	DataStorage Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_DataStorage
	//  -- Security context to be added to spire server pods.
	//  SecurityContext holds pod-level security attributes and common container settings.
	//  ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-pod
	PodSecurityContext Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_PodSecurityContext
	//  -- Security context to be added to spire server containers.
	//  SecurityContext holds pod-level security attributes and common container settings.
	//  ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-container
	SecurityContext Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_SecurityContext
	//  SPIRE CA configuration
	Ca Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server_Ca
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install struct {
	//  -- Enable SPIRE installation.
	//  This will only take effect only if authentication.mutual.spire.enabled is true
	Enabled any
	//  -- SPIRE namespace to install into
	Namespace any
	//  -- SPIRE namespace already exists. Set to true if Helm should not create, manage, and import the SPIRE namespace.
	ExistingNamespace any
	//  -- init container image of SPIRE agent and server
	InitImage Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_InitImage
	//  SPIRE agent configuration
	Agent  Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Agent
	Server Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install_Server
}

type Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire struct {
	//  -- Enable SPIRE integration (beta)
	Enabled any
	//  -- Annotations to be added to all top-level spire objects (resources under templates/spire)
	Annotations Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Annotations
	//  Settings to control the SPIRE installation and configuration
	Install Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire_Install
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- SPIRE server address used by Cilium Operator
	//
	//  If k8s Service DNS along with port number is used (e.g. <service-name>.<namespace>.svc(.*):<port-number> format),
	//  Cilium Operator will resolve its address by looking up the clusterIP from Service resource.
	//
	//  Example values: 10.0.0.1:8081, spire-server.cilium-spire.svc:8081
	ServerAddress any
	//  -- SPIFFE trust domain to use for fetching certificates
	TrustDomain any
	//  -- SPIRE socket path where the SPIRE delegated api agent is listening
	AdminSocketPath any
	//  -- SPIRE socket path where the SPIRE workload agent is listening.
	//  Applies to both the Cilium Agent and Operator
	AgentSocketPath any
	//  -- SPIRE connection timeout
	ConnectionTimeout any
}

type Cilium1163Values_Cilium1163Values_Authentication_Mutual struct {
	//  -- Port on the agent where mutual authentication handshakes between agents will be performed
	Port any
	//  -- Timeout for connecting to the remote node TCP socket
	ConnectTimeout any
	//  Settings for SPIRE
	Spire Cilium1163Values_Cilium1163Values_Authentication_Cilium1163Values_Cilium1163Values_Authentication_Mutual_Spire
}

type Cilium1163Values_Authentication struct {
	//  -- Enable authentication processing and garbage collection.
	//  Note that if disabled, policy enforcement will still block requests that require authentication.
	//  But the resulting authentication requests for these requests will not be processed, therefore the requests not be allowed.
	Enabled any
	//  -- Buffer size of the channel Cilium uses to receive authentication events from the signal map.
	QueueSize any
	//  -- Buffer size of the channel Cilium uses to receive certificate expiration events from auth handlers.
	RotatedIdentitiesQueueSize any
	//  -- Interval for garbage collection of auth map entries.
	GcInterval any
	//  Configuration for Cilium's service-to-service mutual authentication using TLS handshakes.
	//  Note that this is not full mTLS support without also enabling encryption of some form.
	//  Current encryption options are WireGuard or IPsec, configured in encryption block above.
	Mutual Cilium1163Values_Cilium1163Values_Authentication_Mutual
}

type Cilium1163Values struct {
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- upgradeCompatibility helps users upgrading to ensure that the configMap for
	//  Cilium will not change critical values to ensure continued operation
	//  This flag is not required for new installations.
	//  For example: '1.7', '1.8', '1.9'
	UpgradeCompatibility any
	Debug                Cilium1163Values_Debug
	Rbac                 Cilium1163Values_Rbac
	//  -- Configure image pull secrets for pulling container images
	//  - name: "image-pull-secret"
	ImagePullSecrets []any
	//  -- (string) Kubernetes config path
	//  @default -- `"~/.kube/config"`
	KubeConfigPath any
	//  -- (string) Kubernetes service host - use "auto" for automatic lookup from the cluster-info ConfigMap (kubeadm-based clusters only)
	K8SServiceHost any
	//  @schema
	//  type: [string, integer]
	//  @schema
	//  -- (string) Kubernetes service port
	K8SServicePort any
	//  -- Configure the client side rate limit for the agent and operator
	//
	//  If the amount of requests to the Kubernetes API server exceeds the configured
	//  rate limit, the agent and operator will start to throttle requests by delaying
	//  them until there is budget or the request times out.
	K8SClientRateLimit Cilium1163Values_K8SClientRateLimit
	Cluster            Cilium1163Values_Cluster
	//  -- Define serviceAccount names for components.
	//  @default -- Component's fully qualified name.
	ServiceAccounts Cilium1163Values_ServiceAccounts
	//  -- Configure termination grace period for cilium-agent DaemonSet.
	TerminationGracePeriodSeconds any
	//  -- Install the cilium agent resources.
	Agent any
	//  -- Agent container name.
	Name any
	//  -- Roll out cilium agent pods automatically when configmap is updated.
	RollOutCiliumPods any
	//  -- Agent container image.
	Image Cilium1163Values_Image
	//  -- Affinity for cilium-agent.
	Affinity Cilium1163Values_Affinity
	//  -- Node selector for cilium-agent.
	NodeSelector Cilium1163Values_NodeSelector
	//  -- Node tolerations for agent scheduling to nodes with taints
	//  ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []Cilium1163Values_TolerationsItem
	//  -- The priority class to use for cilium-agent.
	PriorityClassName any
	//  -- DNS policy for Cilium agent pods.
	//  Ref: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
	DnsPolicy any
	//  -- Additional containers added to the cilium DaemonSet.
	ExtraContainers []any
	//  -- Additional initContainers added to the cilium Daemonset.
	ExtraInitContainers []any
	//  -- Additional agent container arguments.
	ExtraArgs []any
	//  -- Additional agent container environment variables.
	ExtraEnv []any
	//  -- Additional agent hostPath mounts.
	//  - name: host-mnt-data
	//    mountPath: /host/mnt/data
	//    hostPath: /mnt/data
	//    hostPathType: Directory
	//    readOnly: true
	//    mountPropagation: HostToContainer
	ExtraHostPathMounts []any
	//  -- Additional agent volumes.
	ExtraVolumes []any
	//  -- Additional agent volumeMounts.
	ExtraVolumeMounts []any
	//  -- extraConfig allows you to specify additional configuration parameters to be
	//  included in the cilium-config configmap.
	//   my-config-a: "1234"
	//   my-config-b: |-
	//     test 1
	//     test 2
	//     test 3
	ExtraConfig Cilium1163Values_ExtraConfig
	//  -- Annotations to be added to all top-level cilium-agent objects (resources under templates/cilium-agent)
	Annotations Cilium1163Values_Annotations
	//  -- Security Context for cilium-agent pods.
	PodSecurityContext Cilium1163Values_PodSecurityContext
	//  -- Annotations to be added to agent pods
	PodAnnotations Cilium1163Values_PodAnnotations
	//  -- Labels to be added to agent pods
	PodLabels Cilium1163Values_PodLabels
	//  -- Agent resource limits & requests
	//  ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	//    limits:
	//      cpu: 4000m
	//      memory: 4Gi
	//    requests:
	//      cpu: 100m
	//      memory: 512Mi
	Resources Cilium1163Values_Resources
	//  -- resources & limits for the agent init containers
	InitResources   Cilium1163Values_InitResources
	SecurityContext Cilium1163Values_SecurityContext
	//  -- Cilium agent update strategy
	UpdateStrategy Cilium1163Values_UpdateStrategy
	//  Configuration Values for cilium-agent
	Aksbyocni Cilium1163Values_Aksbyocni
	//  @schema
	//  type: [boolean, string]
	//  @schema
	//  -- Enable installation of PodCIDR routes between worker
	//  nodes if worker nodes share a common L2 network segment.
	AutoDirectNodeRoutes any
	//  -- Enable skipping of PodCIDR routes between worker
	//  nodes if the worker nodes are in a different L2 network segment.
	DirectRoutingSkipUnreachable any
	//  -- Annotate k8s node upon initialization with Cilium's metadata.
	AnnotateK8SNode any
	Azure           Cilium1163Values_Azure
	Alibabacloud    Cilium1163Values_Alibabacloud
	//  -- Enable bandwidth manager to optimize TCP and UDP workloads and allow
	//  for rate-limiting traffic from individual Pods with EDT (Earliest Departure
	//  Time) through the "kubernetes.io/egress-bandwidth" Pod annotation.
	BandwidthManager Cilium1163Values_BandwidthManager
	//  -- Configure standalone NAT46/NAT64 gateway
	Nat46X64Gateway Cilium1163Values_Nat46X64Gateway
	//  -- EnableHighScaleIPcache enables the special ipcache mode for high scale
	//  clusters. The ipcache content will be reduced to the strict minimum and
	//  traffic will be encapsulated to carry security identities.
	HighScaleIpcache Cilium1163Values_HighScaleIpcache
	//  -- Configure L2 announcements
	L2Announcements Cilium1163Values_L2Announcements
	//  -- Configure L2 pod announcements
	L2PodAnnouncements Cilium1163Values_L2PodAnnouncements
	//  -- Configure BGP
	Bgp Cilium1163Values_Bgp
	//  -- This feature set enables virtual BGP routers to be created via
	//  CiliumBGPPeeringPolicy CRDs.
	BgpControlPlane Cilium1163Values_BgpControlPlane
	PmtuDiscovery   Cilium1163Values_PmtuDiscovery
	Bpf             Cilium1163Values_Bpf
	//  -- Enable BPF clock source probing for more efficient tick retrieval.
	BpfClockProbe any
	//  -- Clean all eBPF datapath state from the initContainer of the cilium-agent
	//  DaemonSet.
	//
	//  WARNING: Use with care!
	CleanBpfState any
	//  -- Clean all local Cilium state from the initContainer of the cilium-agent
	//  DaemonSet. Implies cleanBpfState: true.
	//
	//  WARNING: Use with care!
	CleanState any
	//  -- Wait for KUBE-PROXY-CANARY iptables rule to appear in "wait-for-kube-proxy"
	//  init container before launching cilium-agent.
	//  More context can be found in the commit message of below PR
	//  https://github.com/cilium/cilium/pull/20123
	WaitForKubeProxy any
	Cni              Cilium1163Values_Cni
	//  -- (string) Configure how frequently garbage collection should occur for the datapath
	//  connection tracking table.
	//  @default -- `"0s"`
	ConntrackGcinterval any
	//  -- (string) Configure the maximum frequency for the garbage collection of the
	//  connection tracking table. Only affects the automatic computation for the frequency
	//  and has no effect when 'conntrackGCInterval' is set. This can be set to more frequently
	//  clean up unused identities created from ToFQDN policies.
	ConntrackGcmaxInterval any
	//  -- (string) Configure timeout in which Cilium will exit if CRDs are not available
	//  @default -- `"5m"`
	CrdWaitTimeout any
	//  -- Tail call hooks for custom eBPF programs.
	CustomCalls Cilium1163Values_CustomCalls
	//  -- Specify which network interfaces can run the eBPF datapath. This means
	//  that a packet sent from a pod to a destination outside the cluster will be
	//  masqueraded (to an output device IPv4 address), if the output device runs the
	//  program. When not specified, probing will automatically detect devices that have
	//  a non-local route. This should be used only when autodetection is not suitable.
	//  devices: ""
	Daemon Cilium1163Values_Daemon
	//  -- Enables experimental support for the detection of new and removed datapath
	//  devices. When devices change the eBPF datapath is reloaded and services updated.
	//  If "devices" is set then only those devices, or devices matching a wildcard will
	//  be considered.
	//
	//  This option has been deprecated and is a no-op.
	EnableRuntimeDeviceDetection any
	//  -- Forces the auto-detection of devices, even if specific devices are explicitly listed
	//  -- Chains to ignore when installing feeder rules.
	//  disableIptablesFeederRules: ""
	ForceDeviceDetection any
	//  -- Limit iptables-based egress masquerading to interface selector.
	//  egressMasqueradeInterfaces: ""
	//
	//  -- Enable setting identity mark for local traffic.
	//  enableIdentityMark: true
	//
	//  -- Enable Kubernetes EndpointSlice feature in Cilium if the cluster supports it.
	//  enableK8sEndpointSlice: true
	//
	//  -- Enable CiliumEndpointSlice feature (deprecated, please use `ciliumEndpointSlice.enabled` instead).
	EnableCiliumEndpointSlice any
	CiliumEndpointSlice       Cilium1163Values_CiliumEndpointSlice
	EnvoyConfig               Cilium1163Values_EnvoyConfig
	IngressController         Cilium1163Values_IngressController
	GatewayApi                Cilium1163Values_GatewayApi
	//  -- Enables the fallback compatibility solution for when the xt_socket kernel
	//  module is missing and it is needed for the datapath L7 redirection to work
	//  properly. See documentation for details on when this can be disabled:
	//  https://docs.cilium.io/en/stable/operations/system_requirements/#linux-kernel.
	EnableXtsocketFallback any
	Encryption             Cilium1163Values_Encryption
	EndpointHealthChecking Cilium1163Values_EndpointHealthChecking
	EndpointRoutes         Cilium1163Values_EndpointRoutes
	K8SNetworkPolicy       Cilium1163Values_K8SNetworkPolicy
	Eni                    Cilium1163Values_Eni
	ExternalIps            Cilium1163Values_ExternalIps
	//  fragmentTracking enables IPv4 fragment tracking support in the datapath.
	//  fragmentTracking: true
	Gke Cilium1163Values_Gke
	//  -- Enable connectivity health checking.
	HealthChecking any
	//  -- TCP port for the agent health API. This is not the port for cilium-health.
	HealthPort any
	//  -- Configure the host firewall.
	HostFirewall Cilium1163Values_HostFirewall
	HostPort     Cilium1163Values_HostPort
	//  -- Configure socket LB
	SocketLb Cilium1163Values_SocketLb
	//  -- Configure certificate generation for Hubble integration.
	//  If hubble.tls.auto.method=cronJob, these values are used
	//  for the Kubernetes CronJob which will be scheduled regularly to
	//  (re)generate any certificates not provided manually.
	Certgen Cilium1163Values_Certgen
	Hubble  Cilium1163Values_Hubble
	//  -- Method to use for identity allocation (`crd` or `kvstore`).
	IdentityAllocationMode any
	//  -- (string) Time to wait before using new identity on endpoint identity change.
	//  @default -- `"5s"`
	IdentityChangeGracePeriod any
	//  -- Install Iptables rules to skip netfilter connection tracking on all pod
	//  traffic. This option is only effective when Cilium is running in direct
	//  routing and full KPR mode. Moreover, this option cannot be enabled when Cilium
	//  is running in a managed Kubernetes environment or in a chained CNI setup.
	InstallNoConntrackIptablesRules any
	Ipam                            Cilium1163Values_Ipam
	NodeIpam                        Cilium1163Values_NodeIpam
	//  @schema
	//  type: [null, string]
	//  @schema
	//  -- The api-rate-limit option can be used to overwrite individual settings of the default configuration for rate limiting calls to the Cilium Agent API
	ApiRateLimit any
	//  -- Configure the eBPF-based ip-masq-agent
	//  the config of nonMasqueradeCIDRs
	//  config:
	//    nonMasqueradeCIDRs: []
	//    masqLinkLocal: false
	//    masqLinkLocalIPv6: false
	IpMasqAgent Cilium1163Values_IpMasqAgent
	//  iptablesLockTimeout defines the iptables "--wait" option when invoked from Cilium.
	//  iptablesLockTimeout: "5s"
	Ipv4 Cilium1163Values_Ipv4
	Ipv6 Cilium1163Values_Ipv6
	//  -- Configure Kubernetes specific configuration
	K8S Cilium1163Values_K8S
	//  -- Keep the deprecated selector labels when deploying Cilium DaemonSet.
	KeepDeprecatedLabels any
	//  -- Keep the deprecated probes when deploying Cilium DaemonSet
	KeepDeprecatedProbes any
	StartupProbe         Cilium1163Values_StartupProbe
	LivenessProbe        Cilium1163Values_LivenessProbe
	//  -- Configure the kube-proxy replacement in Cilium BPF datapath
	//  Valid options are "true" or "false".
	//  ref: https://docs.cilium.io/en/stable/network/kubernetes/kubeproxy-free/
	// kubeProxyReplacement: "false"
	ReadinessProbe Cilium1163Values_ReadinessProbe
	//  -- healthz server bind address for the kube-proxy replacement.
	//  To enable set the value to '0.0.0.0:10256' for all ipv4
	//  addresses and this '[::]:10256' for all ipv6 addresses.
	//  By default it is disabled.
	KubeProxyReplacementHealthzBindAddr any
	L2NeighDiscovery                    Cilium1163Values_L2NeighDiscovery
	//  -- Enable Layer 7 network policy.
	L7Proxy any
	//  -- Enable Local Redirect Policy.
	//  To include or exclude matched resources from cilium identity evaluation
	//  labels: ""
	LocalRedirectPolicy any
	//  logOptions allows you to define logging options. eg:
	//  logOptions:
	//    format: json
	//
	//  -- Enables periodic logging of system load
	LogSystemLoad any
	//  -- Configure maglev consistent hashing
	//  -- tableSize is the size (parameter M) for the backend table of one
	//  service entry
	//  tableSize:
	Maglev Cilium1163Values_Maglev
	//  -- hashSeed is the cluster-wide base64 encoded seed for the hashing
	//  hashSeed:
	//
	//  -- Enables masquerading of IPv4 traffic leaving the node from endpoints.
	EnableIpv4Masquerade any
	//  -- Enables masquerading of IPv6 traffic leaving the node from endpoints.
	EnableIpv6Masquerade any
	//  -- Enables masquerading to the source of the route for traffic leaving the node from endpoints.
	EnableMasqueradeRouteSource any
	//  -- Enables IPv4 BIG TCP support which increases maximum IPv4 GSO/GRO limits for nodes and pods
	EnableIpv4Bigtcp any
	//  -- Enables IPv6 BIG TCP support which increases maximum IPv6 GSO/GRO limits for nodes and pods
	EnableIpv6Bigtcp any
	Nat              Cilium1163Values_Nat
	EgressGateway    Cilium1163Values_EgressGateway
	Vtep             Cilium1163Values_Vtep
	//  -- (string) Allows to explicitly specify the IPv4 CIDR for native routing.
	//  When specified, Cilium assumes networking for this CIDR is preconfigured and
	//  hands traffic destined for that range to the Linux network stack without
	//  applying any SNAT.
	//  Generally speaking, specifying a native routing CIDR implies that Cilium can
	//  depend on the underlying networking stack to route packets to their
	//  destination. To offer a concrete example, if Cilium is configured to use
	//  direct routing and the Kubernetes CIDR is included in the native routing CIDR,
	//  the user must configure the routes to reach pods, either manually or by
	//  setting the auto-direct-node-routes flag.
	Ipv4NativeRoutingCidr any
	//  -- (string) Allows to explicitly specify the IPv6 CIDR for native routing.
	//  When specified, Cilium assumes networking for this CIDR is preconfigured and
	//  hands traffic destined for that range to the Linux network stack without
	//  applying any SNAT.
	//  Generally speaking, specifying a native routing CIDR implies that Cilium can
	//  depend on the underlying networking stack to route packets to their
	//  destination. To offer a concrete example, if Cilium is configured to use
	//  direct routing and the Kubernetes CIDR is included in the native routing CIDR,
	//  the user must configure the routes to reach pods, either manually or by
	//  setting the auto-direct-node-routes flag.
	Ipv6NativeRoutingCidr any
	//  -- cilium-monitor sidecar.
	Monitor Cilium1163Values_Monitor
	//  -- Configure service load balancing
	LoadBalancer Cilium1163Values_LoadBalancer
	//  -- Configure N-S k8s service loadbalancing
	//  policyAuditMode: false
	NodePort Cilium1163Values_NodePort
	//  -- The agent can be put into one of the three policy enforcement modes:
	//  default, always and never.
	//  ref: https://docs.cilium.io/en/stable/security/policy/intro/#policy-enforcement-modes
	PolicyEnforcementMode any
	//  @schema
	//  type: [null, string, array]
	//  @schema
	//  -- policyCIDRMatchMode is a list of entities that may be selected by CIDR selector.
	//  The possible value is "nodes".
	PolicyCidrmatchMode any
	Pprof               Cilium1163Values_Pprof
	//  -- Configure prometheus metrics on the configured port at /metrics
	Prometheus Cilium1163Values_Prometheus
	//  -- Grafana dashboards for cilium-agent
	//  grafana can import dashboards based on the label and value
	//  ref: https://github.com/grafana/helm-charts/tree/main/charts/grafana#sidecar-for-dashboards
	Dashboards Cilium1163Values_Dashboards
	//  Configure Cilium Envoy options.
	Envoy Cilium1163Values_Envoy
	//  -- Enable/Disable use of node label based identity
	NodeSelectorLabels any
	//  -- Enable resource quotas for priority classes used in the cluster.
	//  Need to document default
	//
	// sessionAffinity: false
	ResourceQuotas Cilium1163Values_ResourceQuotas
	//  -- Do not run Cilium agent when running with clean mode. Useful to completely
	//  uninstall Cilium as it will stop Cilium from starting and create artifacts
	//  in the node.
	SleepAfterInit any
	//  -- Enable check of service source ranges (currently, only for LoadBalancer).
	SvcSourceRangeCheck any
	//  -- Synchronize Kubernetes nodes to kvstore and perform CNP GC.
	SynchronizeK8SNodes any
	//  -- Configure TLS configuration in the agent.
	Tls Cilium1163Values_Tls
	//  -- Tunneling protocol to use in tunneling mode and for ad-hoc tunnels.
	//  Possible values:
	//    - ""
	//    - vxlan
	//    - geneve
	//  @default -- `"vxlan"`
	TunnelProtocol any
	//  -- Enable native-routing mode or tunneling mode.
	//  Possible values:
	//    - ""
	//    - native
	//    - tunnel
	//  @default -- `"tunnel"`
	RoutingMode any
	//  -- Configure VXLAN and Geneve tunnel port.
	//  @default -- Port 8472 for VXLAN, Port 6081 for Geneve
	TunnelPort any
	//  -- Configure what the response should be to traffic for a service without backends.
	//  "reject" only works on kernels >= 5.10, on lower kernels we fallback to "drop".
	//  Possible values:
	//   - reject (default)
	//   - drop
	ServiceNoBackendResponse any
	//  -- Configure the underlying network MTU to overwrite auto-detected MTU.
	//  This value doesn't change the host network interface MTU i.e. eth0 or ens0.
	//  It changes the MTU for cilium_net@cilium_host, cilium_host@cilium_net,
	//  cilium_vxlan and lxc_health interfaces.
	Mtu any
	//  -- Disable the usage of CiliumEndpoint CRD.
	DisableEndpointCrd  any
	WellKnownIdentities Cilium1163Values_WellKnownIdentities
	Etcd                Cilium1163Values_Etcd
	Operator            Cilium1163Values_Operator
	Nodeinit            Cilium1163Values_Nodeinit
	Preflight           Cilium1163Values_Preflight
	//  -- Explicitly enable or disable priority class.
	//  .Capabilities.KubeVersion is unsettable in `helm template` calls,
	//  it depends on k8s libraries version that Helm was compiled against.
	//  This option allows to explicitly disable setting the priority class, which
	//  is useful for rendering charts for gke clusters in advance.
	EnableCriticalPriorityClass any
	//  disableEnvoyVersionCheck removes the check for Envoy, which can be useful
	//  on AArch64 as the images do not currently ship a version of Envoy.
	// disableEnvoyVersionCheck: false
	Clustermesh Cilium1163Values_Clustermesh
	//  -- Configure external workloads support
	ExternalWorkloads Cilium1163Values_ExternalWorkloads
	//  -- Configure cgroup related configuration
	Cgroup Cilium1163Values_Cgroup
	//  -- Configure sysctl override described in #20072.
	Sysctlfix Cilium1163Values_Sysctlfix
	//  -- Configure whether to enable auto detect of terminating state for endpoints
	//  in order to support graceful termination.
	//  -- Configure whether to unload DNS policy rules on graceful shutdown
	//  dnsPolicyUnloadOnShutdown: false
	EnableK8STerminatingEndpoint any
	//  -- Configure the key of the taint indicating that Cilium is not ready on the node.
	//  When set to a value starting with `ignore-taint.cluster-autoscaler.kubernetes.io/`, the Cluster Autoscaler will ignore the taint on its decisions, allowing the cluster to scale up.
	AgentNotReadyTaintKey any
	DnsProxy              Cilium1163Values_DnsProxy
	//  -- SCTP Configuration Values
	Sctp Cilium1163Values_Sctp
	//  Configuration for types of authentication for Cilium (beta)
	Authentication Cilium1163Values_Authentication
}
