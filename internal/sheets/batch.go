package sheets

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"
	sheetsapi "google.golang.org/api/sheets/v4"
)

// --- sheets_batch_update ---

func batchUpdateTool() mcp.Tool {
	return mcp.Tool{
		Name:        "sheets_batch_update",
		Description: "Execute multiple read and write operations in a single batch call for efficient multi-range and multi-tab operations. Reads can mix formatted/formula render options. Writes use USER_ENTERED for formula safety.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"spreadsheet_id": map[string]any{
					"type":        "string",
					"description": "The spreadsheet ID",
				},
				"operations": map[string]any{
					"type":        "array",
					"description": "List of read/write operations to execute in batch",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"action": map[string]any{
								"type":        "string",
								"description": "\"read\" or \"write\"",
								"enum":        []string{"read", "write"},
							},
							"range": map[string]any{
								"type":        "string",
								"description": "A1 notation range including sheet name (e.g., Sheet1!A1:C3)",
							},
							"render_option": map[string]any{
								"type":        "string",
								"description": "For reads: \"formatted\" (default) or \"formula\"",
								"enum":        []string{"formatted", "formula"},
								"default":     "formatted",
							},
							"values": map[string]any{
								"type":        "array",
								"description": "For writes: 2D array of values",
								"items": map[string]any{
									"type": "array",
									"items": map[string]any{
										"type": "string",
									},
								},
							},
						},
						"required": []string{"action", "range"},
					},
				},
			},
			Required: []string{"spreadsheet_id", "operations"},
		},
	}
}

type batchOperation struct {
	Action       string      `json:"action"`
	Range        string      `json:"range"`
	RenderOption string      `json:"render_option"`
	Values       [][]any     `json:"values,omitempty"`
}

type batchOpResult struct {
	Action       string      `json:"action"`
	Range        string      `json:"range"`
	Values       [][]any     `json:"values,omitempty"`
	UpdatedCells int64       `json:"updated_cells,omitempty"`
}

type batchResult struct {
	Results []batchOpResult `json:"results"`
}

// parseOperations extracts and validates the operations array from the request.
func parseOperations(req mcp.CallToolRequest) ([]batchOperation, error) {
	args := req.GetArguments()
	v, ok := args["operations"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: operations")
	}

	opsRaw, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("parameter operations must be an array")
	}

	if len(opsRaw) == 0 {
		return nil, fmt.Errorf("operations array must not be empty")
	}

	ops := make([]batchOperation, len(opsRaw))
	for i, raw := range opsRaw {
		m, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("operations[%d] must be an object", i)
		}

		action, ok := m["action"].(string)
		if !ok || (action != "read" && action != "write") {
			return nil, fmt.Errorf("operations[%d].action must be \"read\" or \"write\"", i)
		}

		rangeStr, ok := m["range"].(string)
		if !ok || rangeStr == "" {
			return nil, fmt.Errorf("operations[%d].range must be a non-empty string", i)
		}

		op := batchOperation{
			Action:       action,
			Range:        rangeStr,
			RenderOption: "formatted",
		}

		if ro, ok := m["render_option"].(string); ok && ro != "" {
			op.RenderOption = ro
		}

		if action == "write" {
			vals, ok := m["values"]
			if !ok {
				return nil, fmt.Errorf("operations[%d]: write operations require \"values\"", i)
			}
			rows, ok := vals.([]any)
			if !ok {
				return nil, fmt.Errorf("operations[%d].values must be an array", i)
			}
			parsed := make([][]any, len(rows))
			for j, row := range rows {
				cols, ok := row.([]any)
				if !ok {
					return nil, fmt.Errorf("operations[%d].values[%d] must be an array", i, j)
				}
				parsed[j] = cols
			}
			op.Values = parsed
		}

		ops[i] = op
	}

	return ops, nil
}

func batchUpdateHandler(ts oauth2.TokenSource) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spreadsheetID, err := requiredString(req, "spreadsheet_id")
		if err != nil {
			return toolError(err.Error()), nil
		}

		ops, err := parseOperations(req)
		if err != nil {
			return toolError(err.Error()), nil
		}

		srv, err := newService(ctx, ts)
		if err != nil {
			return nil, fmt.Errorf("creating sheets service: %w", err)
		}

		results := make([]batchOpResult, len(ops))

		// Group reads by render option for efficient API calls.
		// BatchGet only supports one ValueRenderOption, so we need
		// separate calls for formatted vs formula reads.
		var formattedReads, formulaReads []readRef

		for i, op := range ops {
			if op.Action != "read" {
				continue
			}
			ref := readRef{opIndex: i, range_: op.Range}
			if op.RenderOption == "formula" {
				formulaReads = append(formulaReads, ref)
			} else {
				formattedReads = append(formattedReads, ref)
			}
		}

		// Execute batch reads.
		if err := executeBatchReads(ctx, srv, spreadsheetID, formattedReads, "FORMATTED_VALUE", results); err != nil {
			return toolError(err.Error()), nil
		}
		if err := executeBatchReads(ctx, srv, spreadsheetID, formulaReads, "FORMULA", results); err != nil {
			return toolError(err.Error()), nil
		}

		// Collect writes and execute as batch.
		writeData := []sheetsapi.ValueRange{}
		writeIndices := []int{}
		for i, op := range ops {
			if op.Action != "write" {
				continue
			}
			writeData = append(writeData, sheetsapi.ValueRange{
				Range:  op.Range,
				Values: op.Values,
			})
			writeIndices = append(writeIndices, i)
		}

		if len(writeData) > 0 {
			dataSlice := make([]*sheetsapi.ValueRange, len(writeData))
			for i := range writeData {
				dataSlice[i] = &writeData[i]
			}

			batchReq := &sheetsapi.BatchUpdateValuesRequest{
				ValueInputOption: "USER_ENTERED",
				Data:             dataSlice,
			}

			resp, err := srv.Spreadsheets.Values.BatchUpdate(spreadsheetID, batchReq).
				Context(ctx).
				Do()
			if err != nil {
				return toolError(fmt.Sprintf("failed to batch write: %v", err)), nil
			}

			for j, idx := range writeIndices {
				results[idx] = batchOpResult{
					Action: "write",
					Range:  ops[idx].Range,
				}
				if j < len(resp.Responses) {
					results[idx].Range = resp.Responses[j].UpdatedRange
					results[idx].UpdatedCells = resp.Responses[j].UpdatedCells
				}
			}
		}

		return jsonResult(batchResult{Results: results})
	}
}

type readRef struct {
	opIndex int
	range_  string
}

func executeBatchReads(ctx context.Context, srv *sheetsapi.Service, spreadsheetID string, refs []readRef, renderOption string, results []batchOpResult) error {
	if len(refs) == 0 {
		return nil
	}

	ranges := make([]string, len(refs))
	for i, ref := range refs {
		ranges[i] = ref.range_
	}

	resp, err := srv.Spreadsheets.Values.BatchGet(spreadsheetID).
		Ranges(ranges...).
		ValueRenderOption(renderOption).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to batch read: %v", err)
	}

	for i, ref := range refs {
		results[ref.opIndex] = batchOpResult{
			Action: "read",
		}
		if i < len(resp.ValueRanges) {
			results[ref.opIndex].Range = resp.ValueRanges[i].Range
			results[ref.opIndex].Values = resp.ValueRanges[i].Values
		}
	}

	return nil
}
