package k8s_dqlite

import "os"

func Main() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
