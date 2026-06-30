# Getting Started

## What is this?

`googledrive-mcp` is a [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server written in Go. It gives AI assistants like Claude the ability to interact with your Google Drive, Sheets, and Docs through a standardized tool interface.

When connected to an MCP client, the server exposes tools that let Claude:

- Browse and search your Google Drive
- Read and write spreadsheet data (with full formula support)
- Create and edit Google Docs
- Upload and download files

## Prerequisites

Before you begin, make sure you have:

- **Go 1.21 or later** — [Install Go](https://go.dev/doc/install)
- **A Google account** — with access to Google Drive
- **A Google Cloud project** — free tier is sufficient (see [Google Cloud Setup](setup/google-cloud.md))

## Architecture

The server uses stdio transport, meaning it communicates with MCP clients through standard input/output. This is the simplest and most common MCP transport.

```
MCP Client (Claude Code, Claude Desktop, etc.)
    ↕ stdio
googledrive-mcp binary
    ↕ HTTPS
Google APIs (Drive v3, Sheets v4, Docs v1)
```

Authentication uses OAuth2 with a desktop application flow. On first run, the server opens your browser for Google consent. After authorization, tokens are cached locally and refreshed automatically.

## Next Steps

1. [Create a Google Cloud project and download credentials](setup/google-cloud.md)
2. [Build the binary](setup/installation.md)
3. [Configure and authenticate](setup/configuration.md)
4. [Connect to your MCP client](setup/mcp-clients.md)
