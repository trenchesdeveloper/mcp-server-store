package tools

import (
	"github.com/trenchesdeveloper/mcp-server-store/internal/mcp"
)

// PingTool returns the tool definition for the "ping" tool.
func PingTool() mcp.Tool {
	return mcp.Tool{
		Name:        "ping",
		Description: "A simple ping tool that returns pong.",
		InputSchema: mcp.InputSchema{
			Type: "object",
		},
	}
}

// PingHandler returns a tool handler that simply returns "pong".
func PingHandler() mcp.ToolHandler {
	return func(arguments map[string]interface{}) (*mcp.ToolCallResult, error) {
		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				mcp.NewTextContent("pong"),
			},
		}, nil
	}
}
