package setup

// package setup

// // TODO(neoaggelos): there is currently lots of duplicate code in this package, but it lets us move fast
// // for lack of an easy way to unit test things.

// import (
// 	"fmt"
// 	"os"

// 	"github.com/canonical/k8s/pkg/proxy"
// 	"github.com/canonical/k8s/pkg/snap"
// 	"github.com/canonical/k8s/pkg/utils"
// )

// // initServiceArgs configures service arguments on a node.
// // initServiceArgs uses default values from $SNAP/k8s/args/$service
// // initServiceArgs updates arguments on updateArgs (override if arg exists, append otherwise).
// // initServiceArgs removes arguments from deleteArgs.
// // initServiceArgs writes the resulting arguments file at $SNAP_DATA/args/$service.
// // TODO(neoaggelos): this currently duplicates logic from other helpers, we need to return to this.
// func initServiceArgs(snap snap.Snap, service string, updateArgs map[string]string, deleteArgs []string) error {
// 	args, err := utils.ParseArgumentFile(snap.Path("k8s", "args", service))
// 	if err != nil {
// 		return fmt.Errorf("failed to parse default arguments for %s: %w", "kubelet", err)
// 	}
// 	for key, value := range updateArgs {
// 		args[key] = value
// 	}
// 	for _, key := range deleteArgs {
// 		delete(args, key)
// 	}

// 	if err := utils.SerializeArgumentFile(args, snap.DataPath("args", service)); err != nil {
// 		return fmt.Errorf("failed to write arguments file for kubelet: %w", err)
// 	}
// 	return nil
// }

// // InitKubeletArgs configures kubelet on the node.
// func InitKubeletArgs(snap snap.Snap, extraArgs map[string]string, deleteArgs []string) error {
// 	return initServiceArgs(snap, "kubelet", extraArgs, deleteArgs)
// }

// // InitKubeProxyArgs configures kube-proxy on the node.
// func InitKubeProxyArgs(snap snap.Snap, extraArgs map[string]string, deleteArgs []string) error {
// 	return initServiceArgs(snap, "kube-proxy", extraArgs, deleteArgs)
// }

// // RenderKubeletKubeconfig renders the kubeconfig file for kubelet.
// func RenderKubeletKubeconfig(snap snap.Snap, token string, caPEM string) error {
// 	ip := "127.0.0.1"
// 	port := 6443
// 	return renderKubeconfig(snap, token, []byte(caPEM), "/etc/kubernetes/kubelet.conf", &ip, &port)
// }

// // RenderKubeProxyKubeconfig renders the kubeconfig file for kube-proxy.
// func RenderKubeProxyKubeconfig(snap snap.Snap, token string, caPEM string) error {
// 	ip := "127.0.0.1"
// 	port := 6443
// 	return renderKubeconfig(snap, token, []byte(caPEM), "/etc/kubernetes/proxy.conf", &ip, &port)
// }

// // WriteCA writes the CA certificate of the cluster.
// func WriteCA(snap snap.Snap, crt string) error {
// 	return os.WriteFile("/etc/kubernetes/pki/ca.crt", []byte(crt), 0600)
// }

// func InitAPIServerProxy(snap snap.Snap, servers []string) error {
// 	if err := proxy.WriteEndpointsConfig(servers, "/etc/kubernetes/k8s-apiserver-proxy.json"); err != nil {
// 		return fmt.Errorf("failed to write proxy configuration file: %w", err)
// 	}

// 	return initServiceArgs(snap, "k8s-apiserver-proxy", nil, nil)
// }

// // InitContainerdArgs configures kube-proxy on the node.
// func InitContainerdArgs(snap snap.Snap, extraArgs map[string]string, deleteArgs []string) error {
// 	return initServiceArgs(snap, "containerd", extraArgs, deleteArgs)
// }
