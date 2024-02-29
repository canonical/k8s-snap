package mock

import (
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
)

func mockLogAdapter(format string, v ...any) {
	logrus.Debugf(format, v...)
}

// "Mock" Initializer
type MockHelmConfigProvider struct {
	ActionConfig *action.Configuration
}

func (r *MockHelmConfigProvider) New(namespace string) (*action.Configuration, error) {
	r.ActionConfig.Log = mockLogAdapter
	return r.ActionConfig, nil
}
