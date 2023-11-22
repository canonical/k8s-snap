package k8s_dqlite

import (
	"os"

	"github.com/sirupsen/logrus"
)

func Main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Print("k8s-dqlite command failed")
		os.Exit(1)
	}
}
