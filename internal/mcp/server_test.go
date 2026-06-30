package mcp_test

import (
	"context"
	"encoding/json"
	"testing"

	internalmcp "github.com/pariwork/googledrive-ai/internal/mcp"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewServer_InitializeAndListTools(t *testing.T) {
	s := internalmcp.NewServer()

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

	_, ok = resp2.Result.(mcp.ListToolsResult)
	if !ok {
		t.Fatalf("expected ListToolsResult, got %T", resp2.Result)
	}
}
