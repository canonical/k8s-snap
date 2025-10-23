package mcp

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/mcp/config"
	"github.com/canonical/k8s/pkg/mcp/server"
	"github.com/spf13/cobra"
)

func NewRootCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Canonical Kubernetes MCP server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				cmd.PrintErrf("Failed to load configuration: %v\n", err)
				env.Exit(1)
				return
			}

			var logLevel *slog.Level
			if err := logLevel.UnmarshalText([]byte(cfg.Logging.Level)); err != nil {
				cmd.PrintErrf("Failed to parse log level: %v\n", err)
				env.Exit(1)
				return
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: logLevel,
			}))

			srv, err := server.New(cfg, logger)
			if err != nil {
				cmd.PrintErrf("Failed to create server: %v\n", err)
				env.Exit(1)
				return
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigCh
				cmd.Println("Received shutdown signal")
				cancel()
			}()

			if err := srv.Run(ctx); err != nil {
				cmd.PrintErrf("Server failed: %v\n", err)
				env.Exit(1)
				return
			}

			cmd.Println("Server shutdown gracefully")
		},
	}

	cmd.SetIn(env.Stdin)
	cmd.SetOut(env.Stdout)
	cmd.SetErr(env.Stderr)

	return cmd
}
