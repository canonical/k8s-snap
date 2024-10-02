package k8s

import (
	"os"

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
			if err := doc.GenMarkdownTree(cmd.Parent(), opts.outputDir+"/commands"); err != nil {
				cmd.PrintErrf("Error: Failed to generate markdown documentation for k8s command.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			bootstrap_doc, err := docgen.MarkdownFromJsonStruct(apiv1.BootstrapConfig{})
			if err != nil {
				cmd.PrintErrf("Error: Failed to generate markdown documentation for bootstrap configuration\n\n")
				cmd.PrintErrf("Error: %v", err)
				env.Exit(1)
				return
			}

			bootstrap_doc_path := opts.outputDir + "/bootstrap_config.md"
			err = os.WriteFile(bootstrap_doc_path, []byte(bootstrap_doc), 0644)
			if err != nil {
				cmd.PrintErrf("Error: Failed to write markdown documentation for bootstrap configuration\n\n")
				cmd.PrintErrf("Error: %v")
				env.Exit(1)
				return
			}

			cmd.Printf("Generated documentation in %s\n", opts.outputDir)
		},
	}

	cmd.Flags().StringVar(&opts.outputDir, "output-dir", ".", "directory where the markdown docs will be written")
	return cmd
}
