package k8sd

import (
	"os"

	"github.com/sirupsen/logrus"
)

func Main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Print("k8sd command failed")
		os.Exit(1)
	}
}
