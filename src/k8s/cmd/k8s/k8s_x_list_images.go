package k8s

import (
	"strings"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/images"
	"github.com/spf13/cobra"
)

func newXListImagesCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Hidden: true,
		Use:    "x-list-images",
		Short:  "List all images used by the current version of k8s-snap",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(strings.Join(images.Images(), "\n"))
		},
	}
	return cmd
}
