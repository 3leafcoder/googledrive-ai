# Drive Tools

Tools for managing files and folders in Google Drive.

## drive_list

List files and folders in a directory.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `folder_id` | string | No | Folder ID to list. Defaults to root ("My Drive") |
| `page_size` | integer | No | Maximum number of results to return |

**Example:**

```json
{
  "name": "drive_list",
  "arguments": {
    "folder_id": "1a2b3c4d5e",
    "page_size": 20
  }
}
```

## drive_search

Search Drive by name, type, or content.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query (name, type, or content) |
| `page_size` | integer | No | Maximum number of results to return |

**Example:**

```json
{
  "name": "drive_search",
  "arguments": {
    "query": "budget 2026",
    "page_size": 10
  }
}
```

## drive_get

Get metadata for a specific file.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_id` | string | Yes | The file ID |

**Example:**

```json
{
  "name": "drive_get",
  "arguments": {
    "file_id": "1a2b3c4d5e"
  }
}
```

## drive_create_folder

Create a new folder.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Folder name |
| `parent_id` | string | No | Parent folder ID. Defaults to root |

**Example:**

```json
{
  "name": "drive_create_folder",
  "arguments": {
    "name": "Project Reports",
    "parent_id": "1a2b3c4d5e"
  }
}
```

## drive_upload

Upload a file to Drive.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | File name in Drive |
| `file_path` | string | Yes | Local path of the file to upload |
| `parent_id` | string | No | Parent folder ID. Defaults to root |

**Example:**

```json
{
  "name": "drive_upload",
  "arguments": {
    "name": "report.pdf",
    "file_path": "/tmp/report.pdf",
    "parent_id": "1a2b3c4d5e"
  }
}
```

## drive_download

Download a file's content.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_id` | string | Yes | The file ID to download |

**Example:**

```json
{
  "name": "drive_download",
  "arguments": {
    "file_id": "1a2b3c4d5e"
  }
}
```

## drive_delete

Move a file or folder to the trash.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_id` | string | Yes | The file ID to delete |

**Example:**

```json
{
  "name": "drive_delete",
  "arguments": {
    "file_id": "1a2b3c4d5e"
  }
}
```

## drive_move

Move a file to a different folder.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_id` | string | Yes | The file ID to move |
| `new_parent_id` | string | Yes | Destination folder ID |

**Example:**

```json
{
  "name": "drive_move",
  "arguments": {
    "file_id": "1a2b3c4d5e",
    "new_parent_id": "6f7g8h9i0j"
  }
}
```
