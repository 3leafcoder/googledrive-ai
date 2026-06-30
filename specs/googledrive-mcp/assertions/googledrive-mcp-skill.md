---
id: googledrive-mcp-skill
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 1
status: not_started
depends-on: sheets-batch-operations
---

# Google Drive MCP Claude Skill

A Claude skill (markdown file) that documents exact tool call patterns, range syntax, and common recipes so Claude uses the MCP tools correctly on the first try.

## Success Criteria

- Skill file exists as a markdown file usable by Claude
- Documents every tool with its exact parameter names, types, and valid values
- Includes A1 range notation reference with examples (`Sheet1!A1:C10`, `Sheet1!A:A`, `Sheet1!1:1`)
- Explains `render_option` values: `formatted` vs `formula` with when to use each
- Explains `ValueInputOption: USER_ENTERED` behavior — that writing `=SUM(A1:A10)` creates a formula
- Includes common recipes:
  - Read a summary tab and write results to another tab
  - Append a row to a log sheet
  - Read formulas from one tab, computed values from another
  - Create a new workbook with multiple tabs
  - Batch read across multiple tabs
- Includes gotchas and tips (e.g., "use `formula` render option to audit formulas", "append doesn't overwrite — it finds the last row")
- References correct tool names matching the MCP server implementation
