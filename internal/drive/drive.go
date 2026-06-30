package drive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"
	driveapi "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// RegisterTools adds all Drive tools to the MCP server.
func RegisterTools(s *server.MCPServer, ts oauth2.TokenSource) {
	s.AddTool(listTool(), listHandler(ts))
	s.AddTool(searchTool(), searchHandler(ts))
	s.AddTool(getTool(), getHandler(ts))
	s.AddTool(createFolderTool(), createFolderHandler(ts))
	s.AddTool(uploadTool(), uploadHandler(ts))
	s.AddTool(downloadTool(), downloadHandler(ts))
	s.AddTool(deleteTool(), deleteHandler(ts))
	s.AddTool(moveTool(), moveHandler(ts))
}

func newService(ctx context.Context, ts oauth2.TokenSource) (*driveapi.Service, error) {
	return driveapi.NewService(ctx, option.WithTokenSource(ts))
}

// --- drive_list ---

func listTool() mcp.Tool {
	return mcp.Tool{
		Name:        "drive_list",
		Description: "List files and folders in a Google Drive folder. Defaults to root if no folder_id is provided.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"folder_id": map[string]any{
					"type":        "string",
					"description": "The folder ID to list. Defaults to root.",
				},
			},
		},
	}
}

type fileEntry struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ModifiedTime string `json:"modified_time"`
}

func listHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		folderID := optionalString(req, "folder_id", "root")

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}

		query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
		resp, err := srv.Files.List().
			Q(query).
			Fields("files(id, name, mimeType, modifiedTime)").
			OrderBy("name").
			PageSize(100).
			Context(ctx).
			Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to list files: %v", err)), nil
		}

		entries := make([]fileEntry, len(resp.Files))
		for i, f := range resp.Files {
			entries[i] = fileEntry{
				ID:           f.Id,
				Name:         f.Name,
				Type:         mimeToType(f.MimeType),
				ModifiedTime: f.ModifiedTime,
			}
		}

		return jsonResult(entries)
	}
}

// --- drive_search ---

func searchTool() mcp.Tool {
	return mcp.Tool{
		Name:        "drive_search",
		Description: "Search Google Drive by query string. Searches file names and content.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query (searches name and content)",
				},
			},
			Required: []string{"query"},
		},
	}
}

func searchHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := requiredString(req, "query")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}

		driveQuery := fmt.Sprintf("fullText contains '%s' and trashed = false", escapeQuery(query))
		resp, err := srv.Files.List().
			Q(driveQuery).
			Fields("files(id, name, mimeType, modifiedTime)").
			PageSize(50).
			Context(ctx).
			Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to search: %v", err)), nil
		}

		entries := make([]fileEntry, len(resp.Files))
		for i, f := range resp.Files {
			entries[i] = fileEntry{
				ID:           f.Id,
				Name:         f.Name,
				Type:         mimeToType(f.MimeType),
				ModifiedTime: f.ModifiedTime,
			}
		}

		return jsonResult(entries)
	}
}

// --- drive_get ---

func getTool() mcp.Tool {
	return mcp.Tool{
		Name:        "drive_get",
		Description: "Get metadata for a file or folder by ID.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"file_id": map[string]any{
					"type":        "string",
					"description": "The file or folder ID",
				},
			},
			Required: []string{"file_id"},
		},
	}
}

type fileMetadata struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	MimeType     string   `json:"mime_type"`
	Type         string   `json:"type"`
	Size         int64    `json:"size"`
	ModifiedTime string   `json:"modified_time"`
	CreatedTime  string   `json:"created_time"`
	Parents      []string `json:"parents"`
	WebViewLink  string   `json:"web_view_link"`
}

func getHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		fileID, err := requiredString(req, "file_id")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}

		f, err := srv.Files.Get(fileID).
			Fields("id, name, mimeType, size, modifiedTime, createdTime, parents, webViewLink").
			Context(ctx).
			Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to get file: %v", err)), nil
		}

		return jsonResult(fileMetadata{
			ID:           f.Id,
			Name:         f.Name,
			MimeType:     f.MimeType,
			Type:         mimeToType(f.MimeType),
			Size:         f.Size,
			ModifiedTime: f.ModifiedTime,
			CreatedTime:  f.CreatedTime,
			Parents:      f.Parents,
			WebViewLink:  f.WebViewLink,
		})
	}
}

// --- drive_create_folder ---

func createFolderTool() mcp.Tool {
	return mcp.Tool{
		Name:        "drive_create_folder",
		Description: "Create a new folder in Google Drive.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name of the new folder",
				},
				"parent_id": map[string]any{
					"type":        "string",
					"description": "Optional parent folder ID. Defaults to root.",
				},
			},
			Required: []string{"name"},
		},
	}
}

type createFolderResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func createFolderHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := requiredString(req, "name")
		if err != nil {
			return toolError(err.Error()), nil
		}
		parentID := optionalString(req, "parent_id", "")

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}

		file := &driveapi.File{
			Name:     name,
			MimeType: "application/vnd.google-apps.folder",
		}
		if parentID != "" {
			file.Parents = []string{parentID}
		}

		created, err := srv.Files.Create(file).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to create folder: %v", err)), nil
		}

		return jsonResult(createFolderResult{
			ID:   created.Id,
			Name: created.Name,
		})
	}
}

// --- drive_upload ---

func uploadTool() mcp.Tool {
	return mcp.Tool{
		Name:        "drive_upload",
		Description: "Upload a file to Google Drive. Content should be provided as base64-encoded data.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name for the uploaded file",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Base64-encoded file content",
				},
				"mime_type": map[string]any{
					"type":        "string",
					"description": "MIME type of the file (e.g., text/plain, application/pdf)",
				},
				"parent_id": map[string]any{
					"type":        "string",
					"description": "Optional parent folder ID. Defaults to root.",
				},
			},
			Required: []string{"name", "content", "mime_type"},
		},
	}
}

type uploadResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func uploadHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := requiredString(req, "name")
		if err != nil {
			return toolError(err.Error()), nil
		}
		content, err := requiredString(req, "content")
		if err != nil {
			return toolError(err.Error()), nil
		}
		mimeType, err := requiredString(req, "mime_type")
		if err != nil {
			return toolError(err.Error()), nil
		}
		parentID := optionalString(req, "parent_id", "")

		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return toolError(fmt.Sprintf("invalid base64 content: %v", err)), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}

		file := &driveapi.File{
			Name:     name,
			MimeType: mimeType,
		}
		if parentID != "" {
			file.Parents = []string{parentID}
		}

		created, err := srv.Files.Create(file).
			Media(strings.NewReader(string(decoded))).
			Context(ctx).
			Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to upload file: %v", err)), nil
		}

		return jsonResult(uploadResult{
			ID:   created.Id,
			Name: created.Name,
			Size: created.Size,
		})
	}
}

// --- drive_download ---

func downloadTool() mcp.Tool {
	return mcp.Tool{
		Name:        "drive_download",
		Description: "Download a file's content from Google Drive. Returns base64-encoded content for binary files, or plain text for Google Docs/Sheets exports.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"file_id": map[string]any{
					"type":        "string",
					"description": "The file ID to download",
				},
				"export_mime_type": map[string]any{
					"type":        "string",
					"description": "For Google Workspace files, the MIME type to export as (e.g., text/plain, application/pdf). Required for Docs, Sheets, Slides.",
				},
			},
			Required: []string{"file_id"},
		},
	}
}

type downloadResult struct {
	FileID   string `json:"file_id"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func downloadHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		fileID, err := requiredString(req, "file_id")
		if err != nil {
			return toolError(err.Error()), nil
		}
		exportMime := optionalString(req, "export_mime_type", "")

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}

		var resp *io.ReadCloser

		if exportMime != "" {
			httpResp, err := srv.Files.Export(fileID, exportMime).Context(ctx).Download()
			if err != nil {
				return toolError(fmt.Sprintf("failed to export file: %v", err)), nil
			}
			resp = &httpResp.Body
		} else {
			httpResp, err := srv.Files.Get(fileID).Context(ctx).Download()
			if err != nil {
				return toolError(fmt.Sprintf("failed to download file: %v", err)), nil
			}
			resp = &httpResp.Body
		}
		defer (*resp).Close()

		data, err := io.ReadAll(*resp)
		if err != nil {
			return toolError(fmt.Sprintf("failed to read file content: %v", err)), nil
		}

		// If content is valid UTF-8 text, return as plain text
		if isText(data) {
			return jsonResult(downloadResult{
				FileID:   fileID,
				Content:  string(data),
				Encoding: "text",
			})
		}

		return jsonResult(downloadResult{
			FileID:   fileID,
			Content:  base64.StdEncoding.EncodeToString(data),
			Encoding: "base64",
		})
	}
}

// --- drive_delete ---

func deleteTool() mcp.Tool {
	return mcp.Tool{
		Name:        "drive_delete",
		Description: "Move a file or folder to trash in Google Drive.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"file_id": map[string]any{
					"type":        "string",
					"description": "The file or folder ID to trash",
				},
			},
			Required: []string{"file_id"},
		},
	}
}

type deleteResult struct {
	FileID  string `json:"file_id"`
	Trashed bool   `json:"trashed"`
}

func deleteHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		fileID, err := requiredString(req, "file_id")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}

		_, err = srv.Files.Update(fileID, &driveapi.File{Trashed: true}).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to trash file: %v", err)), nil
		}

		return jsonResult(deleteResult{
			FileID:  fileID,
			Trashed: true,
		})
	}
}

// --- drive_move ---

func moveTool() mcp.Tool {
	return mcp.Tool{
		Name:        "drive_move",
		Description: "Move a file to a different parent folder in Google Drive.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"file_id": map[string]any{
					"type":        "string",
					"description": "The file ID to move",
				},
				"new_parent_id": map[string]any{
					"type":        "string",
					"description": "The destination folder ID",
				},
			},
			Required: []string{"file_id", "new_parent_id"},
		},
	}
}

type moveResult struct {
	FileID      string `json:"file_id"`
	NewParentID string `json:"new_parent_id"`
}

func moveHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		fileID, err := requiredString(req, "file_id")
		if err != nil {
			return toolError(err.Error()), nil
		}
		newParentID, err := requiredString(req, "new_parent_id")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}

		// Get current parents to remove them.
		f, err := srv.Files.Get(fileID).Fields("parents").Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to get file: %v", err)), nil
		}

		previousParents := strings.Join(f.Parents, ",")
		_, err = srv.Files.Update(fileID, &driveapi.File{}).
			AddParents(newParentID).
			RemoveParents(previousParents).
			Context(ctx).
			Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to move file: %v", err)), nil
		}

		return jsonResult(moveResult{
			FileID:      fileID,
			NewParentID: newParentID,
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

// mimeToType converts a Google Drive MIME type to a simple type label.
func mimeToType(mime string) string {
	switch mime {
	case "application/vnd.google-apps.folder":
		return "folder"
	case "application/vnd.google-apps.document":
		return "document"
	case "application/vnd.google-apps.spreadsheet":
		return "spreadsheet"
	case "application/vnd.google-apps.presentation":
		return "presentation"
	case "application/vnd.google-apps.form":
		return "form"
	default:
		return "file"
	}
}

// escapeQuery escapes single quotes in Drive API query strings.
func escapeQuery(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

// isText checks if data appears to be valid UTF-8 text.
func isText(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	// Check for null bytes — a strong indicator of binary data.
	for _, b := range data {
		if b == 0 {
			return false
		}
	}
	return true
}
