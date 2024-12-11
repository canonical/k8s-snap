package main

type Coredns1360Values_Image struct {
	Repository any
	//  Overrides the image tag whose default is the chart appVersion.
	Tag        any
	PullPolicy any
	//  Optionally specify an array of imagePullSecrets.
	//  Secrets must be manually created in the namespace.
	//  ref: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
	//
	//  pullSecrets:
	//    - name: myRegistryKeySecretName
	PullSecrets []any
}

type Coredns1360Values_Coredns1360Values_Resources_Limits struct {
	Cpu    any
	Memory any
}

type Coredns1360Values_Coredns1360Values_Resources_Requests struct {
	Cpu    any
	Memory any
}

type Coredns1360Values_Resources struct {
	Limits   Coredns1360Values_Coredns1360Values_Resources_Limits
	Requests Coredns1360Values_Coredns1360Values_Resources_Requests
}

type Coredns1360Values_RollingUpdate struct {
	MaxUnavailable any
	MaxSurge       any
}

type Coredns1360Values_PodAnnotations struct {
}

type Coredns1360Values_Coredns1360Values_Prometheus_Coredns1360Values_Coredns1360Values_Prometheus_Service_Annotations struct {
	PrometheusIoscrape any
	PrometheusIoport   any
}

type Coredns1360Values_Coredns1360Values_Prometheus_Coredns1360Values_Coredns1360Values_Prometheus_Service_Selector struct {
}

type Coredns1360Values_Coredns1360Values_Prometheus_Service struct {
	Enabled     any
	Annotations Coredns1360Values_Coredns1360Values_Prometheus_Coredns1360Values_Coredns1360Values_Prometheus_Service_Annotations
	Selector    Coredns1360Values_Coredns1360Values_Prometheus_Coredns1360Values_Coredns1360Values_Prometheus_Service_Selector
}

type Coredns1360Values_Coredns1360Values_Prometheus_Coredns1360Values_Coredns1360Values_Prometheus_Monitor_AdditionalLabels struct {
}

type Coredns1360Values_Coredns1360Values_Prometheus_Coredns1360Values_Coredns1360Values_Prometheus_Monitor_Selector struct {
}

type Coredns1360Values_Coredns1360Values_Prometheus_Monitor struct {
	Enabled          any
	AdditionalLabels Coredns1360Values_Coredns1360Values_Prometheus_Coredns1360Values_Coredns1360Values_Prometheus_Monitor_AdditionalLabels
	Namespace        any
	Interval         any
	Selector         Coredns1360Values_Coredns1360Values_Prometheus_Coredns1360Values_Coredns1360Values_Prometheus_Monitor_Selector
}

type Coredns1360Values_Prometheus struct {
	Service Coredns1360Values_Coredns1360Values_Prometheus_Service
	Monitor Coredns1360Values_Coredns1360Values_Prometheus_Monitor
}

type Coredns1360Values_Coredns1360Values_Service_Annotations struct {
}

type Coredns1360Values_Coredns1360Values_Service_Selector struct {
}

type Coredns1360Values_Service struct {
	//  clusterIP: ""
	//  clusterIPs: []
	//  loadBalancerIP: ""
	//  loadBalancerClass: ""
	//  externalIPs: []
	//  externalTrafficPolicy: ""
	//  ipFamilyPolicy: ""
	//  trafficDistribution: PreferClose
	//  The name of the Service
	//  If not set, a name is generated using the fullname template
	Name        any
	Annotations Coredns1360Values_Coredns1360Values_Service_Annotations
	//  Pod selector
	Selector Coredns1360Values_Coredns1360Values_Service_Selector
}

type Coredns1360Values_Coredns1360Values_ServiceAccount_Annotations struct {
}

type Coredns1360Values_ServiceAccount struct {
	Create any
	//  The name of the ServiceAccount to use
	//  If not set and create is true, a name is generated using the fullname template
	Name        any
	Annotations Coredns1360Values_Coredns1360Values_ServiceAccount_Annotations
}

type Coredns1360Values_Rbac struct {
	//  If true, create & use RBAC resources
	Create any
	//  If true, create and use PodSecurityPolicy
	//  The name of the ServiceAccount to use.
	//  If not set and create is true, a name is generated using the fullname template
	//  name:
	PspEnable any
}

type Coredns1360Values_PodSecurityContext struct {
}

type Coredns1360Values_Coredns1360Values_SecurityContext_Coredns1360Values_Coredns1360Values_SecurityContext_Capabilities_AddItem struct {
}

type Coredns1360Values_Coredns1360Values_SecurityContext_Capabilities struct {
	Add []Coredns1360Values_Coredns1360Values_SecurityContext_Coredns1360Values_Coredns1360Values_SecurityContext_Capabilities_AddItem
}

type Coredns1360Values_SecurityContext struct {
	Capabilities Coredns1360Values_Coredns1360Values_SecurityContext_Capabilities
}

type Coredns1360Values_Coredns1360Values_ServersItem_ZonesItem struct {
	Zone any
}

type Coredns1360Values_Coredns1360Values_ServersItem_PluginsItem struct {
	Name any
}

type Coredns1360Values_ServersItem struct {
	Zones []Coredns1360Values_Coredns1360Values_ServersItem_ZonesItem
	Port  any
	//  -- expose the service on a different port
	//  servicePort: 5353
	//  If serviceType is nodePort you can specify nodePort here
	//  nodePort: 30053
	//  hostPort: 53
	Plugins []Coredns1360Values_Coredns1360Values_ServersItem_PluginsItem
}

type Coredns1360Values_ExtraConfig struct {
}

type Coredns1360Values_LivenessProbe struct {
	Enabled             any
	InitialDelaySeconds any
	PeriodSeconds       any
	TimeoutSeconds      any
	FailureThreshold    any
	SuccessThreshold    any
}

type Coredns1360Values_ReadinessProbe struct {
	Enabled             any
	InitialDelaySeconds any
	PeriodSeconds       any
	TimeoutSeconds      any
	FailureThreshold    any
	SuccessThreshold    any
}

type Coredns1360Values_Affinity struct {
}

type Coredns1360Values_NodeSelector struct {
}

type Coredns1360Values_PodDisruptionBudget struct {
}

type Coredns1360Values_CustomLabels struct {
}

type Coredns1360Values_CustomAnnotations struct {
}

type Coredns1360Values_Hpa struct {
	Enabled     any
	MinReplicas any
	MaxReplicas any
	Metrics     []any
}

type Coredns1360Values_Coredns1360Values_Autoscaler_PodAnnotations struct {
}

type Coredns1360Values_Coredns1360Values_Autoscaler_Image struct {
	Repository any
	Tag        any
	PullPolicy any
	//  Optionally specify an array of imagePullSecrets.
	//  Secrets must be manually created in the namespace.
	//  ref: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
	//
	//  pullSecrets:
	//    - name: myRegistryKeySecretName
	PullSecrets []any
}

type Coredns1360Values_Coredns1360Values_Autoscaler_Affinity struct {
}

type Coredns1360Values_Coredns1360Values_Autoscaler_NodeSelector struct {
}

type Coredns1360Values_Coredns1360Values_Autoscaler_Coredns1360Values_Coredns1360Values_Autoscaler_Resources_Requests struct {
	Cpu    any
	Memory any
}

type Coredns1360Values_Coredns1360Values_Autoscaler_Coredns1360Values_Coredns1360Values_Autoscaler_Resources_Limits struct {
	Cpu    any
	Memory any
}

type Coredns1360Values_Coredns1360Values_Autoscaler_Resources struct {
	Requests Coredns1360Values_Coredns1360Values_Autoscaler_Coredns1360Values_Coredns1360Values_Autoscaler_Resources_Requests
	Limits   Coredns1360Values_Coredns1360Values_Autoscaler_Coredns1360Values_Coredns1360Values_Autoscaler_Resources_Limits
}

type Coredns1360Values_Coredns1360Values_Autoscaler_Coredns1360Values_Coredns1360Values_Autoscaler_Configmap_Annotations struct {
}

type Coredns1360Values_Coredns1360Values_Autoscaler_Configmap struct {
	//  Annotations for the coredns-autoscaler configmap
	//  i.e. strategy.spinnaker.io/versioned: "false" to ensure configmap isn't renamed
	Annotations Coredns1360Values_Coredns1360Values_Autoscaler_Coredns1360Values_Coredns1360Values_Autoscaler_Configmap_Annotations
}

type Coredns1360Values_Coredns1360Values_Autoscaler_LivenessProbe struct {
	Enabled             any
	InitialDelaySeconds any
	PeriodSeconds       any
	TimeoutSeconds      any
	FailureThreshold    any
	SuccessThreshold    any
}

type Coredns1360Values_Autoscaler struct {
	//  Enabled the cluster-proportional-autoscaler
	Enabled any
	//  Number of cores in the cluster per coredns replica
	CoresPerReplica any
	//  Number of nodes in the cluster per coredns replica
	NodesPerReplica any
	//  Min size of replicaCount
	Min any
	//  Max size of replicaCount (default of 0 is no max)
	Max any
	//  Whether to include unschedulable nodes in the nodes/cores calculations - this requires version 1.8.0+ of the autoscaler
	IncludeUnschedulableNodes any
	//  If true does not allow single points of failure to form
	PreventSinglePointFailure any
	//  Annotations for the coredns proportional autoscaler pods
	PodAnnotations Coredns1360Values_Coredns1360Values_Autoscaler_PodAnnotations
	//  Optionally specify some extra flags to pass to cluster-proprtional-autoscaler.
	//  Useful for e.g. the nodelabels flag.
	//  customFlags:
	//    - --nodelabels=topology.kubernetes.io/zone=us-east-1a
	//
	Image Coredns1360Values_Coredns1360Values_Autoscaler_Image
	//  Optional priority class to be used for the autoscaler pods. priorityClassName used if not set.
	PriorityClassName any
	//  expects input structure as per specification https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#affinity-v1-core
	Affinity Coredns1360Values_Coredns1360Values_Autoscaler_Affinity
	//  Node labels for pod assignment
	//  Ref: https://kubernetes.io/docs/user-guide/node-selection/
	NodeSelector Coredns1360Values_Coredns1360Values_Autoscaler_NodeSelector
	//  expects input structure as per specification https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#toleration-v1-core
	Tolerations []any
	//  resources for autoscaler pod
	Resources Coredns1360Values_Coredns1360Values_Autoscaler_Resources
	//  Options for autoscaler configmap
	Configmap Coredns1360Values_Coredns1360Values_Autoscaler_Configmap
	//  Enables the livenessProbe for cluster-proportional-autoscaler - this requires version 1.8.0+ of the autoscaler
	LivenessProbe Coredns1360Values_Coredns1360Values_Autoscaler_LivenessProbe
	//  optional array of sidecar containers
	//  - name: some-container-name
	//    image: some-image:latest
	//    imagePullPolicy: Always
	ExtraContainers []any
}

type Coredns1360Values_Coredns1360Values_Deployment_Annotations struct {
}

type Coredns1360Values_Coredns1360Values_Deployment_Selector struct {
}

type Coredns1360Values_Deployment struct {
	SkipConfig any
	Enabled    any
	Name       any
	//  Annotations for the coredns deployment
	Annotations Coredns1360Values_Coredns1360Values_Deployment_Annotations
	//  Pod selector
	Selector Coredns1360Values_Coredns1360Values_Deployment_Selector
}

type Coredns1360Values struct {
	Image                         Coredns1360Values_Image
	ReplicaCount                  any
	Resources                     Coredns1360Values_Resources
	RollingUpdate                 Coredns1360Values_RollingUpdate
	TerminationGracePeriodSeconds any
	//   cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
	PodAnnotations Coredns1360Values_PodAnnotations
	ServiceType    any
	Prometheus     Coredns1360Values_Prometheus
	Service        Coredns1360Values_Service
	ServiceAccount Coredns1360Values_ServiceAccount
	Rbac           Coredns1360Values_Rbac
	//  isClusterService specifies whether chart should be deployed as cluster-service or normal k8s app.
	IsClusterService any
	//  Optional priority class to be used for the coredns pods. Used for autoscaler if autoscaler.priorityClassName not set.
	PriorityClassName any
	//  Configure the pod level securityContext.
	PodSecurityContext Coredns1360Values_PodSecurityContext
	//  Configure SecurityContext for Pod.
	//  Ensure that required linux capability to bind port number below 1024 is assigned (`CAP_NET_BIND_SERVICE`).
	SecurityContext Coredns1360Values_SecurityContext
	//  Default zone is what Kubernetes recommends:
	//  https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/#coredns-configmap-options
	Servers []Coredns1360Values_ServersItem
	//  Complete example with all the options:
	//  - zones:                 # the `zones` block can be left out entirely, defaults to "."
	//    - zone: hello.world.   # optional, defaults to "."
	//      scheme: tls://       # optional, defaults to "" (which equals "dns://" in CoreDNS)
	//    - zone: foo.bar.
	//      scheme: dns://
	//      use_tcp: true        # set this parameter to optionally expose the port on tcp as well as udp for the DNS protocol
	//                           # Note that this will not work if you are also exposing tls or grpc on the same server
	//    port: 12345            # optional, defaults to "" (which equals 53 in CoreDNS)
	//    plugins:               # the plugins to use for this server block
	//    - name: kubernetes     # name of plugin, if used multiple times ensure that the plugin supports it!
	//      parameters: foo bar  # list of parameters after the plugin
	//      configBlock: |-      # if the plugin supports extra block style config, supply it here
	//        hello world
	//        foo bar
	//
	//  Extra configuration that is applied outside of the default zone block.
	//  Example to include additional config files, which may come from extraVolumes:
	//  extraConfig:
	//    import:
	//      parameters: /opt/coredns/*.conf
	ExtraConfig Coredns1360Values_ExtraConfig
	//  To use the livenessProbe, the health plugin needs to be enabled in CoreDNS' server config
	LivenessProbe Coredns1360Values_LivenessProbe
	//  To use the readinessProbe, the ready plugin needs to be enabled in CoreDNS' server config
	ReadinessProbe Coredns1360Values_ReadinessProbe
	//  expects input structure as per specification https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#affinity-v1-core
	//  for example:
	//    affinity:
	//      nodeAffinity:
	//       requiredDuringSchedulingIgnoredDuringExecution:
	//         nodeSelectorTerms:
	//         - matchExpressions:
	//           - key: foo.bar.com/role
	//             operator: In
	//             values:
	//             - master
	Affinity Coredns1360Values_Affinity
	//  expects input structure as per specification https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#topologyspreadconstraint-v1-core
	//  and supports Helm templating.
	//  For example:
	//    topologySpreadConstraints:
	//      - labelSelector:
	//          matchLabels:
	//            app.kubernetes.io/name: '{{ template "coredns.name" . }}'
	//            app.kubernetes.io/instance: '{{ .Release.Name }}'
	//        topologyKey: topology.kubernetes.io/zone
	//        maxSkew: 1
	//        whenUnsatisfiable: ScheduleAnyway
	//      - labelSelector:
	//          matchLabels:
	//            app.kubernetes.io/name: '{{ template "coredns.name" . }}'
	//            app.kubernetes.io/instance: '{{ .Release.Name }}'
	//        topologyKey: kubernetes.io/hostname
	//        maxSkew: 1
	//        whenUnsatisfiable: ScheduleAnyway
	TopologySpreadConstraints []any
	//  Node labels for pod assignment
	//  Ref: https://kubernetes.io/docs/user-guide/node-selection/
	NodeSelector Coredns1360Values_NodeSelector
	//  expects input structure as per specification https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#toleration-v1-core
	//  for example:
	//    tolerations:
	//    - key: foo.bar.com/role
	//      operator: Equal
	//      value: master
	//      effect: NoSchedule
	Tolerations []any
	//  https://kubernetes.io/docs/tasks/run-application/configure-pdb/#specifying-a-poddisruptionbudget
	PodDisruptionBudget Coredns1360Values_PodDisruptionBudget
	//  configure custom zone files as per https://coredns.io/2017/05/08/custom-dns-entries-for-kubernetes/
	//   - filename: example.db
	//     domain: example.com
	//     contents: |
	//       example.com.   IN SOA sns.dns.icann.com. noc.dns.icann.com. 2015082541 7200 3600 1209600 3600
	//       example.com.   IN NS  b.iana-servers.net.
	//       example.com.   IN NS  a.iana-servers.net.
	//       example.com.   IN A   192.168.99.102
	//       *.example.com. IN A   192.168.99.102
	ZoneFiles []any
	//  optional array of sidecar containers
	ExtraContainers []any
	//  - name: some-container-name
	//    image: some-image:latest
	//    imagePullPolicy: Always
	//  optional array of extra volumes to create
	ExtraVolumes []any
	//  - name: some-volume-name
	//    emptyDir: {}
	//  optional array of mount points for extraVolumes
	//  - name: some-volume-name
	//    mountPath: /etc/wherever
	ExtraVolumeMounts []any
	//  optional array of secrets to mount inside coredns container
	//  possible usecase: need for secure connection with etcd backend
	//  - name: etcd-client-certs
	//    mountPath: /etc/coredns/tls/etcd
	//    defaultMode: 420
	//  - name: some-fancy-secret
	//    mountPath: /etc/wherever
	//    defaultMode: 440
	ExtraSecrets []any
	//  optional array of environment variables for coredns container
	//  possible usecase: provides username and password for etcd user authentications
	//  - name: WHATEVER_ENV
	//    value: whatever
	//  - name: SOME_SECRET_ENV
	//    valueFrom:
	//      secretKeyRef:
	//        name: some-secret-name
	//        key: secret-key
	Env []any
	//  To support legacy deployments using CoreDNS with the "k8s-app: kube-dns" label selectors.
	//  See https://github.com/coredns/helm/blob/master/charts/coredns/README.md#adopting-existing-coredns-resources
	//  k8sAppLabelOverride: "kube-dns"
	//
	//  Custom labels to apply to Deployment, Pod, Configmap, Service, ServiceMonitor. Including autoscaler if enabled.
	CustomLabels Coredns1360Values_CustomLabels
	//  Custom annotations to apply to Deployment, Pod, Configmap, Service, ServiceMonitor. Including autoscaler if enabled.
	CustomAnnotations Coredns1360Values_CustomAnnotations
	//  Alternative configuration for HPA deployment if wanted
	//  Create HorizontalPodAutoscaler object.
	//
	//  hpa:
	//    enabled: false
	//    minReplicas: 1
	//    maxReplicas: 10
	//    metrics:
	//     metrics:
	//     - type: Resource
	//       resource:
	//         name: memory
	//         target:
	//           type: Utilization
	//           averageUtilization: 60
	//     - type: Resource
	//       resource:
	//         name: cpu
	//         target:
	//           type: Utilization
	//           averageUtilization: 60
	//
	Hpa Coredns1360Values_Hpa
	//  Configue a cluster-proportional-autoscaler for coredns
	//  See https://github.com/kubernetes-incubator/cluster-proportional-autoscaler
	Autoscaler Coredns1360Values_Autoscaler
	Deployment Coredns1360Values_Deployment
}
