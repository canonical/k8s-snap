package k8s

import (
	"bytes"
	_ "embed"
	"path/filepath"
	"testing"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

var (
	//go:embed testdata/bootstrap-config-full.yaml
	bootstrapConfigFull string
	//go:embed testdata/bootstrap-config-some.yaml
	bootstrapConfigSome string
	//go:embed testdata/bootstrap-config-invalid-keys.yaml
	bootstrapConfigInvalidKeys string
)

type testCase struct {
	name           string
	yamlConfig     string
	expectedConfig apiv1.BootstrapConfig
	expectedError  string
}

var testCases = []testCase{
	{
		name:       "FullConfig",
		yamlConfig: bootstrapConfigFull,
		expectedConfig: apiv1.BootstrapConfig{
			ClusterConfig: apiv1.UserFacingClusterConfig{
				Network: apiv1.NetworkConfig{
					Enabled: utils.Pointer(true),
				},
				DNS: apiv1.DNSConfig{
					Enabled:       utils.Pointer(true),
					ClusterDomain: utils.Pointer("cluster.local"),
				},
				Ingress: apiv1.IngressConfig{
					Enabled: utils.Pointer(true),
				},
				LoadBalancer: apiv1.LoadBalancerConfig{
					Enabled: utils.Pointer(true),
					L2Mode:  utils.Pointer(true),
					CIDRs:   utils.Pointer([]string{"10.0.0.0/24", "10.1.0.10-10.1.0.20"}),
				},
				LocalStorage: apiv1.LocalStorageConfig{
					Enabled:   utils.Pointer(true),
					LocalPath: utils.Pointer("/storage/path"),
					Default:   utils.Pointer(false),
				},
				Gateway: apiv1.GatewayConfig{
					Enabled: utils.Pointer(true),
				},
				MetricsServer: apiv1.MetricsServerConfig{
					Enabled: utils.Pointer(true),
				},
				CloudProvider: utils.Pointer("external"),
				Annotations: map[string]string{
					apiv1_annotations.AnnotationSkipCleanupKubernetesNodeOnRemove: "true",
					apiv1_annotations.AnnotationSkipStopServicesOnRemove:          "true",
				},
			},
			ControlPlaneTaints:                 []string{"node-role.kubernetes.io/control-plane:NoSchedule"},
			PodCIDR:                            utils.Pointer("10.100.0.0/16"),
			ServiceCIDR:                        utils.Pointer("10.200.0.0/16"),
			DisableRBAC:                        utils.Pointer(false),
			SecurePort:                         utils.Pointer(6443),
			K8sDqlitePort:                      utils.Pointer(9090),
			DatastoreType:                      utils.Pointer("k8s-dqlite"),
			ExtraSANs:                          []string{"custom.kubernetes"},
			ExtraNodeConfigFiles:               map[string]string{"extra-node-config-file.yaml": "test-file-content"},
			ExtraNodeKubeAPIServerArgs:         map[string]*string{"--extra-kube-apiserver-arg": utils.Pointer("extra-kube-apiserver-value")},
			ExtraNodeKubeControllerManagerArgs: map[string]*string{"--extra-kube-controller-manager-arg": utils.Pointer("extra-kube-controller-manager-value")},
			ExtraNodeKubeSchedulerArgs:         map[string]*string{"--extra-kube-scheduler-arg": utils.Pointer("extra-kube-scheduler-value")},
			ExtraNodeKubeProxyArgs:             map[string]*string{"--extra-kube-proxy-arg": utils.Pointer("extra-kube-proxy-value")},
			ExtraNodeKubeletArgs:               map[string]*string{"--extra-kubelet-arg": utils.Pointer("extra-kubelet-value")},
			ExtraNodeContainerdArgs:            map[string]*string{"--extra-containerd-arg": utils.Pointer("extra-containerd-value")},
			ExtraNodeK8sDqliteArgs:             map[string]*string{"--extra-k8s-dqlite-arg": utils.Pointer("extra-k8s-dqlite-value")},
		},
	},
	{
		name:       "SomeConfig",
		yamlConfig: bootstrapConfigSome,
		expectedConfig: apiv1.BootstrapConfig{
			PodCIDR:     utils.Pointer("10.100.0.0/16"),
			ServiceCIDR: utils.Pointer("10.152.200.0/24"),
		},
	},
	{
		name:          "InvalidKeys",
		yamlConfig:    bootstrapConfigInvalidKeys,
		expectedError: "field cluster-cidr not found in type apiv1.BootstrapConfig",
	},
	{
		name:          "InvalidYAML",
		yamlConfig:    "this is not valid yaml",
		expectedError: "failed to parse YAML config file",
	},
}

func mustAddConfigToTestDir(t *testing.T, configPath string, data string) {
	t.Helper()
	// Create the cluster bootstrap config file
	err := utils.WriteFile(configPath, []byte(data), 0o644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetConfigYaml(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "init.yaml")

			// Add the test case config to the test directory
			mustAddConfigToTestDir(t, configPath, tc.yamlConfig)

			// Get the config from the test directory
			bootstrapConfig, err := getConfigFromYaml(cmdutil.DefaultExecutionEnvironment(), configPath)

			if tc.expectedError != "" {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tc.expectedError))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(bootstrapConfig).To(Equal(tc.expectedConfig))
			}
		})
	}
}

func TestGetConfigFromYaml_Stdin(t *testing.T) {
	g := NewWithT(t)

	input := `secure-port: 5000`

	// Redirect stdin to the mock input
	env := cmdutil.DefaultExecutionEnvironment()
	env.Stdin = bytes.NewBufferString(input)

	// Call the getConfigFromYaml function with "-" as filePath
	config, err := getConfigFromYaml(env, "-")
	g.Expect(err).ToNot(HaveOccurred())

	expectedConfig := apiv1.BootstrapConfig{SecurePort: utils.Pointer(5000)}
	g.Expect(config).To(Equal(expectedConfig))
}
