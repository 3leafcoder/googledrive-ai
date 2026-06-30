package sheets

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"
	sheetsapi "google.golang.org/api/sheets/v4"
)

// registerWriteTools adds all Sheets write tools to the MCP server.
func registerWriteTools(s *server.MCPServer, ts oauth2.TokenSource) {
	s.AddTool(writeRangeTool(), writeRangeHandler(ts))
	s.AddTool(appendRowsTool(), appendRowsHandler(ts))
	s.AddTool(addSheetTool(), addSheetHandler(ts))
	s.AddTool(createSpreadsheetTool(), createSpreadsheetHandler(ts))
	s.AddTool(batchUpdateTool(), batchUpdateHandler(ts))
}

// --- sheets_write_range ---

func writeRangeTool() mcp.Tool {
	return mcp.Tool{
		Name:        "sheets_write_range",
		Description: "Write values to a specific range on a spreadsheet tab using A1 notation. Formulas (e.g., =SUM(A1:A10)) are interpreted when written.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"spreadsheet_id": map[string]any{
					"type":        "string",
					"description": "The spreadsheet ID",
				},
				"range": map[string]any{
					"type":        "string",
					"description": "A1 notation range including sheet name (e.g., Sheet1!A1:C3)",
				},
				"values": map[string]any{
					"type":        "array",
					"description": "2D array of values to write (rows of columns)",
					"items": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
			},
			Required: []string{"spreadsheet_id", "range", "values"},
		},
	}
}

type writeResult struct {
	UpdatedRange string `json:"updated_range"`
	UpdatedCells int64  `json:"updated_cells"`
}

func writeRangeHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spreadsheetID, err := requiredString(req, "spreadsheet_id")
		if err != nil {
			return toolError(err.Error()), nil
		}
		rangeStr, err := requiredString(req, "range")
		if err != nil {
			return toolError(err.Error()), nil
		}
		values, err := required2DArray(req, "values")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating sheets service: %w", err)
		}

		vr := &sheetsapi.ValueRange{
			Values: values,
		}

		resp, err := srv.Spreadsheets.Values.Update(spreadsheetID, rangeStr, vr).
			ValueInputOption("USER_ENTERED").
			Context(ctx).
			Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to write range: %v", err)), nil
		}

		return jsonResult(writeResult{
			UpdatedRange: resp.UpdatedRange,
			UpdatedCells: resp.UpdatedCells,
		})
	}
}

// --- sheets_append_rows ---

func appendRowsTool() mcp.Tool {
	return mcp.Tool{
		Name:        "sheets_append_rows",
		Description: "Append rows after the last row with data on a specified tab.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"spreadsheet_id": map[string]any{
					"type":        "string",
					"description": "The spreadsheet ID",
				},
				"sheet_name": map[string]any{
					"type":        "string",
					"description": "Name of the tab to append to (e.g., Sheet1)",
				},
				"values": map[string]any{
					"type":        "array",
					"description": "2D array of rows to append",
					"items": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
			},
			Required: []string{"spreadsheet_id", "sheet_name", "values"},
		},
	}
}

type appendResult struct {
	UpdatedRange string `json:"updated_range"`
	UpdatedCells int64  `json:"updated_cells"`
}

func appendRowsHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spreadsheetID, err := requiredString(req, "spreadsheet_id")
		if err != nil {
			return toolError(err.Error()), nil
		}
		sheetName, err := requiredString(req, "sheet_name")
		if err != nil {
			return toolError(err.Error()), nil
		}
		values, err := required2DArray(req, "values")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating sheets service: %w", err)
		}

		vr := &sheetsapi.ValueRange{
			Values: values,
		}

		resp, err := srv.Spreadsheets.Values.Append(spreadsheetID, sheetName, vr).
			ValueInputOption("USER_ENTERED").
			InsertDataOption("INSERT_ROWS").
			Context(ctx).
			Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to append rows: %v", err)), nil
		}

		result := appendResult{}
		if resp.Updates != nil {
			result.UpdatedRange = resp.Updates.UpdatedRange
			result.UpdatedCells = resp.Updates.UpdatedCells
		}

		return jsonResult(result)
	}
}

// --- sheets_add_sheet ---

func addSheetTool() mcp.Tool {
	return mcp.Tool{
		Name:        "sheets_add_sheet",
		Description: "Add a new tab (sheet) to an existing spreadsheet.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"spreadsheet_id": map[string]any{
					"type":        "string",
					"description": "The spreadsheet ID",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Name for the new tab",
				},
			},
			Required: []string{"spreadsheet_id", "title"},
		},
	}
}

type addSheetResult struct {
	SheetID int64  `json:"sheet_id"`
	Title   string `json:"title"`
	Index   int64  `json:"index"`
}

func addSheetHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spreadsheetID, err := requiredString(req, "spreadsheet_id")
		if err != nil {
			return toolError(err.Error()), nil
		}
		title, err := requiredString(req, "title")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating sheets service: %w", err)
		}

		addReq := &sheetsapi.BatchUpdateSpreadsheetRequest{
			Requests: []*sheetsapi.Request{
				{
					AddSheet: &sheetsapi.AddSheetRequest{
						Properties: &sheetsapi.SheetProperties{
							Title: title,
						},
					},
				},
			},
		}

		resp, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, addReq).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to add sheet: %v", err)), nil
		}

		props := resp.Replies[0].AddSheet.Properties
		return jsonResult(addSheetResult{
			SheetID: props.SheetId,
			Title:   props.Title,
			Index:   props.Index,
		})
	}
}

// --- sheets_create ---

func createSpreadsheetTool() mcp.Tool {
	return mcp.Tool{
		Name:        "sheets_create",
		Description: "Create a new spreadsheet (workbook) with a title and optional initial sheet names.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "Title for the new spreadsheet",
				},
				"sheet_names": map[string]any{
					"type":        "array",
					"description": "Optional list of tab names to create (defaults to one sheet named Sheet1)",
					"items": map[string]any{
						"type": "string",
					},
				},
			},
			Required: []string{"title"},
		},
	}
}

type createResult struct {
	SpreadsheetID  string   `json:"spreadsheet_id"`
	SpreadsheetURL string   `json:"spreadsheet_url"`
	Title          string   `json:"title"`
	Sheets         []string `json:"sheets"`
}

func createSpreadsheetHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title, err := requiredString(req, "title")
		if err != nil {
			return toolError(err.Error()), nil
		}

		sheetNames := optionalStringArray(req, "sheet_names")

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating sheets service: %w", err)
		}

		sp := &sheetsapi.Spreadsheet{
			Properties: &sheetsapi.SpreadsheetProperties{
				Title: title,
			},
		}

		if len(sheetNames) > 0 {
			for _, name := range sheetNames {
				sp.Sheets = append(sp.Sheets, &sheetsapi.Sheet{
					Properties: &sheetsapi.SheetProperties{
						Title: name,
					},
				})
			}
		}

		resp, err := srv.Spreadsheets.Create(sp).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to create spreadsheet: %v", err)), nil
		}

		result := createResult{
			SpreadsheetID:  resp.SpreadsheetId,
			SpreadsheetURL: resp.SpreadsheetUrl,
			Title:          resp.Properties.Title,
		}
		for _, s := range resp.Sheets {
			result.Sheets = append(result.Sheets, s.Properties.Title)
		}

		return jsonResult(result)
	}
}

// --- helpers ---

// required2DArray extracts a required 2D array parameter from the request.
func required2DArray(req mcp.CallToolRequest, key string) ([][]any, error) {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: %s", key)
	}

	rows, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("parameter %s must be an array", key)
	}

	result := make([][]any, len(rows))
	for i, row := range rows {
		cols, ok := row.([]any)
		if !ok {
			return nil, fmt.Errorf("parameter %s[%d] must be an array", key, i)
		}
		result[i] = cols
	}

	return result, nil
}

// optionalStringArray extracts an optional string array parameter from the request.
func optionalStringArray(req mcp.CallToolRequest, key string) []string {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok {
		return nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}

	return result
}
