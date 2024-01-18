package k8s_apiserver_proxy

import "os"

func Main() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
