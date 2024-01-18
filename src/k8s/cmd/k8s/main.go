package k8s

import (
	"os"
)

func Main() {
	if err := rootCmd.Execute(); err != nil {
		// TODO: We need to define actionable error message in the future
		// That tell the user what went wrong on a high-level and - if possible - how it can be fixed.
		os.Exit(1)
	}
}
