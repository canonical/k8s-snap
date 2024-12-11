package main

type MetricsServer3122Values_Image struct {
	Repository any
	//  Overrides the image tag whose default is v{{ .Chart.AppVersion }}
	Tag        any
	PullPolicy any
}

type MetricsServer3122Values_MetricsServer3122Values_ServiceAccount_Annotations struct {
}

type MetricsServer3122Values_ServiceAccount struct {
	//  Specifies whether a service account should be created
	Create any
	//  Annotations to add to the service account
	Annotations MetricsServer3122Values_MetricsServer3122Values_ServiceAccount_Annotations
	//  The name of the service account to use.
	//  If not set and create is true, a name is generated using the fullname template
	Name any
	//  The list of secrets mountable by this service account.
	//  See https://kubernetes.io/docs/reference/labels-annotations-taints/#enforce-mountable-secrets
	Secrets []any
}

type MetricsServer3122Values_Rbac struct {
	//  Specifies whether RBAC resources should be created
	Create any
	//  Note: PodSecurityPolicy will not be created when Kubernetes version is 1.25 or later.
	PspEnabled any
}

type MetricsServer3122Values_MetricsServer3122Values_ApiService_Annotations struct {
}

type MetricsServer3122Values_ApiService struct {
	//  Specifies if the v1beta1.metrics.k8s.io API service should be created.
	//
	//  You typically want this enabled! If you disable API service creation you have to
	//  manage it outside of this chart for e.g horizontal pod autoscaling to
	//  work with this release.
	Create any
	//  Annotations to add to the API service
	Annotations MetricsServer3122Values_MetricsServer3122Values_ApiService_Annotations
	//  Specifies whether to skip TLS verification
	InsecureSkipTlsverify any
	//  The PEM encoded CA bundle for TLS verification
	CaBundle any
}

type MetricsServer3122Values_CommonLabels struct {
}

type MetricsServer3122Values_PodLabels struct {
}

type MetricsServer3122Values_PodAnnotations struct {
}

type MetricsServer3122Values_PodSecurityContext struct {
}

type MetricsServer3122Values_MetricsServer3122Values_SecurityContext_SeccompProfile struct {
	Type any
}

type MetricsServer3122Values_MetricsServer3122Values_SecurityContext_MetricsServer3122Values_MetricsServer3122Values_SecurityContext_Capabilities_DropItem struct {
}

type MetricsServer3122Values_MetricsServer3122Values_SecurityContext_Capabilities struct {
	Drop []MetricsServer3122Values_MetricsServer3122Values_SecurityContext_MetricsServer3122Values_MetricsServer3122Values_SecurityContext_Capabilities_DropItem
}

type MetricsServer3122Values_SecurityContext struct {
	AllowPrivilegeEscalation any
	ReadOnlyRootFilesystem   any
	RunAsNonRoot             any
	RunAsUser                any
	SeccompProfile           MetricsServer3122Values_MetricsServer3122Values_SecurityContext_SeccompProfile
	Capabilities             MetricsServer3122Values_MetricsServer3122Values_SecurityContext_Capabilities
}

type MetricsServer3122Values_HostNetwork struct {
	//  Specifies if metrics-server should be started in hostNetwork mode.
	//
	//  You would require this enabled if you use alternate overlay networking for pods and
	//  API server unable to communicate with metrics-server. As an example, this is required
	//  if you use Weave network on EKS
	Enabled any
}

type MetricsServer3122Values_UpdateStrategy struct {
}

type MetricsServer3122Values_PodDisruptionBudget struct {
	//  https://kubernetes.io/docs/tasks/run-application/configure-pdb/
	Enabled        any
	MinAvailable   any
	MaxUnavailable any
}

type MetricsServer3122Values_DefaultArgsItem struct {
}

type MetricsServer3122Values_MetricsServer3122Values_LivenessProbe_HttpGet struct {
	Path   any
	Port   any
	Scheme any
}

type MetricsServer3122Values_LivenessProbe struct {
	HttpGet             MetricsServer3122Values_MetricsServer3122Values_LivenessProbe_HttpGet
	InitialDelaySeconds any
	PeriodSeconds       any
	FailureThreshold    any
}

type MetricsServer3122Values_MetricsServer3122Values_ReadinessProbe_HttpGet struct {
	Path   any
	Port   any
	Scheme any
}

type MetricsServer3122Values_ReadinessProbe struct {
	HttpGet             MetricsServer3122Values_MetricsServer3122Values_ReadinessProbe_HttpGet
	InitialDelaySeconds any
	PeriodSeconds       any
	FailureThreshold    any
}

type MetricsServer3122Values_MetricsServer3122Values_Service_Annotations struct {
}

type MetricsServer3122Values_MetricsServer3122Values_Service_Labels struct {
}

type MetricsServer3122Values_Service struct {
	Type        any
	Port        any
	Annotations MetricsServer3122Values_MetricsServer3122Values_Service_Annotations
	//   Add these labels to have metrics-server show up in `kubectl cluster-info`
	//   kubernetes.io/cluster-service: "true"
	//   kubernetes.io/name: "Metrics-server"
	Labels MetricsServer3122Values_MetricsServer3122Values_Service_Labels
}

type MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Image struct {
	Repository any
	Tag        any
}

type MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext_SeccompProfile struct {
	Type any
}

type MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext_Capabilities_DropItem struct {
}

type MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext_Capabilities struct {
	Drop []MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext_Capabilities_DropItem
}

type MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext struct {
	AllowPrivilegeEscalation any
	ReadOnlyRootFilesystem   any
	RunAsNonRoot             any
	RunAsUser                any
	SeccompProfile           MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext_SeccompProfile
	Capabilities             MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext_Capabilities
}

type MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Resources_Requests struct {
	Cpu    any
	Memory any
}

type MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Resources_Limits struct {
	Cpu    any
	Memory any
}

type MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Resources struct {
	Requests MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Resources_Requests
	Limits   MetricsServer3122Values_MetricsServer3122Values_AddonResizer_MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Resources_Limits
}

type MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Nanny struct {
	Cpu            any
	ExtraCpu       any
	Memory         any
	ExtraMemory    any
	MinClusterSize any
	PollPeriod     any
	Threshold      any
}

type MetricsServer3122Values_AddonResizer struct {
	Enabled         any
	Image           MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Image
	SecurityContext MetricsServer3122Values_MetricsServer3122Values_AddonResizer_SecurityContext
	Resources       MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Resources
	Nanny           MetricsServer3122Values_MetricsServer3122Values_AddonResizer_Nanny
}

type MetricsServer3122Values_Metrics struct {
	Enabled any
}

type MetricsServer3122Values_MetricsServer3122Values_ServiceMonitor_AdditionalLabels struct {
}

type MetricsServer3122Values_ServiceMonitor struct {
	Enabled           any
	AdditionalLabels  MetricsServer3122Values_MetricsServer3122Values_ServiceMonitor_AdditionalLabels
	Interval          any
	ScrapeTimeout     any
	MetricRelabelings []any
	Relabelings       []any
}

type MetricsServer3122Values_MetricsServer3122Values_Resources_Requests struct {
	Cpu    any
	Memory any
}

type MetricsServer3122Values_Resources struct {
	//  limits:
	//    cpu:
	//    memory:
	Requests MetricsServer3122Values_MetricsServer3122Values_Resources_Requests
}

type MetricsServer3122Values_NodeSelector struct {
}

type MetricsServer3122Values_Affinity struct {
}

type MetricsServer3122Values_DnsConfig struct {
}

type MetricsServer3122Values_DeploymentAnnotations struct {
}

type MetricsServer3122Values_MetricsServer3122Values_TmpVolume_EmptyDir struct {
}

type MetricsServer3122Values_TmpVolume struct {
	EmptyDir MetricsServer3122Values_MetricsServer3122Values_TmpVolume_EmptyDir
}

type MetricsServer3122Values struct {
	Image MetricsServer3122Values_Image
	//  - name: registrySecretName
	ImagePullSecrets     []any
	NameOverride         any
	FullnameOverride     any
	ServiceAccount       MetricsServer3122Values_ServiceAccount
	Rbac                 MetricsServer3122Values_Rbac
	ApiService           MetricsServer3122Values_ApiService
	CommonLabels         MetricsServer3122Values_CommonLabels
	PodLabels            MetricsServer3122Values_PodLabels
	PodAnnotations       MetricsServer3122Values_PodAnnotations
	PodSecurityContext   MetricsServer3122Values_PodSecurityContext
	SecurityContext      MetricsServer3122Values_SecurityContext
	PriorityClassName    any
	ContainerPort        any
	HostNetwork          MetricsServer3122Values_HostNetwork
	Replicas             any
	RevisionHistoryLimit any
	//    type: RollingUpdate
	//    rollingUpdate:
	//      maxSurge: 0
	//      maxUnavailable: 1
	UpdateStrategy      MetricsServer3122Values_UpdateStrategy
	PodDisruptionBudget MetricsServer3122Values_PodDisruptionBudget
	DefaultArgs         []MetricsServer3122Values_DefaultArgsItem
	Args                []any
	LivenessProbe       MetricsServer3122Values_LivenessProbe
	ReadinessProbe      MetricsServer3122Values_ReadinessProbe
	Service             MetricsServer3122Values_Service
	AddonResizer        MetricsServer3122Values_AddonResizer
	Metrics             MetricsServer3122Values_Metrics
	ServiceMonitor      MetricsServer3122Values_ServiceMonitor
	//  See https://github.com/kubernetes-sigs/metrics-server#scaling
	Resources                 MetricsServer3122Values_Resources
	ExtraVolumeMounts         []any
	ExtraVolumes              []any
	NodeSelector              MetricsServer3122Values_NodeSelector
	Tolerations               []any
	Affinity                  MetricsServer3122Values_Affinity
	TopologySpreadConstraints []any
	DnsConfig                 MetricsServer3122Values_DnsConfig
	//  Annotations to add to the deployment
	DeploymentAnnotations MetricsServer3122Values_DeploymentAnnotations
	SchedulerName         any
	TmpVolume             MetricsServer3122Values_TmpVolume
}
