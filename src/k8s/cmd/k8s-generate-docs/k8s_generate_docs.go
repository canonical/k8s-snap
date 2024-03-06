package k8s_generate_docs

// TODO(neoaggelos): this should be a sub-command for 'k8s', but is currently here because
// of the k8s cli preRun hook that requires root.

import (
	"github.com/canonical/k8s/cmd/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewRootCmd() *cobra.Command {
	var opts struct {
		outputDir string
	}
	cmd := &cobra.Command{
		Use:    "k8s-generate-docs",
		Hidden: true,
		Short:  "Generate markdown documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doc.GenMarkdownTree(k8s.NewRootCmd(), opts.outputDir)
		},
	}

	cmd.Flags().StringVar(&opts.outputDir, "output-dir", ".", "directory where the markdown docs will be written")
	return cmd
}
