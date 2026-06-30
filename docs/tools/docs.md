# Docs Tools

Tools for creating and editing Google Docs. These provide a simplified interface focused on common text operations.

## docs_create

Create a new Google Doc.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | Yes | Document title |
| `content` | string | No | Initial text content |

**Example:**

```json
{
  "name": "docs_create",
  "arguments": {
    "title": "Meeting Notes - June 2026",
    "content": "## Attendees\n\n## Agenda\n\n## Action Items\n"
  }
}
```

Returns the new document ID.

## docs_read

Read the full text content of a document.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `document_id` | string | Yes | The document ID |

**Example:**

```json
{
  "name": "docs_read",
  "arguments": {
    "document_id": "1a2b3c4d5e"
  }
}
```

## docs_append

Append text to the end of a document.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `document_id` | string | Yes | The document ID |
| `text` | string | Yes | Text to append |

**Example:**

```json
{
  "name": "docs_append",
  "arguments": {
    "document_id": "1a2b3c4d5e",
    "text": "\n\n## Follow-up\n\nSchedule review meeting for next week."
  }
}
```

## docs_replace

Find and replace text within a document.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `document_id` | string | Yes | The document ID |
| `find_text` | string | Yes | Text to search for |
| `replace_text` | string | Yes | Replacement text |

**Example:**

```json
{
  "name": "docs_replace",
  "arguments": {
    "document_id": "1a2b3c4d5e",
    "find_text": "DRAFT",
    "replace_text": "FINAL"
  }
}
```
