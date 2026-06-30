---
id: drive-crud-tools
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 2
status: not_started
depends-on: oauth2-auth
---

# Drive CRUD Tools

Tools for managing files and folders in Google Drive.

## Success Criteria

- Implementation lives in `internal/drive/`
- `drive_list` tool lists files/folders in a given folder (or root), returning name, ID, type, and modified time
- `drive_search` tool searches Drive by query string (name, type, content)
- `drive_get` tool returns metadata for a file by ID
- `drive_create_folder` tool creates a folder with a given name and optional parent folder ID
- `drive_upload` tool uploads a file to Drive with a name and optional parent folder
- `drive_download` tool downloads a file's content by ID
- `drive_delete` tool moves a file/folder to trash by ID
- `drive_move` tool moves a file to a different parent folder
- All tools return structured JSON responses
