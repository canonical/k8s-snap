package controllers_test

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/utils"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	. "github.com/onsi/gomega"
)

// channelSendTimeout is the timeout for pushing to channels for TestControlPlaneConfigController
const channelSendTimeout = 100 * time.Millisecond

type configProvider struct {
	config  types.ClusterConfig
	nodeIPs []string
}

func (c *configProvider) getConfigAndNodeIPs(ctx context.Context) (types.ClusterConfig, []string, error) {
	return c.config, c.nodeIPs, nil
}

func (c *configProvider) getConfig(ctx context.Context) (types.ClusterConfig, error) {
	return c.config, nil
}

func TestControlPlaneConfigController(t *testing.T) {
	t.Run("ControlPlane", func(t *testing.T) {
		dir := t.TempDir()

		s := &mock.Snap{
			Mock: mock.Mock{
				EtcdPKIDir:          path.Join(dir, "etcd-pki"),
				KubernetesPKIDir:    path.Join(dir, "kube-pki"),
				ServiceArgumentsDir: path.Join(dir, "args"),
				UID:                 os.Getuid(),
				GID:                 os.Getgid(),
			},
		}

		g := NewWithT(t)
		g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		triggerCh := make(chan time.Time)
		configProvider := &configProvider{
			nodeIPs: []string{"10.0.0.1", "10.0.0.2"},
		}

		ctrl := controllers.NewControlPlaneConfigurationController(s, func() {}, triggerCh)
		go ctrl.Run(ctx, configProvider.getConfigAndNodeIPs)

		for _, tc := range []struct {
			name   string
			config types.ClusterConfig

			expectKubeAPIServerArgs         map[string]string
			expectKubeControllerManagerArgs map[string]string

			expectServiceRestarts []string
			expectFilesToExist    map[string]bool
		}{
			{
				name: "Default",
				config: types.ClusterConfig{
					Datastore: types.Datastore{
						Type:            utils.Pointer("external"),
						ExternalServers: utils.Pointer([]string{"http://127.0.0.1:2379"}),
					},
				},
				expectKubeAPIServerArgs: map[string]string{
					"--etcd-servers": "http://127.0.0.1:2379",
				},
				expectFilesToExist: map[string]bool{
					path.Join(dir, "etcd-pki", "ca.crt"):     false,
					path.Join(dir, "etcd-pki", "client.crt"): false,
					path.Join(dir, "etcd-pki", "client.key"): false,
				},
				expectServiceRestarts: []string{"kube-apiserver"},
			},
			{
				name: "Certs",
				config: types.ClusterConfig{
					Datastore: types.Datastore{
						Type:               utils.Pointer("external"),
						ExternalServers:    utils.Pointer([]string{"https://127.0.0.1:2379"}),
						ExternalCACert:     utils.Pointer("CA DATA"),
						ExternalClientCert: utils.Pointer("CERT DATA"),
						ExternalClientKey:  utils.Pointer("KEY DATA"),
					},
				},
				expectKubeAPIServerArgs: map[string]string{
					"--etcd-servers":  "https://127.0.0.1:2379",
					"--etcd-cafile":   path.Join(dir, "etcd-pki", "ca.crt"),
					"--etcd-certfile": path.Join(dir, "etcd-pki", "client.crt"),
					"--etcd-keyfile":  path.Join(dir, "etcd-pki", "client.key"),
				},
				expectFilesToExist: map[string]bool{
					path.Join(dir, "etcd-pki", "ca.crt"):     true,
					path.Join(dir, "etcd-pki", "client.crt"): true,
					path.Join(dir, "etcd-pki", "client.key"): true,
				},
				expectServiceRestarts: []string{"kube-apiserver"},
			},
			{
				name: "CloudProvider",
				config: types.ClusterConfig{
					Kubelet: types.Kubelet{
						CloudProvider: utils.Pointer("external"),
					},
				},
				expectKubeControllerManagerArgs: map[string]string{
					"--cloud-provider": "external",
				},
				expectServiceRestarts: []string{"kube-controller-manager"},
			},
			{
				name: "NoUpdates",
				config: types.ClusterConfig{
					Datastore: types.Datastore{
						Type:               utils.Pointer("external"),
						ExternalServers:    utils.Pointer([]string{"https://127.0.0.1:2379"}),
						ExternalCACert:     utils.Pointer("CA DATA"),
						ExternalClientCert: utils.Pointer("CERT DATA"),
						ExternalClientKey:  utils.Pointer("KEY DATA"),
					},
					Kubelet: types.Kubelet{
						CloudProvider: utils.Pointer("external"),
					},
				},
				expectKubeAPIServerArgs: map[string]string{
					"--etcd-servers":  "https://127.0.0.1:2379",
					"--etcd-cafile":   path.Join(dir, "etcd-pki", "ca.crt"),
					"--etcd-certfile": path.Join(dir, "etcd-pki", "client.crt"),
					"--etcd-keyfile":  path.Join(dir, "etcd-pki", "client.key"),
				},
				expectFilesToExist: map[string]bool{
					path.Join(dir, "etcd-pki", "ca.crt"):     true,
					path.Join(dir, "etcd-pki", "client.crt"): true,
					path.Join(dir, "etcd-pki", "client.key"): true,
				},
				expectKubeControllerManagerArgs: map[string]string{
					"--cloud-provider": "external",
				},
			},
			{
				name: "UpdateAll",
				config: types.ClusterConfig{
					Datastore: types.Datastore{
						Type:            utils.Pointer("external"),
						ExternalServers: utils.Pointer([]string{"http://127.0.0.1:2379"}),
					},
					Kubelet: types.Kubelet{
						CloudProvider: utils.Pointer(""),
					},
				},
				expectKubeAPIServerArgs: map[string]string{
					"--etcd-servers":  "http://127.0.0.1:2379",
					"--etcd-cafile":   "",
					"--etcd-certfile": "",
					"--etcd-keyfile":  "",
				},
				expectFilesToExist: map[string]bool{
					path.Join(dir, "etcd-pki", "ca.crt"):     false,
					path.Join(dir, "etcd-pki", "client.crt"): false,
					path.Join(dir, "etcd-pki", "client.key"): false,
				},
				expectKubeControllerManagerArgs: map[string]string{
					"--cloud-provider": "",
				},
				expectServiceRestarts: []string{"kube-apiserver", "kube-controller-manager"},
			},
			{
				name: "Embedded",
				config: types.ClusterConfig{
					Datastore: types.Datastore{
						Type:         utils.Pointer("embedded"),
						EmbeddedPort: utils.Pointer(12379),
					},
				},
				expectKubeAPIServerArgs: map[string]string{
					"--etcd-servers":  "https://10.0.0.1:12379,https://10.0.0.2:12379",
					"--etcd-cafile":   path.Join(dir, "etcd-pki", "ca.crt"),
					"--etcd-certfile": path.Join(dir, "kube-pki", "apiserver-etcd-client.crt"),
					"--etcd-keyfile":  path.Join(dir, "kube-pki", "apiserver-etcd-client.key"),
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)

				s.RestartServiceCalledWith = nil

				configProvider.config = tc.config

				select {
				case triggerCh <- time.Now():
				case <-time.After(channelSendTimeout):
					g.Fail("Timed out while attempting to trigger controller reconcile loop")
				}

				// TODO: this should be changed to call g.Eventually()
				<-time.After(50 * time.Millisecond)

				g.Expect(s.RestartServiceCalledWith).To(ConsistOf(tc.expectServiceRestarts))

				t.Run("APIServerArgs", func(t *testing.T) {
					for earg, eval := range tc.expectKubeAPIServerArgs {
						t.Run(earg, func(t *testing.T) {
							g := NewWithT(t)

							val, err := snaputil.GetServiceArgument(s, "kube-apiserver", earg)
							g.Expect(err).To(BeNil())
							g.Expect(val).To(Equal(eval))
						})
					}
				})

				t.Run("KubeControllerManagerArgs", func(t *testing.T) {
					for earg, eval := range tc.expectKubeControllerManagerArgs {
						t.Run(earg, func(t *testing.T) {
							g := NewWithT(t)

							val, err := snaputil.GetServiceArgument(s, "kube-controller-manager", earg)
							g.Expect(err).To(BeNil())
							g.Expect(val).To(Equal(eval))
						})
					}
				})

				t.Run("Certs", func(t *testing.T) {
					for file, mustExist := range tc.expectFilesToExist {
						t.Run(path.Base(file), func(t *testing.T) {
							g := NewWithT(t)

							_, err := os.Stat(file)
							if mustExist {
								g.Expect(err).To(BeNil())
							} else {
								g.Expect(err).To(MatchError(os.ErrNotExist))
							}
						})
					}
				})
			})
		}
	})

	t.Run("Worker", func(t *testing.T) {
		dir := t.TempDir()

		s := &mock.Snap{
			Mock: mock.Mock{
				EtcdPKIDir:          path.Join(dir, "etcd-pki"),
				ServiceArgumentsDir: path.Join(dir, "args"),
				LockFilesDir:        path.Join(dir, "locks"),
				UID:                 os.Getuid(),
				GID:                 os.Getgid(),
			},
		}

		g := NewWithT(t)
		g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		triggerCh := make(chan time.Time)
		configProvider := &configProvider{}

		ctrl := controllers.NewControlPlaneConfigurationController(s, func() {}, triggerCh)
		go ctrl.Run(ctx, configProvider.getConfigAndNodeIPs)

		// mark as worker node
		g.Expect(snaputil.MarkAsWorkerNode(s, true)).To(Succeed())

		configProvider.config = types.ClusterConfig{
			Datastore: types.Datastore{
				Type:               utils.Pointer("external"),
				ExternalServers:    utils.Pointer([]string{"https://127.0.0.1:2379"}),
				ExternalCACert:     utils.Pointer("CA DATA"),
				ExternalClientCert: utils.Pointer("CERT DATA"),
				ExternalClientKey:  utils.Pointer("KEY DATA"),
			},
			Kubelet: types.Kubelet{
				CloudProvider: utils.Pointer("external"),
			},
		}

		select {
		case triggerCh <- time.Now():
		case <-time.After(channelSendTimeout):
			g.Fail("Timed out while attempting to trigger controller reconcile loop")
		}

		// TODO: this should be changed to call g.Eventually()
		<-time.After(50 * time.Millisecond)

		g.Expect(s.RestartServiceCalledWith).To(BeEmpty())

		t.Run("APIServerArgs", func(t *testing.T) {
			for _, arg := range []string{"--etcd-servers", "--etcd-cafile", "--etcd-certfile", "--etcd-keyfile"} {
				t.Run(arg, func(t *testing.T) {
					g := NewWithT(t)

					val, err := snaputil.GetServiceArgument(s, "kube-apiserver", "--etcd-servers")
					g.Expect(err).To(HaveOccurred())
					g.Expect(val).To(BeEmpty())
				})
			}
		})

		t.Run("KubeControllerManagerArgs", func(t *testing.T) {
			g := NewWithT(t)

			val, err := snaputil.GetServiceArgument(s, "kube-controller-manager", "--cloud-provider")
			g.Expect(err).To(HaveOccurred())
			g.Expect(val).To(BeEmpty())
		})

		t.Run("Certs", func(t *testing.T) {
			for _, cert := range []string{"ca.crt", "client.crt", "client.key"} {
				t.Run(cert, func(t *testing.T) {
					g := NewWithT(t)

					_, err := os.Stat(path.Join(dir, "etcd-pki", cert))
					g.Expect(err).To(MatchError(os.ErrNotExist))
				})
			}
		})
	})
}
