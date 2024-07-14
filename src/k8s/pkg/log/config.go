package log

import (
	"flag"
	"fmt"

	"k8s.io/klog/v2"
)

type Options struct {
	LogLevel     int
	AddDirHeader bool
}

// Configure sets global logging configuration (affects all loggers).
func Configure(o Options) {
	flags := flag.NewFlagSet("klog", flag.ContinueOnError)
	klog.InitFlags(flags)

	flags.Set("v", fmt.Sprintf("%v", o.LogLevel))
	flags.Set("add_dir_header", fmt.Sprintf("%v", o.AddDirHeader))
}
