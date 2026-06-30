package sheets

import (
	"testing"
)

func TestWriteToolDefinitions(t *testing.T) {
	wr := writeRangeTool()
	if wr.Name != "sheets_write_range" {
		t.Errorf("write range tool name = %q", wr.Name)
	}
	if len(wr.InputSchema.Required) != 3 {
		t.Errorf("write range required = %v, want [spreadsheet_id, range, values]", wr.InputSchema.Required)
	}

	ar := appendRowsTool()
	if ar.Name != "sheets_append_rows" {
		t.Errorf("append rows tool name = %q", ar.Name)
	}
	if len(ar.InputSchema.Required) != 3 {
		t.Errorf("append rows required = %v, want [spreadsheet_id, sheet_name, values]", ar.InputSchema.Required)
	}

	as := addSheetTool()
	if as.Name != "sheets_add_sheet" {
		t.Errorf("add sheet tool name = %q", as.Name)
	}
	if len(as.InputSchema.Required) != 2 {
		t.Errorf("add sheet required = %v, want [spreadsheet_id, title]", as.InputSchema.Required)
	}

	cs := createSpreadsheetTool()
	if cs.Name != "sheets_create" {
		t.Errorf("create spreadsheet tool name = %q", cs.Name)
	}
	if len(cs.InputSchema.Required) != 1 || cs.InputSchema.Required[0] != "title" {
		t.Errorf("create spreadsheet required = %v, want [title]", cs.InputSchema.Required)
	}
}

func TestRequired2DArray(t *testing.T) {
	// Missing parameter
	_, err := required2DArray(fakeRequest(nil), "values")
	if err == nil {
		t.Error("expected error for missing parameter")
	}

	// Non-array parameter
	_, err = required2DArray(fakeRequest(map[string]any{"values": "not-array"}), "values")
	if err == nil {
		t.Error("expected error for non-array parameter")
	}

	// Row is not an array
	_, err = required2DArray(fakeRequest(map[string]any{"values": []any{"not-array"}}), "values")
	if err == nil {
		t.Error("expected error for non-array row")
	}

	// Valid 2D array
	input := []any{
		[]any{"a", "b"},
		[]any{"c", "d"},
	}
	vals, err := required2DArray(fakeRequest(map[string]any{"values": input}), "values")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vals) != 2 || len(vals[0]) != 2 {
		t.Errorf("got %d rows with %d cols, want 2x2", len(vals), len(vals[0]))
	}
	if vals[0][0] != "a" || vals[1][1] != "d" {
		t.Errorf("unexpected values: %v", vals)
	}
}

func TestOptionalStringArray(t *testing.T) {
	// Missing returns nil
	result := optionalStringArray(fakeRequest(nil), "names")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}

	// Non-array returns nil
	result = optionalStringArray(fakeRequest(map[string]any{"names": "not-array"}), "names")
	if result != nil {
		t.Errorf("expected nil for non-array, got %v", result)
	}

	// Valid array
	result = optionalStringArray(fakeRequest(map[string]any{
		"names": []any{"Sheet1", "Sheet2"},
	}), "names")
	if len(result) != 2 || result[0] != "Sheet1" || result[1] != "Sheet2" {
		t.Errorf("got %v, want [Sheet1, Sheet2]", result)
	}
}
