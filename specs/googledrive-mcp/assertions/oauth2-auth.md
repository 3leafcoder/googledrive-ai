---
id: oauth2-auth
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 1
status: not_started
depends-on: mcp-server-scaffold
---

# OAuth2 Authentication

The server authenticates with Google APIs using OAuth2 with local token caching.

## Success Criteria

- Auth implementation lives in `internal/auth/`
- Server reads OAuth2 client credentials from a configurable path (default `~/.config/googledrive-mcp/credentials.json`)
- On first run, server launches browser-based OAuth2 consent flow
- After consent, token is cached locally at `~/.config/googledrive-mcp/token.json`
- Subsequent runs reuse the cached token without re-prompting
- Token refresh is handled automatically when the access token expires
- OAuth2 scopes include Drive, Docs, and Sheets access
