package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/pariwork/googledrive-ai/internal/docs"
	"github.com/pariwork/googledrive-ai/internal/drive"
	"github.com/pariwork/googledrive-ai/internal/sheets"
	"golang.org/x/oauth2"
)

const (
	serverName    = "googledrive-mcp"
	serverVersion = "0.1.0"
)

// NewServer creates a configured MCP server with all tools registered.
func NewServer(ts oauth2.TokenSource) *server.MCPServer {
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(false),
	)

	drive.RegisterTools(s, ts)
	sheets.RegisterTools(s, ts)
	docs.RegisterTools(s, ts)

	return s
}
