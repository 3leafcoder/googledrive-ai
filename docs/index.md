# Google Drive MCP

A Go MCP server that exposes Google Drive, Docs, and Sheets operations as tools over stdio transport.

Sheets is the primary use case — the server is formula-aware, handles multi-tab workbooks, and supports safe writes that don't clobber adjacent formulas.

## Features

- **Google Drive** — list, search, upload, download, move, and delete files
- **Google Sheets** — read/write ranges, append rows, batch operations across tabs, formula-safe writes
- **Google Docs** — create, read, append, and find-and-replace text
- **OAuth2 authentication** — browser-based consent flow with automatic token caching and refresh

## Quick Start

1. [Set up a Google Cloud project](setup/google-cloud.md) with Drive, Sheets, and Docs APIs enabled
2. [Install the binary](setup/installation.md) with `go build`
3. [Configure credentials](setup/configuration.md) and run the first-time auth flow
4. [Add the server to your MCP client](setup/mcp-clients.md) (Claude Code, Claude Desktop, etc.)
