---
id: sheets-batch-operations
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 1
status: not_started
depends-on: sheets-write-tools
---

# Sheets Batch Operations

Batch tool for efficient multi-range and multi-tab operations in a single call.

## Success Criteria

- `sheets_batch_update` tool accepts multiple read and write operations in a single request
- Each operation in the batch specifies a tab, range, and action (read or write)
- Batch reads can mix `formatted` and `formula` render options per range
- Batch writes use `ValueInputOption: USER_ENTERED` for formula safety
- The tool returns results for each operation in the batch, preserving order
- Batch operations execute against the Sheets API in a single round-trip where possible (using `batchGet` / `batchUpdate` endpoints)
