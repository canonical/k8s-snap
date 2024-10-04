package k8s

import (
	"path"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/docgen"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newGenerateDocsCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		outputDir string
	}
	cmd := &cobra.Command{
		Use:    "generate-docs",
		Hidden: true,
		Short:  "Generate markdown documentation",
		Run: func(cmd *cobra.Command, args []string) {
			outPath := path.Join(opts.outputDir, "commands")
			if err := doc.GenMarkdownTree(cmd.Parent(), outPath); err != nil {
				cmd.PrintErrf("Error: Failed to generate markdown documentation for k8s command.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			outPath = path.Join(opts.outputDir, "bootstrap_config.md")
			err := docgen.MarkdownFromJsonStructToFile(apiv1.BootstrapConfig{}, outPath)
			if err != nil {
				cmd.PrintErrf("Error: Failed to generate markdown documentation for bootstrap configuration\n\n")
				cmd.PrintErrf("Error: %v", err)
				env.Exit(1)
				return
			}

			outPath = path.Join(opts.outputDir, "control_plane_join_config.md")
			err = docgen.MarkdownFromJsonStructToFile(apiv1.ControlPlaneJoinConfig{}, outPath)
			if err != nil {
				cmd.PrintErrf("Error: Failed to generate markdown documentation for ctrl plane join configuration\n\n")
				cmd.PrintErrf("Error: %v", err)
				env.Exit(1)
				return
			}

			outPath = path.Join(opts.outputDir, "worker_join_config.md")
			err = docgen.MarkdownFromJsonStructToFile(apiv1.WorkerJoinConfig{}, outPath)
			if err != nil {
				cmd.PrintErrf("Error: Failed to generate markdown documentation for worker join configuration\n\n")
				cmd.PrintErrf("Error: %v", err)
				env.Exit(1)
				return
			}

			cmd.Printf("Generated documentation in %s\n", opts.outputDir)
		},
	}

	cmd.Flags().StringVar(&opts.outputDir, "output-dir", ".", "directory where the markdown docs will be written")
	return cmd
}
