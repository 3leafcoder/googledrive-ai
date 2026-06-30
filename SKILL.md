# Google Drive MCP — Tool Reference

Use this reference when calling `googledrive-mcp` MCP tools. All tool names, parameter names, and values are exact.

## Quick Reference

| Tool | Purpose |
|------|---------|
| `sheets_list_sheets` | List all tabs with dimensions |
| `sheets_get_metadata` | Workbook metadata, named ranges, sheet properties |
| `sheets_read_range` | Read cells (values or formulas) |
| `sheets_write_range` | Write to a specific range |
| `sheets_append_rows` | Append rows after last data row |
| `sheets_add_sheet` | Add a new tab |
| `sheets_create` | Create a new workbook |
| `sheets_batch_update` | Multiple reads/writes in one call |

---

## A1 Range Notation

Ranges always include the sheet (tab) name:

| Pattern | Example | Meaning |
|---------|---------|---------|
| Cell | `Sheet1!A1` | Single cell |
| Range | `Sheet1!A1:C10` | Rectangle of cells |
| Entire column | `Sheet1!A:A` | All of column A |
| Multiple columns | `Sheet1!A:C` | Columns A through C |
| Entire row | `Sheet1!1:1` | All of row 1 |
| Multiple rows | `Sheet1!1:5` | Rows 1 through 5 |
| Tab with spaces | `'My Sheet'!A1:C3` | Quote the name |

---

## Key Concepts

### `render_option` — How to Read Cells

Used by `sheets_read_range` and `sheets_batch_update` read operations.

| Value | Returns | Use when |
|-------|---------|----------|
| `"formatted"` (default) | Display values as shown in the UI | You want the computed results — numbers, dates, text |
| `"formula"` | Raw cell contents including formulas | You want to audit/copy formulas like `=SUM(A1:A10)` |

### `USER_ENTERED` — How Writes Work

All write operations (`sheets_write_range`, `sheets_append_rows`, `sheets_batch_update` writes) use `ValueInputOption: USER_ENTERED` internally. This means:

- Writing `"=SUM(A1:A10)"` creates a **live formula**, not literal text
- Writing `"100"` stores the number 100
- Writing `"hello"` stores text
- Formulas in adjacent cells are **not** clobbered

---

## Tool Details

### sheets_list_sheets

List all tabs in a spreadsheet with their names and dimensions.

```json
{
  "spreadsheet_id": "abc123"
}
```

**Parameters:**
- `spreadsheet_id` (string, required) — The spreadsheet ID

---

### sheets_get_metadata

Get spreadsheet title, locale, named ranges, and per-sheet properties (including hidden tabs).

```json
{
  "spreadsheet_id": "abc123"
}
```

**Parameters:**
- `spreadsheet_id` (string, required) — The spreadsheet ID

---

### sheets_read_range

Read a range of cells. Returns display values by default, or raw formulas.

```json
{
  "spreadsheet_id": "abc123",
  "range": "Sheet1!A1:D10",
  "render_option": "formatted"
}
```

**Parameters:**
- `spreadsheet_id` (string, required) — The spreadsheet ID
- `range` (string, required) — A1 notation range, e.g. `Sheet1!A1:D10`
- `render_option` (string, optional) — `"formatted"` (default) or `"formula"`

---

### sheets_write_range

Write values to a specific range. Formulas are interpreted.

```json
{
  "spreadsheet_id": "abc123",
  "range": "Sheet1!A1:C2",
  "values": [
    ["Item", "Qty", "Total"],
    ["Widgets", "100", "=B2*5.99"]
  ]
}
```

**Parameters:**
- `spreadsheet_id` (string, required) — The spreadsheet ID
- `range` (string, required) — A1 notation range, e.g. `Sheet1!A1:C3`
- `values` (string[][], required) — 2D array of values (rows of columns)

---

### sheets_append_rows

Append rows after the last row with data on a tab. Finds the end of existing data automatically.

```json
{
  "spreadsheet_id": "abc123",
  "sheet_name": "Log",
  "values": [
    ["2026-06-30", "Deployment", "success"],
    ["2026-06-30", "Migration", "success"]
  ]
}
```

**Parameters:**
- `spreadsheet_id` (string, required) — The spreadsheet ID
- `sheet_name` (string, required) — Tab name to append to (e.g. `Sheet1`)
- `values` (string[][], required) — 2D array of rows to append

---

### sheets_add_sheet

Add a new tab to an existing spreadsheet.

```json
{
  "spreadsheet_id": "abc123",
  "title": "Q4 Forecast"
}
```

**Parameters:**
- `spreadsheet_id` (string, required) — The spreadsheet ID
- `title` (string, required) — Name for the new tab

---

### sheets_create

Create a new spreadsheet (workbook) with optional initial tabs.

```json
{
  "title": "Q3 Budget",
  "sheet_names": ["Summary", "Revenue", "Expenses"]
}
```

**Parameters:**
- `title` (string, required) — Spreadsheet title
- `sheet_names` (string[], optional) — Tab names to create. Defaults to one tab named `Sheet1`

---

### sheets_batch_update

Execute multiple read and write operations in a single call. Each operation specifies an `action` (`"read"` or `"write"`), a `range`, and optionally `render_option` (for reads) or `values` (for writes).

```json
{
  "spreadsheet_id": "abc123",
  "operations": [
    { "action": "read",  "range": "Revenue!A1:D20" },
    { "action": "read",  "range": "Expenses!A1:D20" },
    { "action": "read",  "range": "Revenue!A1:D20", "render_option": "formula" },
    { "action": "write", "range": "Summary!A1:B2", "values": [["Category", "Total"], ["Revenue", "=Revenue!D20"]] }
  ]
}
```

**Parameters:**
- `spreadsheet_id` (string, required) — The spreadsheet ID
- `operations` (object[], required) — List of operations. Each has:
  - `action` (string, required) — `"read"` or `"write"`
  - `range` (string, required) — A1 notation range
  - `render_option` (string, optional) — For reads: `"formatted"` (default) or `"formula"`
  - `values` (string[][], required for writes) — 2D array of values

---

## Common Recipes

### Read a summary tab and write results to another tab

```json
{
  "spreadsheet_id": "abc123",
  "operations": [
    { "action": "read", "range": "Data!A1:D100" },
    { "action": "write", "range": "Summary!A1:B3", "values": [
      ["Metric", "Value"],
      ["Total Revenue", "=SUM(Data!D2:D100)"],
      ["Row Count", "=COUNTA(Data!A2:A100)"]
    ]}
  ]
}
```

### Append a row to a log sheet

```json
{
  "name": "sheets_append_rows",
  "arguments": {
    "spreadsheet_id": "abc123",
    "sheet_name": "Log",
    "values": [["2026-06-30", "Backup completed", "success"]]
  }
}
```

### Read formulas from one tab, computed values from another

```json
{
  "spreadsheet_id": "abc123",
  "operations": [
    { "action": "read", "range": "Budget!A1:D10", "render_option": "formula" },
    { "action": "read", "range": "Budget!A1:D10", "render_option": "formatted" }
  ]
}
```

Compare the two results: the first shows formulas like `=SUM(B2:B9)`, the second shows the computed value like `42500`.

### Create a new workbook with multiple tabs

```json
{
  "name": "sheets_create",
  "arguments": {
    "title": "Project Tracker",
    "sheet_names": ["Tasks", "Timeline", "Budget", "Notes"]
  }
}
```

### Batch read across multiple tabs

```json
{
  "spreadsheet_id": "abc123",
  "operations": [
    { "action": "read", "range": "Q1!A1:D50" },
    { "action": "read", "range": "Q2!A1:D50" },
    { "action": "read", "range": "Q3!A1:D50" },
    { "action": "read", "range": "Q4!A1:D50" }
  ]
}
```

One API call instead of four separate `sheets_read_range` calls.

---

## Gotchas and Tips

- **Use `"formula"` render option to audit formulas.** Default `"formatted"` only shows computed values — you can't see what formula a cell contains without switching to `"formula"`.

- **`sheets_append_rows` doesn't overwrite — it finds the last row.** It detects where data ends and inserts below. Safe to call repeatedly without tracking row numbers.

- **`sheets_append_rows` uses `sheet_name`, not `range`.** Pass just the tab name (e.g. `"Sheet1"`), not A1 notation. The tool figures out where to append.

- **`sheets_add_sheet` uses `title`, not `sheet_name`.** The parameter name is `title` for the new tab name.

- **Tab names with spaces need quotes in ranges.** Use `'My Sheet'!A1:C3` — single quotes around the tab name inside the range string.

- **All writes interpret formulas.** If you write `"=NOW()"`, it becomes a live formula. To write literal text that starts with `=`, there's no escape — the API always interprets it.

- **Batch operations preserve order.** Results come back in the same order as the operations array, mixing reads and writes freely.

- **Reads in `sheets_batch_update` are grouped by render option.** Internally the API can only use one render option per batch-get call, so formatted and formula reads are split into separate API calls. This is transparent to you — just set `render_option` per operation.

- **The `values` parameter is always a 2D array.** Even for a single cell, wrap it: `[["value"]]`. Even for a single row: `[["a", "b", "c"]]`.
