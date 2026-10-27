package mcp_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/3leafcoder/googledrive-ai/internal/auth"
	internalmcp "github.com/3leafcoder/googledrive-ai/internal/mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"golang.org/x/oauth2"
)

// stubTokenSource satisfies oauth2.TokenSource for testing.
type stubTokenSource struct{}

func (s *stubTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "test-token"}, nil
}

func TestNewServer_InitializeAndListTools(t *testing.T) {
	pm := auth.NewProfileManager(&oauth2.Config{}, "", &stubTokenSource{})
	s := internalmcp.NewServer(pm)

	// Send initialize request as raw JSON.
	initJSON := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}`
	initResp := s.HandleMessage(context.Background(), json.RawMessage(initJSON))

	resp, ok := initResp.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("expected JSONRPCResponse, got %T", initResp)
	}

	result, ok := resp.Result.(mcp.InitializeResult)
	if !ok {
		t.Fatalf("expected InitializeResult, got %T", resp.Result)
	}

	if result.ServerInfo.Name != "googledrive-mcp" {
		t.Errorf("server name = %q, want %q", result.ServerInfo.Name, "googledrive-mcp")
	}
	if result.ServerInfo.Version != "0.1.0" {
		t.Errorf("server version = %q, want %q", result.ServerInfo.Version, "0.1.0")
	}
	if result.Capabilities.Tools == nil {
		t.Fatal("expected tools capability to be declared")
	}

	// Send tools/list request.
	listJSON := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	listResp := s.HandleMessage(context.Background(), json.RawMessage(listJSON))

	resp2, ok := listResp.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("expected JSONRPCResponse, got %T", listResp)
	}

	listResult, ok := resp2.Result.(mcp.ListToolsResult)
	if !ok {
		t.Fatalf("expected ListToolsResult, got %T", resp2.Result)
	}

	wantTools := map[string]bool{
		"sheets_list_sheets":   false,
		"sheets_read_range":    false,
		"sheets_get_metadata":  false,
		"docs_create":          false,
		"docs_read":            false,
		"docs_append":          false,
		"docs_replace":         false,
		"auth_status":          false,
		"auth_list_profiles":   false,
		"auth_switch_profile":  false,
	}
	for _, tool := range listResult.Tools {
		if _, ok := wantTools[tool.Name]; ok {
			wantTools[tool.Name] = true
		}
	}
	for name, found := range wantTools {
		if !found {
			t.Errorf("expected tool %q to be registered", name)
		}
	}
}
