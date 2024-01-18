package k8sd

import "os"

func Main() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
