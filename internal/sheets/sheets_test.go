package sheets

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// fakeRequest creates a CallToolRequest with the given arguments.
func fakeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func TestColLetter(t *testing.T) {
	tests := []struct {
		idx  int64
		want string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
		{701, "ZZ"},
		{702, "AAA"},
	}
	for _, tt := range tests {
		got := colLetter(tt.idx)
		if got != tt.want {
			t.Errorf("colLetter(%d) = %q, want %q", tt.idx, got, tt.want)
		}
	}
}

func TestRequiredString(t *testing.T) {
	// Test missing parameter returns error
	_, err := requiredString(fakeRequest(nil), "missing_key")
	if err == nil {
		t.Error("expected error for missing parameter")
	}

	// Test empty string returns error
	_, err = requiredString(fakeRequest(map[string]any{"key": ""}), "key")
	if err == nil {
		t.Error("expected error for empty string parameter")
	}

	// Test valid string
	val, err := requiredString(fakeRequest(map[string]any{"key": "value"}), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Errorf("got %q, want %q", val, "value")
	}
}

func TestOptionalString(t *testing.T) {
	// Missing key returns default
	val := optionalString(fakeRequest(nil), "key", "default")
	if val != "default" {
		t.Errorf("got %q, want %q", val, "default")
	}

	// Present key returns value
	val = optionalString(fakeRequest(map[string]any{"key": "custom"}), "key", "default")
	if val != "custom" {
		t.Errorf("got %q, want %q", val, "custom")
	}
}

func TestToolDefinitions(t *testing.T) {
	// Verify tool names and required params are correct
	lt := listSheetsTool()
	if lt.Name != "sheets_list_sheets" {
		t.Errorf("list sheets tool name = %q", lt.Name)
	}
	if len(lt.InputSchema.Required) != 1 || lt.InputSchema.Required[0] != "spreadsheet_id" {
		t.Errorf("list sheets required = %v, want [spreadsheet_id]", lt.InputSchema.Required)
	}

	rt := readRangeTool()
	if rt.Name != "sheets_read_range" {
		t.Errorf("read range tool name = %q", rt.Name)
	}
	if len(rt.InputSchema.Required) != 2 {
		t.Errorf("read range required = %v, want [spreadsheet_id, range]", rt.InputSchema.Required)
	}

	mt := getMetadataTool()
	if mt.Name != "sheets_get_metadata" {
		t.Errorf("get metadata tool name = %q", mt.Name)
	}
	if len(mt.InputSchema.Required) != 1 || mt.InputSchema.Required[0] != "spreadsheet_id" {
		t.Errorf("get metadata required = %v, want [spreadsheet_id]", mt.InputSchema.Required)
	}
}
