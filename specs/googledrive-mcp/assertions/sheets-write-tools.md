---
id: sheets-write-tools
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 1
status: in_progress
depends-on: sheets-read-tools
locked-by: builder-Paris-MacBook-Pro-2-51203-1751299200
---

# Sheets Write Tools

Tools for writing data to spreadsheets with formula safety and multi-tab awareness.

## Success Criteria

- `sheets_write_range` tool writes values to a specific range on a named tab using A1 notation
- Writes use `ValueInputOption: USER_ENTERED` so formula strings (e.g., `=SUM(A1:A10)`) are interpreted as formulas
- Writing to a range only affects the targeted cells — adjacent cells and their formulas are untouched
- `sheets_append_rows` tool appends rows after the last row with data on a specified tab
- `sheets_add_sheet` tool adds a new tab to an existing workbook with a given name
- `sheets_create` tool creates a new workbook with a title and optional initial sheet names
- All write tools return confirmation with the affected range and number of cells updated
