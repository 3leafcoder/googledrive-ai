package mcp

import (
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "googledrive-mcp"
	serverVersion = "0.1.0"
)

// NewServer creates a configured MCP server with tool capabilities declared.
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(false),
	)
	return s
}
