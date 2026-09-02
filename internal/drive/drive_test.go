package drive

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func fakeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func TestToolDefinitions(t *testing.T) {
	tools := []struct {
		name     string
		tool     mcp.Tool
		required []string
	}{
		{"drive_list", listTool(), nil},
		{"drive_search", searchTool(), []string{"query"}},
		{"drive_get", getTool(), []string{"file_id"}},
		{"drive_create_folder", createFolderTool(), []string{"name"}},
		{"drive_upload", uploadTool(), nil},
		{"drive_download", downloadTool(), []string{"file_id"}},
		{"drive_delete", deleteTool(), []string{"file_id"}},
		{"drive_move", moveTool(), []string{"file_id", "new_parent_id"}},
	}

	for _, tt := range tools {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tool.Name != tt.name {
				t.Errorf("tool name = %q, want %q", tt.tool.Name, tt.name)
			}
			if tt.tool.Description == "" {
				t.Error("tool description is empty")
			}
			if len(tt.tool.InputSchema.Required) != len(tt.required) {
				t.Errorf("required params = %v, want %v", tt.tool.InputSchema.Required, tt.required)
			}
			for _, r := range tt.required {
				if _, ok := tt.tool.InputSchema.Properties[r]; !ok {
					t.Errorf("missing property definition for required param %q", r)
				}
			}
		})
	}
}

func TestRequiredString(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		req := fakeRequest(map[string]any{"key": "value"})
		got, err := requiredString(req, "key")
		if err != nil {
			t.Fatal(err)
		}
		if got != "value" {
			t.Errorf("got %q, want %q", got, "value")
		}
	})
	t.Run("missing", func(t *testing.T) {
		req := fakeRequest(map[string]any{})
		_, err := requiredString(req, "key")
		if err == nil {
			t.Fatal("expected error for missing key")
		}
	})
	t.Run("empty", func(t *testing.T) {
		req := fakeRequest(map[string]any{"key": ""})
		_, err := requiredString(req, "key")
		if err == nil {
			t.Fatal("expected error for empty string")
		}
	})
}

func TestMimeToType(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"application/vnd.google-apps.folder", "folder"},
		{"application/vnd.google-apps.document", "document"},
		{"application/vnd.google-apps.spreadsheet", "spreadsheet"},
		{"application/vnd.google-apps.presentation", "presentation"},
		{"application/pdf", "file"},
		{"text/plain", "file"},
	}
	for _, tt := range tests {
		if got := mimeToType(tt.mime); got != tt.want {
			t.Errorf("mimeToType(%q) = %q, want %q", tt.mime, got, tt.want)
		}
	}
}

func TestEscapeQuery(t *testing.T) {
	if got := escapeQuery("it's a test"); got != "it\\'s a test" {
		t.Errorf("escapeQuery = %q", got)
	}
}

func TestIsText(t *testing.T) {
	if !isText([]byte("hello world")) {
		t.Error("expected text for ASCII string")
	}
	if !isText([]byte{}) {
		t.Error("expected text for empty data")
	}
	if isText([]byte{0x48, 0x00, 0x65}) {
		t.Error("expected binary for data with null byte")
	}
}
