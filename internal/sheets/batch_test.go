package sheets

import (
	"testing"
)

func TestBatchUpdateToolDefinition(t *testing.T) {
	tool := batchUpdateTool()
	if tool.Name != "sheets_batch_update" {
		t.Errorf("tool name = %q, want sheets_batch_update", tool.Name)
	}
	if len(tool.InputSchema.Required) != 2 {
		t.Errorf("required = %v, want [spreadsheet_id, operations]", tool.InputSchema.Required)
	}
}

func TestParseOperations(t *testing.T) {
	t.Run("missing operations", func(t *testing.T) {
		_, err := parseOperations(fakeRequest(map[string]any{}))
		if err == nil {
			t.Error("expected error for missing operations")
		}
	})

	t.Run("empty operations array", func(t *testing.T) {
		_, err := parseOperations(fakeRequest(map[string]any{
			"operations": []any{},
		}))
		if err == nil {
			t.Error("expected error for empty operations")
		}
	})

	t.Run("invalid operation type", func(t *testing.T) {
		_, err := parseOperations(fakeRequest(map[string]any{
			"operations": []any{"not-an-object"},
		}))
		if err == nil {
			t.Error("expected error for non-object operation")
		}
	})

	t.Run("invalid action", func(t *testing.T) {
		_, err := parseOperations(fakeRequest(map[string]any{
			"operations": []any{
				map[string]any{"action": "delete", "range": "A1"},
			},
		}))
		if err == nil {
			t.Error("expected error for invalid action")
		}
	})

	t.Run("write without values", func(t *testing.T) {
		_, err := parseOperations(fakeRequest(map[string]any{
			"operations": []any{
				map[string]any{"action": "write", "range": "Sheet1!A1"},
			},
		}))
		if err == nil {
			t.Error("expected error for write without values")
		}
	})

	t.Run("valid mixed operations", func(t *testing.T) {
		ops, err := parseOperations(fakeRequest(map[string]any{
			"operations": []any{
				map[string]any{
					"action":        "read",
					"range":         "Sheet1!A1:C3",
					"render_option": "formula",
				},
				map[string]any{
					"action": "write",
					"range":  "Sheet2!A1:B2",
					"values": []any{
						[]any{"x", "y"},
						[]any{"z", "w"},
					},
				},
				map[string]any{
					"action": "read",
					"range":  "Sheet1!D1:D5",
				},
			},
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ops) != 3 {
			t.Fatalf("got %d ops, want 3", len(ops))
		}

		// First op: read with formula render
		if ops[0].Action != "read" || ops[0].Range != "Sheet1!A1:C3" || ops[0].RenderOption != "formula" {
			t.Errorf("op[0] = %+v", ops[0])
		}

		// Second op: write with values
		if ops[1].Action != "write" || ops[1].Range != "Sheet2!A1:B2" || len(ops[1].Values) != 2 {
			t.Errorf("op[1] = %+v", ops[1])
		}
		if ops[1].Values[0][0] != "x" || ops[1].Values[1][1] != "w" {
			t.Errorf("op[1].Values = %v", ops[1].Values)
		}

		// Third op: read with default render (formatted)
		if ops[2].Action != "read" || ops[2].RenderOption != "formatted" {
			t.Errorf("op[2] = %+v", ops[2])
		}
	})
}
