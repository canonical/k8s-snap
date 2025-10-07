package tools

import "github.com/mark3labs/mcp-go/mcp"

var (
	CheckStatus mcp.Tool = mcp.NewTool(
		"checkStatus",
		mcp.WithDescription("Check the status of the cluster"),
	)
	GetResource mcp.Tool = mcp.NewTool(
		"getResource",
		mcp.WithDescription("Get a specific resource by kind, name, and namespace"),
		// TODO: centralize argument definitions and extraction/usage
		mcp.WithString("kind", mcp.Required(), mcp.Description("The kind of the resource")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
	)
	ListResources mcp.Tool = mcp.NewTool(
		"listResources",
		mcp.WithDescription("List all resources in the cluster"),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The kind of the resource")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
		mcp.WithString("labelSelector", mcp.Description("A label selector to filter resources by")),
	)
)
