package docs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"
	docsapi "google.golang.org/api/docs/v1"
	driveapi "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// RegisterTools adds all Docs tools to the MCP server.
func RegisterTools(s *server.MCPServer, ts oauth2.TokenSource) {
	s.AddTool(createTool(), createHandler(ts))
	s.AddTool(readTool(), readHandler(ts))
	s.AddTool(appendTool(), appendHandler(ts))
	s.AddTool(replaceTool(), replaceHandler(ts))
}

func newDocsService(ctx context.Context, ts oauth2.TokenSource) (*docsapi.Service, error) {
	return docsapi.NewService(ctx, option.WithTokenSource(ts))
}

func newDriveService(ctx context.Context, ts oauth2.TokenSource) (*driveapi.Service, error) {
	return driveapi.NewService(ctx, option.WithTokenSource(ts))
}

// --- docs_create ---

func createTool() mcp.Tool {
	return mcp.Tool{
		Name:        "docs_create",
		Description: "Create a new Google Doc with a title and optional initial text content. Returns the new document ID.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "Title of the new document",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Optional initial text content for the document",
				},
			},
			Required: []string{"title"},
		},
	}
}

type createResult struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
}

func createHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title, err := requiredString(req, "title")
		if err != nil {
			return toolError(err.Error()), nil
		}
		content := optionalString(req, "content", "")

		// Create the doc via Drive API (Docs API Create doesn't support setting title directly in a clean way).
		driveSrv, err := newDriveService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}

		file := &driveapi.File{
			Name:     title,
			MimeType: "application/vnd.google-apps.document",
		}
		created, err := driveSrv.Files.Create(file).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to create document: %v", err)), nil
		}

		// If content provided, insert it via Docs API.
		if content != "" {
			docsSrv, err := newDocsService(ctx, ts)
			if err != nil {
				return nil, fmt.Errorf("creating docs service: %w", err)
			}

			insertReq := &docsapi.BatchUpdateDocumentRequest{
				Requests: []*docsapi.Request{
					{
						InsertText: &docsapi.InsertTextRequest{
							Text: content,
							Location: &docsapi.Location{
								Index: 1, // After the initial newline
							},
						},
					},
				},
			}
			_, err = docsSrv.Documents.BatchUpdate(created.Id, insertReq).Context(ctx).Do()
			if err != nil {
				return toolError(fmt.Sprintf("created document but failed to insert content: %v", err)), nil
			}
		}

		return jsonResult(createResult{
			DocumentID: created.Id,
			Title:      title,
		})
	}
}

// --- docs_read ---

func readTool() mcp.Tool {
	return mcp.Tool{
		Name:        "docs_read",
		Description: "Read a Google Doc's full text content as plain text.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"document_id": map[string]any{
					"type":        "string",
					"description": "The document ID",
				},
			},
			Required: []string{"document_id"},
		},
	}
}

func readHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		documentID, err := requiredString(req, "document_id")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newDocsService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating docs service: %w", err)
		}

		doc, err := srv.Documents.Get(documentID).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to get document: %v", err)), nil
		}

		text := extractText(doc)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(text),
			},
		}, nil
	}
}

// extractText walks the document body and concatenates all text runs into plain text.
func extractText(doc *docsapi.Document) string {
	if doc.Body == nil {
		return ""
	}
	var sb strings.Builder
	for _, elem := range doc.Body.Content {
		if elem.Paragraph == nil {
			continue
		}
		for _, pe := range elem.Paragraph.Elements {
			if pe.TextRun != nil {
				sb.WriteString(pe.TextRun.Content)
			}
		}
	}
	return sb.String()
}

// --- docs_append ---

func appendTool() mcp.Tool {
	return mcp.Tool{
		Name:        "docs_append",
		Description: "Append text to the end of an existing Google Doc.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"document_id": map[string]any{
					"type":        "string",
					"description": "The document ID",
				},
				"text": map[string]any{
					"type":        "string",
					"description": "Text to append to the end of the document",
				},
			},
			Required: []string{"document_id", "text"},
		},
	}
}

type appendResult struct {
	DocumentID string `json:"document_id"`
	Appended   bool   `json:"appended"`
}

func appendHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		documentID, err := requiredString(req, "document_id")
		if err != nil {
			return toolError(err.Error()), nil
		}
		text, err := requiredString(req, "text")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newDocsService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating docs service: %w", err)
		}

		// Get the document to find the end index.
		doc, err := srv.Documents.Get(documentID).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to get document: %v", err)), nil
		}

		// The end index of the body content (minus 1 for the trailing newline).
		endIndex := doc.Body.Content[len(doc.Body.Content)-1].EndIndex - 1

		batchReq := &docsapi.BatchUpdateDocumentRequest{
			Requests: []*docsapi.Request{
				{
					InsertText: &docsapi.InsertTextRequest{
						Text: text,
						Location: &docsapi.Location{
							Index: endIndex,
						},
					},
				},
			},
		}
		_, err = srv.Documents.BatchUpdate(documentID, batchReq).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to append text: %v", err)), nil
		}

		return jsonResult(appendResult{
			DocumentID: documentID,
			Appended:   true,
		})
	}
}

// --- docs_replace ---

func replaceTool() mcp.Tool {
	return mcp.Tool{
		Name:        "docs_replace",
		Description: "Find and replace all occurrences of a text string in a Google Doc.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"document_id": map[string]any{
					"type":        "string",
					"description": "The document ID",
				},
				"find_text": map[string]any{
					"type":        "string",
					"description": "The text to find",
				},
				"replace_text": map[string]any{
					"type":        "string",
					"description": "The text to replace with",
				},
			},
			Required: []string{"document_id", "find_text", "replace_text"},
		},
	}
}

type replaceResult struct {
	DocumentID       string `json:"document_id"`
	OccurrencesFound int64  `json:"occurrences_found"`
}

func replaceHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		documentID, err := requiredString(req, "document_id")
		if err != nil {
			return toolError(err.Error()), nil
		}
		findText, err := requiredString(req, "find_text")
		if err != nil {
			return toolError(err.Error()), nil
		}
		replaceText, err := requiredString(req, "replace_text")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newDocsService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating docs service: %w", err)
		}

		batchReq := &docsapi.BatchUpdateDocumentRequest{
			Requests: []*docsapi.Request{
				{
					ReplaceAllText: &docsapi.ReplaceAllTextRequest{
						ContainsText: &docsapi.SubstringMatchCriteria{
							Text:      findText,
							MatchCase: true,
						},
						ReplaceText: replaceText,
					},
				},
			},
		}

		resp, err := srv.Documents.BatchUpdate(documentID, batchReq).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to replace text: %v", err)), nil
		}

		var found int64
		if len(resp.Replies) > 0 && resp.Replies[0].ReplaceAllText != nil {
			found = resp.Replies[0].ReplaceAllText.OccurrencesChanged
		}

		return jsonResult(replaceResult{
			DocumentID:       documentID,
			OccurrencesFound: found,
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

func optionalString(req mcp.CallToolRequest, key, defaultVal string) string {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok {
		return defaultVal
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return defaultVal
	}
	return s
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
