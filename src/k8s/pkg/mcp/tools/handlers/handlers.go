package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/canonical/k8s/pkg/client/k8sd"
	"github.com/canonical/k8s/pkg/mcp/k8s"
	"github.com/mark3labs/mcp-go/mcp"
)

// ToolHandler handles MCP tool calls
type ToolHandler struct {
	k8sClient  *k8s.Client
	k8sdClient k8sd.Client
	logger     *slog.Logger
}

// NewToolHandler creates a new ToolHandler instance
func NewToolHandler(k8sClient *k8s.Client, k8sdClient k8sd.Client, logger *slog.Logger) *ToolHandler {
	return &ToolHandler{
		k8sClient:  k8sClient,
		k8sdClient: k8sdClient,
		logger:     logger,
	}
}

func (h *ToolHandler) ListResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, errors.New("invalid arguments type: expected map[string]any")
	}

	kind, err := getRequiredStringArg(args, "kind")
	if err != nil {
		return nil, err
	}

	namespace := getStringArg(args, "namespace", "")
	labelSelector := getStringArg(args, "labelSelector", "")

	resources, err := h.k8sClient.ListResources(ctx, kind, namespace, labelSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources for kind %q: %w", kind, err)
	}

	jsonResponse, err := resources.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize response: %w", err)
	}

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

func (h *ToolHandler) GetResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, errors.New("invalid arguments type: expected map[string]any")
	}

	// TODO: make centralized with where these args are defined
	kind, err := getRequiredStringArg(args, "kind")
	if err != nil {
		return nil, err
	}

	name, err := getRequiredStringArg(args, "name")
	if err != nil {
		return nil, err
	}

	namespace := getStringArg(args, "namespace", "")

	resource, err := h.k8sClient.GetResource(ctx, kind, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource %q of kind %q: %w", name, kind, err)
	}

	jsonResponse, err := resource.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize response: %w", err)
	}

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

func (h *ToolHandler) CheckStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, errors.New("invalid arguments type: expected map[string]any")
	}

	waitReady := getBoolArg(args, "waitReady", false)
	resp, err := h.k8sdClient.ClusterStatus(ctx, waitReady)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster status: %w", err)
	}

	return mcp.NewToolResultJSON(resp)
}

func getStringArg(args map[string]any, key string, defaultValue string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultValue
}

func getBoolArg(args map[string]any, key string, defaultValue bool) bool {
	if val, ok := args[key].(bool); ok {
		return val
	}
	return defaultValue
}

func getRequiredStringArg(args map[string]any, key string) (string, error) {
	val, ok := args[key].(string)
	if !ok || val == "" {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	return val, nil
}
