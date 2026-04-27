package controllers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
)

// ServiceArgsControllerOpts holds configuration for ServiceArgsController.
type ServiceArgsControllerOpts struct {
	// Snap is the snap interface.
	Snap snap.Snap
	// Services is an optional override for the list of service names to watch.
	// When nil, the list is determined dynamically each reconcile cycle based on
	// whether the node is a worker or control-plane.
	Services []string
	// TriggerCh drives the reconciliation loop. Typically time.NewTicker(<interval>).C.
	TriggerCh <-chan time.Time
	// GetRunningArgs returns the parsed command-line arguments of the running process
	// for the named service. Returns nil, nil when the process is not running.
	// Defaults to a systemd-based implementation when nil.
	GetRunningArgs func(ctx context.Context, serviceName string) (map[string]string, error)
}

// ServiceArgsController periodically compares each service's args file against the
// arguments of the running process and restarts any service whose arguments have drifted.
type ServiceArgsController struct {
	snap           snap.Snap
	services       []string
	triggerCh      <-chan time.Time
	reconciledCh   chan struct{}
	getRunningArgs func(ctx context.Context, serviceName string) (map[string]string, error)
}

// NewServiceArgsController creates a new ServiceArgsController.
func NewServiceArgsController(opts ServiceArgsControllerOpts) *ServiceArgsController {
	if opts.GetRunningArgs == nil {
		opts.GetRunningArgs = utils.RunningServiceArgs
	}
	return &ServiceArgsController{
		snap:           opts.Snap,
		services:       opts.Services,
		triggerCh:      opts.TriggerCh,
		reconciledCh:   make(chan struct{}, 1),
		getRunningArgs: opts.GetRunningArgs,
	}
}

// Run starts the controller and blocks until ctx is cancelled.
func (c *ServiceArgsController) Run(ctx context.Context) {
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "service-args"))
	log := log.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.triggerCh:
		}

		log.Info("checking service arguments for drift")
		if err := c.reconcile(ctx); err != nil {
			log.Error(err, "failed to reconcile service arguments")
		}

		select {
		case c.reconciledCh <- struct{}{}:
		default:
		}
	}
}

func (c *ServiceArgsController) reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)

	services := c.services
	if services == nil {
		isWorker, err := snaputil.IsWorker(c.snap)
		if err != nil {
			return fmt.Errorf("failed to determine node type: %w", err)
		}
		if isWorker {
			services = snaputil.WorkerServices()
		} else {
			services = snaputil.ControlPlaneServices()
		}
	}

	for _, svc := range services {
		differs, err := c.serviceArgsDiffer(ctx, svc)
		if err != nil {
			log.Error(err, "failed to compare arguments", "service", svc)
			continue
		}
		if !differs {
			continue
		}

		log.Info("service arguments have drifted from args file, restarting", "service", svc)
		if err := c.snap.RestartServices(ctx, []string{svc}); err != nil {
			log.Error(err, "failed to restart service", "service", svc)
		}
	}

	return nil
}

// serviceArgsDiffer returns true if the desired args (from the args file) differ
// from the arguments the running service was started with.
// Returns false if the args file does not exist or the process is not running.
func (c *ServiceArgsController) serviceArgsDiffer(ctx context.Context, serviceName string) (bool, error) {
	argsFilePath := filepath.Join(c.snap.ServiceArgumentsDir(), serviceName)
	desiredArgs, err := utils.ParseArgumentFile(argsFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read args file for %q: %w", serviceName, err)
	}

	actualArgs, err := c.getRunningArgs(ctx, serviceName)
	if err != nil {
		if errors.Is(err, utils.ErrUnitNotRunning) {
			// service is not running, args don't differ technically
			return false, nil
		}
		return false, fmt.Errorf("failed to get running args for %q: %w", serviceName, err)
	}

	if len(desiredArgs) != len(actualArgs) {
		return true, nil
	}
	for key, desiredVal := range desiredArgs {
		if actualArgs[key] != desiredVal {
			return true, nil
		}
	}
	return false, nil
}

// ReconciledCh returns the channel that receives a value after each reconciliation loop.
func (c *ServiceArgsController) ReconciledCh() <-chan struct{} {
	return c.reconciledCh
}
