package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/canonical/k8s/pkg/client/k8sd"
	"github.com/canonical/k8s/pkg/mcp/config"
	"github.com/canonical/k8s/pkg/mcp/k8s"
	"github.com/canonical/k8s/pkg/mcp/tools"
	"github.com/canonical/k8s/pkg/mcp/tools/handlers"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents the main MCP server
type Server struct {
	config       *config.Config
	logger       *slog.Logger
	mcpServer    *server.MCPServer
	toolHandlers *handlers.ToolHandler
	httpServer   *server.StreamableHTTPServer
}

// New creates a new server instance
func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	// Create Kubernetes client
	k8sClient, err := k8s.NewClient(cfg.Kubernetes.KubeConfig, cfg.Kubernetes.InCluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	mcpServer := server.NewMCPServer(
		cfg.Server.Name,
		cfg.Server.Version,
		server.WithToolCapabilities(false),            // Tools list does not change
		server.WithResourceCapabilities(false, false), // No resource subscription or list changes
	)

	// Register all tools
	var k8sdClient k8sd.Client = nil // TODO
	toolHandlers := handlers.NewToolHandler(k8sClient, k8sdClient, logger)
	mcpServer.AddTool(tools.CheckStatus, toolHandlers.CheckStatus)
	mcpServer.AddTool(tools.GetResource, toolHandlers.GetResources)
	mcpServer.AddTool(tools.ListResources, toolHandlers.ListResources)

	httpServer := server.NewStreamableHTTPServer(mcpServer)

	return &Server{
		config:       cfg,
		logger:       logger,
		mcpServer:    mcpServer,
		toolHandlers: toolHandlers,
		httpServer:   httpServer,
	}, nil
}

// Run starts the server
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("Starting K8s MCP Server",
		"name", s.config.Server.Name,
		"version", s.config.Server.Version,
	)

	go s.httpServer.Start(s.config.Server.Address)

	<-ctx.Done()
	s.logger.Info("Context cancelled, shutting down K8s MCP Server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Failed to shut down HTTP server", "error", err)
	}

	return nil
}
