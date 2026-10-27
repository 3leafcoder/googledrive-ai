package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/3leafcoder/googledrive-ai/internal/auth"
	"github.com/3leafcoder/googledrive-ai/internal/docs"
	"github.com/3leafcoder/googledrive-ai/internal/drive"
	"github.com/3leafcoder/googledrive-ai/internal/sheets"
)

const (
	serverName    = "googledrive-mcp"
	serverVersion = "0.1.0"
)

// NewServer creates a configured MCP server with all tools registered.
func NewServer(pm *auth.ProfileManager) *server.MCPServer {
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(false),
	)

	drive.RegisterTools(s, pm)
	sheets.RegisterTools(s, pm)
	docs.RegisterTools(s, pm)
	auth.RegisterTools(s, pm)

	return s
}
