package k8s_dqlite

import (
	_ "embed"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	//go:embed dqlite_version.sh
	dqliteVersion string

	versionCmd = &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("--------------------")
			fmt.Println("Go version:", runtime.Version())
			fmt.Println("--------------------")
			fmt.Println(dqliteVersion)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
