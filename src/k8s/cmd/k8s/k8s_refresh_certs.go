package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newRefreshCertsCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		certificates  []string
		externalCerts string
		extraSANs     []string
		expiresIn     string
		timeout       time.Duration
	}
	cmd := &cobra.Command{
		Use:   "refresh-certs",
		Short: "Refresh the certificates of the running node",
		Args:  cobra.NoArgs,
		PreRun: chainPreRunHooks(hookRequireRoot(env), func(cmd *cobra.Command, args []string) {
			if opts.externalCerts == "" {
				if opts.expiresIn == "" {
					cmd.PrintErrln("Error: the --expires-in flag is required when not using --external-certificates.")
					env.Exit(1)
					return
				}
			} else {
				if opts.expiresIn != "" || len(opts.extraSANs) > 0 || len(opts.certificates) > 0 {
					cmd.PrintErrln("Error: --external-certificates cannot be used together with --expires-in, --extra-sans, or --certificates.")
					env.Exit(1)
					return
				}
			}
		}),
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

			// Check if we are dealing with external certs.
			if opts.externalCerts != "" {
				certificates, err := getCertificatesFromYAML(env, opts.externalCerts)
				if err != nil {
					cmd.PrintErrf("Error: Failed to parse certificates file: %v\n", err)
					env.Exit(1)
					return
				}
				if _, err := client.RefreshCertificatesUpdate(ctx, certificates); err != nil {
					cmd.PrintErrf("Error: Failed to refresh external certificates: %v\n", err)
					env.Exit(1)
					return
				}
				cmd.Println("External certificates have been successfully updated.")
				return
			}

			// Default internal certificates refresh.
			ttl, err := utils.TTLToSeconds(opts.expiresIn)
			if err != nil {
				cmd.PrintErrf("Error: Failed to parse TTL. \n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			plan, err := client.RefreshCertificatesPlan(ctx, apiv1.RefreshCertificatesPlanRequest{
				Certificates: opts.certificates,
			})
			if err != nil {
				cmd.PrintErrf("Error: Failed to get the certificates refresh plan.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if len(plan.CertificateSigningRequests) > 0 {
				cmd.Println("The following CertificateSigningRequests should be approved. Run the following commands on any of the control plane nodes of the cluster:")
				for _, csr := range plan.CertificateSigningRequests {
					cmd.Printf("k8s kubectl certificate approve %s\n", csr)
				}
			}

			runRequest := apiv1.RefreshCertificatesRunRequest{
				Certificates:      opts.certificates,
				Seed:              plan.Seed,
				ExpirationSeconds: ttl,
				ExtraSANs:         opts.extraSANs,
			}

			stopHB := cmdutil.StartSpinner(ctx, cmd.ErrOrStderr(), "Waiting for certificates to be created...")

			runResponse, err := client.RefreshCertificatesRun(ctx, runRequest)
			// stop spinner before printing final output or error
			stopHB()
			if err != nil {
				cmd.PrintErrf("Error: Failed to refresh the certificates.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			expiryTimeUNIX := time.Unix(int64(runResponse.ExpirationSeconds), 0)
			cmd.Printf("Certificates have been successfully refreshed, and will expire at %v.\n", expiryTimeUNIX)
		},
	}

	certificateOpts := fmt.Sprintf("Worker nodes: %s\nControl Plane nodes: %s",
		formatCertificatesList(apiv1.ClusterRoleWorker),
		formatCertificatesList(apiv1.ClusterRoleControlPlane),
	)
	cmd.Flags().StringSliceVar(&opts.certificates, "certificates", []string{}, fmt.Sprintf("List of certificates to renew in the cluster (must be used with --expires-in). Defaults to all certificates.\nAllowed values:\n%s", certificateOpts))
	cmd.Flags().StringVar(&opts.externalCerts, "external-certificates", "", "path to a YAML file containing external certificate data in PEM format. If the cluster was bootstrapped with external certificates, the certificates will be updated. Use '-' to read from stdin.")
	cmd.Flags().StringVar(&opts.expiresIn, "expires-in", "", "the time until the certificates expire, e.g., 1h, 2d, 4mo, 5y. Aditionally, any valid time unit for ParseDuration is accepted.")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")
	cmd.Flags().StringArrayVar(&opts.extraSANs, "extra-sans", []string{}, "extra SANs to add to the certificates.")

	return cmd
}

// formatCertificatesList returns a comma separated string of the certificates
// names for a specific cluster role.
func formatCertificatesList(nodeRole apiv1.ClusterRole) string {
	certs, found := apiv1.CertificatesByRole[nodeRole]
	if !found {
		return ""
	}

	var parts []string
	for cert := range certs {
		parts = append(parts, string(cert))
	}
	slices.Sort(parts)
	return strings.Join(parts, ", ")
}

func getCertificatesFromYAML(env cmdutil.ExecutionEnvironment, filePath string) (apiv1.RefreshCertificatesUpdateRequest, error) {
	var b []byte
	var err error

	if filePath == "-" {
		b, err = io.ReadAll(env.Stdin)
		if err != nil {
			return apiv1.RefreshCertificatesUpdateRequest{}, fmt.Errorf("failed to read certificates from stdin: %w", err)
		}
	} else {
		b, err = os.ReadFile(filePath)
		if err != nil {
			return apiv1.RefreshCertificatesUpdateRequest{}, fmt.Errorf("failed to read file: %w", err)
		}
	}

	var config apiv1.RefreshCertificatesUpdateRequest
	if err := yaml.UnmarshalStrict(b, &config); err != nil {
		return apiv1.RefreshCertificatesUpdateRequest{}, fmt.Errorf("failed to parse YAML certificates file: %w", err)
	}

	return config, nil
}
