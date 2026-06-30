---
id: mcp-server-scaffold
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 1
status: not_started
---

# MCP Server Scaffold

The `googledrive-mcp` binary builds, runs, and speaks MCP protocol over stdio.

## Success Criteria

- `go build ./cmd/googledrive-mcp` produces a binary without errors
- Binary starts and communicates over stdio using the MCP protocol
- Server uses `github.com/mark3labs/mcp-go` SDK
- Server responds to `initialize` and `tools/list` MCP requests
- Server declares its tool capabilities in the MCP handshake
- `go.mod` exists with module path and Go version specified
