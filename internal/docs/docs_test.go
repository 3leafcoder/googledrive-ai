package docs

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	docsapi "google.golang.org/api/docs/v1"
)

func fakeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
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

func TestOptionalString(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		req := fakeRequest(map[string]any{"key": "value"})
		got := optionalString(req, "key", "default")
		if got != "value" {
			t.Errorf("got %q, want %q", got, "value")
		}
	})
	t.Run("missing returns default", func(t *testing.T) {
		req := fakeRequest(map[string]any{})
		got := optionalString(req, "key", "default")
		if got != "default" {
			t.Errorf("got %q, want %q", got, "default")
		}
	})
}

func TestExtractText(t *testing.T) {
	doc := &docsapi.Document{
		Body: &docsapi.Body{
			Content: []*docsapi.StructuralElement{
				{
					Paragraph: &docsapi.Paragraph{
						Elements: []*docsapi.ParagraphElement{
							{TextRun: &docsapi.TextRun{Content: "Hello "}},
							{TextRun: &docsapi.TextRun{Content: "World"}},
						},
					},
				},
				{
					Paragraph: &docsapi.Paragraph{
						Elements: []*docsapi.ParagraphElement{
							{TextRun: &docsapi.TextRun{Content: "\n"}},
						},
					},
				},
				{
					Paragraph: &docsapi.Paragraph{
						Elements: []*docsapi.ParagraphElement{
							{TextRun: &docsapi.TextRun{Content: "Second paragraph\n"}},
						},
					},
				},
			},
		},
	}

	got := extractText(doc)
	want := "Hello World\nSecond paragraph\n"
	if got != want {
		t.Errorf("extractText = %q, want %q", got, want)
	}
}

func TestExtractText_NilBody(t *testing.T) {
	doc := &docsapi.Document{}
	got := extractText(doc)
	if got != "" {
		t.Errorf("extractText(nil body) = %q, want empty", got)
	}
}

func TestToolDefinitions(t *testing.T) {
	tools := []struct {
		name     string
		tool     mcp.Tool
		required []string
	}{
		{"docs_create", createTool(), []string{"title"}},
		{"docs_read", readTool(), []string{"document_id"}},
		{"docs_append", appendTool(), []string{"document_id", "text"}},
		{"docs_replace", replaceTool(), []string{"document_id", "find_text", "replace_text"}},
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
			for i, r := range tt.required {
				if i < len(tt.tool.InputSchema.Required) && tt.tool.InputSchema.Required[i] != r {
					t.Errorf("required[%d] = %q, want %q", i, tt.tool.InputSchema.Required[i], r)
				}
			}
			for _, r := range tt.required {
				if _, ok := tt.tool.InputSchema.Properties[r]; !ok {
					t.Errorf("missing property definition for required param %q", r)
				}
			}
		})
	}
}
