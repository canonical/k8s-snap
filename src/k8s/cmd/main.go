package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/canonical/k8s/cmd/k8s"
	k8s_apiserver_proxy "github.com/canonical/k8s/cmd/k8s-apiserver-proxy"
	"github.com/canonical/k8s/cmd/k8sd"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func main() {
	// execution environment
	env := cmdutil.DefaultExecutionEnvironment()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// ensure hooks from all commands are executed
	cobra.EnableTraverseRunHooks = true

	// choose command based on the binary name
	base := path.Base(os.Args[0])
	switch base {
	case "k8s-apiserver-proxy":
		if err := k8s_apiserver_proxy.NewRootCmd(env).ExecuteContext(ctx); err != nil {
			log.Fatalf("error executing k8s-apiserver-proxy command: %v", err)
		}
	case "k8sd":
		k8sd.NewRootCmd(env).ExecuteContext(ctx)
	case "k8s":
		k8s.NewRootCmd(env).ExecuteContext(ctx)
	default:
		panic(fmt.Errorf("invalid entrypoint name %q", base))
	}
}
