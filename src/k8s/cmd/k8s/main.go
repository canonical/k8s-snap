package k8s

import (
	"errors"
	"os"

	"github.com/sirupsen/logrus"
)

func Main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Print(errors.Unwrap(err))
		os.Exit(1)
	}
}
