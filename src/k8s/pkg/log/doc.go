// package log provides logging tools and capabilities to k8s-snap.
//
// Typical usage:
//
//	// When a context is available
//	log := log.FromContext(ctx)
//	log.Info("Info message")
//	log.V(3).Info("Info message that will be printed when log level is 3 or more")
//	log.Error(err, "Failed to do something")
//
//	// When no context is available, you can use log.L() to get a default logger.
//	log := log.L()
//
//	// To add structured fields to your logs
//	log.WithValues("name", "k8sd-config").Info("Reconcile configmap")
//
//	// To add structured fields for multiple logs
//	log = log.WithValues("name", "k8sd-config")
//	log.Info("Start reconcile")
//	log.Info("Looking for changes")
//
//	// You can create a new context with a logger (use in controllers or components)
//	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("key", value))
//
// Log messages include the file name and line number of the log message. In case
// of utility functions, you can print the line number of the caller instead with:
//
//	log.FromContext(ctx).WithCallDepth(1).Error(err, "Failed to format JSON output")
//
// To configure the logger behaviour, you can use this in the main package:
//
//	log.Configure(log.Options{LogLevel: 3, AddDirHeader: true})
package log
