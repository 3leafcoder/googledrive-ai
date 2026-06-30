---
id: sheets-read-tools
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 1
status: in_progress
locked-by: builder-Paris-MacBook-Pro-2.local-27744-1782839751
depends-on: oauth2-auth
---

# Sheets Read Tools

Tools for reading spreadsheet structure and data, with support for formulas vs computed values.

## Success Criteria

- Implementation lives in `internal/sheets/`
- `sheets_list_sheets` tool returns all tabs in a workbook with their names, indexes, and row/column counts
- `sheets_read_range` tool reads a cell range from a named tab using A1 notation (e.g., `Sheet1!A1:C10`)
- `sheets_read_range` accepts a `render_option` parameter: `formatted` (default) returns display values, `formula` returns raw formulas
- `sheets_get_metadata` tool returns workbook title, locale, named ranges, and per-sheet properties
- All tools accept a `spreadsheet_id` parameter to identify the workbook
- Tools return structured JSON responses, not raw API payloads
