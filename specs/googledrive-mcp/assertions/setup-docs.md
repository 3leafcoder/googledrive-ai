---
id: setup-docs
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 1
status: in_progress
locked-by: builder-Paris-MacBook-Pro-2.local-79235-1782839259
depends-on: oauth2-auth
---

# Setup Documentation Site

A Zensical documentation site exists at `docs/` that walks users through setting up and configuring the Google Drive MCP server.

## Success Criteria

- Zensical project is scaffolded with `zensical.toml` configured for this project
- Site name is "Google Drive MCP"
- Documentation covers:
  - **Getting Started** — what this MCP server does, prerequisites (Go, Google Cloud account)
  - **Google Cloud Setup** — step-by-step: create project, enable Drive/Sheets/Docs APIs, create OAuth2 desktop credentials, download `credentials.json`
  - **Installation** — how to build the binary (`go build ./cmd/googledrive-mcp`)
  - **Configuration** — placing `credentials.json` at `~/.config/googledrive-ai/credentials.json`, first-run auth flow, token caching
  - **MCP Client Setup** — how to add the server to Claude Code (`claude mcp add`), Claude Desktop (`claude_desktop_config.json`), or other MCP clients
  - **Tools Reference** — list of all available tools with parameters and examples (Drive, Docs, Sheets)
- Navigation groups docs logically (Getting Started → Setup → Tools Reference)
- `zensical serve` renders the site without errors
