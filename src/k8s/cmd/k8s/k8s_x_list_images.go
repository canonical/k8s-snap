package k8s

import (
	"strings"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/images"
	"github.com/spf13/cobra"
)

func newListImagesCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Hidden:  true,
		Aliases: []string{"list-images"},
		Short:   "List all images used by this build",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(strings.Join(images.Images(), "\n"))
		},
	}
	return cmd
}
