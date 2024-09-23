package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/canonical/k8s/cmd/k8s"
	k8s_apiserver_proxy "github.com/canonical/k8s/cmd/k8s-apiserver-proxy"
	"github.com/canonical/k8s/cmd/k8sd"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/log"
	"github.com/spf13/cobra"
)

func main() {
	// execution environment
	env := cmdutil.DefaultExecutionEnvironment()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// logging
	ctx = log.NewContext(ctx, log.L())

	// ensure hooks from all commands are executed
	cobra.EnableTraverseRunHooks = true

	// choose command based on the binary name
	base := filepath.Base(os.Args[0])
	var err error
	switch base {
	case "k8s-apiserver-proxy":
		err = k8s_apiserver_proxy.NewRootCmd(env).ExecuteContext(ctx)
	case "k8sd":
		err = k8sd.NewRootCmd(env).ExecuteContext(ctx)
	case "k8s":
		err = k8s.NewRootCmd(env).ExecuteContext(ctx)
	default:
		panic(fmt.Errorf("invalid entrypoint name %q", base))
	}

	// Although k8s commands typically use Run instead of RunE and handle
	// errors directly within the command execution, this acts as a safeguard in
	// case any are overlooked.
	//
	// Furthermore, the Cobra framework may not invoke the "Run*" entry points
	// at all in case of argument parsing errors, in which case we *need* to
	// handle the errors here.
	if err != nil {
		env.Exit(1)
	}
}
