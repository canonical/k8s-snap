package main

type Metallb0148Values_Rbac struct {
	//  create specifies whether to install and use RBAC rules.
	Create any
}

type Metallb0148Values_Metallb0148Values_Prometheus_RbacProxy struct {
	Repository any
	Tag        any
	PullPolicy any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PodMonitor_AdditionalLabels struct {
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PodMonitor_Annotations struct {
}

type Metallb0148Values_Metallb0148Values_Prometheus_PodMonitor struct {
	//  enable support for Prometheus Operator
	Enabled any
	//  optional additionnal labels for podMonitors
	AdditionalLabels Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PodMonitor_AdditionalLabels
	//  optional annotations for podMonitors
	Annotations Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PodMonitor_Annotations
	//  Job label for scrape target
	JobLabel any
	//  Scrape interval. If not set, the Prometheus default scrape interval is used.
	Interval any
	//  	metric relabel configs to apply to samples before ingestion.
	//  - action: keep
	//    regex: 'kube_(daemonset|deployment|pod|namespace|node|statefulset).+'
	//    sourceLabels: [__name__]
	MetricRelabelings []any
	//  	relabel configs to apply to samples before ingestion.
	//  - sourceLabels: [__meta_kubernetes_pod_node_name]
	//    separator: ;
	//    regex: ^(.*)$
	//    target_label: nodename
	//    replacement: $1
	//    action: replace
	Relabelings []any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Speaker_AdditionalLabels struct {
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Speaker_Annotations struct {
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Speaker_TlsConfig struct {
	InsecureSkipVerify any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Speaker struct {
	//  optional additional labels for the speaker serviceMonitor
	AdditionalLabels Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Speaker_AdditionalLabels
	//  optional additional annotations for the speaker serviceMonitor
	Annotations Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Speaker_Annotations
	//  optional tls configuration for the speaker serviceMonitor, in case
	//  secure metrics are enabled.
	TlsConfig Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Speaker_TlsConfig
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Controller_AdditionalLabels struct {
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Controller_Annotations struct {
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Controller_TlsConfig struct {
	InsecureSkipVerify any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Controller struct {
	//  optional additional labels for the controller serviceMonitor
	AdditionalLabels Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Controller_AdditionalLabels
	//  optional additional annotations for the controller serviceMonitor
	Annotations Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Controller_Annotations
	//  optional tls configuration for the controller serviceMonitor, in case
	//  secure metrics are enabled.
	TlsConfig Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Controller_TlsConfig
}

type Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor struct {
	//  enable support for Prometheus Operator
	Enabled    any
	Speaker    Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Speaker
	Controller Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor_Controller
	//  Job label for scrape target
	JobLabel any
	//  Scrape interval. If not set, the Prometheus default scrape interval is used.
	Interval any
	//  	metric relabel configs to apply to samples before ingestion.
	//  - action: keep
	//    regex: 'kube_(daemonset|deployment|pod|namespace|node|statefulset).+'
	//    sourceLabels: [__name__]
	MetricRelabelings []any
	//  	relabel configs to apply to samples before ingestion.
	//  - sourceLabels: [__meta_kubernetes_pod_node_name]
	//    separator: ;
	//    regex: ^(.*)$
	//    target_label: nodename
	//    replacement: $1
	//    action: replace
	Relabelings []any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AdditionalLabels struct {
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Annotations struct {
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_StaleConfig_Labels struct {
	Severity any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_StaleConfig struct {
	Enabled any
	Labels  Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_StaleConfig_Labels
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_ConfigNotLoaded_Labels struct {
	Severity any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_ConfigNotLoaded struct {
	Enabled any
	Labels  Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_ConfigNotLoaded_Labels
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolExhausted_Labels struct {
	Severity any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolExhausted struct {
	Enabled any
	Labels  Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolExhausted_Labels
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolUsage_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolUsage_ThresholdsItem_Labels struct {
	Severity any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolUsage_ThresholdsItem struct {
	Percent any
	Labels  Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolUsage_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolUsage_ThresholdsItem_Labels
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolUsage struct {
	Enabled    any
	Thresholds []Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolUsage_ThresholdsItem
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_BgpSessionDown_Labels struct {
	Severity any
}

type Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_BgpSessionDown struct {
	Enabled any
	Labels  Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_BgpSessionDown_Labels
}

type Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule struct {
	//  enable alertmanager alerts
	Enabled any
	//  optional additionnal labels for prometheusRules
	AdditionalLabels Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AdditionalLabels
	//  optional annotations for prometheusRules
	Annotations Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_Annotations
	//  MetalLBStaleConfig
	StaleConfig Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_StaleConfig
	//  MetalLBConfigNotLoaded
	ConfigNotLoaded Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_ConfigNotLoaded
	//  MetalLBAddressPoolExhausted
	AddressPoolExhausted Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolExhausted
	AddressPoolUsage     Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_AddressPoolUsage
	//  MetalLBBGPSessionDown
	BgpSessionDown Metallb0148Values_Metallb0148Values_Prometheus_Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule_BgpSessionDown
	ExtraAlerts    []any
}

type Metallb0148Values_Prometheus struct {
	//  scrape annotations specifies whether to add Prometheus metric
	//  auto-collection annotations to pods. See
	//  https://github.com/prometheus/prometheus/blob/release-2.1/documentation/examples/prometheus-kubernetes.yml
	//  for a corresponding Prometheus configuration. Alternatively, you
	//  may want to use the Prometheus Operator
	//  (https://github.com/coreos/prometheus-operator) for more powerful
	//  monitoring configuration. If you use the Prometheus operator, this
	//  can be left at false.
	ScrapeAnnotations any
	//  port both controller and speaker will listen on for metrics
	MetricsPort any
	//  if set, enables rbac proxy on the controller and speaker to expose
	//  the metrics via tls.
	//  secureMetricsPort: 9120
	//
	//  the name of the secret to be mounted in the speaker pod
	//  to expose the metrics securely. If not present, a self signed
	//  certificate to be used.
	SpeakerMetricsTlssecret any
	//  the name of the secret to be mounted in the controller pod
	//  to expose the metrics securely. If not present, a self signed
	//  certificate to be used.
	ControllerMetricsTlssecret any
	//  prometheus doens't have the permission to scrape all namespaces so we give it permission to scrape metallb's one
	RbacPrometheus any
	//  the service account used by prometheus
	//  required when " .Values.prometheus.rbacPrometheus == true " and " .Values.prometheus.podMonitor.enabled=true or prometheus.serviceMonitor.enabled=true "
	ServiceAccount any
	//  the namespace where prometheus is deployed
	//  required when " .Values.prometheus.rbacPrometheus == true " and " .Values.prometheus.podMonitor.enabled=true or prometheus.serviceMonitor.enabled=true "
	Namespace any
	//  the image to be used for the kuberbacproxy container
	RbacProxy Metallb0148Values_Metallb0148Values_Prometheus_RbacProxy
	//  Prometheus Operator PodMonitors
	PodMonitor Metallb0148Values_Metallb0148Values_Prometheus_PodMonitor
	//  Prometheus Operator ServiceMonitors. To be used as an alternative
	//  to podMonitor, supports secure metrics.
	ServiceMonitor Metallb0148Values_Metallb0148Values_Prometheus_ServiceMonitor
	//  Prometheus Operator alertmanager alerts
	PrometheusRule Metallb0148Values_Metallb0148Values_Prometheus_PrometheusRule
}

type Metallb0148Values_Metallb0148Values_Controller_Image struct {
	Repository any
	Tag        any
	PullPolicy any
}

type Metallb0148Values_Metallb0148Values_Controller_Strategy struct {
	Type any
}

type Metallb0148Values_Metallb0148Values_Controller_Metallb0148Values_Metallb0148Values_Controller_ServiceAccount_Annotations struct {
}

type Metallb0148Values_Metallb0148Values_Controller_ServiceAccount struct {
	//  Specifies whether a ServiceAccount should be created
	Create any
	//  The name of the ServiceAccount to use. If not set and create is
	//  true, a name is generated using the fullname template
	Name        any
	Annotations Metallb0148Values_Metallb0148Values_Controller_Metallb0148Values_Metallb0148Values_Controller_ServiceAccount_Annotations
}

type Metallb0148Values_Metallb0148Values_Controller_SecurityContext struct {
	RunAsNonRoot any
	//  nobody
	RunAsUser any
	FsGroup   any
}

type Metallb0148Values_Metallb0148Values_Controller_Resources struct {
}

type Metallb0148Values_Metallb0148Values_Controller_NodeSelector struct {
}

type Metallb0148Values_Metallb0148Values_Controller_Affinity struct {
}

type Metallb0148Values_Metallb0148Values_Controller_PodAnnotations struct {
}

type Metallb0148Values_Metallb0148Values_Controller_Labels struct {
}

type Metallb0148Values_Metallb0148Values_Controller_LivenessProbe struct {
	Enabled             any
	FailureThreshold    any
	InitialDelaySeconds any
	PeriodSeconds       any
	SuccessThreshold    any
	TimeoutSeconds      any
}

type Metallb0148Values_Metallb0148Values_Controller_ReadinessProbe struct {
	Enabled             any
	FailureThreshold    any
	InitialDelaySeconds any
	PeriodSeconds       any
	SuccessThreshold    any
	TimeoutSeconds      any
}

type Metallb0148Values_Controller struct {
	Enabled any
	//  -- Controller log level. Must be one of: `all`, `debug`, `info`, `warn`, `error` or `none`
	LogLevel any
	//  command: /controller
	//  webhookMode: enabled
	Image Metallb0148Values_Metallb0148Values_Controller_Image
	//  @param controller.updateStrategy.type Metallb controller deployment strategy type.
	//  ref: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy
	//  e.g:
	//  strategy:
	//   type: RollingUpdate
	//   rollingUpdate:
	//     maxSurge: 25%
	//     maxUnavailable: 25%
	//
	Strategy        Metallb0148Values_Metallb0148Values_Controller_Strategy
	ServiceAccount  Metallb0148Values_Metallb0148Values_Controller_ServiceAccount
	SecurityContext Metallb0148Values_Metallb0148Values_Controller_SecurityContext
	Resources       Metallb0148Values_Metallb0148Values_Controller_Resources
	//  limits:
	//  cpu: 100m
	//  memory: 100Mi
	NodeSelector      Metallb0148Values_Metallb0148Values_Controller_NodeSelector
	Tolerations       []any
	PriorityClassName any
	RuntimeClassName  any
	Affinity          Metallb0148Values_Metallb0148Values_Controller_Affinity
	PodAnnotations    Metallb0148Values_Metallb0148Values_Controller_PodAnnotations
	Labels            Metallb0148Values_Metallb0148Values_Controller_Labels
	LivenessProbe     Metallb0148Values_Metallb0148Values_Controller_LivenessProbe
	ReadinessProbe    Metallb0148Values_Metallb0148Values_Controller_ReadinessProbe
	TlsMinVersion     any
	TlsCipherSuites   any
	ExtraContainers   []any
}

type Metallb0148Values_Metallb0148Values_Speaker_Memberlist struct {
	Enabled            any
	MlBindPort         any
	MlBindAddrOverride any
	MlSecretKeyPath    any
}

type Metallb0148Values_Metallb0148Values_Speaker_ExcludeInterfaces struct {
	Enabled any
}

type Metallb0148Values_Metallb0148Values_Speaker_Image struct {
	Repository any
	Tag        any
	PullPolicy any
}

type Metallb0148Values_Metallb0148Values_Speaker_UpdateStrategy struct {
	//  StrategyType
	//  Can be set to RollingUpdate or OnDelete
	//
	Type any
}

type Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_ServiceAccount_Annotations struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_ServiceAccount struct {
	//  Specifies whether a ServiceAccount should be created
	Create any
	//  The name of the ServiceAccount to use. If not set and create is
	//  true, a name is generated using the fullname template
	Name        any
	Annotations Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_ServiceAccount_Annotations
}

type Metallb0148Values_Metallb0148Values_Speaker_SecurityContext struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_Resources struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_NodeSelector struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_Affinity struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_PodAnnotations struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_Labels struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_LivenessProbe struct {
	Enabled             any
	FailureThreshold    any
	InitialDelaySeconds any
	PeriodSeconds       any
	SuccessThreshold    any
	TimeoutSeconds      any
}

type Metallb0148Values_Metallb0148Values_Speaker_ReadinessProbe struct {
	Enabled             any
	FailureThreshold    any
	InitialDelaySeconds any
	PeriodSeconds       any
	SuccessThreshold    any
	TimeoutSeconds      any
}

type Metallb0148Values_Metallb0148Values_Speaker_StartupProbe struct {
	Enabled          any
	FailureThreshold any
	PeriodSeconds    any
}

type Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_Frr_Image struct {
	Repository any
	Tag        any
	PullPolicy any
}

type Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_Frr_Resources struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_Frr struct {
	Enabled     any
	Image       Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_Frr_Image
	MetricsPort any
	//  if set, enables a rbac proxy sidecar container on the speaker to
	//  expose the frr metrics via tls.
	//  secureMetricsPort: 9121
	//
	Resources Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_Frr_Resources
}

type Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_Reloader_Resources struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_Reloader struct {
	Resources Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_Reloader_Resources
}

type Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_FrrMetrics_Resources struct {
}

type Metallb0148Values_Metallb0148Values_Speaker_FrrMetrics struct {
	Resources Metallb0148Values_Metallb0148Values_Speaker_Metallb0148Values_Metallb0148Values_Speaker_FrrMetrics_Resources
}

type Metallb0148Values_Speaker struct {
	Enabled any
	//  command: /speaker
	//  -- Speaker log level. Must be one of: `all`, `debug`, `info`, `warn`, `error` or `none`
	LogLevel          any
	TolerateMaster    any
	Memberlist        Metallb0148Values_Metallb0148Values_Speaker_Memberlist
	ExcludeInterfaces Metallb0148Values_Metallb0148Values_Speaker_ExcludeInterfaces
	//  ignore the exclude-from-external-loadbalancer label
	IgnoreExcludeLb any
	Image           Metallb0148Values_Metallb0148Values_Speaker_Image
	//  @param speaker.updateStrategy.type Speaker daemonset strategy type
	//  ref: https://kubernetes.io/docs/tasks/manage-daemon/update-daemon-set/
	//
	UpdateStrategy  Metallb0148Values_Metallb0148Values_Speaker_UpdateStrategy
	ServiceAccount  Metallb0148Values_Metallb0148Values_Speaker_ServiceAccount
	SecurityContext Metallb0148Values_Metallb0148Values_Speaker_SecurityContext
	//  Defines a secret name for the controller to generate a memberlist encryption secret
	//  By default secretName: {{ "metallb.fullname" }}-memberlist
	//
	//  secretName:
	Resources Metallb0148Values_Metallb0148Values_Speaker_Resources
	//  limits:
	//  cpu: 100m
	//  memory: 100Mi
	NodeSelector      Metallb0148Values_Metallb0148Values_Speaker_NodeSelector
	Tolerations       []any
	PriorityClassName any
	Affinity          Metallb0148Values_Metallb0148Values_Speaker_Affinity
	//  Selects which runtime class will be used by the pod.
	RuntimeClassName any
	PodAnnotations   Metallb0148Values_Metallb0148Values_Speaker_PodAnnotations
	Labels           Metallb0148Values_Metallb0148Values_Speaker_Labels
	LivenessProbe    Metallb0148Values_Metallb0148Values_Speaker_LivenessProbe
	ReadinessProbe   Metallb0148Values_Metallb0148Values_Speaker_ReadinessProbe
	StartupProbe     Metallb0148Values_Metallb0148Values_Speaker_StartupProbe
	//  frr contains configuration specific to the MetalLB FRR container,
	//  for speaker running alongside FRR.
	Frr             Metallb0148Values_Metallb0148Values_Speaker_Frr
	Reloader        Metallb0148Values_Metallb0148Values_Speaker_Reloader
	FrrMetrics      Metallb0148Values_Metallb0148Values_Speaker_FrrMetrics
	ExtraContainers []any
}

type Metallb0148Values_Crds struct {
	Enabled                 any
	ValidationFailurePolicy any
}

type Metallb0148Values_Frrk8S struct {
	//  if set, enables frrk8s as a backend. This is mutually exclusive to frr
	//  mode.
	Enabled   any
	External  any
	Namespace any
}

type Metallb0148Values struct {
	ImagePullSecrets  []any
	NameOverride      any
	FullnameOverride  any
	LoadBalancerClass any
	//  To configure MetalLB, you must specify ONE of the following two
	//  options.
	//
	Rbac       Metallb0148Values_Rbac
	Prometheus Metallb0148Values_Prometheus
	//  controller contains configuration specific to the MetalLB cluster
	//  controller.
	Controller Metallb0148Values_Controller
	//  speaker contains configuration specific to the MetalLB speaker
	//  daemonset.
	Speaker Metallb0148Values_Speaker
	Crds    Metallb0148Values_Crds
	//  frrk8s contains the configuration related to using an frrk8s instance
	//  (github.com/metallb/frr-k8s) as the backend for the BGP implementation.
	//  This allows configuring additional frr parameters in combination to those
	//  applied by MetalLB.
	Frrk8S Metallb0148Values_Frrk8S
}
