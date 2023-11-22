package k8s

import (
	"os"

	"github.com/sirupsen/logrus"
)

func Main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Print("k8s command failed")
		os.Exit(1)
	}
}
