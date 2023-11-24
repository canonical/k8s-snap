package k8s

import (
	"context"
	"sort"

	lxdCmd "github.com/canonical/lxd/shared/cmd"

	cluster "github.com/canonical/k8s/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listClusterCmdOpts struct {
		flagFormat string
	}
	listClusterCmd = &cobra.Command{
		Use:   "list-cluster",
		Short: "List servers in the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.flagLogDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			clusterMembers, err := cluster.GetMembers(context.Background(), cluster.ClusterOpts{
				Address:  clusterCmdOpts.flagAddress,
				StateDir: clusterCmdOpts.flagStateDir,
				Verbose:  rootCmdOpts.flagLogVerbose,
				Debug:    rootCmdOpts.flagLogDebug,
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
			return lxdCmd.RenderTable(listClusterCmdOpts.flagFormat, header, members, clusterMembers)
		},
	}
)

func init() {
	listClusterCmd.Flags().StringVarP(&listClusterCmdOpts.flagFormat, "format", "f", "table", "Format (csv|json|table|yaml|compact)")

	rootCmd.AddCommand(listClusterCmd)
}
