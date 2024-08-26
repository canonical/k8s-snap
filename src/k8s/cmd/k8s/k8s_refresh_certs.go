package k8s

import (
	"context"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/spf13/cobra"
)

func newRefreshCertsCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		extraSANs []string
		expiresIn string
		timeout   time.Duration
	}
	cmd := &cobra.Command{
		Use:    "refresh-certs",
		Short:  "Refresh the certificates of the running node",
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			ttl, err := utils.TTLToSeconds(opts.expiresIn)
			if err != nil {
				cmd.PrintErrf("Error: Failed to parse TTL. \n\nThe error was: %v\n", err)
			}

			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)

			plan, err := client.RefreshCertificatesPlan(ctx, apiv1.RefreshCertificatesPlanRequest{})
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
				Seed:              plan.Seed,
				ExpirationSeconds: ttl,
				ExtraSANs:         opts.extraSANs,
			}

			cmd.Println("Waiting for certificates to be created...")
			runResponse, err := client.RefreshCertificatesRun(ctx, runRequest)
			if err != nil {
				cmd.PrintErrf("Error: Failed to refresh the certificates.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			expiryTimeUNIX := time.Unix(int64(runResponse.ExpirationSeconds), 0)
			cmd.Printf("Certificates have been successfully refreshed, and will expire at %v.\n", expiryTimeUNIX)
		},
	}
	cmd.Flags().StringVar(&opts.expiresIn, "expires-in", "", "the time until the certificates expire")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")
	cmd.Flags().StringArrayVar(&opts.extraSANs, "extra-sans", []string{}, "extra SANs to add to the certificates.")

	cmd.MarkFlagRequired("expires-in")
	return cmd
}
