package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools adds auth management tools to the MCP server.
func RegisterTools(s *server.MCPServer, pm *ProfileManager) {
	s.AddTool(authStatusTool(), authStatusHandler(pm))
	s.AddTool(authListProfilesTool(), authListProfilesHandler(pm))
	s.AddTool(authSwitchProfileTool(), authSwitchProfileHandler(pm))
}

// --- auth_status ---

func authStatusTool() mcp.Tool {
	return mcp.Tool{
		Name:        "auth_status",
		Description: "Returns the active Google account profile name, authenticated email, granted scopes, and token expiry timestamp.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
		},
	}
}

func authStatusHandler(pm *ProfileManager) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := pm.Status(ctx)
		if err != nil {
			return toolError(fmt.Sprintf("failed to get auth status: %v", err)), nil
		}
		return jsonResult(status)
	}
}

// --- auth_list_profiles ---

func authListProfilesTool() mcp.Tool {
	return mcp.Tool{
		Name:        "auth_list_profiles",
		Description: "Lists all configured Google account profiles with their email addresses. Profiles with expired or invalid tokens are marked as such rather than omitted.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
		},
	}
}

func authListProfilesHandler(pm *ProfileManager) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		profiles, err := pm.List(ctx)
		if err != nil {
			return toolError(fmt.Sprintf("failed to list profiles: %v", err)), nil
		}
		return jsonResult(profiles)
	}
}

// --- auth_switch_profile ---

func authSwitchProfileTool() mcp.Tool {
	return mcp.Tool{
		Name:        "auth_switch_profile",
		Description: "Switches the active Google account to the specified profile. All subsequent Drive, Sheets, and Docs calls will use the new account. No server restart required.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"profile": map[string]any{
					"type":        "string",
					"description": "The profile name to switch to (e.g. 'work', 'personal', or 'default').",
				},
			},
			Required: []string{"profile"},
		},
	}
}

func authSwitchProfileHandler(pm *ProfileManager) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		profile, err := requiredString(req, "profile")
		if err != nil {
			return toolError(err.Error()), nil
		}

		if err := pm.SwitchProfile(ctx, profile); err != nil {
			return toolError(err.Error()), nil
		}

		return jsonResult(map[string]string{
			"status":  "switched",
			"profile": profile,
		})
	}
}

// --- helpers ---

func requiredString(req mcp.CallToolRequest, key string) (string, error) {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("parameter %s must be a non-empty string", key)
	}
	return s, nil
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshaling result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(data)),
		},
	}, nil
}

func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			mcp.NewTextContent(msg),
		},
	}
}
