# Sheets Tools

Tools for reading and writing Google Sheets. Sheets is the primary use case for this server — all write operations are formula-safe and use `USER_ENTERED` input mode so formulas are interpreted correctly.

## Reading

### sheets_list_sheets

List all tabs (sheets) in a spreadsheet with their dimensions.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The spreadsheet ID |

**Example:**

```json
{
  "name": "sheets_list_sheets",
  "arguments": {
    "spreadsheet_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms"
  }
}
```

### sheets_get_metadata

Get spreadsheet metadata including named ranges, sheet properties, and dimensions.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The spreadsheet ID |

**Example:**

```json
{
  "name": "sheets_get_metadata",
  "arguments": {
    "spreadsheet_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms"
  }
}
```

### sheets_read_range

Read a range of cells from a specific tab.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The spreadsheet ID |
| `range` | string | Yes | A1 notation range (e.g., `Sheet1!A1:D10`) |
| `render_option` | string | No | `"formatted"` (default) for display values, `"formula"` for raw formulas |

**Example — read formatted values:**

```json
{
  "name": "sheets_read_range",
  "arguments": {
    "spreadsheet_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
    "range": "Sheet1!A1:D10"
  }
}
```

**Example — read formulas:**

```json
{
  "name": "sheets_read_range",
  "arguments": {
    "spreadsheet_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
    "range": "Sheet1!A1:D10",
    "render_option": "formula"
  }
}
```

## Writing

### sheets_create

Create a new spreadsheet.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | Yes | Spreadsheet title |
| `sheet_names` | string[] | No | Names for initial tabs. Defaults to one tab named "Sheet1" |

**Example:**

```json
{
  "name": "sheets_create",
  "arguments": {
    "title": "Q3 Budget",
    "sheet_names": ["Summary", "Revenue", "Expenses"]
  }
}
```

Returns the new spreadsheet ID.

### sheets_add_sheet

Add a new tab to an existing spreadsheet.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The spreadsheet ID |
| `sheet_name` | string | Yes | Name for the new tab |

**Example:**

```json
{
  "name": "sheets_add_sheet",
  "arguments": {
    "spreadsheet_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
    "sheet_name": "Q4 Forecast"
  }
}
```

### sheets_write_range

Write values to a specific range. Uses `USER_ENTERED` input mode — formulas like `=SUM(A1:A10)` are interpreted, not stored as text.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The spreadsheet ID |
| `range` | string | Yes | A1 notation range (e.g., `Sheet1!A1:C3`) |
| `values` | any[][] | Yes | 2D array of values to write |

**Example:**

```json
{
  "name": "sheets_write_range",
  "arguments": {
    "spreadsheet_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
    "range": "Sheet1!A1:C2",
    "values": [
      ["Item", "Quantity", "Total"],
      ["Widgets", 100, "=B2*5.99"]
    ]
  }
}
```

### sheets_append_rows

Append rows after the last row of data on a tab. Automatically finds the end of existing data.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The spreadsheet ID |
| `sheet_name` | string | Yes | Tab name to append to |
| `values` | any[][] | Yes | 2D array of rows to append |

**Example:**

```json
{
  "name": "sheets_append_rows",
  "arguments": {
    "spreadsheet_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
    "sheet_name": "Sheet1",
    "values": [
      ["Gadgets", 50, "=B3*12.99"],
      ["Sprockets", 200, "=B4*3.49"]
    ]
  }
}
```

## Batch Operations

### sheets_batch_update

Perform multiple read and write operations in a single call. Useful for working across multiple tabs in one request.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The spreadsheet ID |
| `reads` | object[] | No | Array of ranges to read |
| `writes` | object[] | No | Array of range/values pairs to write |

**Example:**

```json
{
  "name": "sheets_batch_update",
  "arguments": {
    "spreadsheet_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
    "reads": [
      { "range": "Revenue!A1:D10" },
      { "range": "Expenses!A1:D10" }
    ],
    "writes": [
      {
        "range": "Summary!A1:B2",
        "values": [
          ["Category", "Total"],
          ["Revenue", "=Revenue!D10"]
        ]
      }
    ]
  }
}
```
