---
id: oauth2-auth
parent: googledrive-mcp
created: 2026-06-30T14:00:00Z
priority: 1
status: not_started
depends-on: mcp-server-scaffold
---

# OAuth2 Authentication

The server authenticates with Google APIs using OAuth2 with local token caching. Config directory uses the project name: `~/.config/googledrive-ai/`.

## Setup Requirements

The following Google Cloud setup is required before first run:

1. A Google Cloud project exists with Drive API, Sheets API, and Docs API enabled
2. OAuth2 credentials are created as "Desktop app" type in the Google Cloud Console
3. The downloaded `credentials.json` file is placed at `~/.config/googledrive-ai/credentials.json`

## Success Criteria

- Auth implementation lives in `internal/auth/`
- Config directory is `~/.config/googledrive-ai/` (matching the project name)
- Server reads OAuth2 client credentials from `~/.config/googledrive-ai/credentials.json`
- On first run with no cached token, server launches browser-based OAuth2 consent flow
- After consent, access token and refresh token are cached at `~/.config/googledrive-ai/token.json`
- Subsequent runs reuse the cached token without re-prompting
- Token refresh is handled automatically via the refresh token when the access token expires
- OAuth2 scopes requested: `drive`, `spreadsheets`, `documents`
- If `credentials.json` is missing, server exits with a clear error message explaining the setup steps
