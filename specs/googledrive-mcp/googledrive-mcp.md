---
id: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 1
---

# Google Drive MCP Server

A Go MCP server (`googledrive-mcp`) that exposes Google Drive, Docs, and Sheets operations as tools over stdio transport. Sheets is the primary use case — the server is formula-aware, handles multi-tab workbooks, and supports safe writes that don't clobber adjacent formulas.

## Architecture

- **Language:** Go
- **MCP SDK:** `github.com/mark3labs/mcp-go`
- **Transport:** stdio
- **Auth:** Google OAuth2 with local token caching (`~/.config/googledrive-ai/`)
- **Google APIs:** Drive v3, Docs v1, Sheets v4

## Project Structure

```
cmd/googledrive-mcp/main.go    # entry point, stdio transport
internal/auth/                  # OAuth2 flow + token caching
internal/drive/                 # Drive tools
internal/docs/                  # Docs tools
internal/sheets/                # Sheets tools
internal/mcp/                   # MCP server setup + tool registration
```

## Tool Surface

### Drive (CRUD)
- `drive_list` — list files/folders
- `drive_search` — search by name/type/content
- `drive_get` — get file metadata
- `drive_create_folder` — create folder
- `drive_upload` — upload file
- `drive_download` — download file
- `drive_delete` — delete file/folder
- `drive_move` — move file to different folder

### Docs (Simplified)
- `docs_create` — create new doc with content
- `docs_read` — read doc as plain text or structured
- `docs_append` — append text to doc
- `docs_replace` — find and replace text in doc

### Sheets (Primary — Formula-Aware, Multi-Tab)
- `sheets_create` — create new workbook
- `sheets_list_sheets` — list all tabs with dimensions
- `sheets_add_sheet` — add new tab to workbook
- `sheets_get_metadata` — workbook metadata, named ranges, sheet properties
- `sheets_read_range` — read range from specific tab (values OR formulas)
- `sheets_write_range` — write to specific range (formula-safe, USER_ENTERED)
- `sheets_append_rows` — append rows after last data row on a tab
- `sheets_batch_update` — multiple reads/writes in one call for multi-tab ops

## Companion Skill

A Claude skill (markdown) that documents exact tool call patterns, range syntax, formula handling, and common recipes so Claude gets the syntax right on the first try.
