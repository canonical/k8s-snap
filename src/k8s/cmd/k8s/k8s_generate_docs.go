package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
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
			if err := doc.GenMarkdownTree(cmd.Parent(), opts.outputDir); err != nil {
				cmd.PrintErrf("ERROR: Failed to generate markdown documentation for k8s command.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}
			cmd.Printf("Generated documentation in %s\n", opts.outputDir)
		},
	}

	cmd.Flags().StringVar(&opts.outputDir, "output-dir", ".", "directory where the markdown docs will be written")
	return cmd
}
