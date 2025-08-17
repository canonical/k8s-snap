package k8s

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/duration"
)

func newCertsStatusCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		timeout time.Duration
	}
	cmd := &cobra.Command{
		Use:   "certs-status",
		Short: "Display certificate and certificate authority expiration details",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client, err := env.Snap.K8sdClient()
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

			status, err := client.CertificatesStatus(ctx, apiv1.CertificatesStatusRequest{})
			if err != nil {
				cmd.PrintErrf("Error: Failed to retrieve certificate status.\n\nError: %v\n", err)
				env.Exit(1)
				return
			}
			if err := printCertificatesStatus(cmd.OutOrStdout(), status); err != nil {
				cmd.PrintErrf("Error: Failed to display certificate status.\n\nError: %v\n", err)
				env.Exit(1)
				return
			}
		},
	}
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")

	return cmd
}

// printCertificatesStatus writes certificate and certificate authority statuses to the provided
// writer in a tabulated format. The output includes the certificate name, expiration date,
// residual time until expiration, associated certificate authority, and whether the certificate
// is externally managed.
func printCertificatesStatus(writer io.Writer, status apiv1.CertificatesStatusResponse) error {
	yesNo := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}

	w := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "CERTIFICATE\tEXPIRES\tRESIDUAL TIME\tCERTIFICATE AUTHORITY\tEXTERNALLY MANAGED")
	for _, certificate := range status.Certificates {
		expirationDate, err := time.Parse(time.RFC3339, certificate.Expires)
		if err != nil {
			return fmt.Errorf("failed to parse expiration date for certificate %s: %w", certificate.Name, err)
		}
		residualTime := duration.HumanDuration(time.Until(expirationDate).Truncate(time.Second))
		formattedDate := expirationDate.Format("Jan 02, 2006 15:04 MST")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			certificate.Name,
			formattedDate,
			residualTime,
			certificate.CertificateAuthority,
			yesNo(certificate.ExternallyManaged),
		)
	}

	if len(status.CertificateAuthorities) > 0 {
		fmt.Fprintln(w, "\nCERTIFICATE AUTHORITY\tEXPIRES\tRESIDUAL TIME\tEXTERNALLY MANAGED")
	}
	for _, certificate := range status.CertificateAuthorities {
		expirationDate, err := time.Parse(time.RFC3339, certificate.Expires)
		if err != nil {
			return fmt.Errorf("failed to parse expiration date for certificate authority %s: %w", certificate.Name, err)
		}
		residualTime := duration.HumanDuration(time.Until(expirationDate).Truncate(time.Second))
		formattedDate := expirationDate.Format("Jan 02, 2006 15:04 MST")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			certificate.Name,
			formattedDate,
			residualTime,
			yesNo(certificate.ExternallyManaged),
		)
	}

	w.Flush()
	return nil
}
