package app

import (
	"context"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/canonical/microcluster/v2/state"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

// onPostJoin is called when a control plane node joins the cluster.
// onPostJoin retrieves the cluster config from the database and configures local services.
func (a *App) onPostJoin(ctx context.Context, s state.State, initConfig map[string]string) (rerr error) {
	log := log.FromContext(ctx).WithValues("hook", "postJoin")

	snap := a.Snap()

	// NOTE: Set the notBefore certificate time to the current time.
	notBefore := time.Now()

	// NOTE(neoaggelos): context timeout is passed over configuration, so that hook failures are propagated to the client
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if t := utils.MicroclusterTimeoutFromMap(initConfig); t != 0 {
		ctx, cancel = context.WithTimeout(ctx, t)
		defer cancel()
	}

	joinConfig, err := utils.MicroclusterControlPlaneJoinConfigFromMap(initConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal control plane join config: %w", err)
	}

	cfg, err := databaseutil.GetClusterConfig(ctx, s)
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %w", err)
	}
	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return fmt.Errorf("failed to parse node IP address %q", s.Address().Hostname())
	}

	var localhostAddress string
	if nodeIP.To4() == nil {
		localhostAddress = "[::1]"
	} else {
		localhostAddress = "127.0.0.1"
	}

	// Create directories
	if err := setup.EnsureAllDirectories(snap); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// cfg.Network.ServiceCIDR may be "IPv4CIDR[,IPv6CIDR]". get the first ip from CIDR(s).
	serviceIPs, err := utils.GetKubernetesServiceIPsFromServiceCIDRs(cfg.Network.GetServiceCIDR())
	if err != nil {
		return fmt.Errorf("failed to get IP address(es) from ServiceCIDR %q: %w", cfg.Network.GetServiceCIDR(), err)
	}

	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		// NOTE: Default certificate expiration is set to 20 years.
		certificates := pki.NewK8sDqlitePKI(pki.K8sDqlitePKIOpts{
			Hostname:  s.Name(),
			IPSANs:    []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
			NotBefore: notBefore,
			NotAfter:  notBefore.AddDate(20, 0, 0),
		})
		certificates.K8sDqliteCert = cfg.Datastore.GetK8sDqliteCert()
		certificates.K8sDqliteKey = cfg.Datastore.GetK8sDqliteKey()
		if err := certificates.CompleteCertificates(); err != nil {
			return fmt.Errorf("failed to initialize k8s-dqlite certificates: %w", err)
		}
		if _, err := setup.EnsureK8sDqlitePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write k8s-dqlite certificates: %w", err)
		}
	case "external":
		certificates := &pki.ExternalDatastorePKI{
			DatastoreCACert:     cfg.Datastore.GetExternalCACert(),
			DatastoreClientCert: cfg.Datastore.GetExternalClientCert(),
			DatastoreClientKey:  cfg.Datastore.GetExternalClientKey(),
		}
		if err := certificates.CheckCertificates(); err != nil {
			return fmt.Errorf("failed to initialize external datastore certificates: %w", err)
		}
		if _, err := setup.EnsureExtDatastorePKI(snap, certificates); err != nil {
			return fmt.Errorf("failed to write external datastore certificates: %w", err)
		}
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	// Certificates
	// NOTE: Default certificate expiration is set to 20 years.
	extraIPs, extraNames := utils.SplitIPAndDNSSANs(joinConfig.ExtraSANS)
	certificates := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:                  s.Name(),
		IPSANs:                    append(append([]net.IP{nodeIP}, serviceIPs...), extraIPs...),
		DNSSANs:                   extraNames,
		NotBefore:                 notBefore,
		NotAfter:                  notBefore.AddDate(20, 0, 0),
		IncludeMachineAddressSANs: true,
	})

	// load shared cluster certificates
	certificates.CACert = cfg.Certificates.GetCACert()
	certificates.CAKey = cfg.Certificates.GetCAKey()
	certificates.ClientCACert = cfg.Certificates.GetClientCACert()
	certificates.ClientCAKey = cfg.Certificates.GetClientCAKey()
	certificates.FrontProxyCACert = cfg.Certificates.GetFrontProxyCACert()
	certificates.FrontProxyCAKey = cfg.Certificates.GetFrontProxyCAKey()
	certificates.APIServerKubeletClientCert = cfg.Certificates.GetAPIServerKubeletClientCert()
	certificates.APIServerKubeletClientKey = cfg.Certificates.GetAPIServerKubeletClientKey()
	certificates.ServiceAccountKey = cfg.Certificates.GetServiceAccountKey()
	certificates.K8sdPublicKey = cfg.Certificates.GetK8sdPublicKey()
	certificates.K8sdPrivateKey = cfg.Certificates.GetK8sdPrivateKey()

	// load certificates from joinConfig
	certificates.APIServerCert = joinConfig.GetAPIServerCert()
	certificates.APIServerKey = joinConfig.GetAPIServerKey()
	certificates.FrontProxyClientCert = joinConfig.GetFrontProxyClientCert()
	certificates.FrontProxyClientKey = joinConfig.GetFrontProxyClientKey()
	certificates.KubeletCert = joinConfig.GetKubeletCert()
	certificates.KubeletKey = joinConfig.GetKubeletKey()
	certificates.AdminClientCert = joinConfig.GetAdminClientCert()
	certificates.AdminClientKey = joinConfig.GetAdminClientKey()
	certificates.KubeControllerManagerClientCert = joinConfig.GetKubeControllerManagerClientCert()
	certificates.KubeControllerManagerClientKey = joinConfig.GetKubeControllerManagerClientKey()
	certificates.KubeSchedulerClientCert = joinConfig.GetKubeSchedulerClientCert()
	certificates.KubeSchedulerClientKey = joinConfig.GetKubeSchedulerClientKey()
	certificates.KubeProxyClientCert = joinConfig.GetKubeProxyClientCert()
	certificates.KubeProxyClientKey = joinConfig.GetKubeProxyClientKey()
	certificates.KubeletClientCert = joinConfig.GetKubeletClientCert()
	certificates.KubeletClientKey = joinConfig.GetKubeletClientKey()

	// generate missing certificates
	if err := certificates.CompleteCertificates(); err != nil {
		return fmt.Errorf("failed to initialize control plane certificates: %w", err)
	}

	serviceConfigs := types.K8sServiceConfigs{
		ExtraNodeKubeSchedulerArgs:         joinConfig.ExtraNodeKubeSchedulerArgs,
		ExtraNodeKubeControllerManagerArgs: joinConfig.ExtraNodeKubeControllerManagerArgs,
		ExtraNodeKubeletArgs:               joinConfig.ExtraNodeKubeletArgs,
		ExtraNodeKubeProxyArgs:             joinConfig.ExtraNodeKubeProxyArgs,
	}

	// Pre-init checks
	if err := snap.PreInitChecks(ctx, cfg, serviceConfigs, true); err != nil {
		return fmt.Errorf("pre-init checks failed for joining node: %w", err)
	}

	// Write certificates to disk
	if _, err := setup.EnsureControlPlanePKI(snap, certificates); err != nil {
		return fmt.Errorf("failed to write control plane certificates: %w", err)
	}

	if err := setup.SetupControlPlaneKubeconfigs(snap.KubernetesConfigDir(), localhostAddress, cfg.APIServer.GetSecurePort(), *certificates); err != nil {
		return fmt.Errorf("failed to generate kubeconfigs: %w", err)
	}

	// Configure datastore
	switch cfg.Datastore.GetType() {
	case "k8s-dqlite":
		// TODO(neoaggelos): use cluster.GetInternalClusterMembers() instead
		leader, err := s.Leader()
		if err != nil {
			return fmt.Errorf("failed to get dqlite leader: %w", err)
		}
		members, err := leader.GetClusterMembers(ctx)
		if err != nil {
			return fmt.Errorf("failed to get microcluster members: %w", err)
		}
		cluster := make([]string, len(members))
		for _, member := range members {
			var address string
			if member.Address.Addr().Is6() {
				address = fmt.Sprintf("[%s]", member.Address.Addr())
			} else {
				address = member.Address.Addr().String()
			}
			cluster = append(cluster, fmt.Sprintf("%s:%d", address, cfg.Datastore.GetK8sDqlitePort()))
		}

		address := fmt.Sprintf("%s:%d", utils.ToIPString(nodeIP), cfg.Datastore.GetK8sDqlitePort())
		if err := setup.K8sDqlite(snap, address, cluster, joinConfig.ExtraNodeK8sDqliteArgs); err != nil {
			return fmt.Errorf("failed to configure k8s-dqlite with address=%s cluster=%v: %w", address, cluster, err)
		}
	case "external":
	default:
		return fmt.Errorf("unsupported datastore %s, must be one of %v", cfg.Datastore.GetType(), setup.SupportedDatastores)
	}

	// Configure services
	if err := setup.Containerd(snap, joinConfig.ExtraNodeContainerdConfig, joinConfig.ExtraNodeContainerdArgs); err != nil {
		return fmt.Errorf("failed to configure containerd: %w", err)
	}
	if err := setup.KubeletControlPlane(snap, s.Name(), nodeIP, cfg.Kubelet.GetClusterDNS(), cfg.Kubelet.GetClusterDomain(), cfg.Kubelet.GetCloudProvider(), cfg.Kubelet.GetControlPlaneTaints(), joinConfig.ExtraNodeKubeletArgs); err != nil {
		return fmt.Errorf("failed to configure kubelet: %w", err)
	}
	if err := setup.KubeProxy(ctx, snap, s.Name(), cfg.Network.GetPodCIDR(), localhostAddress, joinConfig.ExtraNodeKubeProxyArgs); err != nil {
		return fmt.Errorf("failed to configure kube-proxy: %w", err)
	}
	if err := setup.KubeControllerManager(snap, joinConfig.ExtraNodeKubeControllerManagerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-controller-manager: %w", err)
	}
	if err := setup.KubeScheduler(snap, joinConfig.ExtraNodeKubeSchedulerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-scheduler: %w", err)
	}
	if err := setup.KubeAPIServer(snap, cfg.APIServer.GetSecurePort(), nodeIP, cfg.Network.GetServiceCIDR(), s.Address().Path("1.0", "kubernetes", "auth", "webhook").String(), true, cfg.Datastore, cfg.APIServer.GetAuthorizationMode(), joinConfig.ExtraNodeKubeAPIServerArgs); err != nil {
		return fmt.Errorf("failed to configure kube-apiserver: %w", err)
	}

	if err := setup.ExtraNodeConfigFiles(snap, joinConfig.ExtraNodeConfigFiles); err != nil {
		return fmt.Errorf("failed to write extra node config files: %w", err)
	}

	if err := snapdconfig.SetSnapdFromK8sd(ctx, cfg.ToUserFacing(), snap); err != nil {
		return fmt.Errorf("failed to set snapd configuration from k8sd: %w", err)
	}

	// Start services
	// This may fail if the node controllers try to restart the services at the same time, hence the retry.
	log.Info("Starting control-plane services")
	if err := control.RetryFor(ctx, 5, 5*time.Second, func() error {
		if err := startControlPlaneServices(ctx, snap, cfg.Datastore.GetType()); err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed after retry: %w", err)
	}

	// Wait until Kube-API server is ready
	if err := waitApiServerReady(ctx, snap); err != nil {
		return fmt.Errorf("failed to wait for kube-apiserver to become ready: %w", err)
	}

	log.Info("Create Kubernetes client")
	k8sClient, err := a.snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	// This is required for backwards compatibility.
	log.Info("Applying custom CRDs")
	if err := k8sClient.ApplyCRDs(ctx, a.snap.K8sCRDDir()); err != nil {
		log.Error(err, "Failed to apply custom CRDs")
	}

	if err := handleRollOutUpgrade(ctx, a.snap, s, k8sClient); err != nil {
		log.Error(err, "Failed to handle rollout-upgrade")
		return fmt.Errorf("failed to handle rollout-upgrade: %w", err)
	}

	return nil
}

func handleRollOutUpgrade(ctx context.Context, snap snap.Snap, s state.State, k8sClient *kubernetes.Client) error {
	log := log.FromContext(ctx).WithValues("step", "rollout-upgrade")

	log.Info("Checking if an upgrade is in progress")
	upgrade, err := k8sClient.GetInProgressUpgrade(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for in-progress upgrade: %w", err)
	}

	thisNodeVersion, err := getNodeVersion(ctx, snap)
	if err != nil {
		return fmt.Errorf("failed to get node Kubernetes version: %w", err)
	}

	nodeVersions, err := k8sClient.NodeVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes versions of cluster: %w", err)
	}

	if upgrade == nil {
		return handleNoUpgradeInProgress(ctx, snap, s, k8sClient, thisNodeVersion, nodeVersions)
	}

	return handleUpgradeInProgress(ctx, s, k8sClient, upgrade, thisNodeVersion, nodeVersions)
}

func getNodeVersion(ctx context.Context, snap snap.Snap) (*versionutil.Version, error) {
	versionStr, err := snap.NodeKubernetesVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node Kubernetes version: %w", err)
	}

	version, err := versionutil.Parse(versionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node Kubernetes version %q: %w", versionStr, err)
	}

	return version, nil
}

func handleNoUpgradeInProgress(ctx context.Context, snap snap.Snap, s state.State, k8sClient *kubernetes.Client, thisNodeVersion *versionutil.Version, nodeVersions map[string]*versionutil.Version) error {
	var clusterK8sVersion *versionutil.Version
	for node, version := range nodeVersions {
		if node == s.Name() {
			continue
		}
		if clusterK8sVersion == nil {
			clusterK8sVersion = version
		}
		if !clusterK8sVersion.EqualTo(version) {
			return fmt.Errorf("the cluster has nodes with different Kubernetes versions %q and %q - upgrade all nodes to the same version before joining a new one", clusterK8sVersion, version)
		}
	}

	if clusterK8sVersion == nil {
		return fmt.Errorf("the cluster has no nodes - cannot determine cluster Kubernetes version")
	}

	if thisNodeVersion.EqualTo(clusterK8sVersion) {
		return nil // Normal join
	}

	if thisNodeVersion.Major() == clusterK8sVersion.Major() && thisNodeVersion.Minor() == clusterK8sVersion.Minor() {
		return fmt.Errorf("the joining node has a different patch version than cluster nodes")
	}

	if thisNodeVersion.Major() == clusterK8sVersion.Major() && thisNodeVersion.Minor() != clusterK8sVersion.Minor() {
		return initiateRollingUpgrade(ctx, snap, s, k8sClient, thisNodeVersion, clusterK8sVersion)
	}

	return nil
}

func initiateRollingUpgrade(ctx context.Context, snap snap.Snap, s state.State, k8sClient *kubernetes.Client, thisNodeVersion, clusterK8sVersion *versionutil.Version) error {
	log := log.FromContext(ctx)
	log.Info("Minor version mismatch between joining node and cluster")

	rev, err := snap.Revision(ctx)
	if err != nil {
		return fmt.Errorf("failed to get snap revision: %w", err)
	}

	var strategy kubernetes.UpgradeStrategy
	if thisNodeVersion.GreaterThan(clusterK8sVersion) {
		log.Info("Joining node has a greater version - rolling upgrade")
		strategy = kubernetes.UpgradeStrategyRollingUpgrade
	} else {
		log.Info("Joining node has a lower version - downgrade")
		strategy = kubernetes.UpgradeStrategyRollingDowngrade
	}

	newUpgrade := kubernetes.NewUpgrade(fmt.Sprintf("cluster-upgrade-to-rev-%s", rev), strategy)
	newUpgrade.Status.UpgradedNodes = []string{s.Name()}
	return k8sClient.CreateUpgrade(ctx, newUpgrade)
}

func handleUpgradeInProgress(ctx context.Context, s state.State, k8sClient *kubernetes.Client, upgrade *kubernetes.Upgrade, thisNodeVersion *versionutil.Version, nodeVersions map[string]*versionutil.Version) error {
	log := log.FromContext(ctx)
	nodeName := s.Name()
	lowest, highest := lowestHighestK8sVersions(nodeVersions)

	switch upgrade.Status.Strategy {
	case kubernetes.UpgradeStrategyRollingUpgrade:
		log.Info("Rolling upgrade in progress")
		if !thisNodeVersion.EqualTo(highest) {
			return fmt.Errorf("joining node version %q needs to match highest version %q", thisNodeVersion, highest)
		}
	case kubernetes.UpgradeStrategyRollingDowngrade:
		log.Info("Rolling downgrade in progress")
		if !thisNodeVersion.EqualTo(lowest) {
			return fmt.Errorf("joining node version %q needs to match lowest version %q", thisNodeVersion, lowest)
		}
	default:
		return fmt.Errorf("unknown upgrade strategy in progress: %q", upgrade.Status.Strategy)
	}

	log.Info("Marking node as upgraded", "node", nodeName)
	upgradedNodes := upgrade.Status.UpgradedNodes
	if !slices.Contains(upgradedNodes, nodeName) {
		upgradedNodes = append(upgradedNodes, nodeName)
	}
	return k8sClient.PatchUpgradeStatus(ctx, upgrade.Name, kubernetes.UpgradeStatus{UpgradedNodes: upgradedNodes})
}

// lowestHighestK8sVersions returns the lowest and highest Kubernetes versions from the given map.
func lowestHighestK8sVersions(k8sVersionMap map[string]*versionutil.Version) (lowest, highest *versionutil.Version) {
	for _, version := range k8sVersionMap {
		if lowest == nil || version.LessThan(lowest) {
			lowest = version
		}
		if highest == nil || version.GreaterThan(highest) {
			highest = version
		}
	}
	return lowest, highest
}
