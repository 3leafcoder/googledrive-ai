---
id: docs-tools
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 2
status: done
depends-on: oauth2-auth
---

# Docs Tools (Simplified)

High-level tools for creating and editing Google Docs without exposing the raw batch update API complexity.

## Success Criteria

- Implementation lives in `internal/docs/`
- `docs_create` tool creates a new Google Doc with a title and optional initial text content
- `docs_read` tool reads a document's full text content by ID, returned as plain text
- `docs_append` tool appends text to the end of an existing document
- `docs_replace` tool finds and replaces all occurrences of a text string in a document
- All tools accept a `document_id` parameter (except `docs_create` which returns one)
- Tools abstract away the Docs API batch update model — callers don't need to know about positional indexes
