package log

import "k8s.io/klog/v2"

var logger = klog.NewKlogr().WithName("k8sd")

// L is the default logger to use.
func L() Logger {
	return logger
}
