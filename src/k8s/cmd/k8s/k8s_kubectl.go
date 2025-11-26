package k8s

import (
	"os"
	"path/filepath"
	"syscall"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newKubectlCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	return &cobra.Command{
		Use:                "kubectl",
		Short:              "Integrated Kubernetes kubectl client",
		DisableFlagParsing: true,
		PreRun:             chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			binary := filepath.Join(env.Snap.K8sBinDir(), "kubectl")
			kubeconfigEnvKey := "KUBECONFIG"
			adminKubeconfigPath := filepath.Join(env.Snap.KubernetesConfigDir(), "admin.conf")
			kubeletKubeconfigPath := filepath.Join(env.Snap.KubernetesConfigDir(), "kubelet.conf")
			var kubeconfigFallback string

			if _, err := os.Stat(adminKubeconfigPath); err == nil {
				kubeconfigFallback = adminKubeconfigPath
			} else if _, err := os.Stat(kubeletKubeconfigPath); err == nil {
				kubeconfigFallback = kubeletKubeconfigPath
			} else if !cmdutil.ExistsInEnviron(env.Environ, kubeconfigEnvKey) {
				cmd.PrintErrf("Error: %s and %s do not exist; please set KUBECONFIG.\n", adminKubeconfigPath, kubeletKubeconfigPath)
				env.Exit(1)
				return
			}

			command := append([]string{"kubectl"}, args...)
			keyValues := []string{"EDITOR", "nano"}
			if kubeconfigFallback != "" {
				keyValues = append([]string{kubeconfigEnvKey, kubeconfigFallback}, keyValues...)
			}
			environ := cmdutil.EnvironWithDefaults(
				env.Environ,
				keyValues...,
			)
			if err := syscall.Exec(binary, command, environ); err != nil {
				cmd.PrintErrf("Failed to run %s.\n\nThe error was: %v\n", command, err)
				env.Exit(1)
				return
			}
		},
	}
}
