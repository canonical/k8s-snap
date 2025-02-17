package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newUpdateCertsCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		file    string
		timeout time.Duration
	}
	cmd := &cobra.Command{
		Use:    "update-certs",
		Short:  "Update the running node's certificates with user provided certificates",
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)

			if _, initialized, err := client.NodeStatus(cmd.Context()); err != nil {
				cmd.PrintErrf("Error: Failed to check the current node status.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			} else if !initialized {
				cmd.PrintErrln("Error: The node is not part of a Kubernetes cluster. You can bootstrap a new cluster with:\n\n  sudo k8s bootstrap")
				env.Exit(1)
				return
			}

			config, err := getCertificatesFromYAML(env, opts.file)
			if err != nil {
				cmd.PrintErrf("Error: Failed to get the certificates from the YAML file.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			_, err = client.UpdateCertificates(ctx, config)
			if err != nil {
				cmd.PrintErrf("Error: Failed to update the certificates.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			cmd.Printf("Certificates updated successfully. The node will be restarted to apply the changes.\n")
		},
	}
	cmd.Flags().StringVar(&opts.file, "file", "", "path to the YAML file containing all certificates and key pairs.")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")
	cmd.MarkFlagRequired("file")

	return cmd
}

func getCertificatesFromYAML(env cmdutil.ExecutionEnvironment, filePath string) (apiv1.UpdateCertificatesRequest, error) {
	var b []byte
	var err error

	if filePath == "-" {
		b, err = io.ReadAll(env.Stdin)
		if err != nil {
			return apiv1.UpdateCertificatesRequest{}, fmt.Errorf("failed to read config from stdin: %w", err)
		}
	} else {
		b, err = os.ReadFile(filePath)
		if err != nil {
			return apiv1.UpdateCertificatesRequest{}, fmt.Errorf("failed to read file: %w", err)
		}
	}

	var config apiv1.UpdateCertificatesRequest
	if err := yaml.UnmarshalStrict(b, &config); err != nil {
		return apiv1.UpdateCertificatesRequest{}, fmt.Errorf("failed to parse YAML config file: %w", err)
	}

	return config, nil
}
