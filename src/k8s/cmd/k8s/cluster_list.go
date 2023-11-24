package k8s

import (
	"context"
	"sort"

	lxdCmd "github.com/canonical/lxd/shared/cmd"

	"github.com/canonical/k8s/pkg/k8s/cluster"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listClusterCmdOpts struct {
		format string
	}
	listClusterCmd = &cobra.Command{
		Use:   "list-cluster",
		Short: "List servers in the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.logDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			clusterMembers, err := cluster.GetMembers(context.Background(), cluster.ClusterOpts{
				Address:  clusterCmdOpts.address,
				StateDir: clusterCmdOpts.stateDir,
				Verbose:  rootCmdOpts.logVerbose,
				Debug:    rootCmdOpts.logDebug,
			})
			if err != nil {
				return err
			}

			members := make([][]string, len(clusterMembers))
			for i, clusterMember := range clusterMembers {
				members[i] = []string{
					clusterMember.Name,
					clusterMember.Address,
					clusterMember.Role,
					clusterMember.Fingerprint,
					clusterMember.Status,
				}
			}

			header := []string{"NAME", "ADDRESS", "ROLE", "FINGERPRINT", "STATUS"}
			sort.Sort(lxdCmd.SortColumnsNaturally(members))
			return lxdCmd.RenderTable(listClusterCmdOpts.format, header, members, clusterMembers)
		},
	}
)

func init() {
	listClusterCmd.Flags().StringVarP(&listClusterCmdOpts.format, "format", "f", "table", "Format (csv|json|table|yaml|compact)")

	rootCmd.AddCommand(listClusterCmd)
}
