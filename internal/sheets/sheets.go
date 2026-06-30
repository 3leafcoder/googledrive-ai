package sheets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	sheetsapi "google.golang.org/api/sheets/v4"
)

// RegisterTools adds all Sheets read tools to the MCP server.
func RegisterTools(s *server.MCPServer, ts oauth2.TokenSource) {
	s.AddTool(listSheetsTool(), listSheetsHandler(ts))
	s.AddTool(readRangeTool(), readRangeHandler(ts))
	s.AddTool(getMetadataTool(), getMetadataHandler(ts))
}

func newService(ctx context.Context, ts oauth2.TokenSource) (*sheetsapi.Service, error) {
	return sheetsapi.NewService(ctx, option.WithTokenSource(ts))
}

// --- sheets_list_sheets ---

func listSheetsTool() mcp.Tool {
	return mcp.Tool{
		Name:        "sheets_list_sheets",
		Description: "List all tabs (sheets) in a spreadsheet with their names, indexes, and dimensions.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"spreadsheet_id": map[string]any{
					"type":        "string",
					"description": "The spreadsheet ID",
				},
			},
			Required: []string{"spreadsheet_id"},
		},
	}
}

type sheetInfo struct {
	Name     string `json:"name"`
	Index    int64  `json:"index"`
	RowCount int64  `json:"row_count"`
	ColCount int64  `json:"col_count"`
}

func listSheetsHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spreadsheetID, err := requiredString(req, "spreadsheet_id")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating sheets service: %w", err)
		}

		sp, err := srv.Spreadsheets.Get(spreadsheetID).Fields("sheets.properties").Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to get spreadsheet: %v", err)), nil
		}

		sheets := make([]sheetInfo, len(sp.Sheets))
		for i, s := range sp.Sheets {
			sheets[i] = sheetInfo{
				Name:     s.Properties.Title,
				Index:    s.Properties.Index,
				RowCount: s.Properties.GridProperties.RowCount,
				ColCount: s.Properties.GridProperties.ColumnCount,
			}
		}

		return jsonResult(sheets)
	}
}

// --- sheets_read_range ---

func readRangeTool() mcp.Tool {
	return mcp.Tool{
		Name:        "sheets_read_range",
		Description: "Read a range of cells from a spreadsheet tab using A1 notation (e.g., Sheet1!A1:C10). Returns display values by default, or raw formulas with render_option=formula.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"spreadsheet_id": map[string]any{
					"type":        "string",
					"description": "The spreadsheet ID",
				},
				"range": map[string]any{
					"type":        "string",
					"description": "A1 notation range (e.g., Sheet1!A1:D10)",
				},
				"render_option": map[string]any{
					"type":        "string",
					"description": "\"formatted\" (default) for display values, \"formula\" for raw formulas",
					"enum":        []string{"formatted", "formula"},
					"default":     "formatted",
				},
			},
			Required: []string{"spreadsheet_id", "range"},
		},
	}
}

type rangeResult struct {
	Range  string          `json:"range"`
	Values [][]any `json:"values"`
}

func readRangeHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spreadsheetID, err := requiredString(req, "spreadsheet_id")
		if err != nil {
			return toolError(err.Error()), nil
		}
		rangeStr, err := requiredString(req, "range")
		if err != nil {
			return toolError(err.Error()), nil
		}
		renderOption := optionalString(req, "render_option", "formatted")

		var apiRenderOption string
		switch renderOption {
		case "formula":
			apiRenderOption = "FORMULA"
		default:
			apiRenderOption = "FORMATTED_VALUE"
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating sheets service: %w", err)
		}

		resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeStr).
			ValueRenderOption(apiRenderOption).
			Context(ctx).
			Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to read range: %v", err)), nil
		}

		result := rangeResult{
			Range:  resp.Range,
			Values: resp.Values,
		}

		return jsonResult(result)
	}
}

// --- sheets_get_metadata ---

func getMetadataTool() mcp.Tool {
	return mcp.Tool{
		Name:        "sheets_get_metadata",
		Description: "Get spreadsheet metadata including title, locale, named ranges, and per-sheet properties.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"spreadsheet_id": map[string]any{
					"type":        "string",
					"description": "The spreadsheet ID",
				},
			},
			Required: []string{"spreadsheet_id"},
		},
	}
}

type metadataResult struct {
	Title       string           `json:"title"`
	Locale      string           `json:"locale"`
	NamedRanges []namedRangeInfo `json:"named_ranges"`
	Sheets      []sheetDetail    `json:"sheets"`
}

type namedRangeInfo struct {
	Name    string `json:"name"`
	Range   string `json:"range"`
	SheetID int64  `json:"sheet_id"`
}

type sheetDetail struct {
	SheetID  int64  `json:"sheet_id"`
	Title    string `json:"title"`
	Index    int64  `json:"index"`
	RowCount int64  `json:"row_count"`
	ColCount int64  `json:"col_count"`
	Hidden   bool   `json:"hidden"`
}

func getMetadataHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spreadsheetID, err := requiredString(req, "spreadsheet_id")
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating sheets service: %w", err)
		}

		sp, err := srv.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
		if err != nil {
			return toolError(fmt.Sprintf("failed to get spreadsheet: %v", err)), nil
		}

		result := metadataResult{
			Title:  sp.Properties.Title,
			Locale: sp.Properties.Locale,
		}

		for _, nr := range sp.NamedRanges {
			info := namedRangeInfo{
				Name:    nr.Name,
				SheetID: nr.Range.SheetId,
			}
			if nr.Range != nil {
				info.Range = formatGridRange(nr.Range, sp.Sheets)
			}
			result.NamedRanges = append(result.NamedRanges, info)
		}

		for _, s := range sp.Sheets {
			detail := sheetDetail{
				SheetID: s.Properties.SheetId,
				Title:   s.Properties.Title,
				Index:   s.Properties.Index,
				Hidden:  s.Properties.Hidden,
			}
			if s.Properties.GridProperties != nil {
				detail.RowCount = s.Properties.GridProperties.RowCount
				detail.ColCount = s.Properties.GridProperties.ColumnCount
			}
			result.Sheets = append(result.Sheets, detail)
		}

		return jsonResult(result)
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

// formatGridRange converts a GridRange to a human-readable string.
func formatGridRange(gr *sheetsapi.GridRange, sheets []*sheetsapi.Sheet) string {
	sheetTitle := ""
	for _, s := range sheets {
		if s.Properties.SheetId == gr.SheetId {
			sheetTitle = s.Properties.Title
			break
		}
	}
	if sheetTitle == "" {
		sheetTitle = fmt.Sprintf("Sheet%d", gr.SheetId)
	}

	startCol := colLetter(gr.StartColumnIndex)
	endCol := colLetter(gr.EndColumnIndex - 1)
	return fmt.Sprintf("%s!%s%d:%s%d", sheetTitle, startCol, gr.StartRowIndex+1, endCol, gr.EndRowIndex)
}

// colLetter converts a 0-based column index to A1-notation letters (0=A, 25=Z, 26=AA).
func colLetter(idx int64) string {
	result := ""
	idx++ // convert to 1-based
	for idx > 0 {
		idx--
		result = string(rune('A'+idx%26)) + result
		idx /= 26
	}
	return result
}
