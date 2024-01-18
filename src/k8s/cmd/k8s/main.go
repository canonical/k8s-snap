package k8s

import "os"

func Main() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
